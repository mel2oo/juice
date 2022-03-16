package zap

import (
	"testing"

	"github.com/mel2oo/juice/pkg/logger"

	"go.uber.org/zap"
)

func Test_ZapLogger(t *testing.T) {
	log := NewZapLogger(
		logger.WithMaxSize(1),
		logger.WithMaxBackups(10),
		logger.WithMaxAge(1),
	)
	for i := 0; i < 100000; i++ {
		log.Info("hello world hello world hello world hello world hello world")
	}
}

func Test_ZapLogFile(t *testing.T) {
	log := NewZapLogger(
		logger.WithDevelopment(),
	)

	log.(*ZapLogger).ReplaceGlobals()
	zap.L().Info("hello")
}
