package logger

import "go.uber.org/fx"

var Module = fx.Module("shared/logger", fx.Provide(New))
