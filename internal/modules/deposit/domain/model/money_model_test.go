package model_test

import (
	"errors"
	"testing"

	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMoneyModel(t *testing.T) {
	t.Run("should store the amount in cents", func(t *testing.T) {
		// Arrange / Act
		money := model.NewMoneyModel(1234)

		// Assert
		assert.Equal(t, uint64(1234), money.AmountInCents())
	})
}

func TestMoneyModel_IsZero(t *testing.T) {
	t.Run("should be zero for amount zero", func(t *testing.T) {
		// Arrange / Act / Assert
		assert.True(t, model.NewMoneyModel(0).IsZero())
	})

	t.Run("should not be zero for positive amount", func(t *testing.T) {
		// Arrange / Act / Assert
		assert.False(t, model.NewMoneyModel(1).IsZero())
	})
}

func TestMoneyModel_Comparisons(t *testing.T) {
	t.Run("IsGreaterThan should compare correctly", func(t *testing.T) {
		// Arrange
		big := model.NewMoneyModel(200)
		small := model.NewMoneyModel(100)

		// Act / Assert
		assert.True(t, big.IsGreaterThan(small))
		assert.False(t, small.IsGreaterThan(big))
		assert.False(t, big.IsGreaterThan(big))
	})

	t.Run("IsGreaterThanOrEqual should compare correctly", func(t *testing.T) {
		// Arrange
		big := model.NewMoneyModel(200)
		small := model.NewMoneyModel(100)

		// Act / Assert
		assert.True(t, big.IsGreaterThanOrEqual(small))
		assert.True(t, big.IsGreaterThanOrEqual(big))
		assert.False(t, small.IsGreaterThanOrEqual(big))
	})

	t.Run("IsLessThan should compare correctly", func(t *testing.T) {
		// Arrange
		big := model.NewMoneyModel(200)
		small := model.NewMoneyModel(100)

		// Act / Assert
		assert.True(t, small.IsLessThan(big))
		assert.False(t, big.IsLessThan(small))
	})
}

func TestMoneyModel_Add(t *testing.T) {
	t.Run("should add amounts immutably", func(t *testing.T) {
		// Arrange
		first := model.NewMoneyModel(100)
		second := model.NewMoneyModel(50)

		// Act
		sum := first.Add(second)

		// Assert
		assert.Equal(t, uint64(150), sum.AmountInCents())
		assert.Equal(t, uint64(100), first.AmountInCents())
	})
}

func TestMoneyModel_Subtract(t *testing.T) {
	t.Run("should subtract amounts immutably", func(t *testing.T) {
		// Arrange
		first := model.NewMoneyModel(100)
		second := model.NewMoneyModel(40)

		// Act
		difference, err := first.Subtract(second)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, uint64(60), difference.AmountInCents())
		assert.Equal(t, uint64(100), first.AmountInCents())
	})

	t.Run("should return error on underflow", func(t *testing.T) {
		// Arrange
		first := model.NewMoneyModel(40)
		second := model.NewMoneyModel(100)

		// Act
		difference, err := first.Subtract(second)

		// Assert
		assert.ErrorIs(t, err, errs.ErrInsufficientBalance)
		assert.Equal(t, uint64(0), difference.AmountInCents())
	})
}

func TestMoneyModel_Subtract_ErrorValue(t *testing.T) {
	// Arrange
	first := model.NewMoneyModel(10)
	second := model.NewMoneyModel(20)

	// Act
	_, err := first.Subtract(second)

	// Assert
	assert.True(t, errors.Is(err, errs.ErrInsufficientBalance))
}
