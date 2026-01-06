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

type BidMapperTestSuite struct {
	suite.Suite
	sut *mapper.BidMapper
}

func (s *BidMapperTestSuite) SetupTest() {
	s.sut = mapper.NewBidMapper()
}

func TestBidMapperSuite(t *testing.T) {
	suite.Run(t, new(BidMapperTestSuite))
}

func (s *BidMapperTestSuite) TestToDomain_ValidEntity_ReturnsBidModel() {
	// Arrange
	now := time.Now().UTC()
	e := entity.BidEntity{
		ID:            1,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		Currency:      "USD",
		Status:        "accepted",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().NoError(err)
	s.Equal(e.ID, result.ID())
	s.Equal(e.AuctionID, result.AuctionID())
	s.Equal(e.UserID, result.UserID())
	s.Equal(e.AmountInCents, result.Amount().AmountInCents())
	s.Equal(e.Currency, result.Amount().Currency())
	status := result.Status()
	s.Equal(e.Status, status.String())
	s.Equal(e.CreatedAt.UTC(), result.CreatedAt())
	s.Equal(e.UpdatedAt.UTC(), result.UpdatedAt())
}

func (s *BidMapperTestSuite) TestToDomain_InvalidCurrency_ReturnsError() {
	// Arrange
	now := time.Now().UTC()
	e := entity.BidEntity{
		ID:            1,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		Currency:      "INVALID",
		Status:        "accepted",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().Error(err)
	s.Equal(model.BidModel{}, result)
}

func (s *BidMapperTestSuite) TestToDomain_InvalidStatus_ReturnsError() {
	// Arrange
	now := time.Now().UTC()
	e := entity.BidEntity{
		ID:            1,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		Currency:      "USD",
		Status:        "invalid_status",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Act
	result, err := s.sut.ToDomain(e)

	// Assert
	s.Require().Error(err)
	s.Equal(model.BidModel{}, result)
}

func (s *BidMapperTestSuite) TestToEntity_ValidBidModel_ReturnsBidEntity() {
	// Arrange
	now := time.Now().UTC()
	amount, err := model.NewMoneyModel(50000, "USD")
	s.Require().NoError(err)

	status, err := enum.NewBidStatusEnum("accepted")
	s.Require().NoError(err)

	bid, err := model.RestoreBidModel(
		1,
		10,
		20,
		amount,
		status,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToEntity(bid)

	// Assert
	s.Equal(bid.ID(), result.ID)
	s.Equal(bid.AuctionID(), result.AuctionID)
	s.Equal(bid.UserID(), result.UserID)
	s.Equal(bid.Amount().AmountInCents(), result.AmountInCents)
	s.Equal(bid.Amount().Currency(), result.Currency)
	bidStatus := bid.Status()
	s.Equal(bidStatus.String(), result.Status)
	s.Equal(bid.CreatedAt(), result.CreatedAt)
	s.Equal(bid.UpdatedAt(), result.UpdatedAt)
}

func (s *BidMapperTestSuite) TestToEntity_SupersededStatus_ReturnsEntityWithSupersededStatus() {
	// Arrange
	now := time.Now().UTC()
	amount, err := model.NewMoneyModel(50000, "EUR")
	s.Require().NoError(err)

	status, err := enum.NewBidStatusEnum("superseded")
	s.Require().NoError(err)

	bid, err := model.RestoreBidModel(
		1,
		10,
		20,
		amount,
		status,
		now,
		now,
	)
	s.Require().NoError(err)

	// Act
	result := s.sut.ToEntity(bid)

	// Assert
	s.Equal("superseded", result.Status)
	s.Equal("EUR", result.Currency)
}


