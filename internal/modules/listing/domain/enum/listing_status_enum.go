package enum

import (
	"auction/internal/modules/listing/domain/errs"
)

const (
	EnumListingStatusDraft     string = "draft"
	EnumListingStatusPublished string = "published"
	EnumListingStatusOffShelf  string = "off_shelf"
)

// ListingStatusEnum represents the lifecycle status of an SPU or SKU
type ListingStatusEnum struct {
	value string
}

// NewListingStatusEnum creates a new ListingStatusEnum
func NewListingStatusEnum(value string) (ListingStatusEnum, error) {
	if err := validateListingStatusEnum(value); err != nil {
		return ListingStatusEnum{}, err
	}

	return ListingStatusEnum{value: value}, nil
}

// String returns the string representation of the enum
func (e *ListingStatusEnum) String() string {
	return e.value
}

// IsDraft reports whether the status is draft
func (e *ListingStatusEnum) IsDraft() bool {
	return e.value == EnumListingStatusDraft
}

// IsPublished reports whether the status is published
func (e *ListingStatusEnum) IsPublished() bool {
	return e.value == EnumListingStatusPublished
}

// IsOffShelf reports whether the status is off_shelf
func (e *ListingStatusEnum) IsOffShelf() bool {
	return e.value == EnumListingStatusOffShelf
}

// validateListingStatusEnum validates the listing status value
func validateListingStatusEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumListingStatusDraft:     {},
		EnumListingStatusPublished: {},
		EnumListingStatusOffShelf:  {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errs.ErrInvalidListingStatus
	}

	return nil
}
