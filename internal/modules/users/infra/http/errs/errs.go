package errs

import (
	"errors"
	"net/http"

	domainerrs "auction/internal/modules/users/domain/errs"
	"auction/pkg/errs"
)

var (
	ErrUserNotFound       = errs.New("USER_01", "User not found", http.StatusNotFound, nil)
	ErrEmailAlreadyExists = errs.New("USER_02", "Email already exists", http.StatusConflict, nil)
	ErrInvalidCredentials = errs.New("USER_03", "Invalid credentials", http.StatusUnauthorized, nil)
	ErrInvalidRequest     = errs.New("USER_04", "Invalid request body", http.StatusBadRequest, nil)
	ErrInvalidEmail       = errs.New("USER_05", "Invalid email address", http.StatusBadRequest, nil)
	ErrPasswordTooShort   = errs.New(
		"USER_06",
		"Password must have at least 8 characters",
		http.StatusBadRequest,
		nil,
	)
	ErrPasswordTooLong = errs.New(
		"USER_07",
		"Password must have at most 72 bytes",
		http.StatusBadRequest,
		nil,
	)
	ErrNameRequired = errs.New(
		"USER_08",
		"Name must have at least 2 characters",
		http.StatusBadRequest,
		nil,
	)
	ErrInvalidRole         = errs.New("USER_09", "Invalid role", http.StatusBadRequest, nil)
	ErrUserInactive        = errs.New("USER_10", "User is not active", http.StatusForbidden, nil)
	ErrRefreshTokenInvalid = errs.New(
		"USER_11",
		"Refresh token is invalid or expired",
		http.StatusUnauthorized,
		nil,
	)
	ErrInvalidUserID        = errs.New("USER_12", "Invalid user ID", http.StatusBadRequest, nil)
	ErrOptimisticLockFailed = errs.New(
		"USER_13",
		"Resource was modified by another transaction",
		http.StatusConflict,
		nil,
	)
	ErrInvalidPolicy = errs.New(
		"USER_14",
		"Invalid policy: sub, obj and act are required",
		http.StatusBadRequest,
		nil,
	)
	ErrInvalidRoleAssignment = errs.New(
		"USER_15",
		"Invalid role assignment: user id must be greater than zero and role must be admin, seller or bidder",
		http.StatusBadRequest,
		nil,
	)
)

var domainToHTTPErrorMap = []struct {
	domainError error
	httpError   error
}{
	{domainerrs.ErrUserNotFound, ErrUserNotFound},
	{domainerrs.ErrEmailAlreadyExists, ErrEmailAlreadyExists},
	{domainerrs.ErrInvalidCredentials, ErrInvalidCredentials},
	{domainerrs.ErrInvalidEmail, ErrInvalidEmail},
	{domainerrs.ErrNameRequired, ErrNameRequired},
	{domainerrs.ErrPasswordTooShort, ErrPasswordTooShort},
	{domainerrs.ErrPasswordTooLong, ErrPasswordTooLong},
	{domainerrs.ErrPasswordHashRequired, ErrInvalidRequest},
	{domainerrs.ErrInvalidRole, ErrInvalidRole},
	{domainerrs.ErrInvalidUserStatus, ErrInvalidRequest},
	{domainerrs.ErrUserInactive, ErrUserInactive},
	{domainerrs.ErrUserIDRequired, ErrInvalidUserID},
	{domainerrs.ErrRefreshTokenNotFound, ErrRefreshTokenInvalid},
	{domainerrs.ErrRefreshTokenInvalid, ErrRefreshTokenInvalid},
	{domainerrs.ErrConcurrencyConflict, ErrOptimisticLockFailed},
	{domainerrs.ErrInvalidPolicy, ErrInvalidPolicy},
	{domainerrs.ErrInvalidRoleAssignment, ErrInvalidRoleAssignment},
}

func MapDomainError(err error) error {
	for _, mapping := range domainToHTTPErrorMap {
		if errors.Is(err, mapping.domainError) {
			return mapping.httpError
		}
	}
	return err
}
