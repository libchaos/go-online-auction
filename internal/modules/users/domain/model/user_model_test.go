package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
)

func TestNewUserModel(t *testing.T) {
	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)

	t.Run("valid input returns user", func(t *testing.T) {
		// Act
		user, err := model.NewUserModel("John Doe", "John@Example.com", "hashed-password", role)

		// Assert
		require.NoError(t, err)
		require.Equal(t, "John Doe", user.Name())
		require.Equal(t, "john@example.com", user.Email())
		require.NotNil(t, user.PasswordHash())
		require.Equal(t, "hashed-password", *user.PasswordHash())
		require.Equal(t, uint64(1), user.Version())
		require.True(t, user.IsActive())
	})

	t.Run("invalid email returns error", func(t *testing.T) {
		// Act
		_, err := model.NewUserModel("John Doe", "not-an-email", "hashed-password", role)

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidEmail)
	})

	t.Run("short name returns error", func(t *testing.T) {
		// Act
		_, err := model.NewUserModel("J", "john@example.com", "hashed-password", role)

		// Assert
		require.ErrorIs(t, err, errs.ErrNameRequired)
	})

	t.Run("empty password hash returns error", func(t *testing.T) {
		// Act
		_, err := model.NewUserModel("John Doe", "john@example.com", "", role)

		// Assert
		require.ErrorIs(t, err, errs.ErrPasswordHashRequired)
	})
}

func TestRestoreUserModel(t *testing.T) {
	role, _ := enum.NewRoleEnum(enum.EnumRoleSeller)
	status, _ := enum.NewUserStatusEnum(enum.EnumUserStatusActive)
	now := time.Now().UTC()

	t.Run("valid input returns user", func(t *testing.T) {
		// Arrange
		passwordHash := "hashed-password"

		// Act
		user, err := model.RestoreUserModel(
			1, "John Doe", "john@example.com", &passwordHash, role, status, nil, nil, 2, now, now,
		)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), user.ID())
		require.Equal(t, uint64(2), user.Version())
	})

	t.Run("nil password hash is allowed for restore", func(t *testing.T) {
		// Arrange
		provider := "google"
		providerID := "google-123"

		// Act
		user, err := model.RestoreUserModel(
			1, "John Doe", "john@example.com", nil, role, status, &provider, &providerID, 1, now, now,
		)

		// Assert
		require.NoError(t, err)
		require.Nil(t, user.PasswordHash())
	})

	t.Run("zero id returns error", func(t *testing.T) {
		// Act
		_, err := model.RestoreUserModel(
			0, "John Doe", "john@example.com", nil, role, status, nil, nil, 1, now, now,
		)

		// Assert
		require.ErrorIs(t, err, errs.ErrUserIDRequired)
	})
}

func TestUserModel_UpdateProfile(t *testing.T) {
	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)

	t.Run("valid name updates and bumps version", func(t *testing.T) {
		// Arrange
		user, _ := model.NewUserModel("John Doe", "john@example.com", "hash", role)

		// Act
		err := user.UpdateProfile("Jane Doe")

		// Assert
		require.NoError(t, err)
		require.Equal(t, "Jane Doe", user.Name())
		require.Equal(t, uint64(2), user.Version())
	})

	t.Run("short name returns error", func(t *testing.T) {
		// Arrange
		user, _ := model.NewUserModel("John Doe", "john@example.com", "hash", role)

		// Act
		err := user.UpdateProfile(" ")

		// Assert
		require.ErrorIs(t, err, errs.ErrNameRequired)
	})
}

func TestUserModel_ChangePasswordHash(t *testing.T) {
	role, _ := enum.NewRoleEnum(enum.EnumRoleBidder)

	t.Run("valid hash updates and bumps version", func(t *testing.T) {
		// Arrange
		user, _ := model.NewUserModel("John Doe", "john@example.com", "hash", role)

		// Act
		err := user.ChangePasswordHash("new-hash")

		// Assert
		require.NoError(t, err)
		require.Equal(t, "new-hash", *user.PasswordHash())
		require.Equal(t, uint64(2), user.Version())
	})

	t.Run("empty hash returns error", func(t *testing.T) {
		// Arrange
		user, _ := model.NewUserModel("John Doe", "john@example.com", "hash", role)

		// Act
		err := user.ChangePasswordHash("")

		// Assert
		require.ErrorIs(t, err, errs.ErrPasswordHashRequired)
	})
}

func TestUserModel_ChangeRole(t *testing.T) {
	t.Run("role changes and bumps version", func(t *testing.T) {
		// Arrange
		bidderRole, _ := enum.NewRoleEnum(enum.EnumRoleBidder)
		sellerRole, _ := enum.NewRoleEnum(enum.EnumRoleSeller)
		user, _ := model.NewUserModel("John Doe", "john@example.com", "hash", bidderRole)

		// Act
		user.ChangeRole(sellerRole)

		// Assert
		updatedRole := user.Role()
		require.Equal(t, enum.EnumRoleSeller, updatedRole.String())
		require.Equal(t, uint64(2), user.Version())
	})
}
