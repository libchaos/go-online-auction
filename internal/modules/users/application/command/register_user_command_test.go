package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/tests/mocks"
)

type RegisterUserCommandTestSuite struct {
	suite.Suite
	sut                *command.RegisterUserCommand
	userRepositoryMock *mocks.MockUserRepository
	passwordHasherMock *mocks.MockPasswordHasher
	loggerMock         *mocks.MockLogger
	mockPersistedUser  model.UserModel
}

func (s *RegisterUserCommandTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.passwordHasherMock = mocks.NewMockPasswordHasher(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewRegisterUserCommand(
		s.userRepositoryMock,
		s.passwordHasherMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	passwordHash := "hashed-password"
	now := time.Now().UTC()
	s.mockPersistedUser, _ = model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 1, now, now,
	)
}

func TestRegisterUserCommandSuite(t *testing.T) {
	suite.Run(t, new(RegisterUserCommandTestSuite))
}

func (s *RegisterUserCommandTestSuite) TestExecute_ValidInput_ReturnsCreatedUser() {
	// Arrange
	ctx := context.Background()
	input := command.RegisterUserCommandInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "super-secret",
	}

	s.passwordHasherMock.On("Hash", input.Password).Return("hashed-password", nil)
	s.userRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.UserModel")).
		Return(s.mockPersistedUser, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.ID)
	s.Equal("john@example.com", output.Email)
	s.Equal(enum.EnumRoleBidder, output.Role)
	s.Equal(enum.EnumUserStatusActive, output.Status)
}

func (s *RegisterUserCommandTestSuite) TestExecute_ShortPassword_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.RegisterUserCommandInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "short",
	}

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordTooShort)
	s.Equal(command.RegisterUserCommandOutput{}, output)
}

func (s *RegisterUserCommandTestSuite) TestExecute_AdminRole_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.RegisterUserCommandInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "super-secret",
		Role:     enum.EnumRoleAdmin,
	}

	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidRole)
	s.Equal(command.RegisterUserCommandOutput{}, output)
}

func (s *RegisterUserCommandTestSuite) TestExecute_DuplicateEmail_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.RegisterUserCommandInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "super-secret",
	}

	emptyUser := model.UserModel{}
	s.passwordHasherMock.On("Hash", input.Password).Return("hashed-password", nil)
	s.userRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.UserModel")).
		Return(emptyUser, errs.ErrEmailAlreadyExists)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrEmailAlreadyExists)
	s.Equal(command.RegisterUserCommandOutput{}, output)
}

func (s *RegisterUserCommandTestSuite) TestExecute_HasherError_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.RegisterUserCommandInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "super-secret",
	}

	hasherErr := errors.New("hasher error")
	s.passwordHasherMock.On("Hash", input.Password).Return("", hasherErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, hasherErr)
	s.Equal(command.RegisterUserCommandOutput{}, output)
}
