package authz

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/stretchr/testify/suite"

	"auction/internal/shared/modules/authn"
)

// casbinRule is a typed policy line (ptype + fields) used by the in-memory
// adapter so both "p" and "g" rules can be loaded by tests.
type casbinRule struct {
	ptype string
	rule  []string
}

// memAdapter is a minimal in-memory persist.Adapter used only by tests so the
// middleware can be exercised without a Postgres instance.
type memAdapter struct {
	rules []casbinRule
}

func (a *memAdapter) LoadPolicy(m model.Model) error {
	for _, r := range a.rules {
		line := r.ptype + ", " + strings.Join(r.rule, ", ")
		if err := persist.LoadPolicyLine(line, m); err != nil {
			return err
		}
	}

	return nil
}

func (a *memAdapter) SavePolicy(model.Model) error { return nil }

func (a *memAdapter) AddPolicy(_ string, ptype string, rule []string) error {
	a.rules = append(a.rules, casbinRule{ptype: ptype, rule: rule})

	return nil
}

func (a *memAdapter) RemovePolicy(_ string, ptype string, rule []string) error {
	kept := a.rules[:0]
	for _, r := range a.rules {
		if r.ptype == ptype && sameRule(r.rule, rule) {
			continue
		}
		kept = append(kept, r)
	}
	a.rules = kept

	return nil
}

func (a *memAdapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return nil
}

func sameRule(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func newTestEnforcer(rules ...casbinRule) *casbin.Enforcer {
	m, err := model.NewModelFromString(modelConfig)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer(m, &memAdapter{rules: rules})
	if err != nil {
		panic(err)
	}

	return e
}

type MiddlewareTestSuite struct {
	suite.Suite
	sut *Middleware
}

func (s *MiddlewareTestSuite) SetupTest() {
	enforcer := newTestEnforcer(
		casbinRule{"p", []string{"admin", "/api/v1/*", "*"}},
		casbinRule{"p", []string{"seller", "/api/v1/auctions", "POST"}},
		casbinRule{"p", []string{"seller", "/api/v1/auctions/{id}/start", "PUT"}},
		casbinRule{"g", []string{"1", "admin"}},
		casbinRule{"g", []string{"2", "seller"}},
		casbinRule{"g", []string{"3", "bidder"}},
	)
	s.sut = NewMiddleware(enforcer, nil)
}

func TestAuthzMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) serve(userID uint64, path, method string) int {
	handler := s.sut.RequirePermission()(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest(method, path, nil)
	req = req.WithContext(authn.WithClaims(req.Context(), authn.Claims{UserID: userID, Role: "ignored"}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	return rec.Code
}

// Arrange/Act/Assert

func (s *MiddlewareTestSuite) TestAdmin_User_GetsFullWildcardAccess() {
	// User 1 is bound to the admin role via g, which owns /api/v1/** with *.
	s.Equal(http.StatusOK, s.serve(1, "/api/v1/users", http.MethodGet))
	s.Equal(http.StatusOK, s.serve(1, "/api/v1/anything/deep/path", http.MethodDelete))
}

func (s *MiddlewareTestSuite) TestSeller_User_AuctionCreate_Allowed() {
	s.Equal(http.StatusOK, s.serve(2, "/api/v1/auctions", http.MethodPost))
}

func (s *MiddlewareTestSuite) TestSeller_User_UserList_Forbidden() {
	// User 2 is bound to seller, which has no policy covering /api/v1/users.
	s.Equal(http.StatusForbidden, s.serve(2, "/api/v1/users", http.MethodGet))
}

func (s *MiddlewareTestSuite) TestBidder_User_NoMatchingPolicy_Forbidden() {
	s.Equal(http.StatusForbidden, s.serve(3, "/api/v1/users", http.MethodGet))
}

func (s *MiddlewareTestSuite) TestUnbound_User_Forbidden() {
	// User 99 has no g binding, so every guarded route is denied.
	s.Equal(http.StatusForbidden, s.serve(99, "/api/v1/users", http.MethodGet))
}

func (s *MiddlewareTestSuite) TestNoClaims_Unauthorized() {
	handler := s.sut.RequirePermission()(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *MiddlewareTestSuite) TestEnforce_UsesSubject_NotRole() {
	// The direct Enforce helper resolves the subject (user id) via g.
	allowed, err := s.sut.Enforce(strconv.FormatUint(2, 10), "/api/v1/auctions", "POST")
	s.Require().NoError(err)
	s.True(allowed)
}
