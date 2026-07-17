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

type ChangePasswordCommandTestSuite struct {
	suite.Suite
	sut                        *command.ChangePasswordCommand
	userRepositoryMock         *mocks.MockUserRepository
	refreshTokenRepositoryMock *mocks.MockRefreshTokenRepository
	passwordHasherMock         *mocks.MockPasswordHasher
	loggerMock                 *mocks.MockLogger
	mockUser                   model.UserModel
}

func (s *ChangePasswordCommandTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.refreshTokenRepositoryMock = mocks.NewMockRefreshTokenRepository(s.T())
	s.passwordHasherMock = mocks.NewMockPasswordHasher(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewChangePasswordCommand(
		s.userRepositoryMock,
		s.refreshTokenRepositoryMock,
		s.passwordHasherMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	passwordHash := "old-hash"
	now := time.Now().UTC()
	s.mockUser, _ = model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 1, now, now,
	)
}

func TestChangePasswordCommandSuite(t *testing.T) {
	suite.Run(t, new(ChangePasswordCommandTestSuite))
}

func (s *ChangePasswordCommandTestSuite) TestExecute_ValidInput_ChangesPasswordAndRevokesSessions() {
	// Arrange
	ctx := context.Background()
	input := command.ChangePasswordCommandInput{
		UserID:          1,
		CurrentPassword: "old-password",
		NewPassword:     "new-password",
	}

	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockUser, nil)
	s.passwordHasherMock.On("Compare", "old-hash", input.CurrentPassword).Return(nil)
	s.passwordHasherMock.On("Hash", input.NewPassword).Return("new-hash", nil)
	s.userRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.UserModel")).
		Return(nil)
	s.refreshTokenRepositoryMock.On("RevokeAllForUser", mock.Anything, uint64(1)).Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}

func (s *ChangePasswordCommandTestSuite) TestExecute_WrongCurrentPassword_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.ChangePasswordCommandInput{
		UserID:          1,
		CurrentPassword: "wrong-password",
		NewPassword:     "new-password",
	}

	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockUser, nil)
	s.passwordHasherMock.
		On("Compare", "old-hash", input.CurrentPassword).
		Return(errs.ErrInvalidCredentials)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidCredentials)
}

func (s *ChangePasswordCommandTestSuite) TestExecute_ShortNewPassword_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.ChangePasswordCommandInput{
		UserID:          1,
		CurrentPassword: "old-password",
		NewPassword:     "short",
	}

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrPasswordTooShort)
}
