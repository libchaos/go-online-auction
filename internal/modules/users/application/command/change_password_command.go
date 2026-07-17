package command

import (
	"context"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

type ChangePasswordCommandInput struct {
	UserID          uint64
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordCommand struct {
	userRepository         ports.UserRepository
	refreshTokenRepository ports.RefreshTokenRepository
	passwordHasher         ports.PasswordHasher
	logger                 logger.Logger
}

func NewChangePasswordCommand(
	userRepository ports.UserRepository,
	refreshTokenRepository ports.RefreshTokenRepository,
	passwordHasher ports.PasswordHasher,
	logger logger.Logger,
) *ChangePasswordCommand {
	return &ChangePasswordCommand{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		passwordHasher:         passwordHasher,
		logger:                 logger,
	}
}

func (c *ChangePasswordCommand) Execute(ctx context.Context, input ChangePasswordCommandInput) error {
	if input.UserID == 0 {
		return errs.ErrUserIDRequired
	}

	if err := ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	user, err := c.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to find user")
		return err
	}

	passwordHash := user.PasswordHash()
	if passwordHash == nil {
		return errs.ErrInvalidCredentials
	}

	if compareErr := c.passwordHasher.Compare(*passwordHash, input.CurrentPassword); compareErr != nil {
		return errs.ErrInvalidCredentials
	}

	newHash, err := c.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to hash password")
		return err
	}

	if changeErr := user.ChangePasswordHash(newHash); changeErr != nil {
		c.logger.Error().Err(changeErr).Msg("failed to change password hash")
		return changeErr
	}

	if updateErr := c.userRepository.Update(ctx, user); updateErr != nil {
		c.logger.Error().Err(updateErr).Msg("failed to persist user")
		return updateErr
	}

	// invalidate every active session after a password change
	if revokeErr := c.refreshTokenRepository.RevokeAllForUser(ctx, user.ID()); revokeErr != nil {
		c.logger.Error().Err(revokeErr).Msg("failed to revoke refresh tokens after password change")
		return revokeErr
	}

	return nil
}
