package scheduler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/domain/enum"
	domainerrs "auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/scheduler"
	"auction/tests/mocks"
)

type AuctionSchedulerTestSuite struct {
	suite.Suite
	sut                      *scheduler.AuctionScheduler
	auctionRepositoryMock    *mocks.MockAuctionRepository
	uowFactoryMock           *mocks.MockAuctionUnitOfWorkFactory
	uowMock                  *mocks.MockAuctionUnitOfWork
	uowAuctionRepositoryMock *mocks.MockAuctionRepository
	uowBidRepositoryMock     *mocks.MockBidRepository
	outboxRepositoryMock     *mocks.MockOutboxRepository
	loggerMock               *mocks.MockLogger
}

func (s *AuctionSchedulerTestSuite) SetupTest() {
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.uowAuctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.uowBidRepositoryMock = mocks.NewMockBidRepository(s.T())
	s.outboxRepositoryMock = mocks.NewMockOutboxRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	startCommand := command.NewStartAuctionCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)
	closeCommand := command.NewCloseAuctionCommand(
		s.uowFactoryMock,
		s.loggerMock,
	)

	s.sut = scheduler.NewAuctionScheduler(
		s.auctionRepositoryMock,
		startCommand,
		closeCommand,
		s.loggerMock,
		scheduler.Config{Interval: time.Minute, BatchSize: 10},
	)
}

func TestAuctionSchedulerSuite(t *testing.T) {
	suite.Run(t, new(AuctionSchedulerTestSuite))
}

func (s *AuctionSchedulerTestSuite) draftAuction(id uint64) model.AuctionModel {
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, err := model.RestoreAuctionModel(
		id,
		100,
		nil,
		time.Now().UTC().Add(24*time.Hour),
		draftState,
		nil,
		1,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	s.Require().NoError(err)
	return auction
}

func (s *AuctionSchedulerTestSuite) activeAuction(id uint64) model.AuctionModel {
	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	startTime := time.Now().UTC().Add(-2 * time.Hour)
	auction, err := model.RestoreAuctionModel(
		id,
		100,
		&startTime,
		time.Now().UTC().Add(time.Hour),
		activeState,
		nil,
		2,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	s.Require().NoError(err)
	return auction
}

func (s *AuctionSchedulerTestSuite) TestTick_DueDraftAuction_StartsAuction() {
	// Arrange
	auctionID := uint64(1)
	s.auctionRepositoryMock.On("FindIDsDueToStart", mock.Anything, 10).Return([]uint64{auctionID}, nil)
	s.auctionRepositoryMock.On("FindIDsDueToClose", mock.Anything, 10).Return([]uint64{}, nil)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.uowAuctionRepositoryMock)
	s.uowAuctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).
		Return(s.draftAuction(auctionID), nil)
	s.uowAuctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock)
	s.outboxRepositoryMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.loggerMock.On("Info").Return(nil)

	// Act
	s.sut.Tick(context.Background())

	// Assert
	s.uowAuctionRepositoryMock.AssertCalled(s.T(), "FindByIDForUpdate", mock.Anything, auctionID)
	s.uowMock.AssertCalled(s.T(), "Complete", mock.Anything)
}

func (s *AuctionSchedulerTestSuite) TestTick_ExpiredActiveAuction_ClosesAuction() {
	// Arrange
	auctionID := uint64(2)
	s.auctionRepositoryMock.On("FindIDsDueToStart", mock.Anything, 10).Return([]uint64{}, nil)
	s.auctionRepositoryMock.On("FindIDsDueToClose", mock.Anything, 10).Return([]uint64{auctionID}, nil)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.uowAuctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.uowBidRepositoryMock)
	s.uowAuctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).
		Return(s.activeAuction(auctionID), nil)
	s.uowBidRepositoryMock.On("FindByAuctionID", mock.Anything, auctionID).
		Return([]model.BidModel{}, nil)
	s.uowAuctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("OutboxRepository").Return(s.outboxRepositoryMock)
	s.outboxRepositoryMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.loggerMock.On("Info").Return(nil)

	// Act
	s.sut.Tick(context.Background())

	// Assert
	s.uowMock.AssertCalled(s.T(), "Complete", mock.Anything)
	s.outboxRepositoryMock.AssertCalled(s.T(), "Save", mock.Anything, mock.Anything)
}

func (s *AuctionSchedulerTestSuite) TestTick_ConcurrencyConflict_SkipsSilently() {
	// Arrange: another scheduler instance already holds the row lock
	auctionID := uint64(3)
	s.auctionRepositoryMock.On("FindIDsDueToStart", mock.Anything, 10).Return([]uint64{auctionID}, nil)
	s.auctionRepositoryMock.On("FindIDsDueToClose", mock.Anything, 10).Return([]uint64{}, nil)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.uowAuctionRepositoryMock)
	s.uowAuctionRepositoryMock.On("FindByIDForUpdate", mock.Anything, auctionID).
		Return(model.AuctionModel{}, domainerrs.ErrConcurrencyConflict)
	s.loggerMock.On("Error").Return(nil)

	// Act: must not panic and must not attempt Update/Complete
	s.sut.Tick(context.Background())

	// Assert
	s.uowAuctionRepositoryMock.AssertNotCalled(s.T(), "Update", mock.Anything, mock.Anything)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}

func (s *AuctionSchedulerTestSuite) TestTick_FindDueToStartFails_LogsAndContinuesWithClose() {
	// Arrange
	s.auctionRepositoryMock.On("FindIDsDueToStart", mock.Anything, 10).
		Return(nil, errors.New("db down"))
	s.auctionRepositoryMock.On("FindIDsDueToClose", mock.Anything, 10).Return([]uint64{}, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	s.sut.Tick(context.Background())

	// Assert: close pass still ran despite the start pass failing
	s.auctionRepositoryMock.AssertCalled(s.T(), "FindIDsDueToClose", mock.Anything, 10)
}
