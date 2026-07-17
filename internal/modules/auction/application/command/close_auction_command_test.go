package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/model"
	"auction/pkg/logger"
	"auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CloseAuctionCommandTestSuite struct {
	suite.Suite
	sut             *command.CloseAuctionCommand
	uowFactoryMock  *mocks.MockAuctionUnitOfWorkFactory
	uowMock         *mocks.MockAuctionUnitOfWork
	auctionRepoMock *mocks.MockAuctionRepository
	outboxRepoMock  *mocks.MockOutboxRepository
	bidRepoMock     *mocks.MockBidRepository
}

func (s *CloseAuctionCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepoMock = mocks.NewMockAuctionRepository(s.T())
	s.outboxRepoMock = mocks.NewMockOutboxRepository(s.T())
	s.bidRepoMock = mocks.NewMockBidRepository(s.T())
	s.uowMock.On("BidRepository").Return(s.bidRepoMock).Maybe()
	s.uowMock.On("OutboxRepository").Return(s.outboxRepoMock).Maybe()
	s.bidRepoMock.On("FindByAuctionID", mock.Anything, mock.Anything).Return([]model.BidModel{}, nil).Maybe()

	log := logger.New(logger.Config{LogLevel: logger.MustLogLevel(logger.LogLevelFatal)})

	s.sut = command.NewCloseAuctionCommand(
		s.uowFactoryMock,
		log,
	)
}

func TestCloseAuctionCommandSuite(t *testing.T) {
	suite.Run(t, new(CloseAuctionCommandTestSuite))
}

func (s *CloseAuctionCommandTestSuite) TestExecute_SuccessfullyClosesAuctionWithHighestBid() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)
	highestBidAmount := uint64(5000)
	now := time.Now()

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	startTime := now
	auctionMock, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		&startTime,
		now.Add(24*time.Hour),
		activeState,
		&highestBidAmount,
		1,
		now,
		now,
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auctionMock, nil)
	s.auctionRepoMock.On("Update", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.AnythingOfType("ports.OutboxEvent")).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(auctionID, output.ID)
	s.Equal(listingID, output.ListingID)
	s.Equal("closed", output.State)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_SuccessfullyClosesAuctionWithoutHighestBid() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)
	now := time.Now()

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	startTime := now
	auctionMock, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		&startTime,
		now.Add(24*time.Hour),
		activeState,
		nil,
		1,
		now,
		now,
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auctionMock, nil)
	s.auctionRepoMock.On("Update", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.AnythingOfType("ports.OutboxEvent")).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(auctionID, output.ID)
	s.Equal(listingID, output.ListingID)
	s.Equal("closed", output.State)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_FailsWhenBeginUnitOfWorkFails() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	expectedErr := errors.New("failed to begin transaction")

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(nil, expectedErr)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
	s.Equal(command.CloseAuctionCommandOutput{}, output)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_FailsWhenFindAuctionFails() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	expectedErr := errors.New("auction not found")

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(model.AuctionModel{}, expectedErr)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
	s.Equal(command.CloseAuctionCommandOutput{}, output)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_FailsWhenCloseAuctionFails() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)
	now := time.Now()

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auctionMock, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		nil,
		now.Add(24*time.Hour),
		draftState,
		nil,
		1,
		now,
		now,
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auctionMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(command.CloseAuctionCommandOutput{}, output)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_FailsWhenUpdateAuctionFails() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)
	now := time.Now()
	expectedErr := errors.New("failed to update")

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	startTime := now
	auctionMock, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		&startTime,
		now.Add(24*time.Hour),
		activeState,
		nil,
		1,
		now,
		now,
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auctionMock, nil)
	s.auctionRepoMock.On("Update", mock.Anything, mock.Anything).Return(expectedErr)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
	s.Equal(command.CloseAuctionCommandOutput{}, output)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_FailsWhenCompleteUnitOfWorkFails() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)
	now := time.Now()
	expectedErr := errors.New("failed to commit")

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	startTime := now
	auctionMock, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		&startTime,
		now.Add(24*time.Hour),
		activeState,
		nil,
		1,
		now,
		now,
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auctionMock, nil)
	s.auctionRepoMock.On("Update", mock.Anything, mock.Anything).Return(nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.AnythingOfType("ports.OutboxEvent")).Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(expectedErr)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
	s.Equal(command.CloseAuctionCommandOutput{}, output)
}

func (s *CloseAuctionCommandTestSuite) TestExecute_FailsWhenSaveOutboxFails() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	listingID := uint64(100)
	now := time.Now()
	expectedErr := errors.New("failed to save outbox event")

	input := command.CloseAuctionCommandInput{
		AuctionID: auctionID,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	startTime := now
	auctionMock, _ := model.RestoreAuctionModel(
		auctionID,
		listingID,
		&startTime,
		now.Add(24*time.Hour),
		activeState,
		nil,
		1,
		now,
		now,
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepoMock)
	s.auctionRepoMock.On("FindByIDForUpdate", mock.Anything, auctionID).Return(auctionMock, nil)
	s.auctionRepoMock.On("Update", mock.Anything, mock.Anything).Return(nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.outboxRepoMock.On("Save", mock.Anything, mock.AnythingOfType("ports.OutboxEvent")).
		Return(expectedErr)

	// Act
	output, err := s.sut.Execute(ctx, input)

	// Assert: transaction is rolled back, the event is never delivered
	s.Require().ErrorIs(err, expectedErr)
	s.Equal(command.CloseAuctionCommandOutput{}, output)
	s.uowMock.AssertNotCalled(s.T(), "Complete", mock.Anything)
}
