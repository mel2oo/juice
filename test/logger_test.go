package test

import (
	"testing"

	"github.com/mel2oo/juice/pkg/logger"
	"github.com/mel2oo/juice/pkg/logger/zap"
)

func TestZapLogger(t *testing.T) {
	log := zap.NewZapLogger(
		logger.WithOutputPath("./log"),
	)

	log.Info("hi")
	log.Warn("error")
}
