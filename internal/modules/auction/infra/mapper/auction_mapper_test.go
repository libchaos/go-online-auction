package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/sqlcgen"
)

type AuctionMapperTestSuite struct {
	suite.Suite
	sut *mapper.AuctionMapper
}

func (s *AuctionMapperTestSuite) SetupTest() {
	s.sut = mapper.NewAuctionMapper(strategy.NewDefaultResolver())
}

func TestAuctionMapperSuite(t *testing.T) {
	suite.Run(t, new(AuctionMapperTestSuite))
}

func (s *AuctionMapperTestSuite) TestToDomain_ValidRow_ReturnsAuctionModel() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := int64(5000)
	a := sqlcgen.Auction{
		ID:                      1,
		ListingID:               10,
		TradingMode:             "english",
		StartTime:               &now,
		EndTime:                 now.Add(24 * time.Hour),
		State:                   sqlcgen.AuctionStateActive,
		HighestBidAmountInCents: &highestBidAmount,
		ExtensionWindowSec:      300,
		Version:                 5,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Act
	result, err := s.sut.ToDomain(a)

	// Assert
	s.Require().NoError(err)
	s.Equal(uint64(a.ID), result.ID())
	s.Equal(uint64(a.ListingID), result.ListingID())
	tradingMode := result.TradingMode()
	s.Equal("english", tradingMode.String())
	s.NotNil(result.StartTime())
	s.Equal(a.StartTime.UTC(), *result.StartTime())
	s.Equal(a.EndTime.UTC(), result.EndTime())
	state := result.State()
	s.Equal(string(a.State), state.String())
	s.NotNil(result.HighestBidAmount())
	s.Equal(uint64(highestBidAmount), *result.HighestBidAmount())
	s.Equal(uint64(a.Version), result.Version())
	s.Equal(a.CreatedAt.UTC(), result.CreatedAt())
	s.Equal(a.UpdatedAt.UTC(), result.UpdatedAt())
}

func (s *AuctionMapperTestSuite) TestToDomain_NilHighestBidAmount_ReturnsNilAmount() {
	// Arrange
	now := time.Now().UTC()
	a := sqlcgen.Auction{
		ID:                 1,
		ListingID:          10,
		TradingMode:        "english",
		StartTime:          &now,
		EndTime:            now.Add(24 * time.Hour),
		State:              sqlcgen.AuctionStateDraft,
		ExtensionWindowSec: 300,
		Version:            0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Act
	result, err := s.sut.ToDomain(a)

	// Assert
	s.Require().NoError(err)
	s.Nil(result.HighestBidAmount())
}

func (s *AuctionMapperTestSuite) TestToDomain_InvalidState_ReturnsError() {
	// Arrange
	now := time.Now().UTC()
	a := sqlcgen.Auction{
		ID:                 1,
		ListingID:          10,
		TradingMode:        "english",
		StartTime:          &now,
		EndTime:            now.Add(24 * time.Hour),
		State:              "invalid_state",
		ExtensionWindowSec: 300,
		Version:            0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Act
	result, err := s.sut.ToDomain(a)

	// Assert
	s.Require().Error(err)
	s.Equal(model.AuctionModel{}, result)
}

func (s *AuctionMapperTestSuite) TestToCreateParams_ValidAuctionModel_ReturnsParams() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := uint64(5000)
	auction := s.restoreAuction(now, &highestBidAmount, "active", 5)

	// Act
	result := s.sut.ToCreateParams(auction)

	// Assert
	s.Equal(int64(auction.ListingID()), result.ListingID)
	tradingMode := auction.TradingMode()
	s.Equal(tradingMode.String(), result.TradingMode)
	s.Equal(auction.EndTime(), result.EndTime)
	s.Equal(sqlcgen.AuctionStateActive, result.State)
	s.NotNil(result.HighestBidAmountInCents)
	s.Equal(int64(highestBidAmount), *result.HighestBidAmountInCents)
	s.Equal(int64(auction.Version()), result.Version)
	s.Equal(auction.CreatedAt(), result.CreatedAt)
	s.Equal(auction.UpdatedAt(), result.UpdatedAt)
}

func (s *AuctionMapperTestSuite) TestToCreateParams_NilHighestBidAmount_ReturnsNilAmount() {
	// Arrange
	now := time.Now().UTC()
	auction := s.restoreAuction(now, nil, "draft", 0)

	// Act
	result := s.sut.ToCreateParams(auction)

	// Assert
	s.Nil(result.HighestBidAmountInCents)
}

func (s *AuctionMapperTestSuite) TestToUpdateParams_ValidAuctionModel_ReturnsParams() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := uint64(5000)
	auction := s.restoreAuction(now, &highestBidAmount, "active", 5)

	// Act
	result := s.sut.ToUpdateParams(auction)

	// Assert
	s.Equal(int64(auction.ID()), result.ID)
	s.Equal(int64(auction.ListingID()), result.ListingID)
	s.Equal(auction.StartTime(), result.StartTime)
	s.Equal(auction.EndTime(), result.EndTime)
	s.Equal(sqlcgen.AuctionStateActive, result.State)
	s.NotNil(result.HighestBidAmountInCents)
	s.Equal(int64(highestBidAmount), *result.HighestBidAmountInCents)
	s.Equal(int64(auction.Version()), result.Version)
	s.Equal(auction.UpdatedAt(), result.UpdatedAt)
}

func (s *AuctionMapperTestSuite) restoreAuction(
	now time.Time,
	highestBidAmount *uint64,
	state string,
	version uint64,
) model.AuctionModel {
	stateEnum, err := enum.NewAuctionStateEnum(state)
	s.Require().NoError(err)

	auction, err := model.RestoreAuctionModel(
		1,
		10,
		&now,
		now.Add(24*time.Hour),
		stateEnum,
		highestBidAmount,
		version,
		now,
		now,
	)
	s.Require().NoError(err)
	return auction
}
