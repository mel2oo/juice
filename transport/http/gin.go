package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cors "github.com/rs/cors/wrapper/gin"
	"github.com/switch-li/juice/pkg/logger"
	dlog "github.com/switch-li/juice/pkg/logger/default"
	"github.com/switch-li/juice/transport/http/middleware/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const _MaxBurstSize = 100000

var withoutTracePaths = map[string]bool{
	"/metrics": true,

	"/debug/pprof/":             true,
	"/debug/pprof/cmdline":      true,
	"/debug/pprof/profile":      true,
	"/debug/pprof/symbol":       true,
	"/debug/pprof/trace":        true,
	"/debug/pprof/allocs":       true,
	"/debug/pprof/block":        true,
	"/debug/pprof/goroutine":    true,
	"/debug/pprof/heap":         true,
	"/debug/pprof/mutex":        true,
	"/debug/pprof/threadcreate": true,

	"/favicon.ico": true,

	"/system/health": true,
}

func NewMux(logger logger.Logger, options ...Option) (*Mux, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DisableBindValidation()
	mux := &Mux{
		engine: gin.New(),
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	if !opt.disablePProf {
		pprof.Register(mux.engine)
		dlog.DefaultLogger.Info("[register pprof]")
	}

	// if !opt.disableSwagger {
	// 	mux.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 	dlog.DefaultLogger.Info("[register swagger]")
	// }

	if !opt.disablePrometheus {
		mux.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
		dlog.DefaultLogger.Info("[register prometheus]")
	}

	if opt.enableCors {
		mux.engine.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:     []string{"*"},
			AllowCredentials:   true,
			OptionsPassthrough: true,
		}))
	}

	mux.engine.Use(func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", string(debug.Stack())))
			}
		}()

		ctx.Next()
	})

	mux.engine.Use(func(ctx *gin.Context) {
		ts := time.Now()

		context := newContext(ctx)
		defer releaseContext(context)

		context.init()
		context.setLogger(logger)

		if !withoutTracePaths[ctx.Request.URL.Path] {
			if traceId := context.GetHeader(trace.Header); traceId != "" {
				context.setTrace(trace.New(traceId))
			} else {
				context.setTrace(trace.New(""))
			}
		}

		defer func() {
			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))
				context.AbortWithError(NewError(
					http.StatusInternalServerError,
					ServerError,
					Text(ServerError)),
				)

				if notify := opt.panicNotify; notify != nil && opt.mailOptions != nil {
					notify(context, opt.mailOptions, err, stackInfo)
				}
			}

			if ctx.Writer.Status() == http.StatusNotFound {
				return
			}

			var (
				response        interface{}
				businessCode    int
				businessCodeMsg string
				abortErr        error
				traceId         string
				graphResponse   interface{}
			)

			if ctx.IsAborted() {
				for i := range ctx.Errors {
					multierr.AppendInto(&abortErr, ctx.Errors[i])
				}

				if err := context.abortError(); err != nil {
					multierr.AppendInto(&abortErr, err.GetErr())
					response = err
					businessCode = err.GetBusinessCode()
					businessCodeMsg = err.GetMsg()

					if x := context.Trace(); x != nil {
						context.SetHeader(trace.Header, x.ID())
						traceId = x.ID()
					}

					ctx.JSON(err.GetHttpCode(), &Failure{
						Code:    businessCode,
						Message: businessCodeMsg,
					})
				}
			} else {
				response = context.getPayload()
				if response != nil {
					if x := context.Trace(); x != nil {
						context.SetHeader(trace.Header, x.ID())
						traceId = x.ID()
					}
					ctx.JSON(http.StatusOK, response)
				}
			}

			graphResponse = context.getGraphPayload()

			if opt.recordMetrics != nil {
				uri := context.URI()
				if alias := context.Alias(); alias != "" {
					uri = alias
				}

				opt.recordMetrics(
					context.Method(),
					uri,
					!ctx.IsAborted() && ctx.Writer.Status() == http.StatusOK,
					ctx.Writer.Status(),
					businessCode,
					time.Since(ts).Seconds(),
					traceId,
				)
			}

			var t *trace.Trace
			if x := context.Trace(); x != nil {
				t = x.(*trace.Trace)
			} else {
				return
			}

			decodedURL, _ := url.QueryUnescape(ctx.Request.URL.RequestURI())
			t.WithRequest(&trace.Request{
				TTL:        "un-limit",
				Method:     ctx.Request.Method,
				DecodedURL: decodedURL,
				Header:     ctx.Request.Header,
				Body:       string(context.RawData()),
			})

			var responseBody interface{}

			if response != nil {
				responseBody = response
			}

			if graphResponse != nil {
				responseBody = graphResponse
			}

			t.WithResponse(&trace.Response{
				Header:          ctx.Writer.Header(),
				HttpCode:        ctx.Writer.Status(),
				HttpCodeMsg:     http.StatusText(ctx.Writer.Status()),
				BusinessCode:    businessCode,
				BusinessCodeMsg: businessCodeMsg,
				Body:            responseBody,
				CostSeconds:     time.Since(ts).Seconds(),
			})

			t.Success = !ctx.IsAborted() && ctx.Writer.Status() == http.StatusOK
			t.CostSeconds = time.Since(ts).Seconds()

			if !opt.disableLogger {
				if opt.simpleLogger {
					logger.Debug(
						fmt.Sprintf("interceptor | method: %s | path: %s | http_code: %d",
							ctx.Request.Method,
							decodedURL,
							ctx.Writer.Status(),
						),
					)
				} else {
					logger.Debug("interceptor",
						zap.Any("method", ctx.Request.Method),
						zap.Any("path", decodedURL),
						zap.Any("http_code", ctx.Writer.Status()),
						zap.Any("business_code", businessCode),
						zap.Any("success", t.Success),
						zap.Any("cost_seconds", t.CostSeconds),
						zap.Any("trace_id", t.Identifier),
						zap.Any("trace_info", t),
						zap.Error(abortErr),
					)
				}
			}
		}()

		ctx.Next()
	})

	if opt.enableRate {
		limiter := rate.NewLimiter(rate.Every(time.Second*1), _MaxBurstSize)
		mux.engine.Use(func(ctx *gin.Context) {
			context := newContext(ctx)
			defer releaseContext(context)

			if !limiter.Allow() {
				context.AbortWithError(NewError(
					http.StatusTooManyRequests,
					TooManyRequests,
					Text(TooManyRequests)),
				)
				return
			}

			ctx.Next()
		})
	}

	mux.engine.NoMethod(wrapHandlers(DisableTrace)...)
	mux.engine.NoRoute(wrapHandlers(DisableTrace)...)
	system := mux.Group("/system")
	{
		system.GET("/health", func(ctx Context) {
			resp := &struct {
				Timestamp time.Time `json:"timestamp"`
				Host      string    `json:"host"`
				Status    string    `json:"status"`
			}{
				Timestamp: time.Now(),
				Host:      ctx.Host(),
				Status:    "ok",
			}
			ctx.Payload(resp)
		})
	}

	return mux, nil
}
