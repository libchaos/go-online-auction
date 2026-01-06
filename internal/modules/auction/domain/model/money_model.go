package model

import (
	"errors"
	"strings"
	"unicode"
)

// MoneyModel represents an immutable monetary amount with currency
type MoneyModel struct {
	amountInCents uint64
	currency      string
}

// NewMoneyModel creates a new MoneyModel value object
// amountInCents: the amount in cents
// currency: 3-letter currency code (e.g., USD, EUR, GBP)
func NewMoneyModel(amountInCents uint64, currency string) (MoneyModel, error) {
	currency = strings.TrimSpace(currency)
	currency = strings.ToUpper(currency)

	if err := validateMoneyModel(amountInCents, currency); err != nil {
		return MoneyModel{}, err
	}

	return MoneyModel{
		amountInCents: amountInCents,
		currency:      currency,
	}, nil
}

// AmountInCents returns the amount in cents
func (m MoneyModel) AmountInCents() uint64 {
	return m.amountInCents
}

// Currency returns the currency code
func (m MoneyModel) Currency() string {
	return m.currency
}

// IsGreaterThan compares this MoneyModel with another
// Returns error if currencies don't match
func (m MoneyModel) IsGreaterThan(other MoneyModel) (bool, error) {
	if m.currency != other.currency {
		return false, errors.New("cannot compare amounts with different currencies")
	}
	return m.amountInCents > other.amountInCents, nil
}

// validateMoneyModel validates the currency
func validateMoneyModel(amountInCents uint64, currency string) error {
	if currency == "" {
		return errors.New("currency cannot be empty")
	}

	if len(currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}

	// Validate that currency code contains only letters
	for _, ch := range currency {
		if !unicode.IsLetter(ch) {
			return errors.New("currency must contain only letters")
		}
	}

	return nil
}
