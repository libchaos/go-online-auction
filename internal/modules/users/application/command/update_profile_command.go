package command

import (
	"context"
	"time"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

type UpdateProfileCommandInput struct {
	UserID uint64
	Name   string
}

type UpdateProfileCommandOutput struct {
	ID        uint64
	Name      string
	Email     string
	Role      string
	Status    string
	UpdatedAt time.Time
}

type UpdateProfileCommand struct {
	userRepository ports.UserRepository
	logger         logger.Logger
}

func NewUpdateProfileCommand(
	userRepository ports.UserRepository,
	logger logger.Logger,
) *UpdateProfileCommand {
	return &UpdateProfileCommand{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *UpdateProfileCommand) Execute(
	ctx context.Context,
	input UpdateProfileCommandInput,
) (UpdateProfileCommandOutput, error) {
	if input.UserID == 0 {
		return UpdateProfileCommandOutput{}, errs.ErrUserIDRequired
	}

	user, err := c.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to find user")
		return UpdateProfileCommandOutput{}, err
	}

	if updateErr := user.UpdateProfile(input.Name); updateErr != nil {
		c.logger.Error().Err(updateErr).Msg("failed to update profile")
		return UpdateProfileCommandOutput{}, updateErr
	}

	if updateErr := c.userRepository.Update(ctx, user); updateErr != nil {
		c.logger.Error().Err(updateErr).Msg("failed to persist user")
		return UpdateProfileCommandOutput{}, updateErr
	}

	role := user.Role()
	status := user.Status()
	return UpdateProfileCommandOutput{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		Role:      role.String(),
		Status:    status.String(),
		UpdatedAt: user.UpdatedAt(),
	}, nil
}
