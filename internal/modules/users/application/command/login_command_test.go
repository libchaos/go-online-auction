package command_test

import (
	"context"
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

type LoginCommandTestSuite struct {
	suite.Suite
	sut                        *command.LoginCommand
	userRepositoryMock         *mocks.MockUserRepository
	refreshTokenRepositoryMock *mocks.MockRefreshTokenRepository
	passwordHasherMock         *mocks.MockPasswordHasher
	tokenServiceMock           *mocks.MockTokenService
	loggerMock                 *mocks.MockLogger
	mockUser                   model.UserModel
	mockInactiveUser           model.UserModel
	mockAccessExpiresAt        time.Time
	mockRefreshExpiresAt       time.Time
}

func (s *LoginCommandTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.refreshTokenRepositoryMock = mocks.NewMockRefreshTokenRepository(s.T())
	s.passwordHasherMock = mocks.NewMockPasswordHasher(s.T())
	s.tokenServiceMock = mocks.NewMockTokenService(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewLoginCommand(
		s.userRepositoryMock,
		s.refreshTokenRepositoryMock,
		s.passwordHasherMock,
		s.tokenServiceMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
	activeStatus, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	blockedStatus, _ := enum.NewUserStatusEnum(enum.EnumUserStatusBlocked)
	passwordHash := "hashed-password"
	now := time.Now().UTC()
	s.mockAccessExpiresAt = now.Add(15 * time.Minute)
	s.mockRefreshExpiresAt = now.Add(168 * time.Hour)

	s.mockUser, _ = model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, activeStatus, nil, nil, 1, now, now,
	)
	s.mockInactiveUser, _ = model.RestoreUserModel(
		2, "Jane Doe", "jane@example.com", &passwordHash, role, blockedStatus, nil, nil, 1, now, now,
	)
}

func TestLoginCommandSuite(t *testing.T) {
	suite.Run(t, new(LoginCommandTestSuite))
}

func (s *LoginCommandTestSuite) TestExecute_ValidCredentials_ReturnsTokens() {
	// Arrange
	ctx := context.Background()
	input := command.LoginCommandInput{Email: "john@example.com", Password: "super-secret"}

	persistedToken, _ := model.RestoreRefreshTokenModel(
		10, 1, "refresh-hash", s.mockRefreshExpiresAt, nil, nil, time.Now().UTC(),
	)

	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(s.mockUser, nil)
	s.passwordHasherMock.On("Compare", "hashed-password", input.Password).Return(nil)
	s.tokenServiceMock.
		On("GenerateAccessToken", uint64(1), enum.EnumRoleBidder, "john@example.com").
		Return("access-token", s.mockAccessExpiresAt, nil)
	s.tokenServiceMock.
		On("GenerateRefreshToken").
		Return("raw-refresh-token", "refresh-hash", s.mockRefreshExpiresAt, nil)
	s.refreshTokenRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.RefreshTokenModel")).
		Return(persistedToken, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), output.UserID)
	s.Equal("access-token", output.AccessToken)
	s.Equal("raw-refresh-token", output.RefreshToken)
	s.Equal(s.mockAccessExpiresAt, output.AccessTokenExpiresAt)
}

func (s *LoginCommandTestSuite) TestExecute_UnknownEmail_ReturnsInvalidCredentials() {
	// Arrange
	ctx := context.Background()
	input := command.LoginCommandInput{Email: "unknown@example.com", Password: "super-secret"}

	emptyUser := model.UserModel{}
	s.userRepositoryMock.
		On("FindByEmail", mock.Anything, input.Email).
		Return(emptyUser, errs.ErrUserNotFound)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Equal(command.LoginCommandOutput{}, output)
}

func (s *LoginCommandTestSuite) TestExecute_WrongPassword_ReturnsInvalidCredentials() {
	// Arrange
	ctx := context.Background()
	input := command.LoginCommandInput{Email: "john@example.com", Password: "wrong-password"}

	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(s.mockUser, nil)
	s.passwordHasherMock.
		On("Compare", "hashed-password", input.Password).
		Return(errs.ErrInvalidCredentials)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
	s.Equal(command.LoginCommandOutput{}, output)
}

func (s *LoginCommandTestSuite) TestExecute_InactiveUser_ReturnsUserInactive() {
	// Arrange
	ctx := context.Background()
	input := command.LoginCommandInput{Email: "jane@example.com", Password: "super-secret"}

	s.userRepositoryMock.On("FindByEmail", mock.Anything, input.Email).Return(s.mockInactiveUser, nil)
	s.passwordHasherMock.On("Compare", "hashed-password", input.Password).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserInactive)
	s.Equal(command.LoginCommandOutput{}, output)
}
