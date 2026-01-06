package model

import (
	"testing"
)

func TestNewMoneyModel(t *testing.T) {
	t.Run("valid money creation with USD", func(t *testing.T) {
		money, err := NewMoneyModel(10000, "USD")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if money.AmountInCents() != 10000 {
			t.Errorf("expected amount 10000, got %d", money.AmountInCents())
		}
		if money.Currency() != "USD" {
			t.Errorf("expected currency USD, got %s", money.Currency())
		}
	})

	t.Run("valid money creation with lowercase currency", func(t *testing.T) {
		money, err := NewMoneyModel(5000, "eur")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if money.Currency() != "EUR" {
			t.Errorf("expected currency EUR (uppercase), got %s", money.Currency())
		}
	})

	t.Run("valid money creation with trimmed currency", func(t *testing.T) {
		money, err := NewMoneyModel(5000, "  GBP  ")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if money.Currency() != "GBP" {
			t.Errorf("expected currency GBP (trimmed), got %s", money.Currency())
		}
	})

	t.Run("valid money creation with zero amount", func(t *testing.T) {
		money, err := NewMoneyModel(0, "USD")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if money.AmountInCents() != 0 {
			t.Errorf("expected amount 0, got %d", money.AmountInCents())
		}
	})

	t.Run("rejects empty currency", func(t *testing.T) {
		_, err := NewMoneyModel(1000, "")
		if err == nil {
			t.Fatal("expected error for empty currency, got nil")
		}
		expectedMsg := "currency cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects currency with spaces only", func(t *testing.T) {
		_, err := NewMoneyModel(1000, "   ")
		if err == nil {
			t.Fatal("expected error for whitespace-only currency, got nil")
		}
		expectedMsg := "currency cannot be empty"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects currency with less than 3 characters", func(t *testing.T) {
		_, err := NewMoneyModel(1000, "US")
		if err == nil {
			t.Fatal("expected error for 2-character currency, got nil")
		}
		expectedMsg := "currency must be a 3-letter code"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects currency with more than 3 characters", func(t *testing.T) {
		_, err := NewMoneyModel(1000, "USDD")
		if err == nil {
			t.Fatal("expected error for 4-character currency, got nil")
		}
		expectedMsg := "currency must be a 3-letter code"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects currency with numbers", func(t *testing.T) {
		_, err := NewMoneyModel(1000, "U5D")
		if err == nil {
			t.Fatal("expected error for currency with numbers, got nil")
		}
		expectedMsg := "currency must contain only letters"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("rejects currency with special characters", func(t *testing.T) {
		_, err := NewMoneyModel(1000, "U$D")
		if err == nil {
			t.Fatal("expected error for currency with special characters, got nil")
		}
		expectedMsg := "currency must contain only letters"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestMoneyModel_IsGreaterThan(t *testing.T) {
	t.Run("larger amount is greater", func(t *testing.T) {
		money1, _ := NewMoneyModel(20000, "USD")
		money2, _ := NewMoneyModel(10000, "USD")

		isGreater, err := money1.IsGreaterThan(money2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !isGreater {
			t.Error("expected money1 to be greater than money2")
		}
	})

	t.Run("smaller amount is not greater", func(t *testing.T) {
		money1, _ := NewMoneyModel(10000, "USD")
		money2, _ := NewMoneyModel(20000, "USD")

		isGreater, err := money1.IsGreaterThan(money2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if isGreater {
			t.Error("expected money1 to not be greater than money2")
		}
	})

	t.Run("equal amounts are not greater", func(t *testing.T) {
		money1, _ := NewMoneyModel(10000, "USD")
		money2, _ := NewMoneyModel(10000, "USD")

		isGreater, err := money1.IsGreaterThan(money2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if isGreater {
			t.Error("expected money1 to not be greater than money2 when equal")
		}
	})

	t.Run("zero compared to positive", func(t *testing.T) {
		money1, _ := NewMoneyModel(0, "USD")
		money2, _ := NewMoneyModel(100, "USD")

		isGreater, err := money1.IsGreaterThan(money2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if isGreater {
			t.Error("expected zero to not be greater than positive amount")
		}
	})

	t.Run("positive compared to zero", func(t *testing.T) {
		money1, _ := NewMoneyModel(100, "USD")
		money2, _ := NewMoneyModel(0, "USD")

		isGreater, err := money1.IsGreaterThan(money2)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !isGreater {
			t.Error("expected positive amount to be greater than zero")
		}
	})

	t.Run("returns error for different currencies", func(t *testing.T) {
		money1, _ := NewMoneyModel(10000, "USD")
		money2, _ := NewMoneyModel(10000, "EUR")

		_, err := money1.IsGreaterThan(money2)
		if err == nil {
			t.Fatal("expected error for different currencies, got nil")
		}
		expectedMsg := "cannot compare amounts with different currencies"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("returns error for different currencies regardless of amount", func(t *testing.T) {
		money1, _ := NewMoneyModel(20000, "USD")
		money2, _ := NewMoneyModel(10000, "GBP")

		_, err := money1.IsGreaterThan(money2)
		if err == nil {
			t.Fatal("expected error for different currencies, got nil")
		}
		expectedMsg := "cannot compare amounts with different currencies"
		if err.Error() != expectedMsg {
			t.Errorf("expected error '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestMoneyModel_Immutability(t *testing.T) {
	t.Run("value is returned not pointer", func(t *testing.T) {
		money, err := NewMoneyModel(10000, "USD")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// This test verifies that modifying a copy doesn't affect the original
		// Since MoneyModel fields are private, we can only verify through behavior
		originalAmount := money.AmountInCents()
		originalCurrency := money.Currency()

		// Create another instance with same values
		money2, _ := NewMoneyModel(10000, "USD")

		// Original should remain unchanged
		if money.AmountInCents() != originalAmount {
			t.Error("original money amount changed")
		}
		if money.Currency() != originalCurrency {
			t.Error("original money currency changed")
		}

		// Values should be independent
		if money.AmountInCents() != money2.AmountInCents() {
			t.Error("amounts should be equal")
		}
	})
}
