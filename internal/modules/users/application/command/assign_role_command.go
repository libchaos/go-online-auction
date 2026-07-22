package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/shared/modules/authn"
)

type AssignRoleCommandInput struct {
	UserID uint64
	Role   string
}

type AssignRoleCommand struct {
	enforcer *casbin.Enforcer
}

func NewAssignRoleCommand(enforcer *casbin.Enforcer) *AssignRoleCommand {
	return &AssignRoleCommand{enforcer: enforcer}
}

func (c *AssignRoleCommand) Execute(_ context.Context, input AssignRoleCommandInput) error {
	if input.UserID == 0 || !isValidRoleAssignmentRole(input.Role) {
		return errs.ErrInvalidRoleAssignment
	}

	if _, err := c.enforcer.AddGroupingPolicy(strconv.FormatUint(input.UserID, 10), input.Role); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	return nil
}

func isValidRoleAssignmentRole(role string) bool {
	switch role {
	case authn.RoleAdmin, authn.RoleSeller, authn.RoleBidder:
		return true
	default:
		return false
	}
}
