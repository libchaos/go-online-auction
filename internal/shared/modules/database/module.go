package database

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"shared/database",
	fx.Provide(New),
)
