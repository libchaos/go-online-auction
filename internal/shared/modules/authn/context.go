package authn

import "context"

type contextKey struct{}

var claimsContextKey = contextKey{}

// WithClaims returns a new context carrying the given claims
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// ClaimsFromContext extracts the authenticated user claims from the context
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	return claims, ok
}
