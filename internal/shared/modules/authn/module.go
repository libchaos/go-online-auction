package authn

import "go.uber.org/fx"

var Module = fx.Module(
	"shared/authn",

	fx.Provide(NewMiddleware),
)
