package command

import (
	"context"
	"errors"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

type LogoutCommandInput struct {
	RefreshToken string
}

type LogoutCommand struct {
	refreshTokenRepository ports.RefreshTokenRepository
	tokenService           ports.TokenService
	logger                 logger.Logger
}

func NewLogoutCommand(
	refreshTokenRepository ports.RefreshTokenRepository,
	tokenService ports.TokenService,
	logger logger.Logger,
) *LogoutCommand {
	return &LogoutCommand{
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		logger:                 logger,
	}
}

func (c *LogoutCommand) Execute(ctx context.Context, input LogoutCommandInput) error {
	tokenHash := c.tokenService.HashRefreshToken(input.RefreshToken)

	token, err := c.refreshTokenRepository.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, errs.ErrRefreshTokenNotFound) {
			// logout is idempotent
			return nil
		}
		c.logger.Error().Err(err).Msg("failed to find refresh token")
		return err
	}

	if token.IsRevoked() {
		return nil
	}

	token.Revoke(nil)
	if updateErr := c.refreshTokenRepository.Update(ctx, token); updateErr != nil {
		c.logger.Error().Err(updateErr).Msg("failed to revoke refresh token")
		return updateErr
	}

	return nil
}
