package main

import (
	"fmt"
	"mime/multipart"
	http2 "net/http"

	_ "github.com/switch-li/juice/examples/httpserver/docs"
	"github.com/switch-li/juice/pkg/logger/zap"
	"github.com/switch-li/juice/transport/http"
	"github.com/switch-li/juice/transport/http/middleware/metrics"
	"github.com/switch-li/juice/transport/http/middleware/notify"
)

// @Summary
// @Description
// @Tags Demo
// @Accept  json
// @Produce  json
// @Success 200
// @Router /hello [post]
func Hello() http.HandlerFunc {
	return func(c http.Context) {}
}

func Upload() http.HandlerFunc {
	type request struct {
		Aa    string                  `form:"aa"`
		Files []*multipart.FileHeader `form:"file"`
	}
	return func(c http.Context) {
		req := new(request)
		if err := c.ShouldBindFormMultipart(req); err != nil {
			c.AbortWithError(
				http.NewError(
					http2.StatusBadRequest,
					http.ParamBindError,
					http.Text(http.ParamBindError),
				).WithErr(err),
			)
			return
		}

		for _, f := range req.Files {
			c.SaveUploadedFile(f, f.Filename)
		}

	}
}

// @title juice docs api
// @version
// @description

// @contact.name
// @contact.url
// @contact.email

// @host 127.0.0.1:8880
// @BasePath
func main() {
	log := zap.NewZapLogger()
	mux, err := http.New(log,
		http.WithEnableCors(),
		http.WithEnableRate(),
		http.WithPanicNotify(notify.OnPanicNotify),
		http.WithRecordMetrics(metrics.RecordMetrics),
		http.WithSimplelogger(),
	)
	if err != nil {
		panic(err)
	}

	demo := mux.Group("/demo")
	demo.GET("/hello", Hello())
	demo.POST("/upload", Upload())

	srv := http.NewServer(
		mux,
		http.Address(":8880"),
	)

	err = srv.Start()
	if err != nil {
		fmt.Println(err)
	}
}
