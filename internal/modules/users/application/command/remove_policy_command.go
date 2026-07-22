package command

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"

	"auction/internal/modules/users/domain/errs"
)

type RemovePolicyCommandInput struct {
	Sub string
	Obj string
	Act string
}

type RemovePolicyCommand struct {
	enforcer *casbin.Enforcer
}

func NewRemovePolicyCommand(enforcer *casbin.Enforcer) *RemovePolicyCommand {
	return &RemovePolicyCommand{enforcer: enforcer}
}

func (c *RemovePolicyCommand) Execute(_ context.Context, input RemovePolicyCommandInput) error {
	if input.Sub == "" || input.Obj == "" || input.Act == "" {
		return errs.ErrInvalidPolicy
	}

	if _, err := c.enforcer.RemovePolicy(input.Sub, input.Obj, input.Act); err != nil {
		return fmt.Errorf("remove policy: %w", err)
	}

	return nil
}
