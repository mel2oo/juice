package zap

import (
	"testing"

	"github.com/switch-li/juice/pkg/logger"

	"go.uber.org/zap"
)

func Test_ZapLogger(t *testing.T) {
	log := NewZapLogger()
	log.Info("hello world")
}

func Test_ZapLogFile(t *testing.T) {
	log := NewZapLogger(
		logger.WithDevelopment(false),
	)

	log.(*ZapLogger).ReplaceGlobals()
	zap.L().Info("hello")
}
