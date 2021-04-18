package main

import (
	"fmt"

	"github.com/switch-li/juice/pkg/logger/zap"
	"github.com/switch-li/juice/transport/http"
	"github.com/switch-li/juice/transport/http/middleware/metrics"
	"github.com/switch-li/juice/transport/http/middleware/notify"
)

func Hello() http.HandlerFunc {
	return func(c http.Context) {
		c.Payload("hello world")
	}
}

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
