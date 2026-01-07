package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/application/command"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/tests/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PlaceBidCommandTestSuite struct {
	suite.Suite
	sut                          *command.PlaceBidCommand
	uowFactoryMock               *mocks.MockAuctionUnitOfWorkFactory
	uowMock                      *mocks.MockAuctionUnitOfWork
	auctionRepositoryMock        *mocks.MockAuctionRepository
	bidRepositoryMock            *mocks.MockBidRepository
	bidPlacedEventDispatcherMock *mocks.MockBidPlacedEventDispatcher
	loggerMock                   *mocks.MockLogger
}

func (s *PlaceBidCommandTestSuite) SetupTest() {
	s.uowFactoryMock = mocks.NewMockAuctionUnitOfWorkFactory(s.T())
	s.uowMock = mocks.NewMockAuctionUnitOfWork(s.T())
	s.auctionRepositoryMock = mocks.NewMockAuctionRepository(s.T())
	s.bidRepositoryMock = mocks.NewMockBidRepository(s.T())
	s.bidPlacedEventDispatcherMock = mocks.NewMockBidPlacedEventDispatcher(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())

	s.sut = command.NewPlaceBidCommand(
		s.uowFactoryMock,
		s.bidPlacedEventDispatcherMock,
		s.loggerMock,
	)
}

func TestPlaceBidCommandSuite(t *testing.T) {
	suite.Run(t, new(PlaceBidCommandTestSuite))
}

func (s *PlaceBidCommandTestSuite) TestExecute_SuccessfulBidPlacement_ReturnsOutput() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	userID := uint64(100)
	amountInCents := uint64(5000)
	bidID := uint64(10)

	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: amountInCents,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)
	_ = auction.Start()

	money := model.NewMoneyModel(amountInCents)
	persistedBid, _ := model.RestoreBidModel(
		bidID,
		auctionID,
		userID,
		money,
		time.Now(),
		time.Now(),
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel")).
		Return(persistedBid, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.bidPlacedEventDispatcherMock.On("Dispatch", mock.Anything, mock.Anything).Return(nil)

	// Act
	result, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(bidID, result.ID)
	s.Equal(auctionID, result.AuctionID)
	s.Equal(userID, result.UserID)
	s.Equal(amountInCents, result.AmountInCents)
}

func (s *PlaceBidCommandTestSuite) TestExecute_BeginUnitOfWorkFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	input := command.PlaceBidCommandInput{
		AuctionID:     1,
		UserID:        100,
		AmountInCents: 5000,
	}

	expectedErr := errors.New("begin uow error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(nil, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PlaceBidCommandTestSuite) TestExecute_FindAuctionFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        100,
		AmountInCents: 5000,
	}

	expectedErr := errors.New("auction not found")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(model.AuctionModel{}, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PlaceBidCommandTestSuite) TestExecute_PlaceBidOnAuctionFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	userID := uint64(100)
	amountInCents := uint64(5000)

	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: amountInCents,
	}

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		draftState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)

	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().Error(err)
}

func (s *PlaceBidCommandTestSuite) TestExecute_CreateBidFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	userID := uint64(100)
	amountInCents := uint64(5000)

	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: amountInCents,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)
	_ = auction.Start()

	expectedErr := errors.New("create bid error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel")).
		Return(model.BidModel{}, expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PlaceBidCommandTestSuite) TestExecute_UpdateAuctionFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	userID := uint64(100)
	amountInCents := uint64(5000)
	bidID := uint64(10)

	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: amountInCents,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)
	_ = auction.Start()

	money := model.NewMoneyModel(amountInCents)
	persistedBid, _ := model.RestoreBidModel(
		bidID,
		auctionID,
		userID,
		money,
		time.Now(),
		time.Now(),
	)

	expectedErr := errors.New("update auction error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel")).
		Return(persistedBid, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PlaceBidCommandTestSuite) TestExecute_CompleteUnitOfWorkFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	userID := uint64(100)
	amountInCents := uint64(5000)
	bidID := uint64(10)

	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: amountInCents,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)
	_ = auction.Start()

	money := model.NewMoneyModel(amountInCents)
	persistedBid, _ := model.RestoreBidModel(
		bidID,
		auctionID,
		userID,
		money,
		time.Now(),
		time.Now(),
	)

	expectedErr := errors.New("complete uow error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel")).
		Return(persistedBid, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}

func (s *PlaceBidCommandTestSuite) TestExecute_DispatchEventFails_ReturnsError() {
	// Arrange
	ctx := context.Background()
	auctionID := uint64(1)
	userID := uint64(100)
	amountInCents := uint64(5000)
	bidID := uint64(10)

	input := command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: amountInCents,
	}

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	futureTime := time.Now().Add(24 * time.Hour)
	auction, _ := model.RestoreAuctionModel(
		auctionID,
		1,
		nil,
		futureTime,
		activeState,
		nil,
		1,
		time.Now(),
		time.Now(),
	)
	_ = auction.Start()

	money := model.NewMoneyModel(amountInCents)
	persistedBid, _ := model.RestoreBidModel(
		bidID,
		auctionID,
		userID,
		money,
		time.Now(),
		time.Now(),
	)

	expectedErr := errors.New("dispatch event error")
	s.uowFactoryMock.On("Begin", mock.Anything).Return(s.uowMock, nil)
	s.uowMock.On("Rollback", mock.Anything).Return(nil)
	s.uowMock.On("AuctionRepository").Return(s.auctionRepositoryMock)
	s.uowMock.On("BidRepository").Return(s.bidRepositoryMock)
	s.auctionRepositoryMock.On("FindByID", mock.Anything, auctionID).Return(auction, nil)
	s.bidRepositoryMock.On("Create", mock.Anything, mock.AnythingOfType("model.BidModel")).
		Return(persistedBid, nil)
	s.auctionRepositoryMock.On("Update", mock.Anything, mock.AnythingOfType("model.AuctionModel")).
		Return(nil)
	s.uowMock.On("Complete", mock.Anything).Return(nil)
	s.bidPlacedEventDispatcherMock.On("Dispatch", mock.Anything, mock.Anything).Return(expectedErr)
	s.loggerMock.On("Error").Return(nil)

	// Act
	_, err := s.sut.Execute(ctx, input)

	// Assert
	s.Require().ErrorIs(err, expectedErr)
}
