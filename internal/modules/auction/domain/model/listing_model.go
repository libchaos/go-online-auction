package model

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// ListingModel represents a listing entity in the auction domain
type ListingModel struct {
	id          uint64
	ownerUserID uint64
	title       string
	description string
	createdAt   time.Time
	updatedAt   time.Time
}

// NewListingModel creates a new listing entity without ID
// ownerUserID: the ID of the user who owns this listing
// title: the listing title
// description: the listing description (can be empty)
func NewListingModel(ownerUserID uint64, title, description string) (ListingModel, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if err := validateListingTitle(title); err != nil {
		return ListingModel{}, err
	}

	if err := validateListingDescription(description); err != nil {
		return ListingModel{}, err
	}

	now := time.Now().UTC()
	return ListingModel{
		ownerUserID: ownerUserID,
		title:       title,
		description: description,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// RestoreListingModel reconstitutes an existing listing entity with all fields
func RestoreListingModel(id, ownerUserID uint64, title, description string, createdAt, updatedAt time.Time) (ListingModel, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if err := validateListingTitle(title); err != nil {
		return ListingModel{}, err
	}

	if err := validateListingDescription(description); err != nil {
		return ListingModel{}, err
	}

	return ListingModel{
		id:          id,
		ownerUserID: ownerUserID,
		title:       title,
		description: description,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// ID returns the listing's ID
func (l ListingModel) ID() uint64 {
	return l.id
}

// OwnerUserID returns the owner's user ID
func (l ListingModel) OwnerUserID() uint64 {
	return l.ownerUserID
}

// Title returns the listing title
func (l ListingModel) Title() string {
	return l.title
}

// Description returns the listing description
func (l ListingModel) Description() string {
	return l.description
}

// CreatedAt returns the creation timestamp
func (l ListingModel) CreatedAt() time.Time {
	return l.createdAt
}

// UpdatedAt returns the last update timestamp
func (l ListingModel) UpdatedAt() time.Time {
	return l.updatedAt
}

// validateListingTitle validates the listing title
func validateListingTitle(title string) error {
	if title == "" {
		return errors.New("listing title cannot be empty")
	}

	if utf8.RuneCountInString(title) > 200 {
		return errors.New("listing title cannot exceed 200 characters")
	}

	return nil
}

// validateListingDescription validates the listing description
func validateListingDescription(description string) error {
	if utf8.RuneCountInString(description) > 2000 {
		return errors.New("listing description cannot exceed 2000 characters")
	}

	return nil
}
