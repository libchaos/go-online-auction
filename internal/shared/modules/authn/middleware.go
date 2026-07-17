package authn

import (
	"net/http"
	"slices"
	"strings"

	"auction/internal/shared/sdk/http/response"
)

const bearerPrefix = "Bearer "

type Middleware struct {
	verifier TokenVerifier
}

func NewMiddleware(verifier TokenVerifier) *Middleware {
	return &Middleware{verifier: verifier}
}

// RequireAuth parses the Authorization header, verifies the bearer token
// and injects the claims into the request context.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			response.Error(w, ErrUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)
		claims, err := m.verifier.Verify(token)
		if err != nil {
			response.Error(w, ErrUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
	})
}

// RequireRole allows the request only when the authenticated user's role
// is in the given list. Must be composed after RequireAuth.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				response.Error(w, ErrUnauthorized)
				return
			}

			if slices.Contains(roles, claims.Role) {
				next.ServeHTTP(w, r)
				return
			}

			response.Error(w, ErrForbidden)
		})
	}
}
