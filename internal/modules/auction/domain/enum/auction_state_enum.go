package enum

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
)

const (
	EnumAuctionStateDraft     string = "draft"
	EnumAuctionStateActive    string = "active"
	EnumAuctionStateClosed    string = "closed"
	EnumAuctionStateCancelled string = "cancelled"
)

// AuctionStateEnum represents the lifecycle state of an auction
type AuctionStateEnum struct {
	value string
}

// NewAuctionStateEnum creates a new AuctionStateEnum
func NewAuctionStateEnum(value string) (AuctionStateEnum, error) {
	if err := validateAuctionStateEnum(value); err != nil {
		return AuctionStateEnum{}, err
	}

	return AuctionStateEnum{value: value}, nil
}

// String returns the string representation of the enum
func (e *AuctionStateEnum) String() string {
	return e.value
}

// validateAuctionStateEnum validates the auction state value
func validateAuctionStateEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumAuctionStateDraft:     {},
		EnumAuctionStateActive:    {},
		EnumAuctionStateClosed:    {},
		EnumAuctionStateCancelled: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errs.ErrInvalidAuctionState
	}

	return nil
}
