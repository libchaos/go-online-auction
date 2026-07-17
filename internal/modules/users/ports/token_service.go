package ports

import "time"

type TokenService interface {
	GenerateAccessToken(userID uint64, role, email string) (token string, expiresAt time.Time, err error)
	GenerateRefreshToken() (raw, hash string, expiresAt time.Time, err error)
	HashRefreshToken(raw string) string
}
