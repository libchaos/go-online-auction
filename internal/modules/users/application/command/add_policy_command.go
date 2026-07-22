package command

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"

	"auction/internal/modules/users/domain/errs"
)

type AddPolicyCommandInput struct {
	Sub string
	Obj string
	Act string
}

type AddPolicyCommand struct {
	enforcer *casbin.Enforcer
}

func NewAddPolicyCommand(enforcer *casbin.Enforcer) *AddPolicyCommand {
	return &AddPolicyCommand{enforcer: enforcer}
}

func (c *AddPolicyCommand) Execute(_ context.Context, input AddPolicyCommandInput) error {
	if input.Sub == "" || input.Obj == "" || input.Act == "" {
		return errs.ErrInvalidPolicy
	}

	if _, err := c.enforcer.AddPolicy(input.Sub, input.Obj, input.Act); err != nil {
		return fmt.Errorf("add policy: %w", err)
	}

	return nil
}
