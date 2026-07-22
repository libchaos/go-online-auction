package command

import (
	"context"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/stretchr/testify/suite"

	domainerrs "auction/internal/modules/users/domain/errs"
	"auction/internal/shared/modules/authn"
)

// gCasbinRule is a typed policy line so both "p" and "g" rules load in tests.
type gCasbinRule struct {
	ptype string
	rule  []string
}

type gMemAdapter struct {
	rules []gCasbinRule
}

func (a *gMemAdapter) LoadPolicy(m model.Model) error {
	for _, r := range a.rules {
		line := r.ptype + ", " + strings.Join(r.rule, ", ")
		if err := persist.LoadPolicyLine(line, m); err != nil {
			return err
		}
	}

	return nil
}

func (a *gMemAdapter) SavePolicy(model.Model) error { return nil }

func (a *gMemAdapter) AddPolicy(_ string, ptype string, rule []string) error {
	a.rules = append(a.rules, gCasbinRule{ptype: ptype, rule: rule})

	return nil
}

func (a *gMemAdapter) RemovePolicy(_ string, ptype string, rule []string) error {
	kept := a.rules[:0]
	for _, r := range a.rules {
		if r.ptype == ptype && gSameRule(r.rule, rule) {
			continue
		}
		kept = append(kept, r)
	}
	a.rules = kept

	return nil
}

func (a *gMemAdapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return nil
}

func gSameRule(a, b []string) bool {
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

func newGTestEnforcer(rules ...gCasbinRule) *casbin.Enforcer {
	m, err := model.NewModelFromString(modelConfigForGTest)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer(m, &gMemAdapter{rules: rules})
	if err != nil {
		panic(err)
	}

	return e
}

const modelConfigForGTest = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch4(r.obj, p.obj) && (r.act == p.act || p.act == '*')
`

type AssignRoleCommandTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	sut      *AssignRoleCommand
}

func (s *AssignRoleCommandTestSuite) SetupTest() {
	s.enforcer = newGTestEnforcer(
		gCasbinRule{"p", []string{"admin", "/api/v1/*", "*"}},
		gCasbinRule{"p", []string{"seller", "/api/v1/auctions", "POST"}},
		gCasbinRule{"g", []string{"1", "admin"}},
	)
	s.sut = NewAssignRoleCommand(s.enforcer)
}

func TestAssignRoleCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssignRoleCommandTestSuite))
}

// Arrange/Act/Assert

func (s *AssignRoleCommandTestSuite) TestExecute_ValidInput_AddsGroupingAndGrantsAccess() {
	// Act
	err := s.sut.Execute(context.Background(), AssignRoleCommandInput{UserID: 2, Role: authn.RoleSeller})

	// Assert
	s.Require().NoError(err)
	allowed, enforceErr := s.enforcer.Enforce("2", "/api/v1/auctions", "POST")
	s.Require().NoError(enforceErr)
	s.True(allowed)
}

func (s *AssignRoleCommandTestSuite) TestExecute_ZeroUserID_ReturnsErrInvalidRoleAssignment() {
	err := s.sut.Execute(context.Background(), AssignRoleCommandInput{UserID: 0, Role: authn.RoleSeller})

	s.ErrorIs(err, domainerrs.ErrInvalidRoleAssignment)
}

func (s *AssignRoleCommandTestSuite) TestExecute_UnknownRole_ReturnsErrInvalidRoleAssignment() {
	err := s.sut.Execute(context.Background(), AssignRoleCommandInput{UserID: 2, Role: "ghost"})

	s.ErrorIs(err, domainerrs.ErrInvalidRoleAssignment)
}

type RevokeRoleCommandTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	sut      *RevokeRoleCommand
}

func (s *RevokeRoleCommandTestSuite) SetupTest() {
	s.enforcer = newGTestEnforcer(
		gCasbinRule{"p", []string{"seller", "/api/v1/auctions", "POST"}},
		gCasbinRule{"g", []string{"2", "seller"}},
	)
	s.sut = NewRevokeRoleCommand(s.enforcer)
}

func TestRevokeRoleCommandTestSuite(t *testing.T) {
	suite.Run(t, new(RevokeRoleCommandTestSuite))
}

// Arrange/Act/Assert

func (s *RevokeRoleCommandTestSuite) TestExecute_ExistingBinding_RemovesIt() {
	// Act
	err := s.sut.Execute(context.Background(), RevokeRoleCommandInput{UserID: 2, Role: authn.RoleSeller})

	// Assert
	s.Require().NoError(err)
	allowed, enforceErr := s.enforcer.Enforce("2", "/api/v1/auctions", "POST")
	s.Require().NoError(enforceErr)
	s.False(allowed)
}

func (s *RevokeRoleCommandTestSuite) TestExecute_MissingField_ReturnsErrInvalidRoleAssignment() {
	err := s.sut.Execute(context.Background(), RevokeRoleCommandInput{UserID: 0, Role: authn.RoleSeller})

	s.ErrorIs(err, domainerrs.ErrInvalidRoleAssignment)
}

type RevokeAllRolesCommandTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	sut      *RevokeAllRolesCommand
}

func (s *RevokeAllRolesCommandTestSuite) SetupTest() {
	s.enforcer = newGTestEnforcer(
		gCasbinRule{"g", []string{"2", "seller"}},
		gCasbinRule{"g", []string{"2", "bidder"}},
	)
	s.sut = NewRevokeAllRolesCommand(s.enforcer)
}

func TestRevokeAllRolesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(RevokeAllRolesCommandTestSuite))
}

// Arrange/Act/Assert

func (s *RevokeAllRolesCommandTestSuite) TestExecute_RemovesEveryBindingForUser() {
	// Act
	err := s.sut.Execute(context.Background(), RevokeAllRolesCommandInput{UserID: 2})

	// Assert
	s.Require().NoError(err)
	bindings, enforceErr := s.enforcer.GetFilteredGroupingPolicy(0, "2")
	s.Require().NoError(enforceErr)
	s.Empty(bindings)
}

func (s *RevokeAllRolesCommandTestSuite) TestExecute_ZeroUserID_ReturnsErrInvalidRoleAssignment() {
	err := s.sut.Execute(context.Background(), RevokeAllRolesCommandInput{UserID: 0})

	s.ErrorIs(err, domainerrs.ErrInvalidRoleAssignment)
}
