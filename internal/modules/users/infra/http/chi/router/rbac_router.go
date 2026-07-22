package router

import (
	"github.com/go-chi/chi/v5"

	"auction/internal/modules/users/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"
)

// RegisterRBACRoutes exposes the RBAC policy management API. It is intentionally
// gated by authn.RequireRole(admin) rather than Casbin itself, so the policy
// store can always be repaired by a superuser even if enforcement is broken.
func RegisterRBACRoutes(
	server *httpserver.Server,
	rbacHandler *handler.RBACHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	admin := authn.RequireRole(authn.RoleAdmin)

	router.Route("/api/v1/rbac", func(r chi.Router) {
		r.With(middleware.RequireAuth, admin).Get("/policies", rbacHandler.ListPolicies)
		r.With(middleware.RequireAuth, admin).Post("/policies", rbacHandler.AddPolicy)
		r.With(middleware.RequireAuth, admin).Delete("/policies", rbacHandler.RemovePolicy)

		r.With(middleware.RequireAuth, admin).Get("/role-assignments", rbacHandler.ListRoleAssignments)
		r.With(middleware.RequireAuth, admin).Post("/role-assignments", rbacHandler.AssignRole)
		r.With(middleware.RequireAuth, admin).Delete("/role-assignments", rbacHandler.RevokeRole)
		r.With(middleware.RequireAuth, admin).Delete("/role-assignments/{userID}", rbacHandler.RevokeAllRoles)
	})
}
