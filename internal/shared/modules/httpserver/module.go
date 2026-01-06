package httpserver

import "go.uber.org/fx"

var Module = fx.Module("shared/httpserver", fx.Provide(New))


