package test

import (
	"testing"

	"github.com/switch-li/juice/pkg/logger"
	"github.com/switch-li/juice/pkg/logger/zap"
)

func TestZapLogger(t *testing.T) {
	log := zap.NewZapLogger(
		logger.WithOutputPath("./log"),
	)

	log.Info("hi")
	log.Warn("error")
}
