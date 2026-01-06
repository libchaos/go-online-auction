package enum

import "errors"

const (
	EnumBidStatusAccepted   string = "accepted"
	EnumBidStatusRejected   string = "rejected"
	EnumBidStatusSuperseded string = "superseded"
)

// BidStatusEnum represents the status of a bid
type BidStatusEnum struct {
	value string
}

// NewBidStatusEnum creates a new BidStatusEnum
func NewBidStatusEnum(value string) (BidStatusEnum, error) {
	if err := validateBidStatusEnum(value); err != nil {
		return BidStatusEnum{}, err
	}

	return BidStatusEnum{value: value}, nil
}

// String returns the string representation of the enum
func (e *BidStatusEnum) String() string {
	return e.value
}

// validateBidStatusEnum validates the bid status value
func validateBidStatusEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumBidStatusAccepted:   {},
		EnumBidStatusRejected:   {},
		EnumBidStatusSuperseded: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid bid status: " + value)
	}

	return nil
}
