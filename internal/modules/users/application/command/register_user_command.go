package command

import (
	"context"
	"time"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72 // bcrypt input limit
)

type RegisterUserCommandInput struct {
	Name     string
	Email    string
	Password string
	Role     string
}

type RegisterUserCommandOutput struct {
	ID        uint64
	Name      string
	Email     string
	Role      string
	Status    string
	CreatedAt time.Time
}

type RegisterUserCommand struct {
	userRepository ports.UserRepository
	passwordHasher ports.PasswordHasher
	logger         logger.Logger
}

func NewRegisterUserCommand(
	userRepository ports.UserRepository,
	passwordHasher ports.PasswordHasher,
	logger logger.Logger,
) *RegisterUserCommand {
	return &RegisterUserCommand{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		logger:         logger,
	}
}

func (c *RegisterUserCommand) Execute(
	ctx context.Context,
	input RegisterUserCommandInput,
) (RegisterUserCommandOutput, error) {
	if err := ValidatePassword(input.Password); err != nil {
		c.logger.Error().Err(err).Msg("invalid password")
		return RegisterUserCommandOutput{}, err
	}

	roleValue := input.Role
	if roleValue == "" {
		roleValue = enum.EnumRoleBidder
	}

	// admin accounts must be created out of band (create-admin command)
	if roleValue == enum.EnumRoleAdmin {
		c.logger.Error().Msg("attempt to register with admin role")
		return RegisterUserCommandOutput{}, errs.ErrInvalidRole
	}

	role, err := enum.NewRoleEnum(roleValue)
	if err != nil {
		c.logger.Error().Err(err).Msg("invalid role")
		return RegisterUserCommandOutput{}, err
	}

	passwordHash, err := c.passwordHasher.Hash(input.Password)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to hash password")
		return RegisterUserCommandOutput{}, err
	}

	user, err := model.NewUserModel(input.Name, input.Email, passwordHash, role)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create user domain model")
		return RegisterUserCommandOutput{}, err
	}

	persistedUser, err := c.userRepository.Create(ctx, user)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to persist user")
		return RegisterUserCommandOutput{}, err
	}

	persistedRole := persistedUser.Role()
	persistedStatus := persistedUser.Status()
	return RegisterUserCommandOutput{
		ID:        persistedUser.ID(),
		Name:      persistedUser.Name(),
		Email:     persistedUser.Email(),
		Role:      persistedRole.String(),
		Status:    persistedStatus.String(),
		CreatedAt: persistedUser.CreatedAt(),
	}, nil
}

// ValidatePassword validates the plaintext password rules
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return errs.ErrPasswordTooShort
	}

	if len(password) > maxPasswordBytes {
		return errs.ErrPasswordTooLong
	}

	return nil
}
