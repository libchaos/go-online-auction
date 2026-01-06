package redis

import "go.uber.org/fx"

var Module = fx.Module("shared/redis", fx.Provide(New))
