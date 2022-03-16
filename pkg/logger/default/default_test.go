package dlog

import (
	"testing"

	"github.com/mel2oo/juice/pkg/logger"
)

func Test_DefaultLogger(t *testing.T) {

	// [test] 2021/03/30 10:10:29 INFO /Users/switch/Project/own/whl-shepherd/pkg/logger/default/default_test.go:14 hello world
	// [test] 2021/03/30 10:10:29 INFO /Users/switch/Project/own/whl-shepherd/pkg/logger/default/default_test.go:15 hello - switch

	log := NewDefaultLogger()
	log.Info("hello world")
}

func Test_DisableDevelopment(t *testing.T) {
	log := NewDefaultLogger(
		logger.WithDevelopment(),
	)
	log.Info("hello world")
}

func Test_DisableCaller(t *testing.T) {
	log := NewDefaultLogger(
		logger.WithDisableCaller(),
	)
	log.Info("hello world")
}
