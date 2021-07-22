package http

import (
	"github.com/switch-li/juice/pkg/logger"
	"github.com/switch-li/juice/pkg/logger/zap"
	"github.com/switch-li/juice/pkg/mail"
)

type Option func(*option)

type option struct {
	log          logger.Logger
	disablePProf bool
	// disableSwagger    bool
	disablePrometheus bool
	disableLogger     bool
	simpleLogger      bool
	panicNotify       OnPanicNotify
	mailOptions       *mail.Options
	recordMetrics     RecordMetrics
	enableCors        bool
	enableRate        bool
}

type OnPanicNotify func(ctx Context, opts *mail.Options, err interface{}, stackInfo string)

type RecordMetrics func(method, uri string, success bool, httpCode, businessCode int, costSeconds float64, traceId string)

func newOptions() *option {
	return &option{
		log: zap.DefaultLogger,
	}
}

func WithLogger(log logger.Logger) Option {
	return func(opt *option) {
		opt.log = log
	}
}

func WithDisablePProf() Option {
	return func(opt *option) {
		opt.disablePProf = true
	}
}

// func WithDisableSwagger() Option {
// 	return func(opt *option) {
// 		opt.disableSwagger = true
// 	}
// }

func WithDisableproPrometheus() Option {
	return func(opt *option) {
		opt.disablePrometheus = true
	}
}

func WithDisableLogger() Option {
	return func(o *option) {
		o.disableLogger = true
	}
}

func WithSimplelogger() Option {
	return func(o *option) {
		o.simpleLogger = true
	}
}

func WithPanicNotify(notify OnPanicNotify) Option {
	return func(opt *option) {
		opt.panicNotify = notify
		zap.DefaultLogger.Info("register panic notify")
	}
}

func WithMailOptions(mailOptions *mail.Options) Option {
	return func(o *option) {
		o.mailOptions = mailOptions
	}
}

func WithRecordMetrics(record RecordMetrics) Option {
	return func(opt *option) {
		opt.recordMetrics = record
	}
}

func WithEnableCors() Option {
	return func(opt *option) {
		opt.enableCors = true
		zap.DefaultLogger.Info("register cors")
	}
}

func WithEnableRate() Option {
	return func(opt *option) {
		opt.enableRate = true
		zap.DefaultLogger.Info("register rate")
	}
}

func DisableTrace(ctx Context) {
	ctx.disableTrace()
}

// AliasForRecordMetrics 对请求uri起个别名，用于prometheus记录指标。
// 如：Get /user/:username 这样的uri，因为username会有非常多的情况，这样记录prometheus指标会非常的不有好。
func AliasForRecordMetrics(path string) HandlerFunc {
	return func(ctx Context) {
		ctx.setAlias(path)
	}
}

// WrapAuthHandler 用来处理 Auth 的入口，在之后的handler中只需 ctx.UserID() ctx.UserName() 即可。
func WrapAuthHandler(handler func(Context) (userID int64, userName string, err Error)) HandlerFunc {
	return func(ctx Context) {
		userID, userName, err := handler(ctx)
		if err != nil {
			ctx.AbortWithError(err)
			return
		}
		ctx.setUserID(userID)
		ctx.setUserName(userName)
	}
}
