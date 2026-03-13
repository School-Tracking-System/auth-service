package logger

import (
	"github.com/fercho/school-tracking/services/auth/pkg/env"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a Zap logger configured by the service environment and log level.
func NewLogger(cfg *env.Config) (*zap.Logger, error) {
	var zapCfg zap.Config
	if cfg.Environment == "development" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	return zapCfg.Build()
}

var Module = fx.Module("logger", fx.Provide(NewLogger))
