package main

import (
	"fmt"
	"mime/multipart"
	http2 "net/http"

	"github.com/switch-li/juice/pkg/logger"
	"github.com/switch-li/juice/pkg/logger/zap"
	"github.com/switch-li/juice/transport/http"
)

func Hello() http.HandlerFunc {
	return func(c http.Context) {
		c.FileFromMultipart("1.txt", []byte("hello"))
	}
}

func Hi() http.HandlerFunc {
	return func(c http.Context) {
		fmt.Println("hi")
	}
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

func main() {
	log := zap.NewZapLogger(
		logger.WithDevelopment(),
	)
	mux, err := http.NewMux(
		http.WithEnableCors(),
		http.WithEnableRate(),
		http.WithSimplelogger(),
	)
	if err != nil {
		panic(err)
	}

	demo := mux.Group("/demo")
	demo.GET("/hello", Hello())
	demo.GET("/hi", Hi())
	demo.POST("/upload", Upload())

	srv := http.NewServer(
		mux,
		http.Network("tcp"),
		http.Address(":10002"),
		http.Logger(log),
	)

	err = srv.Start()
}
