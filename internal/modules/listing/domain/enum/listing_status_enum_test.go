package enum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
)

func TestNewListingStatusEnum(t *testing.T) {
	t.Run("valid values return enum", func(t *testing.T) {
		// Arrange
		validValues := []string{
			enum.EnumListingStatusDraft,
			enum.EnumListingStatusPublished,
			enum.EnumListingStatusOffShelf,
		}

		for _, value := range validValues {
			// Act
			status, err := enum.NewListingStatusEnum(value)

			// Assert
			require.NoError(t, err)
			require.Equal(t, value, status.String())
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		// Act
		_, err := enum.NewListingStatusEnum("archived")

		// Assert
		require.ErrorIs(t, err, errs.ErrInvalidListingStatus)
	})
}

func TestListingStatusEnum_Predicates(t *testing.T) {
	t.Run("draft predicates", func(t *testing.T) {
		// Arrange
		status, err := enum.NewListingStatusEnum(enum.EnumListingStatusDraft)
		require.NoError(t, err)

		// Assert
		require.True(t, status.IsDraft())
		require.False(t, status.IsPublished())
		require.False(t, status.IsOffShelf())
	})

	t.Run("published predicates", func(t *testing.T) {
		// Arrange
		status, err := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
		require.NoError(t, err)

		// Assert
		require.False(t, status.IsDraft())
		require.True(t, status.IsPublished())
		require.False(t, status.IsOffShelf())
	})

	t.Run("off_shelf predicates", func(t *testing.T) {
		// Arrange
		status, err := enum.NewListingStatusEnum(enum.EnumListingStatusOffShelf)
		require.NoError(t, err)

		// Assert
		require.False(t, status.IsDraft())
		require.False(t, status.IsPublished())
		require.True(t, status.IsOffShelf())
	})
}
