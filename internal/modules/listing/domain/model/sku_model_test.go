package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
)

var validSpecValues = map[string]string{"颜色": "红", "尺寸": "L"}

func newDraftSku(t *testing.T) model.SkuModel {
	t.Helper()
	sku, err := model.NewSkuModel(1, validSpecValues, 19900, 5)
	require.NoError(t, err)
	return sku
}

func restoreSkuWithStatus(t *testing.T, status string, quantity uint64) model.SkuModel {
	t.Helper()
	statusEnum, err := enum.NewListingStatusEnum(status)
	require.NoError(t, err)
	now := time.Now().UTC()
	sku, err := model.RestoreSkuModel(1, 1, validSpecValues, 19900, quantity, statusEnum, 1, now, now)
	require.NoError(t, err)
	return sku
}

func TestNewSkuModel(t *testing.T) {
	t.Run("valid input returns draft sku", func(t *testing.T) {
		// Act
		sku, err := model.NewSkuModel(1, validSpecValues, 19900, 5)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), sku.SpuID())
		require.Equal(t, validSpecValues, sku.SpecValues())
		require.Equal(t, uint64(19900), sku.PriceInCents())
		require.Equal(t, uint64(5), sku.Quantity())
		status := sku.Status()
		require.True(t, status.IsDraft())
		require.Equal(t, uint64(1), sku.Version())
	})

	t.Run("zero spu id returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSkuModel(0, validSpecValues, 19900, 5)

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuIDRequired)
	})

	t.Run("empty spec values returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSkuModel(1, map[string]string{}, 19900, 5)

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuSpecValuesRequired)
	})

	t.Run("zero price returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSkuModel(1, validSpecValues, 0, 5)

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuPriceInvalid)
	})
}

func TestRestoreSkuModel(t *testing.T) {
	status, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	now := time.Now().UTC()

	t.Run("valid input returns sku", func(t *testing.T) {
		// Act
		sku, err := model.RestoreSkuModel(1, 2, validSpecValues, 19900, 5, status, 3, now, now)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), sku.ID())
		require.Equal(t, uint64(2), sku.SpuID())
		require.Equal(t, uint64(3), sku.Version())
	})

	t.Run("zero id returns error", func(t *testing.T) {
		// Act
		_, err := model.RestoreSkuModel(0, 2, validSpecValues, 19900, 5, status, 1, now, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuIDRequired)
	})
}

func TestSkuModel_IsAuctionable(t *testing.T) {
	t.Run("published with quantity is auctionable", func(t *testing.T) {
		// Arrange
		sku := restoreSkuWithStatus(t, enum.EnumListingStatusPublished, 5)

		// Assert
		require.True(t, sku.IsAuctionable())
	})

	t.Run("published without quantity is not auctionable", func(t *testing.T) {
		// Arrange
		sku := restoreSkuWithStatus(t, enum.EnumListingStatusPublished, 0)

		// Assert
		require.False(t, sku.IsAuctionable())
	})

	t.Run("draft is not auctionable", func(t *testing.T) {
		// Arrange
		sku := newDraftSku(t)

		// Assert
		require.False(t, sku.IsAuctionable())
	})
}

func TestSkuModel_Publish(t *testing.T) {
	t.Run("draft sku publishes", func(t *testing.T) {
		// Arrange
		sku := newDraftSku(t)

		// Act
		err := sku.Publish()

		// Assert
		require.NoError(t, err)
		status := sku.Status()
		require.True(t, status.IsPublished())
	})

	t.Run("off_shelf sku republishes", func(t *testing.T) {
		// Arrange
		sku := restoreSkuWithStatus(t, enum.EnumListingStatusOffShelf, 5)

		// Act
		err := sku.Publish()

		// Assert
		require.NoError(t, err)
		status := sku.Status()
		require.True(t, status.IsPublished())
	})

	t.Run("published sku returns error", func(t *testing.T) {
		// Arrange
		sku := restoreSkuWithStatus(t, enum.EnumListingStatusPublished, 5)

		// Act
		err := sku.Publish()

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuAlreadyPublished)
	})
}

func TestSkuModel_OffShelf(t *testing.T) {
	t.Run("published sku goes off shelf", func(t *testing.T) {
		// Arrange
		sku := restoreSkuWithStatus(t, enum.EnumListingStatusPublished, 5)

		// Act
		err := sku.OffShelf()

		// Assert
		require.NoError(t, err)
		status := sku.Status()
		require.True(t, status.IsOffShelf())
	})

	t.Run("draft sku returns error", func(t *testing.T) {
		// Arrange
		sku := newDraftSku(t)

		// Act
		err := sku.OffShelf()

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuNotPublished)
	})
}

func TestSkuModel_Update(t *testing.T) {
	t.Run("draft sku updates", func(t *testing.T) {
		// Arrange
		sku := newDraftSku(t)
		newSpecs := map[string]string{"颜色": "蓝"}

		// Act
		err := sku.Update(newSpecs, 29900, 10)

		// Assert
		require.NoError(t, err)
		require.Equal(t, newSpecs, sku.SpecValues())
		require.Equal(t, uint64(29900), sku.PriceInCents())
		require.Equal(t, uint64(10), sku.Quantity())
	})

	t.Run("published sku returns error", func(t *testing.T) {
		// Arrange
		sku := restoreSkuWithStatus(t, enum.EnumListingStatusPublished, 5)

		// Act
		err := sku.Update(validSpecValues, 29900, 10)

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuNotEditable)
	})

	t.Run("zero price returns error", func(t *testing.T) {
		// Arrange
		sku := newDraftSku(t)

		// Act
		err := sku.Update(validSpecValues, 0, 10)

		// Assert
		require.ErrorIs(t, err, errs.ErrSkuPriceInvalid)
	})
}
