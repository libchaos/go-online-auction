package model

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// UserModel represents a user entity in the auction domain
type UserModel struct {
	id         uint64
	identifier string // username or email
	createdAt  time.Time
	updatedAt  time.Time
}

// NewUserModel creates a new user entity without ID
// identifier: username or email identifier
func NewUserModel(identifier string) (UserModel, error) {
	identifier = strings.TrimSpace(identifier)

	if err := validateUserIdentifier(identifier); err != nil {
		return UserModel{}, err
	}

	now := time.Now().UTC()
	return UserModel{
		identifier: identifier,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// RestoreUserModel reconstitutes an existing user entity with all fields
func RestoreUserModel(id uint64, identifier string, createdAt, updatedAt time.Time) (UserModel, error) {
	identifier = strings.TrimSpace(identifier)

	if err := validateUserIdentifier(identifier); err != nil {
		return UserModel{}, err
	}

	return UserModel{
		id:         id,
		identifier: identifier,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}, nil
}

// ID returns the user's ID
func (u UserModel) ID() uint64 {
	return u.id
}

// Identifier returns the user's identifier
func (u UserModel) Identifier() string {
	return u.identifier
}

// CreatedAt returns the creation timestamp
func (u UserModel) CreatedAt() time.Time {
	return u.createdAt
}

// UpdatedAt returns the last update timestamp
func (u UserModel) UpdatedAt() time.Time {
	return u.updatedAt
}

// validateUserIdentifier validates the user identifier
func validateUserIdentifier(identifier string) error {
	if identifier == "" {
		return errors.New("user identifier cannot be empty")
	}

	if utf8.RuneCountInString(identifier) > 254 {
		return errors.New("user identifier cannot exceed 254 characters")
	}

	return nil
}
