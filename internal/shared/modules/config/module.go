package config

import "go.uber.org/fx"

var Module = fx.Module("shared/config", fx.Provide(GetConfig))
