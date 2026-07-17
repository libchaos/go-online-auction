package model_test

import (
	"testing"
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"github.com/stretchr/testify/require"
)

func TestNewAuctionModel(t *testing.T) {
	t.Run("valid input creates draft auction with nil highest bid", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)

		// Act
		auction, err := model.NewAuctionModel(listingID, endTime)

		// Assert
		require.NoError(t, err)
		require.Equal(t, listingID, auction.ListingID())
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateDraft, state.String())
		require.Nil(t, auction.HighestBidAmount())
		require.Equal(t, uint64(1), auction.Version())
	})

	t.Run("returns error when listing id is zero", func(t *testing.T) {
		// Arrange
		listingID := uint64(0)
		endTime := time.Now().UTC().Add(24 * time.Hour)

		// Act
		_, err := model.NewAuctionModel(listingID, endTime)

		// Assert
		require.ErrorIs(t, err, errs.ErrListingIDRequired)
	})

	t.Run("returns error when end time is zero", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Time{}

		// Act
		_, err := model.NewAuctionModel(listingID, endTime)

		// Assert
		require.ErrorIs(t, err, errs.ErrEndTimeRequired)
	})
}

func TestNewAuctionModelWithMode_ScheduledStartTime(t *testing.T) {
	newAuction := func(startTime *time.Time, endTime time.Time) (model.AuctionModel, error) {
		tradingMode, err := enum.NewTradingModeEnum(enum.EnumTradingModeEnglish)
		require.NoError(t, err)
		return model.NewAuctionModelWithMode(
			100,
			endTime,
			tradingMode,
			nil,
			nil,
			nil,
			false,
			300,
			startTime,
		)
	}

	t.Run("valid future start time before end time is stored in UTC", func(t *testing.T) {
		// Arrange
		startTime := time.Now().Add(time.Hour)
		endTime := time.Now().UTC().Add(24 * time.Hour)

		// Act
		auction, err := newAuction(&startTime, endTime)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, auction.StartTime())
		require.Equal(t, startTime.UTC(), *auction.StartTime())
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateDraft, state.String())
	})

	t.Run("nil start time keeps auction manual", func(t *testing.T) {
		// Arrange
		endTime := time.Now().UTC().Add(24 * time.Hour)

		// Act
		auction, err := newAuction(nil, endTime)

		// Assert
		require.NoError(t, err)
		require.Nil(t, auction.StartTime())
	})

	t.Run("returns error when start time is in the past", func(t *testing.T) {
		// Arrange
		startTime := time.Now().UTC().Add(-time.Hour)
		endTime := time.Now().UTC().Add(24 * time.Hour)

		// Act
		_, err := newAuction(&startTime, endTime)

		// Assert
		require.ErrorIs(t, err, errs.ErrStartTimeMustBeInFuture)
	})

	t.Run("returns error when start time is not before end time", func(t *testing.T) {
		// Arrange
		endTime := time.Now().UTC().Add(time.Hour)
		startTime := endTime.Add(time.Minute)

		// Act
		_, err := newAuction(&startTime, endTime)

		// Assert
		require.ErrorIs(t, err, errs.ErrStartTimeMustBeBeforeEndTime)
	})
}

func TestRestoreAuctionModel(t *testing.T) {
	t.Run("restores auction with all fields", func(t *testing.T) {
		// Arrange
		now := time.Now().UTC()
		id := uint64(1)
		listingID := uint64(100)
		startTime := now
		endTime := now.Add(24 * time.Hour)
		state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
		highestBidAmount := uint64(5000)
		version := uint64(3)

		// Act
		auction, err := model.RestoreAuctionModel(
			id,
			listingID,
			&startTime,
			endTime,
			state,
			&highestBidAmount,
			version,
			now,
			now,
		)

		// Assert
		require.NoError(t, err)
		require.Equal(t, id, auction.ID())
		require.Equal(t, listingID, auction.ListingID())
		require.NotNil(t, auction.StartTime())
		require.Equal(t, startTime, *auction.StartTime())
		require.Equal(t, endTime, auction.EndTime())
		require.Equal(t, version, auction.Version())
		require.Equal(t, now, auction.CreatedAt())
		require.Equal(t, now, auction.UpdatedAt())
		require.NotNil(t, auction.HighestBidAmount())
		require.Equal(t, highestBidAmount, *auction.HighestBidAmount())
	})

	t.Run("returns error when id is zero", func(t *testing.T) {
		// Arrange
		now := time.Now().UTC()
		state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)

		// Act
		_, err := model.RestoreAuctionModel(
			0,
			100,
			&now,
			now.Add(24*time.Hour),
			state,
			nil,
			0,
			now,
			now,
		)

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionIDRequired)
	})

	t.Run("returns error when listing id is zero", func(t *testing.T) {
		// Arrange
		now := time.Now().UTC()
		state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)

		// Act
		_, err := model.RestoreAuctionModel(
			1,
			0,
			&now,
			now.Add(24*time.Hour),
			state,
			nil,
			0,
			now,
			now,
		)

		// Assert
		require.ErrorIs(t, err, errs.ErrListingIDRequired)
	})

	t.Run("returns error when end time is zero", func(t *testing.T) {
		// Arrange
		now := time.Now().UTC()
		state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)

		// Act
		_, err := model.RestoreAuctionModel(
			1,
			100,
			&now,
			time.Time{},
			state,
			nil,
			0,
			now,
			now,
		)

		// Assert
		require.ErrorIs(t, err, errs.ErrEndTimeRequired)
	})
}

func TestAuctionModel_Start(t *testing.T) {
	t.Run("starts draft auction successfully", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)

		// Act
		err := auction.Start()

		// Assert
		require.NoError(t, err)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateActive, state.String())
		require.NotZero(t, auction.StartTime())
		require.Equal(t, uint64(2), auction.Version())
	})

	t.Run("returns error when starting non-draft auction", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		// Act
		err := auction.Start()

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionCanOnlyStartFromDraft)
	})
}

func TestAuctionModel_PlaceBid(t *testing.T) {
	t.Run("places first bid successfully", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		amountInCents := uint64(5000)
		amount := model.NewMoneyModel(amountInCents)

		// Act
		err := auction.PlaceBid(amount)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, auction.HighestBidAmount())
		require.Equal(t, amountInCents, *auction.HighestBidAmount())
		require.Equal(t, uint64(3), auction.Version())
	})

	t.Run("updates existing highest bid", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		firstAmountInCents := uint64(1000)
		firstAmount := model.NewMoneyModel(firstAmountInCents)
		_ = auction.PlaceBid(firstAmount)

		secondAmountInCents := uint64(2000)
		secondAmount := model.NewMoneyModel(secondAmountInCents)

		// Act
		err := auction.PlaceBid(secondAmount)

		// Assert
		require.NoError(t, err)
		require.Equal(t, secondAmountInCents, *auction.HighestBidAmount())
	})

	t.Run("returns error when bidding on non-active auction", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)

		amount := model.NewMoneyModel(5000)

		// Act
		err := auction.PlaceBid(amount)

		// Assert
		require.ErrorIs(t, err, errs.ErrBidsOnlyOnActiveAuctions)
	})

	t.Run("returns error when auction is expired", func(t *testing.T) {
		// Arrange
		now := time.Now().UTC()
		pastEndTime := now.Add(-1 * time.Hour)
		state, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
		startTime := now.Add(-2 * time.Hour)

		auction, _ := model.RestoreAuctionModel(
			1,
			100,
			&startTime,
			pastEndTime,
			state,
			nil,
			0,
			now,
			now,
		)

		amount := model.NewMoneyModel(5000)

		// Act
		err := auction.PlaceBid(amount)

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionExpired)
	})

	t.Run("returns error when first bid is zero", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		amount := model.NewMoneyModel(0)

		// Act
		err := auction.PlaceBid(amount)

		// Assert
		require.ErrorIs(t, err, errs.ErrFirstBidMustBePositive)
	})

	t.Run("returns error when bid does not exceed highest", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		firstAmount := model.NewMoneyModel(1000)
		_ = auction.PlaceBid(firstAmount)

		secondAmount := model.NewMoneyModel(500)

		// Act
		err := auction.PlaceBid(secondAmount)

		// Assert
		require.ErrorIs(t, err, errs.ErrBidMustExceedHighest)
	})
}

func TestAuctionModel_Close(t *testing.T) {
	t.Run("closes active auction successfully", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		// Act
		err := auction.Close([]model.BidModel{})

		// Assert
		require.NoError(t, err)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateClosed, state.String())
		require.Equal(t, uint64(3), auction.Version())
	})

	t.Run("returns error when closing non-active auction", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)

		// Act
		err := auction.Close([]model.BidModel{})

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionCanOnlyCloseFromActive)
	})
}

func TestAuctionModel_Cancel(t *testing.T) {
	t.Run("cancels draft auction successfully", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)

		// Act
		err := auction.Cancel()

		// Assert
		require.NoError(t, err)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateCancelled, state.String())
		require.Equal(t, uint64(2), auction.Version())
	})

	t.Run("cancels active auction successfully", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		// Act
		err := auction.Cancel()

		// Assert
		require.NoError(t, err)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateCancelled, state.String())
	})

	t.Run("returns error when cancelling closed auction", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()
		_ = auction.Close([]model.BidModel{})

		// Act
		err := auction.Cancel()

		// Assert
		require.ErrorIs(t, err, errs.ErrAuctionCanOnlyCancelFromDraftOrActive)
	})
}

func TestAuctionModel_CheckAndCloseIfExpired(t *testing.T) {
	t.Run("closes expired active auction", func(t *testing.T) {
		// Arrange
		now := time.Now().UTC()
		pastEndTime := now.Add(-1 * time.Hour)
		activeState, _ := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
		startTime := now.Add(-2 * time.Hour)

		auction, _ := model.RestoreAuctionModel(
			1,
			100,
			&startTime,
			pastEndTime,
			activeState,
			nil,
			0,
			now,
			now,
		)

		// Act
		closed, err := auction.CheckAndCloseIfExpired([]model.BidModel{})

		// Assert
		require.NoError(t, err)
		require.True(t, closed)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateClosed, state.String())
	})

	t.Run("does not close non-expired active auction", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)
		_ = auction.Start()

		// Act
		closed, err := auction.CheckAndCloseIfExpired([]model.BidModel{})

		// Assert
		require.NoError(t, err)
		require.False(t, closed)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateActive, state.String())
	})

	t.Run("does not close non-active auction", func(t *testing.T) {
		// Arrange
		listingID := uint64(100)
		endTime := time.Now().UTC().Add(24 * time.Hour)
		auction, _ := model.NewAuctionModel(listingID, endTime)

		// Act
		closed, err := auction.CheckAndCloseIfExpired([]model.BidModel{})

		// Assert
		require.NoError(t, err)
		require.False(t, closed)
		state := auction.State()
		require.Equal(t, enum.EnumAuctionStateDraft, state.String())
	})
}
