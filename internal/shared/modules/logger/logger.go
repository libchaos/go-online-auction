package logger

import (
	"auction/internal/shared/modules/config"
	"auction/pkg/logger"
)

type Logger logger.Logger

func New(cfg config.Config) Logger {
	logLevel := logger.MustLogLevel(cfg.Log.LogLevel)
	logConfig := logger.Config{
		LogLevel: logLevel,
	}
	return logger.New(logConfig)
}
