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
)

// memAdapter is a minimal in-memory persist.Adapter used only by tests so the
// policy commands can be exercised without a Postgres instance.
type memAdapter struct {
	rules [][]string
}

func (a *memAdapter) LoadPolicy(m model.Model) error {
	for _, r := range a.rules {
		if err := persist.LoadPolicyLine("p, "+strings.Join(r, ", "), m); err != nil {
			return err
		}
	}

	return nil
}

func (a *memAdapter) SavePolicy(model.Model) error { return nil }

func (a *memAdapter) AddPolicy(_ string, _ string, rule []string) error {
	a.rules = append(a.rules, rule)

	return nil
}

func (a *memAdapter) RemovePolicy(_ string, _ string, rule []string) error {
	kept := a.rules[:0]
	for _, r := range a.rules {
		if len(r) != len(rule) {
			kept = append(kept, r)

			continue
		}
		match := true
		for i := range rule {
			if r[i] != rule[i] {
				match = false

				break
			}
		}
		if !match {
			kept = append(kept, r)
		}
	}
	a.rules = kept

	return nil
}

func (a *memAdapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return nil
}

func newTestEnforcer(rules ...[]string) *casbin.Enforcer {
	m, err := model.NewModelFromString(modelConfigForTest)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer(m, &memAdapter{rules: rules})
	if err != nil {
		panic(err)
	}

	return e
}

const modelConfigForTest = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch4(r.obj, p.obj) && (r.act == p.act || p.act == '*')
`

type AddPolicyCommandTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	sut      *AddPolicyCommand
}

func (s *AddPolicyCommandTestSuite) SetupTest() {
	s.enforcer = newTestEnforcer()
	s.sut = NewAddPolicyCommand(s.enforcer)
}

func TestAddPolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AddPolicyCommandTestSuite))
}

// Arrange/Act/Assert

func (s *AddPolicyCommandTestSuite) TestExecute_ValidInput_AddsPolicy() {
	// Act
	err := s.sut.Execute(context.Background(), AddPolicyCommandInput{
		Sub: "seller",
		Obj: "/api/v1/spus",
		Act: "POST",
	})

	// Assert
	s.Require().NoError(err)
	allowed, enforceErr := s.enforcer.Enforce("seller", "/api/v1/spus", "POST")
	s.Require().NoError(enforceErr)
	s.True(allowed)
}

func (s *AddPolicyCommandTestSuite) TestExecute_MissingSub_ReturnsErrInvalidPolicy() {
	err := s.sut.Execute(context.Background(), AddPolicyCommandInput{Sub: "", Obj: "/x", Act: "GET"})

	s.ErrorIs(err, domainerrs.ErrInvalidPolicy)
}

func (s *AddPolicyCommandTestSuite) TestExecute_MissingObj_ReturnsErrInvalidPolicy() {
	err := s.sut.Execute(context.Background(), AddPolicyCommandInput{Sub: "seller", Obj: "", Act: "GET"})

	s.ErrorIs(err, domainerrs.ErrInvalidPolicy)
}

func (s *AddPolicyCommandTestSuite) TestExecute_MissingAct_ReturnsErrInvalidPolicy() {
	err := s.sut.Execute(context.Background(), AddPolicyCommandInput{Sub: "seller", Obj: "/x", Act: ""})

	s.ErrorIs(err, domainerrs.ErrInvalidPolicy)
}

type RemovePolicyCommandTestSuite struct {
	suite.Suite
	enforcer *casbin.Enforcer
	sut      *RemovePolicyCommand
}

func (s *RemovePolicyCommandTestSuite) SetupTest() {
	s.enforcer = newTestEnforcer([]string{"seller", "/api/v1/spus", "POST"})
	s.sut = NewRemovePolicyCommand(s.enforcer)
}

func TestRemovePolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(RemovePolicyCommandTestSuite))
}

// Arrange/Act/Assert

func (s *RemovePolicyCommandTestSuite) TestExecute_ExistingPolicy_RemovesIt() {
	// Act
	err := s.sut.Execute(context.Background(), RemovePolicyCommandInput{
		Sub: "seller",
		Obj: "/api/v1/spus",
		Act: "POST",
	})

	// Assert
	s.Require().NoError(err)
	allowed, enforceErr := s.enforcer.Enforce("seller", "/api/v1/spus", "POST")
	s.Require().NoError(enforceErr)
	s.False(allowed)
}

func (s *RemovePolicyCommandTestSuite) TestExecute_MissingField_ReturnsErrInvalidPolicy() {
	err := s.sut.Execute(context.Background(), RemovePolicyCommandInput{Sub: "", Obj: "/x", Act: "GET"})

	s.ErrorIs(err, domainerrs.ErrInvalidPolicy)
}
