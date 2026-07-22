package authz

import "go.uber.org/fx"

// Module wires the Casbin enforcer and its HTTP middleware into the fx graph.
// Any router that needs RBAC can depend on *authz.Middleware; the enforcer is
// built once from the shared connection pool and the embedded model.
var Module = fx.Module(
	"authz",
	fx.Provide(NewEnforcer),
	fx.Provide(NewMiddleware),
)
