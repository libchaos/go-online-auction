package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/sqlcgen"
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

func (s *BidMapperTestSuite) TestToDomain_ValidRow_ReturnsBidModel() {
	now := time.Now().UTC()
	b := sqlcgen.Bid{
		ID:            1,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := s.sut.ToDomain(b)

	s.Require().NoError(err)
	s.Equal(uint64(b.ID), result.ID())
	s.Equal(uint64(b.AuctionID), result.AuctionID())
	s.Equal(uint64(b.UserID), result.UserID())
	s.Equal(uint64(b.AmountInCents), result.Amount().AmountInCents())
	s.Equal(b.CreatedAt.UTC(), result.CreatedAt())
	s.Equal(b.UpdatedAt.UTC(), result.UpdatedAt())
}

func (s *BidMapperTestSuite) TestToDomain_WithMaxAmount_MapsMaxAmount() {
	now := time.Now().UTC()
	maxAmount := int64(80000)
	b := sqlcgen.Bid{
		ID:               1,
		AuctionID:        10,
		UserID:           20,
		AmountInCents:    50000,
		MaxAmountInCents: &maxAmount,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	result, err := s.sut.ToDomain(b)

	s.Require().NoError(err)
	s.Require().NotNil(result.MaxAmount())
	s.Equal(uint64(maxAmount), result.MaxAmount().AmountInCents())
}

func (s *BidMapperTestSuite) TestToDomain_ZeroID_ReturnsError() {
	now := time.Now().UTC()
	b := sqlcgen.Bid{
		ID:            0,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := s.sut.ToDomain(b)

	s.Require().Error(err)
	s.Equal(model.BidModel{}, result)
}

func (s *BidMapperTestSuite) TestToCreateParams_ValidBidModel_ReturnsParams() {
	now := time.Now().UTC()
	amount := model.NewMoneyModel(50000)

	bid, err := model.RestoreBidModel(
		1,
		10,
		20,
		amount,
		now,
		now,
	)
	s.Require().NoError(err)

	result := s.sut.ToCreateParams(bid, "idem-key")

	s.Equal(int64(bid.AuctionID()), result.AuctionID)
	s.Equal(int64(bid.UserID()), result.UserID)
	s.Equal(int64(bid.Amount().AmountInCents()), result.AmountInCents)
	s.Nil(result.MaxAmountInCents)
	s.Equal("idem-key", result.IdempotencyKey)
	s.Equal(bid.CreatedAt(), result.CreatedAt)
	s.Equal(bid.UpdatedAt(), result.UpdatedAt)
}
