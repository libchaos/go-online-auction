package query

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
)

// PolicyView is a single RBAC policy rule as exposed by the management API.
type PolicyView struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}

// minPolicyFields is the number of components in a "p" rule (sub, obj, act).
const minPolicyFields = 3

type ListPoliciesQuery struct {
	enforcer *casbin.Enforcer
}

func NewListPoliciesQuery(enforcer *casbin.Enforcer) *ListPoliciesQuery {
	return &ListPoliciesQuery{enforcer: enforcer}
}

func (q *ListPoliciesQuery) Execute(_ context.Context) ([]PolicyView, error) {
	rules, err := q.enforcer.GetPolicy()
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	views := make([]PolicyView, 0, len(rules))
	for _, rule := range rules {
		if len(rule) < minPolicyFields {
			continue
		}

		views = append(views, PolicyView{
			Sub: rule[0],
			Obj: rule[1],
			Act: rule[2],
		})
	}

	return views, nil
}
