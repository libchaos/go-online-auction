package errs

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrNameRequired          = errors.New("name must have at least 2 characters")
	ErrPasswordTooShort      = errors.New("password must have at least 8 characters")
	ErrPasswordTooLong       = errors.New("password must have at most 72 bytes")
	ErrPasswordHashRequired  = errors.New("password hash is required")
	ErrInvalidRole           = errors.New("invalid role")
	ErrInvalidUserStatus     = errors.New("invalid user status")
	ErrUserInactive          = errors.New("user is not active")
	ErrUserIDRequired        = errors.New("user ID must be greater than zero")
	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenInvalid   = errors.New("refresh token is invalid")
	ErrTokenHashRequired     = errors.New("token hash is required")
	ErrConcurrencyConflict   = errors.New("concurrency conflict: resource was modified by another transaction")
	ErrInvalidPolicy         = errors.New("invalid policy: sub, obj and act are required")
	ErrInvalidRoleAssignment = errors.New(
		"invalid role assignment: user id must be greater than zero and role must be one of admin, seller, bidder",
	)
)
