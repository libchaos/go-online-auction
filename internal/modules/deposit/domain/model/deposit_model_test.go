package model_test

import (
	"testing"
	"time"

	"auction/internal/modules/deposit/domain/enum"
	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID    = uint64(100)
	testAuctionID = uint64(200)
	testAmount    = uint64(5000)
	testCurrency  = "CNY"
	testReference = "ref-1"
)

func newHeldDeposit(t *testing.T) model.DepositModel {
	t.Helper()
	deposit, err := model.RestoreDepositModel(
		1,
		testUserID,
		testAuctionID,
		testAmount,
		testCurrency,
		enum.EnumDepositStatusHeld,
		"ext-ref",
		testReference,
		2,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)

	return deposit
}

func TestNewDeposit(t *testing.T) {
	t.Run("should create a pending deposit with version one", func(t *testing.T) {
		// Arrange / Act
		deposit, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)

		// Assert
		require.NoError(t, err)
		status := deposit.Status()
		assert.Equal(t, enum.EnumDepositStatusPending, status.String())
		assert.Equal(t, uint64(1), deposit.Version())
		assert.Equal(t, testUserID, deposit.UserID())
		assert.Equal(t, testAuctionID, deposit.AuctionID())
		assert.Equal(t, testAmount, deposit.Amount().AmountInCents())
		assert.Equal(t, testCurrency, deposit.Currency())
		assert.Equal(t, testReference, deposit.Reference())
	})

	t.Run("should reject zero user id", func(t *testing.T) {
		// Act
		_, err := model.NewDeposit(0, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)

		// Assert
		assert.ErrorIs(t, err, errs.ErrDepositUserRequired)
	})

	t.Run("should reject zero auction id", func(t *testing.T) {
		// Act
		_, err := model.NewDeposit(testUserID, 0, model.NewMoneyModel(testAmount), testCurrency, testReference)

		// Assert
		assert.ErrorIs(t, err, errs.ErrDepositAuctionRequired)
	})

	t.Run("should reject zero amount", func(t *testing.T) {
		// Act
		_, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(0), testCurrency, testReference)

		// Assert
		assert.ErrorIs(t, err, errs.ErrDepositAmountRequired)
	})

	t.Run("should reject empty currency", func(t *testing.T) {
		// Act
		_, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), "", testReference)

		// Assert
		assert.ErrorIs(t, err, errs.ErrDepositCurrencyRequired)
	})

	t.Run("should reject empty reference", func(t *testing.T) {
		// Act
		_, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, "")

		// Assert
		assert.ErrorIs(t, err, errs.ErrDepositReferenceRequired)
	})
}

func TestDepositModel_ConfirmHold(t *testing.T) {
	t.Run("should move pending to held and bump version", func(t *testing.T) {
		// Arrange
		deposit, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)
		require.NoError(t, err)

		// Act
		err = deposit.ConfirmHold("ext-ref-1")

		// Assert
		require.NoError(t, err)
		status := deposit.Status()
		assert.Equal(t, enum.EnumDepositStatusHeld, status.String())
		assert.Equal(t, "ext-ref-1", deposit.ExternalReference())
		assert.Equal(t, uint64(2), deposit.Version())
	})

	t.Run("should reject confirm when not pending", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		err := deposit.ConfirmHold("ext-ref-2")

		// Assert
		assert.ErrorIs(t, err, errs.ErrInvalidDepositTransition)
	})
}

func TestDepositModel_Release(t *testing.T) {
	t.Run("should move held to released", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		err := deposit.Release()

		// Assert
		require.NoError(t, err)
		status := deposit.Status()
		assert.Equal(t, enum.EnumDepositStatusReleased, status.String())
		assert.Equal(t, uint64(3), deposit.Version())
	})

	t.Run("should reject release when not held", func(t *testing.T) {
		// Arrange
		deposit, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)
		require.NoError(t, err)

		// Act
		err = deposit.Release()

		// Assert
		assert.ErrorIs(t, err, errs.ErrInvalidDepositTransition)
	})
}

func TestDepositModel_ApplyToWinning(t *testing.T) {
	t.Run("should move held to applied", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		err := deposit.ApplyToWinning()

		// Assert
		require.NoError(t, err)
		status := deposit.Status()
		assert.Equal(t, enum.EnumDepositStatusApplied, status.String())
	})

	t.Run("should reject apply when not held", func(t *testing.T) {
		// Arrange
		deposit, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)
		require.NoError(t, err)

		// Act
		err = deposit.ApplyToWinning()

		// Assert
		assert.ErrorIs(t, err, errs.ErrInvalidDepositTransition)
	})
}

func TestDepositModel_Forfeit(t *testing.T) {
	t.Run("should move held to forfeited", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		err := deposit.Forfeit()

		// Assert
		require.NoError(t, err)
		status := deposit.Status()
		assert.Equal(t, enum.EnumDepositStatusForfeited, status.String())
	})
}

func TestDepositModel_Cancel(t *testing.T) {
	t.Run("should move pending to released", func(t *testing.T) {
		// Arrange
		deposit, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)
		require.NoError(t, err)

		// Act
		err = deposit.Cancel()

		// Assert
		require.NoError(t, err)
		status := deposit.Status()
		assert.Equal(t, enum.EnumDepositStatusReleased, status.String())
	})

	t.Run("should reject cancel when held", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		err := deposit.Cancel()

		// Assert
		assert.ErrorIs(t, err, errs.ErrInvalidDepositTransition)
	})
}

func TestDepositModel_IsEligible(t *testing.T) {
	t.Run("should be eligible when held and sufficient", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		eligible := deposit.IsEligible(model.NewMoneyModel(testAmount))

		// Assert
		assert.True(t, eligible)
	})

	t.Run("should be ineligible when held but insufficient", func(t *testing.T) {
		// Arrange
		deposit := newHeldDeposit(t)

		// Act
		eligible := deposit.IsEligible(model.NewMoneyModel(testAmount + 1))

		// Assert
		assert.False(t, eligible)
	})

	t.Run("should be ineligible when not held", func(t *testing.T) {
		// Arrange
		deposit, err := model.NewDeposit(testUserID, testAuctionID, model.NewMoneyModel(testAmount), testCurrency, testReference)
		require.NoError(t, err)

		// Act
		eligible := deposit.IsEligible(model.NewMoneyModel(1))

		// Assert
		assert.False(t, eligible)
	})
}
