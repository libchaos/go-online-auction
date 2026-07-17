package model_test

import (
	"testing"
	"time"

	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBidModel(t *testing.T) {
	t.Run("should create a new bid model with valid input", func(t *testing.T) {
		// Arrange
		auctionID := uint64(100)
		userID := uint64(200)
		amount := model.NewMoneyModel(5000)

		// Act
		bid, err := model.NewBidModel(auctionID, userID, amount)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, auctionID, bid.AuctionID())
		assert.Equal(t, userID, bid.UserID())
		assert.Equal(t, amount, bid.Amount())
		assert.False(t, bid.CreatedAt().IsZero())
		assert.False(t, bid.UpdatedAt().IsZero())
	})

	t.Run("should return error when auction id is zero", func(t *testing.T) {
		// Arrange
		auctionID := uint64(0)
		userID := uint64(200)
		amount := model.NewMoneyModel(5000)

		// Act
		_, err := model.NewBidModel(auctionID, userID, amount)

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionIDRequired)
	})

	t.Run("should return error when user id is zero", func(t *testing.T) {
		// Arrange
		auctionID := uint64(100)
		userID := uint64(0)
		amount := model.NewMoneyModel(5000)

		// Act
		_, err := model.NewBidModel(auctionID, userID, amount)

		// Assert
		require.ErrorIs(t, err, errs.ErrUserIDRequired)
	})
}

func TestRestoreBidModel(t *testing.T) {
	t.Run("should restore bid model with valid input", func(t *testing.T) {
		// Arrange
		id := uint64(1)
		auctionID := uint64(100)
		userID := uint64(200)
		amount := model.NewMoneyModel(5000)
		now := time.Now().UTC()

		// Act
		bid, err := model.RestoreBidModel(id, auctionID, userID, amount, now, now)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, bid.ID())
		assert.Equal(t, auctionID, bid.AuctionID())
		assert.Equal(t, userID, bid.UserID())
		assert.Equal(t, amount, bid.Amount())
		assert.Equal(t, now, bid.CreatedAt())
		assert.Equal(t, now, bid.UpdatedAt())
	})

	t.Run("should return error when id is zero", func(t *testing.T) {
		// Arrange
		id := uint64(0)
		auctionID := uint64(100)
		userID := uint64(200)
		amount := model.NewMoneyModel(5000)
		now := time.Now().UTC()

		// Act
		_, err := model.RestoreBidModel(id, auctionID, userID, amount, now, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrBidIDRequired)
	})

	t.Run("should return error when auction id is zero", func(t *testing.T) {
		// Arrange
		id := uint64(1)
		auctionID := uint64(0)
		userID := uint64(200)
		amount := model.NewMoneyModel(5000)
		now := time.Now().UTC()

		// Act
		_, err := model.RestoreBidModel(id, auctionID, userID, amount, now, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionIDRequired)
	})

	t.Run("should return error when user id is zero", func(t *testing.T) {
		// Arrange
		id := uint64(1)
		auctionID := uint64(100)
		userID := uint64(0)
		amount := model.NewMoneyModel(5000)
		now := time.Now().UTC()

		// Act
		_, err := model.RestoreBidModel(id, auctionID, userID, amount, now, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrUserIDRequired)
	})
}
