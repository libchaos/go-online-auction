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

type RefreshTokenCommandTestSuite struct {
	suite.Suite
	sut                        *command.RefreshTokenCommand
	userRepositoryMock         *mocks.MockUserRepository
	refreshTokenRepositoryMock *mocks.MockRefreshTokenRepository
	tokenServiceMock           *mocks.MockTokenService
	loggerMock                 *mocks.MockLogger
	mockUser                   model.UserModel
	mockValidToken             model.RefreshTokenModel
	mockRevokedToken           model.RefreshTokenModel
	mockExpiredToken           model.RefreshTokenModel
}

func (s *RefreshTokenCommandTestSuite) SetupTest() {
	s.userRepositoryMock = mocks.NewMockUserRepository(s.T())
	s.refreshTokenRepositoryMock = mocks.NewMockRefreshTokenRepository(s.T())
	s.tokenServiceMock = mocks.NewMockTokenService(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewRefreshTokenCommand(
		s.userRepositoryMock,
		s.refreshTokenRepositoryMock,
		s.tokenServiceMock,
		s.loggerMock,
	)

	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	passwordHash := "hashed-password"
	now := time.Now().UTC()

	s.mockUser, _ = model.RestoreUserModel(
		1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 1, now, now,
	)

	s.mockValidToken, _ = model.RestoreRefreshTokenModel(
		10, 1, "current-hash", now.Add(time.Hour), nil, nil, now,
	)

	revokedAt := now.Add(-time.Minute)
	s.mockRevokedToken, _ = model.RestoreRefreshTokenModel(
		11, 1, "revoked-hash", now.Add(time.Hour), &revokedAt, nil, now,
	)

	s.mockExpiredToken, _ = model.RestoreRefreshTokenModel(
		12, 1, "expired-hash", now.Add(-time.Hour), nil, nil, now,
	)
}

func TestRefreshTokenCommandSuite(t *testing.T) {
	suite.Run(t, new(RefreshTokenCommandTestSuite))
}

func (s *RefreshTokenCommandTestSuite) TestExecute_ValidToken_RotatesAndReturnsNewPair() {
	// Arrange
	ctx := context.Background()
	input := command.RefreshTokenCommandInput{RefreshToken: "raw-current-token"}

	now := time.Now().UTC()
	accessExpiresAt := now.Add(15 * time.Minute)
	refreshExpiresAt := now.Add(168 * time.Hour)
	persistedNewToken, _ := model.RestoreRefreshTokenModel(
		20, 1, "new-hash", refreshExpiresAt, nil, nil, now,
	)

	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("current-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "current-hash").
		Return(s.mockValidToken, nil)
	s.userRepositoryMock.On("FindByID", mock.Anything, uint64(1)).Return(s.mockUser, nil)
	s.tokenServiceMock.
		On("GenerateAccessToken", uint64(1), enum.EnumRoleBidder, "john@example.com").
		Return("new-access-token", accessExpiresAt, nil)
	s.tokenServiceMock.
		On("GenerateRefreshToken").
		Return("raw-new-token", "new-hash", refreshExpiresAt, nil)
	s.refreshTokenRepositoryMock.
		On("Create", mock.Anything, mock.AnythingOfType("model.RefreshTokenModel")).
		Return(persistedNewToken, nil)
	s.refreshTokenRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.RefreshTokenModel")).
		Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal("new-access-token", output.AccessToken)
	s.Equal("raw-new-token", output.RefreshToken)
	s.Equal(uint64(1), output.UserID)
}

func (s *RefreshTokenCommandTestSuite) TestExecute_RevokedToken_RevokesAllSessions() {
	// Arrange
	ctx := context.Background()
	input := command.RefreshTokenCommandInput{RefreshToken: "raw-revoked-token"}

	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("revoked-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "revoked-hash").
		Return(s.mockRevokedToken, nil)
	s.refreshTokenRepositoryMock.On("RevokeAllForUser", mock.Anything, uint64(1)).Return(nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrRefreshTokenInvalid)
	s.Equal(command.RefreshTokenCommandOutput{}, output)
}

func (s *RefreshTokenCommandTestSuite) TestExecute_ExpiredToken_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.RefreshTokenCommandInput{RefreshToken: "raw-expired-token"}

	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("expired-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "expired-hash").
		Return(s.mockExpiredToken, nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrRefreshTokenInvalid)
	s.Equal(command.RefreshTokenCommandOutput{}, output)
}

func (s *RefreshTokenCommandTestSuite) TestExecute_UnknownToken_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.RefreshTokenCommandInput{RefreshToken: "raw-unknown-token"}

	emptyToken := model.RefreshTokenModel{}
	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("unknown-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "unknown-hash").
		Return(emptyToken, errs.ErrRefreshTokenNotFound)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrRefreshTokenInvalid)
	s.Equal(command.RefreshTokenCommandOutput{}, output)
}
