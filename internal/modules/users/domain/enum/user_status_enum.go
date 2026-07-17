package enum

import (
	"auction/internal/modules/users/domain/errs"
)

const (
	EnumUserStatusActive   string = "active"
	EnumUserStatusInactive string = "inactive"
	EnumUserStatusBlocked  string = "blocked"
)

// UserStatusEnum represents the status of a user account
type UserStatusEnum struct {
	value string
}

// NewUserStatusEnum creates a new UserStatusEnum
func NewUserStatusEnum(value string) (UserStatusEnum, error) {
	if err := validateUserStatusEnum(value); err != nil {
		return UserStatusEnum{}, err
	}

	return UserStatusEnum{value: value}, nil
}

// String returns the string representation of the enum
func (e *UserStatusEnum) String() string {
	return e.value
}

// validateUserStatusEnum validates the user status value
func validateUserStatusEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumUserStatusActive:   {},
		EnumUserStatusInactive: {},
		EnumUserStatusBlocked:  {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errs.ErrInvalidUserStatus
	}

	return nil
}
