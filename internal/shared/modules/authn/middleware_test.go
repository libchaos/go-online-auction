package authn_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"auction/internal/shared/modules/authn"
)

type stubVerifier struct {
	claims authn.Claims
	err    error
}

func (v *stubVerifier) Verify(_ string) (authn.Claims, error) {
	return v.claims, v.err
}

type MiddlewareTestSuite struct {
	suite.Suite
	verifier *stubVerifier
	sut      *authn.Middleware
}

func (s *MiddlewareTestSuite) SetupTest() {
	s.verifier = &stubVerifier{
		claims: authn.Claims{UserID: 1, Role: authn.RoleBidder, Email: "a@b.com"},
	}
	s.sut = authn.NewMiddleware(s.verifier)
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) TestRequireAuth_MissingHeader_Returns401() {
	// Arrange
	handler := s.sut.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *MiddlewareTestSuite) TestRequireAuth_InvalidToken_Returns401() {
	// Arrange
	s.verifier.err = errors.New("invalid token")
	handler := s.sut.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *MiddlewareTestSuite) TestRequireAuth_ValidToken_InjectsClaims() {
	// Arrange
	var gotClaims authn.Claims
	var gotOK bool
	handler := s.sut.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims, gotOK = authn.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	s.Equal(http.StatusOK, rec.Code)
	s.True(gotOK)
	s.Equal(s.verifier.claims, gotClaims)
}

func (s *MiddlewareTestSuite) TestRequireRole_AllowedRole_CallsNext() {
	// Arrange
	handler := authn.RequireRole(authn.RoleSeller, authn.RoleAdmin)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	claims := authn.Claims{UserID: 1, Role: authn.RoleSeller}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(authn.WithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	s.Equal(http.StatusOK, rec.Code)
}

func (s *MiddlewareTestSuite) TestRequireRole_DisallowedRole_Returns403() {
	// Arrange
	handler := authn.RequireRole(authn.RoleAdmin)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	claims := authn.Claims{UserID: 1, Role: authn.RoleBidder}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(authn.WithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *MiddlewareTestSuite) TestRequireRole_NoClaims_Returns401() {
	// Arrange
	handler := authn.RequireRole(authn.RoleAdmin)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	s.Equal(http.StatusUnauthorized, rec.Code)
}
