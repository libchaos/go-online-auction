package authz

import (
	"net/http"
	"strconv"

	"github.com/casbin/casbin/v2"

	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/errs"
)

// ErrAuthorizationFailed is returned when the enforcer itself errors (for example
// the policy store is unreachable). It is distinct from ErrForbidden, which means
// the request was evaluated and denied.
var ErrAuthorizationFailed = errs.New("AUTHZ_01", "Authorization check failed", http.StatusInternalServerError, nil)

// Middleware enforces RBAC decisions produced by the Casbin enforcer.
type Middleware struct {
	enforcer *casbin.Enforcer
	logger   logger.Logger
}

func NewMiddleware(enforcer *casbin.Enforcer, logger logger.Logger) *Middleware {
	return &Middleware{enforcer: enforcer, logger: logger}
}

// RequirePermission enforces the request against the Casbin enforcer. The
// authenticated user's id is the subject; the enforcer resolves the user to one
// or more roles through the g (grouping) relation, then matches the role
// against the request path (object) and HTTP method (action). It must be
// composed after authn.RequireAuth so the claims are present in the context.
func (m *Middleware) RequirePermission() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := authn.ClaimsFromContext(r.Context())
			if !ok {
				response.Error(w, authn.ErrUnauthorized)
				return
			}

			subject := strconv.FormatUint(claims.UserID, 10)
			allowed, err := m.enforcer.Enforce(subject, r.URL.Path, r.Method)
			if err != nil {
				if m.logger != nil {
					m.logger.Error().Err(err).Msg("authz enforce failed")
				}
				response.Error(w, ErrAuthorizationFailed)
				return
			}

			if !allowed {
				response.Error(w, authn.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Enforce exposes a direct authorization check, useful for non-HTTP call sites.
// subject is the user id; roles are resolved through the g relation.
func (m *Middleware) Enforce(subject, object, action string) (bool, error) {
	return m.enforcer.Enforce(subject, object, action)
}
