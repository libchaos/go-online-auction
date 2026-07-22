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

type gQueryRule struct {
	ptype string
	rule  []string
}

type gQueryMemAdapter struct {
	rules []gQueryRule
}

func (a *gQueryMemAdapter) LoadPolicy(m model.Model) error {
	for _, r := range a.rules {
		line := r.ptype + ", " + strings.Join(r.rule, ", ")
		if err := persist.LoadPolicyLine(line, m); err != nil {
			return err
		}
	}

	return nil
}

func (a *gQueryMemAdapter) SavePolicy(model.Model) error { return nil }

func (a *gQueryMemAdapter) AddPolicy(_ string, ptype string, rule []string) error {
	a.rules = append(a.rules, gQueryRule{ptype: ptype, rule: rule})

	return nil
}

func (a *gQueryMemAdapter) RemovePolicy(_ string, ptype string, rule []string) error {
	kept := a.rules[:0]
	for _, r := range a.rules {
		if r.ptype == ptype && gQuerySameRule(r.rule, rule) {
			continue
		}
		kept = append(kept, r)
	}
	a.rules = kept

	return nil
}

func (a *gQueryMemAdapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return nil
}

func gQuerySameRule(a, b []string) bool {
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

func newGQueryEnforcer(rules ...gQueryRule) *casbin.Enforcer {
	m, err := model.NewModelFromString(modelConfigForGQueryTest)
	if err != nil {
		panic(err)
	}
	e, err := casbin.NewEnforcer(m, &gQueryMemAdapter{rules: rules})
	if err != nil {
		panic(err)
	}

	return e
}

const modelConfigForGQueryTest = `
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

type ListRoleAssignmentsQueryTestSuite struct {
	suite.Suite
	sut *ListRoleAssignmentsQuery
}

func (s *ListRoleAssignmentsQueryTestSuite) SetupTest() {
	enforcer := newGQueryEnforcer(
		gQueryRule{"g", []string{"1", "admin"}},
		gQueryRule{"g", []string{"2", "seller"}},
	)
	s.sut = NewListRoleAssignmentsQuery(enforcer)
}

func TestListRoleAssignmentsQueryTestSuite(t *testing.T) {
	suite.Run(t, new(ListRoleAssignmentsQueryTestSuite))
}

// Arrange/Act/Assert

func (s *ListRoleAssignmentsQueryTestSuite) TestExecute_NoFilter_ReturnsAllBindings() {
	views, err := s.sut.Execute(context.Background(), 0)

	s.Require().NoError(err)
	s.Len(views, 2)

	byUser := map[uint64]string{}
	for _, v := range views {
		byUser[v.UserID] = v.Role
	}
	s.Equal("admin", byUser[1])
	s.Equal("seller", byUser[2])
}

func (s *ListRoleAssignmentsQueryTestSuite) TestExecute_ByUserID_ReturnsOnlyThatUser() {
	views, err := s.sut.Execute(context.Background(), 1)

	s.Require().NoError(err)
	s.Len(views, 1)
	s.Equal(uint64(1), views[0].UserID)
	s.Equal("admin", views[0].Role)
}
