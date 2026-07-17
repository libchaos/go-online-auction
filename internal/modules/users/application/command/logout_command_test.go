package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/tests/mocks"
)

type LogoutCommandTestSuite struct {
	suite.Suite
	sut                        *command.LogoutCommand
	refreshTokenRepositoryMock *mocks.MockRefreshTokenRepository
	tokenServiceMock           *mocks.MockTokenService
	loggerMock                 *mocks.MockLogger
}

func (s *LogoutCommandTestSuite) SetupTest() {
	s.refreshTokenRepositoryMock = mocks.NewMockRefreshTokenRepository(s.T())
	s.tokenServiceMock = mocks.NewMockTokenService(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewLogoutCommand(
		s.refreshTokenRepositoryMock,
		s.tokenServiceMock,
		s.loggerMock,
	)
}

func TestLogoutCommandSuite(t *testing.T) {
	suite.Run(t, new(LogoutCommandTestSuite))
}

func (s *LogoutCommandTestSuite) TestExecute_ValidToken_RevokesToken() {
	// Arrange
	ctx := context.Background()
	input := command.LogoutCommandInput{RefreshToken: "raw-token"}

	now := time.Now().UTC()
	token, _ := model.RestoreRefreshTokenModel(10, 1, "token-hash", now.Add(time.Hour), nil, nil, now)

	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("token-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "token-hash").
		Return(token, nil)
	s.refreshTokenRepositoryMock.
		On("Update", mock.Anything, mock.AnythingOfType("model.RefreshTokenModel")).
		Return(nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}

func (s *LogoutCommandTestSuite) TestExecute_UnknownToken_IsIdempotent() {
	// Arrange
	ctx := context.Background()
	input := command.LogoutCommandInput{RefreshToken: "raw-unknown"}

	emptyToken := model.RefreshTokenModel{}
	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("unknown-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "unknown-hash").
		Return(emptyToken, errs.ErrRefreshTokenNotFound)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}

func (s *LogoutCommandTestSuite) TestExecute_AlreadyRevokedToken_IsIdempotent() {
	// Arrange
	ctx := context.Background()
	input := command.LogoutCommandInput{RefreshToken: "raw-token"}

	now := time.Now().UTC()
	revokedAt := now.Add(-time.Minute)
	token, _ := model.RestoreRefreshTokenModel(
		10, 1, "token-hash", now.Add(time.Hour), &revokedAt, nil, now,
	)

	s.tokenServiceMock.On("HashRefreshToken", input.RefreshToken).Return("token-hash")
	s.refreshTokenRepositoryMock.
		On("FindByTokenHash", mock.Anything, "token-hash").
		Return(token, nil)

	// Act
	err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
}
