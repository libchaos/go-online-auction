package model_test

import (
	"testing"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuctionModel_InitializesWithNilHighestBidAmount(t *testing.T) {
	// Arrange
	listingID := uint64(100)
	endTime := time.Now().UTC().Add(24 * time.Hour)

	// Act
	auction, err := model.NewAuctionModel(listingID, endTime)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, auction.HighestBidAmount())
	assert.Nil(t, auction.HighestBidID())
}

func TestPlaceBid_SetsHighestBidIDAndAmount(t *testing.T) {
	// Arrange
	listingID := uint64(100)
	endTime := time.Now().UTC().Add(24 * time.Hour)
	auction, _ := model.NewAuctionModel(listingID, endTime)
	_ = auction.Start()

	bidID := uint64(200)
	amountInCents := uint64(5000)
	amount := model.NewMoneyModel(amountInCents)

	// Act
	err := auction.PlaceBid(bidID, amount, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, auction.HighestBidID())
	assert.Equal(t, bidID, *auction.HighestBidID())

	assert.NotNil(t, auction.HighestBidAmount())
	assert.Equal(t, amountInCents, *auction.HighestBidAmount())
}

func TestPlaceBid_UpdatesExistingHighestBidAmount(t *testing.T) {
	// Arrange
	listingID := uint64(100)
	endTime := time.Now().UTC().Add(24 * time.Hour)
	auction, _ := model.NewAuctionModel(listingID, endTime)
	_ = auction.Start()

	// Place first bid
	firstBidID := uint64(200)
	firstAmountInCents := uint64(1000)
	firstAmount := model.NewMoneyModel(firstAmountInCents)
	_ = auction.PlaceBid(firstBidID, firstAmount, nil)

	// Verify first bid set correctly
	assert.Equal(t, firstBidID, *auction.HighestBidID())
	assert.Equal(t, firstAmountInCents, *auction.HighestBidAmount())

	// Create a current highest bid model to simulate existing state checking (optional but good context)
	// Note: We don't construct the full BidModel here because PlaceBid takes it as a dependency for validation
	// but the actual update happens on the auction model.

	currentHighestBid, _ := model.RestoreBidModel(
		firstBidID,
		auction.ID(),
		1, // userID
		firstAmount,
		time.Now(),
		time.Now(),
	)

	// Place second bid
	secondBidID := uint64(201)
	secondAmountInCents := uint64(2000)
	secondAmount := model.NewMoneyModel(secondAmountInCents)

	// Act
	err := auction.PlaceBid(secondBidID, secondAmount, &currentHighestBid)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, secondBidID, *auction.HighestBidID())
	assert.Equal(t, secondAmountInCents, *auction.HighestBidAmount())
}

func TestRestoreAuctionModel_SetsHighestBidAmount(t *testing.T) {
	// Arrange
	now := time.Now().UTC()
	listingID := uint64(100)
	state, _ := enum.NewAuctionStateEnum("active")
	highestBidID := uint64(200)
	highestBidAmount := uint64(5000)

	// Act
	auction, err := model.RestoreAuctionModel(
		1,
		listingID,
		now,
		now.Add(24*time.Hour),
		state,
		&highestBidID,
		&highestBidAmount,
		1,
		now,
		now,
	)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, auction.HighestBidID())
	assert.Equal(t, highestBidID, *auction.HighestBidID())

	assert.NotNil(t, auction.HighestBidAmount())
	assert.Equal(t, highestBidAmount, *auction.HighestBidAmount())
}
