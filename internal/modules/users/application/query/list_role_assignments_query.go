package query

import (
	"context"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"
)

// RoleAssignmentView is a single user->role binding as exposed by the API.
type RoleAssignmentView struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
}

// minGroupingFields is the number of components in a "g" rule (user, role).
const minGroupingFields = 2

type ListRoleAssignmentsQuery struct {
	enforcer *casbin.Enforcer
}

func NewListRoleAssignmentsQuery(enforcer *casbin.Enforcer) *ListRoleAssignmentsQuery {
	return &ListRoleAssignmentsQuery{enforcer: enforcer}
}

// Execute returns the user->role bindings. When userID is zero, every binding is
// returned; otherwise only the bindings for that user are returned.
func (q *ListRoleAssignmentsQuery) Execute(_ context.Context, userID uint64) ([]RoleAssignmentView, error) {
	var rules [][]string
	var err error
	if userID == 0 {
		rules, err = q.enforcer.GetGroupingPolicy()
	} else {
		rules, err = q.enforcer.GetFilteredGroupingPolicy(0, strconv.FormatUint(userID, 10))
	}
	if err != nil {
		return nil, fmt.Errorf("list role assignments: %w", err)
	}

	views := make([]RoleAssignmentView, 0, len(rules))
	for _, rule := range rules {
		if len(rule) < minGroupingFields {
			continue
		}

		id, parseErr := strconv.ParseUint(rule[0], 10, 64)
		if parseErr != nil {
			continue
		}

		views = append(views, RoleAssignmentView{UserID: id, Role: rule[1]})
	}

	return views, nil
}
