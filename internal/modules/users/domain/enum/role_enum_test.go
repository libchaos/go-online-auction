package enum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
)

func TestNewRoleEnum(t *testing.T) {
	t.Run("valid values return enum", func(t *testing.T) {
		// Arrange
		validValues := []string{
			enum.EnumRoleAdmin,
			enum.EnumRoleSeller,
			enum.EnumRoleBidder,
		}

		for _, value := range validValues {
			// Act
			role, err := enum.NewRoleEnum(value)

			// Assert
			require.NoError(t, err)
			require.Equal(t, value, role.String())
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		// Act
		_, err := enum.NewRoleEnum("superuser")

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidRole)
	})

	t.Run("empty value returns error", func(t *testing.T) {
		// Act
		_, err := enum.NewRoleEnum("")

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidRole)
	})
}
