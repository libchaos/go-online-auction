package enum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/errs"
)

func TestNewUserStatusEnum(t *testing.T) {
	t.Run("valid values return enum", func(t *testing.T) {
		// Arrange
		validValues := []string{
			enum.EnumUserStatusActive,
			enum.EnumUserStatusInactive,
			enum.EnumUserStatusBlocked,
		}

		for _, value := range validValues {
			// Act
			status, err := enum.NewUserStatusEnum(value)

			// Assert
			require.NoError(t, err)
			require.Equal(t, value, status.String())
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		// Act
		_, err := enum.NewUserStatusEnum("deleted")

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidUserStatus)
	})
}
