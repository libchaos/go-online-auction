package logger

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
	"github.com/cristiano-pacheco/go-online-auction/pkg/logger"
)

type Logger logger.Logger

func New(cfg config.Config) Logger {
	logLevel := logger.MustLogLevel(cfg.Log.LogLevel)
	logConfig := logger.Config{
		LogLevel: logLevel,
	}
	return logger.New(logConfig)
}
