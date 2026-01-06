package mapper_test

import (
	"testing"
	"time"

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
	now := time.Now().UTC()
	e := entity.BidEntity{
		ID:            1,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := s.sut.ToDomain(e)

	s.Require().NoError(err)
	s.Equal(e.ID, result.ID())
	s.Equal(e.AuctionID, result.AuctionID())
	s.Equal(e.UserID, result.UserID())
	s.Equal(e.AmountInCents, result.Amount().AmountInCents())
	s.Equal(e.CreatedAt.UTC(), result.CreatedAt())
	s.Equal(e.UpdatedAt.UTC(), result.UpdatedAt())
}

func (s *BidMapperTestSuite) TestToDomain_ZeroID_ReturnsError() {
	now := time.Now().UTC()
	e := entity.BidEntity{
		ID:            0,
		AuctionID:     10,
		UserID:        20,
		AmountInCents: 50000,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result, err := s.sut.ToDomain(e)

	s.Require().Error(err)
	s.Equal(model.BidModel{}, result)
}

func (s *BidMapperTestSuite) TestToEntity_ValidBidModel_ReturnsBidEntity() {
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

	result := s.sut.ToEntity(bid)

	s.Equal(bid.ID(), result.ID)
	s.Equal(bid.AuctionID(), result.AuctionID)
	s.Equal(bid.UserID(), result.UserID)
	s.Equal(bid.Amount().AmountInCents(), result.AmountInCents)
	s.Equal(bid.CreatedAt(), result.CreatedAt)
	s.Equal(bid.UpdatedAt(), result.UpdatedAt)
}
