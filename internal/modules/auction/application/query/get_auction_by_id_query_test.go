package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auction/internal/modules/auction/application/query"
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/tests/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GetAuctionByIDQueryTestSuite struct {
	suite.Suite
	sut             *query.GetAuctionByIDQuery
	auctionRepoMock *mocks.MockAuctionRepository
	bidRepoMock     *mocks.MockBidRepository
	loggerMock      *mocks.MockLogger
	ctx             context.Context
}

func (s *GetAuctionByIDQueryTestSuite) SetupTest() {
	s.auctionRepoMock = mocks.NewMockAuctionRepository(s.T())
	s.bidRepoMock = mocks.NewMockBidRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.ctx = context.Background()

	s.sut = query.NewGetAuctionByIDQuery(s.auctionRepoMock, s.bidRepoMock, s.loggerMock)
}

func (s *GetAuctionByIDQueryTestSuite) mockLoggerError() {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	event := logger.Error()
	s.loggerMock.On("Error").Return(event)
}

func TestGetAuctionByIDQuerySuite(t *testing.T) {
	suite.Run(t, new(GetAuctionByIDQueryTestSuite))
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WithValidID_ReturnsAuctionAndBids() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	startTime := now.Add(-1 * time.Hour)
	highestBid := uint64(10000)

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	auction, _ := model.RestoreAuctionModel(
		1,
		100,
		&startTime,
		endTime,
		activeState,
		&highestBid,
		1,
		now,
		now,
	)

	amount1 := model.NewMoneyModel(10000)
	amount2 := model.NewMoneyModel(9000)
	bid1, _ := model.RestoreBidModel(1, 1, 100, amount1, now, now)
	bid2, _ := model.RestoreBidModel(2, 1, 101, amount2, now, now)
	bids := []model.BidModel{bid1, bid2}

	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).Return(auction, nil)
	s.bidRepoMock.On("FindTopBidsByAuctionID", mock.Anything, uint64(1), 10).Return(bids, nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), result.Auction.ID)
	s.Equal(uint64(100), result.Auction.ListingID)
	s.Equal(enum.EnumAuctionStateActive, result.Auction.State)
	s.Equal(&startTime, result.Auction.StartTime)
	s.Equal(endTime, result.Auction.EndTime)
	s.Equal(&highestBid, result.Auction.HighestBidAmountInCents)
	s.Equal(2, len(result.Bids))
	s.Equal(uint64(1), result.Bids[0].ID)
	s.Equal(uint64(100), result.Bids[0].UserID)
	s.Equal(uint64(10000), result.Bids[0].AmountInCents)
	s.Equal(uint64(2), result.Bids[1].ID)
	s.Equal(uint64(101), result.Bids[1].UserID)
	s.Equal(uint64(9000), result.Bids[1].AmountInCents)
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WithValidIDAndNoBids_ReturnsAuctionWithEmptyBids() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(1, 100, nil, endTime, draftState, nil, 1, now, now)

	bids := []model.BidModel{}

	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).Return(auction, nil)
	s.bidRepoMock.On("FindTopBidsByAuctionID", mock.Anything, uint64(1), 10).Return(bids, nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), result.Auction.ID)
	s.Equal(uint64(100), result.Auction.ListingID)
	s.Equal(enum.EnumAuctionStateDraft, result.Auction.State)
	s.Nil(result.Auction.StartTime)
	s.Nil(result.Auction.HighestBidAmountInCents)
	s.Equal(0, len(result.Bids))
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WhenAuctionNotFound_ReturnsError() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 999,
	}

	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(999)).
		Return(model.AuctionModel{}, errs.ErrAuctionNotFound)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrAuctionNotFound)
	s.Equal(query.GetAuctionByIDQueryOutput{}, result)
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WhenFindByIDFails_ReturnsError() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	expectedErr := errors.New("database error")
	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).
		Return(model.AuctionModel{}, expectedErr)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.Equal(query.GetAuctionByIDQueryOutput{}, result)
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WhenFindTopBidsFails_ReturnsError() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(1, 100, nil, endTime, draftState, nil, 1, now, now)

	expectedErr := errors.New("database error")
	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).Return(auction, nil)
	s.bidRepoMock.On("FindTopBidsByAuctionID", mock.Anything, uint64(1), 10).
		Return([]model.BidModel(nil), expectedErr)
	s.mockLoggerError()

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.Equal(query.GetAuctionByIDQueryOutput{}, result)
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WithClosedAuction_ReturnsAuctionAndBids() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	now := time.Now().UTC()
	endTime := now.Add(-1 * time.Hour)
	startTime := now.Add(-25 * time.Hour)
	highestBid := uint64(15000)

	closedState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateClosed)
	auction, _ := model.RestoreAuctionModel(
		1,
		100,
		&startTime,
		endTime,
		closedState,
		&highestBid,
		1,
		now.Add(-25*time.Hour),
		now,
	)

	amount := model.NewMoneyModel(15000)
	bid, _ := model.RestoreBidModel(1, 1, 100, amount, now, now)
	bids := []model.BidModel{bid}

	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).Return(auction, nil)
	s.bidRepoMock.On("FindTopBidsByAuctionID", mock.Anything, uint64(1), 10).Return(bids, nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), result.Auction.ID)
	s.Equal(enum.EnumAuctionStateClosed, result.Auction.State)
	s.Equal(&highestBid, result.Auction.HighestBidAmountInCents)
	s.Equal(1, len(result.Bids))
	s.Equal(uint64(15000), result.Bids[0].AmountInCents)
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WithCancelledAuction_ReturnsAuctionAndBids() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	startTime := now.Add(-1 * time.Hour)
	highestBid := uint64(5000)

	cancelledState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateCancelled)
	auction, _ := model.RestoreAuctionModel(
		1,
		100,
		&startTime,
		endTime,
		cancelledState,
		&highestBid,
		1,
		now.Add(-2*time.Hour),
		now,
	)

	amount := model.NewMoneyModel(5000)
	bid, _ := model.RestoreBidModel(1, 1, 100, amount, now, now)
	bids := []model.BidModel{bid}

	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).Return(auction, nil)
	s.bidRepoMock.On("FindTopBidsByAuctionID", mock.Anything, uint64(1), 10).Return(bids, nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(1), result.Auction.ID)
	s.Equal(enum.EnumAuctionStateCancelled, result.Auction.State)
	s.Equal(&highestBid, result.Auction.HighestBidAmountInCents)
	s.Equal(1, len(result.Bids))
}

func (s *GetAuctionByIDQueryTestSuite) TestExecute_WithMultipleBids_ReturnsAllBids() {
	// Arrange
	input := query.GetAuctionByIDQueryInput{
		ID: 1,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	startTime := now.Add(-1 * time.Hour)
	highestBid := uint64(10000)

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	auction, _ := model.RestoreAuctionModel(
		1,
		100,
		&startTime,
		endTime,
		activeState,
		&highestBid,
		1,
		now,
		now,
	)

	bids := make([]model.BidModel, 0, 5)
	for i := 1; i <= 5; i++ {
		amount := model.NewMoneyModel(uint64(5000 + i*1000))
		bid, _ := model.RestoreBidModel(uint64(i), 1, uint64(100+i), amount, now, now)
		bids = append(bids, bid)
	}

	s.auctionRepoMock.On("FindByID", mock.Anything, uint64(1)).Return(auction, nil)
	s.bidRepoMock.On("FindTopBidsByAuctionID", mock.Anything, uint64(1), 10).Return(bids, nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(5, len(result.Bids))
	for i := range 5 {
		s.Equal(uint64(i+1), result.Bids[i].ID)
		s.Equal(uint64(100+i+1), result.Bids[i].UserID)
		s.Equal(uint64(5000+(i+1)*1000), result.Bids[i].AmountInCents)
	}
}
