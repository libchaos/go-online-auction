package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"

	"auction/internal/modules/users/domain/errs"
)

type RevokeRoleCommandInput struct {
	UserID uint64
	Role   string
}

type RevokeRoleCommand struct {
	enforcer *casbin.Enforcer
}

func NewRevokeRoleCommand(enforcer *casbin.Enforcer) *RevokeRoleCommand {
	return &RevokeRoleCommand{enforcer: enforcer}
}

func (c *RevokeRoleCommand) Execute(_ context.Context, input RevokeRoleCommandInput) error {
	if input.UserID == 0 || input.Role == "" {
		return errs.ErrInvalidRoleAssignment
	}

	if _, err := c.enforcer.RemoveGroupingPolicy(strconv.FormatUint(input.UserID, 10), input.Role); err != nil {
		return fmt.Errorf("revoke role: %w", err)
	}

	return nil
}
