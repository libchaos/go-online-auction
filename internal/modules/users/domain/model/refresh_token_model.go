package model

import (
	"time"

	"auction/internal/modules/users/domain/errs"
)

type RefreshTokenModel struct {
	id         uint64
	userID     uint64
	tokenHash  string
	expiresAt  time.Time
	revokedAt  *time.Time
	replacedBy *uint64
	createdAt  time.Time
}

func NewRefreshTokenModel(userID uint64, tokenHash string, expiresAt time.Time) (RefreshTokenModel, error) {
	if userID == 0 {
		return RefreshTokenModel{}, errs.ErrUserIDRequired
	}

	if tokenHash == "" {
		return RefreshTokenModel{}, errs.ErrTokenHashRequired
	}

	return RefreshTokenModel{
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		createdAt: time.Now().UTC(),
	}, nil
}

func RestoreRefreshTokenModel(
	id, userID uint64,
	tokenHash string,
	expiresAt time.Time,
	revokedAt *time.Time,
	replacedBy *uint64,
	createdAt time.Time,
) (RefreshTokenModel, error) {
	if id == 0 {
		return RefreshTokenModel{}, errs.ErrRefreshTokenNotFound
	}

	if userID == 0 {
		return RefreshTokenModel{}, errs.ErrUserIDRequired
	}

	if tokenHash == "" {
		return RefreshTokenModel{}, errs.ErrTokenHashRequired
	}

	return RefreshTokenModel{
		id:         id,
		userID:     userID,
		tokenHash:  tokenHash,
		expiresAt:  expiresAt,
		revokedAt:  revokedAt,
		replacedBy: replacedBy,
		createdAt:  createdAt,
	}, nil
}

func (m *RefreshTokenModel) ID() uint64 {
	return m.id
}

func (m *RefreshTokenModel) UserID() uint64 {
	return m.userID
}

func (m *RefreshTokenModel) TokenHash() string {
	return m.tokenHash
}

func (m *RefreshTokenModel) ExpiresAt() time.Time {
	return m.expiresAt
}

func (m *RefreshTokenModel) RevokedAt() *time.Time {
	return m.revokedAt
}

func (m *RefreshTokenModel) ReplacedBy() *uint64 {
	return m.replacedBy
}

func (m *RefreshTokenModel) CreatedAt() time.Time {
	return m.createdAt
}

func (m *RefreshTokenModel) IsRevoked() bool {
	return m.revokedAt != nil
}

func (m *RefreshTokenModel) IsExpired(now time.Time) bool {
	return now.After(m.expiresAt)
}

// IsValid reports whether the token can still be used
func (m *RefreshTokenModel) IsValid(now time.Time) bool {
	return !m.IsRevoked() && !m.IsExpired(now)
}

// Revoke marks the token as revoked, optionally linking its replacement
func (m *RefreshTokenModel) Revoke(replacedBy *uint64) {
	now := time.Now().UTC()
	m.revokedAt = &now
	m.replacedBy = replacedBy
}
