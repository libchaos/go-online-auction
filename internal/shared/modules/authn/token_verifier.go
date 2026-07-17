package authn

// TokenVerifier verifies an access token and extracts the claims.
// The concrete implementation is provided by the users module.
type TokenVerifier interface {
	Verify(tokenString string) (Claims, error)
}
