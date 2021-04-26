package main

import (
	"fmt"

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
	return func(c http.Context) {
		c.Payload("hello world")
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
	)
	if err != nil {
		panic(err)
	}

	demo := mux.Group("/demo")
	demo.GET("hello", Hello())

	srv := http.NewServer(
		mux,
		http.Address(":8880"),
	)

	err = srv.Start()
	if err != nil {
		fmt.Println(err)
	}
}
