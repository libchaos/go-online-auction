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

type LoginCommandInput struct {
	Email    string
	Password string
}

type LoginCommandOutput struct {
	UserID               uint64
	Role                 string
	AccessToken          string
	AccessTokenExpiresAt time.Time
	RefreshToken         string
}

type LoginCommand struct {
	userRepository         ports.UserRepository
	refreshTokenRepository ports.RefreshTokenRepository
	passwordHasher         ports.PasswordHasher
	tokenService           ports.TokenService
	logger                 logger.Logger
}

func NewLoginCommand(
	userRepository ports.UserRepository,
	refreshTokenRepository ports.RefreshTokenRepository,
	passwordHasher ports.PasswordHasher,
	tokenService ports.TokenService,
	logger logger.Logger,
) *LoginCommand {
	return &LoginCommand{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		passwordHasher:         passwordHasher,
		tokenService:           tokenService,
		logger:                 logger,
	}
}

func (c *LoginCommand) Execute(ctx context.Context, input LoginCommandInput) (LoginCommandOutput, error) {
	user, err := c.userRepository.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			// do not leak whether the email exists
			return LoginCommandOutput{}, errs.ErrInvalidCredentials
		}
		c.logger.Error().Err(err).Msg("failed to find user by email")
		return LoginCommandOutput{}, err
	}

	passwordHash := user.PasswordHash()
	if passwordHash == nil {
		return LoginCommandOutput{}, errs.ErrInvalidCredentials
	}

	if compareErr := c.passwordHasher.Compare(*passwordHash, input.Password); compareErr != nil {
		return LoginCommandOutput{}, errs.ErrInvalidCredentials
	}

	if !user.IsActive() {
		return LoginCommandOutput{}, errs.ErrUserInactive
	}

	return c.issueTokens(ctx, user)
}

func (c *LoginCommand) issueTokens(ctx context.Context, user model.UserModel) (LoginCommandOutput, error) {
	role := user.Role()

	accessToken, accessExpiresAt, err := c.tokenService.GenerateAccessToken(
		user.ID(),
		role.String(),
		user.Email(),
	)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to generate access token")
		return LoginCommandOutput{}, err
	}

	rawRefreshToken, refreshTokenHash, refreshExpiresAt, err := c.tokenService.GenerateRefreshToken()
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to generate refresh token")
		return LoginCommandOutput{}, err
	}

	refreshToken, err := model.NewRefreshTokenModel(user.ID(), refreshTokenHash, refreshExpiresAt)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create refresh token domain model")
		return LoginCommandOutput{}, err
	}

	if _, createErr := c.refreshTokenRepository.Create(ctx, refreshToken); createErr != nil {
		c.logger.Error().Err(createErr).Msg("failed to persist refresh token")
		return LoginCommandOutput{}, createErr
	}

	return LoginCommandOutput{
		UserID:               user.ID(),
		Role:                 role.String(),
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessExpiresAt,
		RefreshToken:         rawRefreshToken,
	}, nil
}
