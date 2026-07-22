package command_test

import (
	"context"
	"testing"
	"time"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/domain/model"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreateWatchCommandTestSuite struct {
	suite.Suite
	sut            *command.CreateWatchCommand
	watchlistsMock *mocks.MockWatchlistRepository
	loggerMock     *mocks.MockLogger
}

func (s *CreateWatchCommandTestSuite) SetupTest() {
	s.watchlistsMock = mocks.NewMockWatchlistRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	nopLogger := zerolog.Nop()
	s.loggerMock.On("Info").Return(nopLogger.Info()).Maybe()
	s.sut = command.NewCreateWatchCommand(s.watchlistsMock, s.loggerMock)
}

func TestCreateWatchCommandSuite(t *testing.T) {
	suite.Run(t, new(CreateWatchCommandTestSuite))
}

func (s *CreateWatchCommandTestSuite) TestExecute_Valid_SavesWatchlist() {
	ctx := context.Background()
	saved := model.ReconstructWatchlist(1, 100, 42, time.Now().UTC())
	s.watchlistsMock.EXPECT().Save(ctx, mock.Anything).Return(saved, nil)

	output, err := s.sut.Execute(ctx, command.CreateWatchCommandInput{UserID: 100, SpuID: 42})

	s.Require().NoError(err)
	s.Equal(uint64(1), output.Watchlist.ID())
}

func (s *CreateWatchCommandTestSuite) TestExecute_ZeroSpu_ReturnsErrorWithoutSave() {
	ctx := context.Background()

	_, err := s.sut.Execute(ctx, command.CreateWatchCommandInput{UserID: 100, SpuID: 0})

	s.Require().Error(err)
	s.watchlistsMock.AssertNotCalled(s.T(), "Save", mock.Anything, mock.Anything)
}
