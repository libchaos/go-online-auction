package mapper_test

import (
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/entity"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/mapper"
	"github.com/stretchr/testify/suite"
)

type AuctionMapperTestSuite struct {
	suite.Suite
	sut *mapper.AuctionMapper
}

func (s *AuctionMapperTestSuite) SetupTest() {
	s.sut = mapper.NewAuctionMapper()
}

func TestAuctionMapperSuite(t *testing.T) {
	suite.Run(t, new(AuctionMapperTestSuite))
}

func (s *AuctionMapperTestSuite) TestToDomain_ValidEntity_ReturnsAuctionModel() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := uint64(5000)
	e := entity.AuctionEntity{
		ID:                      1,
		ListingID:               10,
		StartTime:               &now,
		EndTime:                 now.Add(24 * time.Hour),
		State:                   "active",
		HighestBidAmountInCents: &highestBidAmount,
		Version:                 5,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().NoError(err)
	s.Equal(e.ID, result.ID())
	s.Equal(e.ListingID, result.ListingID())
	s.NotNil(result.StartTime())
	s.Equal(e.StartTime.UTC(), *result.StartTime())
	s.Equal(e.EndTime.UTC(), result.EndTime())
	state := result.State()
	s.Equal(e.State, state.String())
	s.NotNil(result.HighestBidAmount())
	s.Equal(*e.HighestBidAmountInCents, *result.HighestBidAmount())
	s.Equal(e.Version, result.Version())
	s.Equal(e.CreatedAt.UTC(), result.CreatedAt())
	s.Equal(e.UpdatedAt.UTC(), result.UpdatedAt())
}

func (s *AuctionMapperTestSuite) TestToDomain_NilHighestBidID_ReturnsAuctionModelWithNilHighestBidID() {
	// Arrange
	now := time.Now().UTC()
	e := entity.AuctionEntity{
		ID:                      1,
		ListingID:               10,
		StartTime:               &now,
		EndTime:                 now.Add(24 * time.Hour),
		State:                   "draft",
		HighestBidAmountInCents: nil,
		Version:                 0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().NoError(err)
	s.Nil(result.HighestBidAmount())
}

func (s *AuctionMapperTestSuite) TestToDomain_InvalidState_ReturnsError() {
	// Arrange
	now := time.Now().UTC()
	e := entity.AuctionEntity{
		ID:                      1,
		ListingID:               10,
		StartTime:               &now,
		EndTime:                 now.Add(24 * time.Hour),
		State:                   "invalid_state",
		HighestBidAmountInCents: nil,
		Version:                 0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().Error(err)
	s.Equal(model.AuctionModel{}, result)
}

func (s *AuctionMapperTestSuite) TestToEntity_ValidAuctionModel_ReturnsAuctionEntity() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := uint64(5000)
	state, _ := enum.NewAuctionStateEnum("active")
	auction, err := model.RestoreAuctionModel(
		1,
		10,
		&now,
		now.Add(24*time.Hour),
		state,
		&highestBidAmount,
		5,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToEntity(auction)

	// Assert
	s.Equal(auction.ID(), result.ID)
	s.Equal(auction.ListingID(), result.ListingID)
	s.Equal(auction.StartTime(), result.StartTime)
	s.Equal(auction.EndTime(), result.EndTime)
	auctionState := auction.State()
	s.Equal(auctionState.String(), result.State)
	s.NotNil(result.HighestBidAmountInCents)
	s.Equal(*auction.HighestBidAmount(), *result.HighestBidAmountInCents)
	s.Equal(auction.Version(), result.Version)
	s.Equal(auction.CreatedAt(), result.CreatedAt)
	s.Equal(auction.UpdatedAt(), result.UpdatedAt)
}

func (s *AuctionMapperTestSuite) TestToEntity_NilHighestBidID_ReturnsEntityWithNilHighestBidID() {
	// Arrange
	now := time.Now().UTC()
	state, _ := enum.NewAuctionStateEnum("draft")
	auction, err := model.RestoreAuctionModel(
		1,
		10,
		nil,
		now.Add(24*time.Hour),
		state,
		nil,
		0,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToEntity(auction)

	// Assert
	s.Nil(result.HighestBidAmountInCents)
}

func (s *AuctionMapperTestSuite) TestToDomain_WithHighestBidAmount_MapsCorrectly() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := uint64(5000)
	e := entity.AuctionEntity{
		ID:                      1,
		ListingID:               10,
		StartTime:               &now,
		EndTime:                 now.Add(24 * time.Hour),
		State:                   "active",
		HighestBidAmountInCents: &highestBidAmount,
		Version:                 5,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().NoError(err)
	s.NotNil(result.HighestBidAmount())
	s.Equal(highestBidAmount, *result.HighestBidAmount())
}

func (s *AuctionMapperTestSuite) TestToDomain_NilHighestBidAmount_ReturnsNilAmount() {
	// Arrange
	now := time.Now().UTC()
	e := entity.AuctionEntity{
		ID:                      1,
		ListingID:               10,
		StartTime:               &now,
		EndTime:                 now.Add(24 * time.Hour),
		State:                   "draft",
		HighestBidAmountInCents: nil,
		Version:                 0,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().NoError(err)
	s.Nil(result.HighestBidAmount())
}

func (s *AuctionMapperTestSuite) TestToEntity_WithHighestBidAmount_MapsCorrectly() {
	// Arrange
	now := time.Now().UTC()
	highestBidAmount := uint64(5000)
	state, _ := enum.NewAuctionStateEnum("active")
	auction, err := model.RestoreAuctionModel(
		1,
		10,
		&now,
		now.Add(24*time.Hour),
		state,
		&highestBidAmount,
		5,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToEntity(auction)

	// Assert
	s.NotNil(result.HighestBidAmountInCents)
	s.Equal(highestBidAmount, *result.HighestBidAmountInCents)
}
