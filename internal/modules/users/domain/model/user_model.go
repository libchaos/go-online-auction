package model

import (
	"regexp"
	"strings"
	"time"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
)

const minNameLength = 2

var emailRegexp = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type UserModel struct {
	id              uint64
	name            string
	email           string
	passwordHash    *string
	role            enum.RoleEnum
	status          enum.UserStatusEnum
	oauthProvider   *string
	oauthProviderID *string
	version         uint64
	createdAt       time.Time
	updatedAt       time.Time
}

func NewUserModel(name, email, passwordHash string, role enum.RoleEnum) (UserModel, error) {
	normalizedEmail := normalizeEmail(email)
	if err := validateUser(name, normalizedEmail); err != nil {
		return UserModel{}, err
	}

	if passwordHash == "" {
		return UserModel{}, errs.ErrPasswordHashRequired
	}

	status, err := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	if err != nil {
		return UserModel{}, err
	}

	now := time.Now().UTC()
	return UserModel{
		name:         name,
		email:        normalizedEmail,
		passwordHash: &passwordHash,
		role:         role,
		status:       status,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func RestoreUserModel(
	id uint64,
	name, email string,
	passwordHash *string,
	role enum.RoleEnum,
	status enum.UserStatusEnum,
	oauthProvider, oauthProviderID *string,
	version uint64,
	createdAt, updatedAt time.Time,
) (UserModel, error) {
	if id == 0 {
		return UserModel{}, errs.ErrUserIDRequired
	}

	normalizedEmail := normalizeEmail(email)
	if err := validateUser(name, normalizedEmail); err != nil {
		return UserModel{}, err
	}

	return UserModel{
		id:              id,
		name:            name,
		email:           normalizedEmail,
		passwordHash:    passwordHash,
		role:            role,
		status:          status,
		oauthProvider:   oauthProvider,
		oauthProviderID: oauthProviderID,
		version:         version,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}, nil
}

func (u *UserModel) ID() uint64 {
	return u.id
}

func (u *UserModel) Name() string {
	return u.name
}

func (u *UserModel) Email() string {
	return u.email
}

func (u *UserModel) PasswordHash() *string {
	return u.passwordHash
}

func (u *UserModel) Role() enum.RoleEnum {
	return u.role
}

func (u *UserModel) Status() enum.UserStatusEnum {
	return u.status
}

func (u *UserModel) OauthProvider() *string {
	return u.oauthProvider
}

func (u *UserModel) OauthProviderID() *string {
	return u.oauthProviderID
}

func (u *UserModel) Version() uint64 {
	return u.version
}

func (u *UserModel) CreatedAt() time.Time {
	return u.createdAt
}

func (u *UserModel) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *UserModel) IsActive() bool {
	return u.status.String() == enum.EnumUserStatusActive
}

// UpdateProfile updates the user's profile information
func (u *UserModel) UpdateProfile(name string) error {
	if len(strings.TrimSpace(name)) < minNameLength {
		return errs.ErrNameRequired
	}

	u.name = name
	u.touch()
	return nil
}

// ChangePasswordHash replaces the user's password hash
func (u *UserModel) ChangePasswordHash(passwordHash string) error {
	if passwordHash == "" {
		return errs.ErrPasswordHashRequired
	}

	u.passwordHash = &passwordHash
	u.touch()
	return nil
}

// ChangeRole updates the user's role
func (u *UserModel) ChangeRole(role enum.RoleEnum) {
	u.role = role
	u.touch()
}

func (u *UserModel) touch() {
	u.version++
	u.updatedAt = time.Now().UTC()
}

func validateUser(name, email string) error {
	if len(strings.TrimSpace(name)) < minNameLength {
		return errs.ErrNameRequired
	}

	if !emailRegexp.MatchString(email) {
		return errs.ErrInvalidEmail
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
