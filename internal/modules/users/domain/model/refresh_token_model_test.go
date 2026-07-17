package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
)

func TestNewRefreshTokenModel(t *testing.T) {
	t.Run("valid input returns token", func(t *testing.T) {
		// Arrange
		expiresAt := time.Now().UTC().Add(24 * time.Hour)

		// Act
		token, err := model.NewRefreshTokenModel(1, "token-hash", expiresAt)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), token.UserID())
		require.Equal(t, "token-hash", token.TokenHash())
		require.False(t, token.IsRevoked())
	})

	t.Run("zero user id returns error", func(t *testing.T) {
		// Act
		_, err := model.NewRefreshTokenModel(0, "token-hash", time.Now().UTC())

		// Assert
		require.ErrorIs(t, err, errs.ErrUserIDRequired)
	})

	t.Run("empty token hash returns error", func(t *testing.T) {
		// Act
		_, err := model.NewRefreshTokenModel(1, "", time.Now().UTC())

		// Assert
		require.ErrorIs(t, err, errs.ErrTokenHashRequired)
	})
}

func TestRefreshTokenModel_IsValid(t *testing.T) {
	now := time.Now().UTC()

	t.Run("active unexpired token is valid", func(t *testing.T) {
		// Arrange
		token, _ := model.NewRefreshTokenModel(1, "token-hash", now.Add(time.Hour))

		// Act
		valid := token.IsValid(now)

		// Assert
		require.True(t, valid)
	})

	t.Run("expired token is invalid", func(t *testing.T) {
		// Arrange
		token, _ := model.NewRefreshTokenModel(1, "token-hash", now.Add(-time.Hour))

		// Act
		valid := token.IsValid(now)

		// Assert
		require.False(t, valid)
		require.True(t, token.IsExpired(now))
	})

	t.Run("revoked token is invalid", func(t *testing.T) {
		// Arrange
		token, _ := model.NewRefreshTokenModel(1, "token-hash", now.Add(time.Hour))
		replacedBy := uint64(2)

		// Act
		token.Revoke(&replacedBy)

		// Assert
		require.False(t, token.IsValid(now))
		require.True(t, token.IsRevoked())
		require.NotNil(t, token.ReplacedBy())
		require.Equal(t, replacedBy, *token.ReplacedBy())
	})
}

func TestRestoreRefreshTokenModel(t *testing.T) {
	now := time.Now().UTC()

	t.Run("valid input returns token", func(t *testing.T) {
		// Act
		token, err := model.RestoreRefreshTokenModel(1, 2, "token-hash", now.Add(time.Hour), nil, nil, now)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), token.ID())
		require.Equal(t, uint64(2), token.UserID())
	})

	t.Run("zero id returns error", func(t *testing.T) {
		// Act
		_, err := model.RestoreRefreshTokenModel(0, 2, "token-hash", now, nil, nil, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrRefreshTokenNotFound)
	})
}
