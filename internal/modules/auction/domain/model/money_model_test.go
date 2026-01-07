package model_test

import (
	"testing"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestNewMoneyModel(t *testing.T) {
	t.Run("should create a new money model with correct amount", func(t *testing.T) {
		// Arrange
		amount := uint64(100)

		// Act
		m := model.NewMoneyModel(amount)

		// Assert
		assert.Equal(t, amount, m.AmountInCents())
	})
}

func TestMoneyModel_IsGreaterThan(t *testing.T) {
	t.Run("should return true when amount is greater than other", func(t *testing.T) {
		// Arrange
		m1 := model.NewMoneyModel(200)
		m2 := model.NewMoneyModel(100)

		// Act
		result := m1.IsGreaterThan(m2)

		// Assert
		assert.True(t, result)
	})

	t.Run("should return false when amount is less than other", func(t *testing.T) {
		// Arrange
		m1 := model.NewMoneyModel(100)
		m2 := model.NewMoneyModel(200)

		// Act
		result := m1.IsGreaterThan(m2)

		// Assert
		assert.False(t, result)
	})

	t.Run("should return false when amount is equal to other", func(t *testing.T) {
		// Arrange
		m1 := model.NewMoneyModel(100)
		m2 := model.NewMoneyModel(100)

		// Act
		result := m1.IsGreaterThan(m2)

		// Assert
		assert.False(t, result)
	})
}

func TestMoneyModel_IsGreaterThanOrEqual(t *testing.T) {
	t.Run("should return true when amount is greater than other", func(t *testing.T) {
		// Arrange
		m1 := model.NewMoneyModel(200)
		m2 := model.NewMoneyModel(100)

		// Act
		result := m1.IsGreaterThanOrEqual(m2)

		// Assert
		assert.True(t, result)
	})

	t.Run("should return true when amount is equal to other", func(t *testing.T) {
		// Arrange
		m1 := model.NewMoneyModel(100)
		m2 := model.NewMoneyModel(100)

		// Act
		result := m1.IsGreaterThanOrEqual(m2)

		// Assert
		assert.True(t, result)
	})

	t.Run("should return false when amount is less than other", func(t *testing.T) {
		// Arrange
		m1 := model.NewMoneyModel(100)
		m2 := model.NewMoneyModel(200)

		// Act
		result := m1.IsGreaterThanOrEqual(m2)

		// Assert
		assert.False(t, result)
	})
}
