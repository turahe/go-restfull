package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = zap.Logger

func New(env string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if env == "dev" || env == "local" || env == "" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return cfg.Build()
}

