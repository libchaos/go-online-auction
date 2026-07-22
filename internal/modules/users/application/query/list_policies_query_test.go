package query

import (
	"context"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/stretchr/testify/suite"
)

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
	m, err := model.NewModelFromString(modelConfigForQueryTest)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer(m, &memAdapter{rules: rules})
	if err != nil {
		panic(err)
	}

	return e
}

const modelConfigForQueryTest = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch4(r.obj, p.obj) && (r.act == p.act || p.act == '*')
`

type ListPoliciesQueryTestSuite struct {
	suite.Suite
	sut *ListPoliciesQuery
}

func (s *ListPoliciesQueryTestSuite) SetupTest() {
	enforcer := newTestEnforcer(
		[]string{"admin", "/api/v1/*", "*"},
		[]string{"seller", "/api/v1/auctions", "POST"},
	)
	s.sut = NewListPoliciesQuery(enforcer)
}

func TestListPoliciesQueryTestSuite(t *testing.T) {
	suite.Run(t, new(ListPoliciesQueryTestSuite))
}

// Arrange/Act/Assert

func (s *ListPoliciesQueryTestSuite) TestExecute_ReturnsAllStoredPolicies() {
	// Act
	views, err := s.sut.Execute(context.Background())

	// Assert
	s.Require().NoError(err)
	s.Len(views, 2)

	bySubject := map[string]PolicyView{}
	for _, v := range views {
		bySubject[v.Sub] = v
	}
	s.Equal("/api/v1/*", bySubject["admin"].Obj)
	s.Equal("*", bySubject["admin"].Act)
	s.Equal("/api/v1/auctions", bySubject["seller"].Obj)
	s.Equal("POST", bySubject["seller"].Act)
}

func (s *ListPoliciesQueryTestSuite) TestExecute_NoPolicies_ReturnsEmptySlice() {
	s.sut = NewListPoliciesQuery(newTestEnforcer())

	views, err := s.sut.Execute(context.Background())

	s.Require().NoError(err)
	s.Empty(views)
}
