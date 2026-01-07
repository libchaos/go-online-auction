package model_test

import (
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/stretchr/testify/suite"
)

type BidModelTestSuite struct {
	suite.Suite
}

func TestBidModelSuite(t *testing.T) {
	suite.Run(t, new(BidModelTestSuite))
}

func (s *BidModelTestSuite) TestNewBidModel_ValidInput_ReturnsBidModel() {
	// Arrange
	auctionID := uint64(1)
	userID := uint64(2)
	amount := model.NewMoneyModel(10000)

	// Act
	result, err := model.NewBidModel(auctionID, userID, amount)

	// Assert
	s.Require().NoError(err)
	s.Equal(auctionID, result.AuctionID())
	s.Equal(userID, result.UserID())
	s.Equal(amount, result.Amount())
	s.NotZero(result.CreatedAt())
	s.NotZero(result.UpdatedAt())
	s.Equal(result.CreatedAt(), result.UpdatedAt())
}

func (s *BidModelTestSuite) TestNewBidModel_ZeroAuctionID_ReturnsError() {
	// Arrange
	auctionID := uint64(0)
	userID := uint64(2)
	amount := model.NewMoneyModel(10000)

	// Act
	_, err := model.NewBidModel(auctionID, userID, amount)

	// Assert
	s.Require().ErrorIs(err, errs.ErrAuctionIDRequired)
}

func (s *BidModelTestSuite) TestNewBidModel_ZeroUserID_ReturnsError() {
	// Arrange
	auctionID := uint64(1)
	userID := uint64(0)
	amount := model.NewMoneyModel(10000)

	// Act
	_, err := model.NewBidModel(auctionID, userID, amount)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserIDRequired)
}

func (s *BidModelTestSuite) TestRestoreBidModel_ValidInput_ReturnsBidModel() {
	// Arrange
	id := uint64(100)
	auctionID := uint64(1)
	userID := uint64(2)
	amount := model.NewMoneyModel(10000)
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

	// Act
	result, err := model.RestoreBidModel(id, auctionID, userID, amount, createdAt, updatedAt)

	// Assert
	s.Require().NoError(err)
	s.Equal(id, result.ID())
	s.Equal(auctionID, result.AuctionID())
	s.Equal(userID, result.UserID())
	s.Equal(amount, result.Amount())
	s.Equal(createdAt, result.CreatedAt())
	s.Equal(updatedAt, result.UpdatedAt())
}

func (s *BidModelTestSuite) TestRestoreBidModel_ZeroID_ReturnsError() {
	// Arrange
	id := uint64(0)
	auctionID := uint64(1)
	userID := uint64(2)
	amount := model.NewMoneyModel(10000)
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()

	// Act
	_, err := model.RestoreBidModel(id, auctionID, userID, amount, createdAt, updatedAt)

	// Assert
	s.Require().ErrorIs(err, errs.ErrBidIDRequired)
}

func (s *BidModelTestSuite) TestRestoreBidModel_ZeroAuctionID_ReturnsError() {
	// Arrange
	id := uint64(100)
	auctionID := uint64(0)
	userID := uint64(2)
	amount := model.NewMoneyModel(10000)
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()

	// Act
	_, err := model.RestoreBidModel(id, auctionID, userID, amount, createdAt, updatedAt)

	// Assert
	s.Require().ErrorIs(err, errs.ErrAuctionIDRequired)
}

func (s *BidModelTestSuite) TestRestoreBidModel_ZeroUserID_ReturnsError() {
	// Arrange
	id := uint64(100)
	auctionID := uint64(1)
	userID := uint64(0)
	amount := model.NewMoneyModel(10000)
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()

	// Act
	_, err := model.RestoreBidModel(id, auctionID, userID, amount, createdAt, updatedAt)

	// Assert
	s.Require().ErrorIs(err, errs.ErrUserIDRequired)
}

func (s *BidModelTestSuite) TestGetters_ReturnCorrectValues() {
	// Arrange
	id := uint64(100)
	auctionID := uint64(1)
	userID := uint64(2)
	amount := model.NewMoneyModel(10000)
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	bidModel, _ := model.RestoreBidModel(id, auctionID, userID, amount, createdAt, updatedAt)

	// Act & Assert
	s.Equal(id, bidModel.ID())
	s.Equal(auctionID, bidModel.AuctionID())
	s.Equal(userID, bidModel.UserID())
	s.Equal(amount, bidModel.Amount())
	s.Equal(createdAt, bidModel.CreatedAt())
	s.Equal(updatedAt, bidModel.UpdatedAt())
}
