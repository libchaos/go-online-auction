package command

import (
	"context"
	"time"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

type UpdateUserRoleCommandInput struct {
	UserID uint64
	Role   string
}

type UpdateUserRoleCommandOutput struct {
	ID        uint64
	Name      string
	Email     string
	Role      string
	Status    string
	UpdatedAt time.Time
}

type UpdateUserRoleCommand struct {
	userRepository ports.UserRepository
	logger         logger.Logger
}

func NewUpdateUserRoleCommand(
	userRepository ports.UserRepository,
	logger logger.Logger,
) *UpdateUserRoleCommand {
	return &UpdateUserRoleCommand{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *UpdateUserRoleCommand) Execute(
	ctx context.Context,
	input UpdateUserRoleCommandInput,
) (UpdateUserRoleCommandOutput, error) {
	if input.UserID == 0 {
		return UpdateUserRoleCommandOutput{}, errs.ErrUserIDRequired
	}

	role, err := enum.NewRoleEnum(input.Role)
	if err != nil {
		c.logger.Error().Err(err).Msg("invalid role")
		return UpdateUserRoleCommandOutput{}, err
	}

	user, err := c.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to find user")
		return UpdateUserRoleCommandOutput{}, err
	}

	user.ChangeRole(role)

	if updateErr := c.userRepository.Update(ctx, user); updateErr != nil {
		c.logger.Error().Err(updateErr).Msg("failed to persist user")
		return UpdateUserRoleCommandOutput{}, updateErr
	}

	updatedRole := user.Role()
	status := user.Status()
	return UpdateUserRoleCommandOutput{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		Role:      updatedRole.String(),
		Status:    status.String(),
		UpdatedAt: user.UpdatedAt(),
	}, nil
}
