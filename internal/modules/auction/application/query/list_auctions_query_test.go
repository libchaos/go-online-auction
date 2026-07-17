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

type ListAuctionsQueryTestSuite struct {
	suite.Suite
	sut             *query.ListAuctionsQuery
	auctionRepoMock *mocks.MockAuctionRepository
	loggerMock      *mocks.MockLogger
	ctx             context.Context
}

func (s *ListAuctionsQueryTestSuite) SetupTest() {
	s.auctionRepoMock = mocks.NewMockAuctionRepository(s.T())
	s.loggerMock = mocks.NewMockLogger(s.T())
	s.ctx = context.Background()

	s.sut = query.NewListAuctionsQuery(s.auctionRepoMock, s.loggerMock)
}

func (s *ListAuctionsQueryTestSuite) mockLoggerError() {
	logger := zerolog.New(nil).Level(zerolog.Disabled)
	event := logger.Error()
	s.loggerMock.On("Error").Return(event)
}

func TestListAuctionsQuerySuite(t *testing.T) {
	suite.Run(t, new(ListAuctionsQueryTestSuite))
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithDefaultPagination_ReturnsAuctionsList() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  0,
		Offset: 0,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	startTime := now.Add(-1 * time.Hour)
	highestBid := uint64(10000)

	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction1, _ := model.RestoreAuctionModel(
		1,
		100,
		nil,
		endTime,
		draftState,
		nil,
		1,
		now,
		now,
	)

	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	auction2, _ := model.RestoreAuctionModel(
		2,
		200,
		&startTime,
		endTime,
		activeState,
		&highestBid,
		1,
		now,
		now,
	)

	auctions := []model.AuctionModel{auction1, auction2}

	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 20, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(2), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(2, len(result.Auctions))
	s.Equal(uint64(2), result.TotalCount)
	s.Equal(20, result.Limit)
	s.Equal(0, result.Offset)
	s.Equal(uint64(1), result.Auctions[0].ID)
	s.Equal(uint64(100), result.Auctions[0].ListingID)
	s.Equal(enum.EnumAuctionStateDraft, result.Auctions[0].State)
	s.Nil(result.Auctions[0].HighestBidAmountInCents)
	s.Equal(uint64(2), result.Auctions[1].ID)
	s.Equal(uint64(200), result.Auctions[1].ListingID)
	s.Equal(enum.EnumAuctionStateActive, result.Auctions[1].State)
	s.Equal(&highestBid, result.Auctions[1].HighestBidAmountInCents)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithCustomLimit_ReturnsLimitedResults() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  10,
		Offset: 0,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(1, 100, nil, endTime, draftState, nil, 1, now, now)
	auctions := []model.AuctionModel{auction}

	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 10, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(1), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(10, result.Limit)
	s.Equal(1, len(result.Auctions))
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithLimitExceedingMax_CapsAtMaxLimit() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  150,
		Offset: 0,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(1, 100, nil, endTime, draftState, nil, 1, now, now)
	auctions := []model.AuctionModel{auction}

	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 100, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(1), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(100, result.Limit)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithNegativeOffset_ResetsToZero() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  20,
		Offset: -10,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(1, 100, nil, endTime, draftState, nil, 1, now, now)
	auctions := []model.AuctionModel{auction}

	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 20, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(1), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(0, result.Offset)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithStateFilter_ReturnsFilteredAuctions() {
	// Arrange
	stateStr := enum.EnumAuctionStateActive
	input := query.ListAuctionsQueryInput{
		State:  &stateStr,
		Limit:  20,
		Offset: 0,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	startTime := now.Add(-1 * time.Hour)
	activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	auction, _ := model.RestoreAuctionModel(1, 100, &startTime, endTime, activeState, nil, 1, now, now)
	auctions := []model.AuctionModel{auction}

	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, &activeState, 20, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, &activeState).
		Return(uint64(1), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(1, len(result.Auctions))
	s.Equal(enum.EnumAuctionStateActive, result.Auctions[0].State)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithInvalidState_ReturnsError() {
	// Arrange
	invalidState := "invalid_state"
	input := query.ListAuctionsQueryInput{
		State:  &invalidState,
		Limit:  20,
		Offset: 0,
	}

	s.mockLoggerError()

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().ErrorIs(err, errs.ErrInvalidAuctionState)
	s.Equal(query.ListAuctionsQueryOutput{}, result)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WhenFindAllPaginatedFails_ReturnsError() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  20,
		Offset: 0,
	}

	expectedErr := errors.New("database error")
	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 20, 0).
		Return([]model.AuctionModel(nil), expectedErr)
	s.mockLoggerError()

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.Equal(query.ListAuctionsQueryOutput{}, result)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WhenCountFails_ReturnsError() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  20,
		Offset: 0,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(1, 100, nil, endTime, draftState, nil, 1, now, now)
	auctions := []model.AuctionModel{auction}

	expectedErr := errors.New("count error")
	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 20, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(0), expectedErr)
	s.mockLoggerError()

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().Error(err)
	s.Equal(expectedErr, err)
	s.Equal(query.ListAuctionsQueryOutput{}, result)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithEmptyResults_ReturnsEmptyList() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  20,
		Offset: 0,
	}

	auctions := []model.AuctionModel{}
	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 20, 0).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(0), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(0, len(result.Auctions))
	s.Equal(uint64(0), result.TotalCount)
	s.Equal(20, result.Limit)
	s.Equal(0, result.Offset)
}

func (s *ListAuctionsQueryTestSuite) TestExecute_WithCustomOffset_ReturnsCorrectPage() {
	// Arrange
	input := query.ListAuctionsQueryInput{
		State:  nil,
		Limit:  10,
		Offset: 20,
	}

	now := time.Now().UTC()
	endTime := now.Add(24 * time.Hour)
	draftState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	auction, _ := model.RestoreAuctionModel(21, 121, nil, endTime, draftState, nil, 1, now, now)
	auctions := []model.AuctionModel{auction}

	s.auctionRepoMock.On("FindAllPaginated", mock.Anything, (*enum.AuctionStateEnum)(nil), 10, 20).
		Return(auctions, nil)
	s.auctionRepoMock.On("Count", mock.Anything, (*enum.AuctionStateEnum)(nil)).
		Return(uint64(25), nil)

	// Act
	result, err := s.sut.Execute(s.ctx, input)

	// Assert
	s.Require().NoError(err)
	s.Equal(1, len(result.Auctions))
	s.Equal(uint64(25), result.TotalCount)
	s.Equal(10, result.Limit)
	s.Equal(20, result.Offset)
}
