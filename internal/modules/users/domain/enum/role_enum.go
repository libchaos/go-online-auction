package enum

import (
	"auction/internal/modules/users/domain/errs"
)

const (
	EnumRoleAdmin  string = "admin"
	EnumRoleSeller string = "seller"
	EnumRoleBidder string = "bidder"
)

// RoleEnum represents the role of a user
type RoleEnum struct {
	value string
}

// NewRoleEnum creates a new RoleEnum
func NewRoleEnum(value string) (RoleEnum, error) {
	if err := validateRoleEnum(value); err != nil {
		return RoleEnum{}, err
	}

	return RoleEnum{value: value}, nil
}

// String returns the string representation of the enum
func (e *RoleEnum) String() string {
	return e.value
}

// validateRoleEnum validates the role value
func validateRoleEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumRoleAdmin:  {},
		EnumRoleSeller: {},
		EnumRoleBidder: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errs.ErrInvalidRole
	}

	return nil
}
