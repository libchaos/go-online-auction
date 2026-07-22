package command_test

import (
	"context"
	"testing"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/domain/errs"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type DeleteWatchCommandTestSuite struct {
	suite.Suite
	sut            *command.DeleteWatchCommand
	watchlistsMock *mocks.MockWatchlistRepository
	loggerMock     *mocks.MockLogger
}

func (s *DeleteWatchCommandTestSuite) SetupTest() {
	s.watchlistsMock = mocks.NewMockWatchlistRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.sut = command.NewDeleteWatchCommand(s.watchlistsMock, s.loggerMock)
}

func TestDeleteWatchCommandSuite(t *testing.T) {
	suite.Run(t, new(DeleteWatchCommandTestSuite))
}

func (s *DeleteWatchCommandTestSuite) TestExecute_Existing_RemovesWatchlist() {
	ctx := context.Background()
	s.watchlistsMock.EXPECT().Remove(ctx, uint64(100), uint64(42)).Return(nil)

	err := s.sut.Execute(ctx, command.DeleteWatchCommandInput{UserID: 100, SpuID: 42})

	s.Require().NoError(err)
}

func (s *DeleteWatchCommandTestSuite) TestExecute_NotWatched_ReturnsDomainError() {
	ctx := context.Background()
	s.watchlistsMock.EXPECT().Remove(ctx, uint64(100), uint64(42)).Return(errs.ErrWatchNotFound)

	err := s.sut.Execute(ctx, command.DeleteWatchCommandInput{UserID: 100, SpuID: 42})

	s.Require().ErrorIs(err, errs.ErrWatchNotFound)
}
