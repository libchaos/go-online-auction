package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"

	"auction/internal/modules/users/domain/errs"
)

type RevokeAllRolesCommandInput struct {
	UserID uint64
}

type RevokeAllRolesCommand struct {
	enforcer *casbin.Enforcer
}

func NewRevokeAllRolesCommand(enforcer *casbin.Enforcer) *RevokeAllRolesCommand {
	return &RevokeAllRolesCommand{enforcer: enforcer}
}

func (c *RevokeAllRolesCommand) Execute(_ context.Context, input RevokeAllRolesCommandInput) error {
	if input.UserID == 0 {
		return errs.ErrInvalidRoleAssignment
	}

	if _, err := c.enforcer.RemoveFilteredGroupingPolicy(0, strconv.FormatUint(input.UserID, 10)); err != nil {
		return fmt.Errorf("revoke all roles: %w", err)
	}

	return nil
}
