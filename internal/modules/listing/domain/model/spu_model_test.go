package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
)

func newDraftSpu(t *testing.T) model.SpuModel {
	t.Helper()
	spu, err := model.NewSpuModel("iPhone 15", "旗舰手机", 1, nil, []string{"https://cdn.example.com/1.jpg"})
	require.NoError(t, err)
	return spu
}

func restoreSpuWithStatus(t *testing.T, status string) model.SpuModel {
	t.Helper()
	statusEnum, err := enum.NewListingStatusEnum(status)
	require.NoError(t, err)
	now := time.Now().UTC()
	spu, err := model.RestoreSpuModel(1, "iPhone 15", "旗舰手机", 1, nil, nil, statusEnum, 1, now, now)
	require.NoError(t, err)
	return spu
}

func TestNewSpuModel(t *testing.T) {
	t.Run("valid input returns draft spu", func(t *testing.T) {
		// Arrange
		brand := "Apple"

		// Act
		spu, err := model.NewSpuModel(
			"  iPhone 15  ", "旗舰手机", 1, &brand,
			[]string{"https://cdn.example.com/1.jpg", "http://cdn.example.com/2.jpg"},
		)

		// Assert
		require.NoError(t, err)
		require.Equal(t, "iPhone 15", spu.Title())
		require.Equal(t, "旗舰手机", spu.Description())
		require.Equal(t, uint64(1), spu.CategoryID())
		require.Equal(t, "Apple", *spu.Brand())
		require.Len(t, spu.Images(), 2)
		status := spu.Status()
		require.True(t, status.IsDraft())
		require.Equal(t, uint64(1), spu.Version())
	})

	t.Run("blank title returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSpuModel("  ", "", 1, nil, nil)

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuTitleRequired)
	})

	t.Run("zero category id returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSpuModel("iPhone 15", "", 0, nil, nil)

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuCategoryRequired)
	})

	t.Run("invalid image url returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSpuModel("iPhone 15", "", 1, nil, []string{"not-a-url"})

		// Assert
		require.ErrorIs(t, err, errs.ErrImageURLInvalid)
	})

	t.Run("non http scheme returns error", func(t *testing.T) {
		// Act
		_, err := model.NewSpuModel("iPhone 15", "", 1, nil, []string{"ftp://cdn.example.com/1.jpg"})

		// Assert
		require.ErrorIs(t, err, errs.ErrImageURLInvalid)
	})
}

func TestRestoreSpuModel(t *testing.T) {
	status, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
	now := time.Now().UTC()

	t.Run("valid input returns spu", func(t *testing.T) {
		// Act
		spu, err := model.RestoreSpuModel(1, "iPhone 15", "", 1, nil, nil, status, 3, now, now)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), spu.ID())
		require.Equal(t, uint64(3), spu.Version())
		spuStatus := spu.Status()
		require.True(t, spuStatus.IsPublished())
	})

	t.Run("zero id returns error", func(t *testing.T) {
		// Act
		_, err := model.RestoreSpuModel(0, "iPhone 15", "", 1, nil, nil, status, 1, now, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuIDRequired)
	})
}

func TestSpuModel_Publish(t *testing.T) {
	t.Run("draft spu publishes", func(t *testing.T) {
		// Arrange
		spu := newDraftSpu(t)

		// Act
		err := spu.Publish()

		// Assert
		require.NoError(t, err)
		status := spu.Status()
		require.True(t, status.IsPublished())
	})

	t.Run("off_shelf spu republishes", func(t *testing.T) {
		// Arrange
		spu := restoreSpuWithStatus(t, enum.EnumListingStatusOffShelf)

		// Act
		err := spu.Publish()

		// Assert
		require.NoError(t, err)
		status := spu.Status()
		require.True(t, status.IsPublished())
	})

	t.Run("published spu returns error", func(t *testing.T) {
		// Arrange
		spu := restoreSpuWithStatus(t, enum.EnumListingStatusPublished)

		// Act
		err := spu.Publish()

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuAlreadyPublished)
	})
}

func TestSpuModel_OffShelf(t *testing.T) {
	t.Run("published spu goes off shelf", func(t *testing.T) {
		// Arrange
		spu := restoreSpuWithStatus(t, enum.EnumListingStatusPublished)

		// Act
		err := spu.OffShelf()

		// Assert
		require.NoError(t, err)
		status := spu.Status()
		require.True(t, status.IsOffShelf())
	})

	t.Run("draft spu returns error", func(t *testing.T) {
		// Arrange
		spu := newDraftSpu(t)

		// Act
		err := spu.OffShelf()

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuNotPublished)
	})

	t.Run("off_shelf spu returns error", func(t *testing.T) {
		// Arrange
		spu := restoreSpuWithStatus(t, enum.EnumListingStatusOffShelf)

		// Act
		err := spu.OffShelf()

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuNotPublished)
	})
}

func TestSpuModel_Update(t *testing.T) {
	t.Run("draft spu updates", func(t *testing.T) {
		// Arrange
		spu := newDraftSpu(t)
		brand := "Apple"

		// Act
		err := spu.Update("iPhone 15 Pro", "更强", 2, &brand, []string{"https://cdn.example.com/3.jpg"})

		// Assert
		require.NoError(t, err)
		require.Equal(t, "iPhone 15 Pro", spu.Title())
		require.Equal(t, uint64(2), spu.CategoryID())
	})

	t.Run("off_shelf spu updates", func(t *testing.T) {
		// Arrange
		spu := restoreSpuWithStatus(t, enum.EnumListingStatusOffShelf)

		// Act
		err := spu.Update("iPhone 15 Pro", "", 1, nil, nil)

		// Assert
		require.NoError(t, err)
	})

	t.Run("published spu returns error", func(t *testing.T) {
		// Arrange
		spu := restoreSpuWithStatus(t, enum.EnumListingStatusPublished)

		// Act
		err := spu.Update("iPhone 15 Pro", "", 1, nil, nil)

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuNotEditable)
		require.Equal(t, "iPhone 15", spu.Title())
	})

	t.Run("blank title returns error", func(t *testing.T) {
		// Arrange
		spu := newDraftSpu(t)

		// Act
		err := spu.Update("", "", 1, nil, nil)

		// Assert
		require.ErrorIs(t, err, errs.ErrSpuTitleRequired)
	})
}
