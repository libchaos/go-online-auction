package command

import (
	"context"
	"errors"
	"time"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

type RefreshTokenCommandInput struct {
	RefreshToken string
}

type RefreshTokenCommandOutput struct {
	UserID               uint64
	Role                 string
	AccessToken          string
	AccessTokenExpiresAt time.Time
	RefreshToken         string
}

type RefreshTokenCommand struct {
	userRepository         ports.UserRepository
	refreshTokenRepository ports.RefreshTokenRepository
	tokenService           ports.TokenService
	logger                 logger.Logger
}

func NewRefreshTokenCommand(
	userRepository ports.UserRepository,
	refreshTokenRepository ports.RefreshTokenRepository,
	tokenService ports.TokenService,
	logger logger.Logger,
) *RefreshTokenCommand {
	return &RefreshTokenCommand{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		tokenService:           tokenService,
		logger:                 logger,
	}
}

func (c *RefreshTokenCommand) Execute(
	ctx context.Context,
	input RefreshTokenCommandInput,
) (RefreshTokenCommandOutput, error) {
	tokenHash := c.tokenService.HashRefreshToken(input.RefreshToken)

	currentToken, err := c.refreshTokenRepository.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, errs.ErrRefreshTokenNotFound) {
			return RefreshTokenCommandOutput{}, errs.ErrRefreshTokenInvalid
		}
		c.logger.Error().Err(err).Msg("failed to find refresh token")
		return RefreshTokenCommandOutput{}, err
	}

	if currentToken.IsRevoked() {
		// reuse of a rotated token: revoke every session for this user
		c.logger.Error().Msg("refresh token reuse detected, revoking all sessions for user")
		if revokeErr := c.refreshTokenRepository.RevokeAllForUser(ctx, currentToken.UserID()); revokeErr != nil {
			c.logger.Error().Err(revokeErr).Msg("failed to revoke all refresh tokens for user")
		}
		return RefreshTokenCommandOutput{}, errs.ErrRefreshTokenInvalid
	}

	if currentToken.IsExpired(time.Now().UTC()) {
		return RefreshTokenCommandOutput{}, errs.ErrRefreshTokenInvalid
	}

	user, err := c.userRepository.FindByID(ctx, currentToken.UserID())
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to find user for refresh token")
		return RefreshTokenCommandOutput{}, err
	}

	if !user.IsActive() {
		return RefreshTokenCommandOutput{}, errs.ErrUserInactive
	}

	return c.rotateTokens(ctx, user, currentToken)
}

func (c *RefreshTokenCommand) rotateTokens(
	ctx context.Context,
	user model.UserModel,
	currentToken model.RefreshTokenModel,
) (RefreshTokenCommandOutput, error) {
	role := user.Role()

	accessToken, accessExpiresAt, err := c.tokenService.GenerateAccessToken(
		user.ID(),
		role.String(),
		user.Email(),
	)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to generate access token")
		return RefreshTokenCommandOutput{}, err
	}

	rawRefreshToken, refreshTokenHash, refreshExpiresAt, err := c.tokenService.GenerateRefreshToken()
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to generate refresh token")
		return RefreshTokenCommandOutput{}, err
	}

	newToken, err := model.NewRefreshTokenModel(user.ID(), refreshTokenHash, refreshExpiresAt)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create refresh token domain model")
		return RefreshTokenCommandOutput{}, err
	}

	persistedToken, err := c.refreshTokenRepository.Create(ctx, newToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist refresh token")
		return RefreshTokenCommandOutput{}, err
	}

	newTokenID := persistedToken.ID()
	currentToken.Revoke(&newTokenID)
	if updateErr := c.refreshTokenRepository.Update(ctx, currentToken); updateErr != nil {
		c.logger.Error().Err(updateErr).Msg("failed to revoke rotated refresh token")
		return RefreshTokenCommandOutput{}, updateErr
	}

	return RefreshTokenCommandOutput{
		UserID:               user.ID(),
		Role:                 role.String(),
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessExpiresAt,
		RefreshToken:         rawRefreshToken,
	}, nil
}
