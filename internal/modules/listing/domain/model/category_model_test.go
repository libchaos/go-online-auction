package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
)

func TestNewCategoryModel(t *testing.T) {
	t.Run("valid input returns category", func(t *testing.T) {
		// Act
		category, err := model.NewCategoryModel("  数码  ", nil, 10)

		// Assert
		require.NoError(t, err)
		require.Equal(t, "数码", category.Name())
		require.Nil(t, category.ParentID())
		require.Equal(t, int32(10), category.SortOrder())
		require.Equal(t, uint64(1), category.Version())
	})

	t.Run("valid input with parent returns category", func(t *testing.T) {
		// Arrange
		parentID := uint64(1)

		// Act
		category, err := model.NewCategoryModel("手机", &parentID, 0)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, category.ParentID())
		require.Equal(t, uint64(1), *category.ParentID())
	})

	t.Run("blank name returns error", func(t *testing.T) {
		// Act
		_, err := model.NewCategoryModel("   ", nil, 0)

		// Assert
		require.ErrorIs(t, err, errs.ErrCategoryNameRequired)
	})
}

func TestRestoreCategoryModel(t *testing.T) {
	now := time.Now().UTC()

	t.Run("valid input returns category", func(t *testing.T) {
		// Act
		category, err := model.RestoreCategoryModel(1, "数码", nil, 0, 2, now, now)

		// Assert
		require.NoError(t, err)
		require.Equal(t, uint64(1), category.ID())
		require.Equal(t, uint64(2), category.Version())
	})

	t.Run("zero id returns error", func(t *testing.T) {
		// Act
		_, err := model.RestoreCategoryModel(0, "数码", nil, 0, 1, now, now)

		// Assert
		require.ErrorIs(t, err, errs.ErrCategoryIDRequired)
	})
}

func TestCategoryModel_Update(t *testing.T) {
	t.Run("valid input updates name and sort order", func(t *testing.T) {
		// Arrange
		category, err := model.NewCategoryModel("数码", nil, 0)
		require.NoError(t, err)

		// Act
		err = category.Update("电子产品", 5)

		// Assert
		require.NoError(t, err)
		require.Equal(t, "电子产品", category.Name())
		require.Equal(t, int32(5), category.SortOrder())
	})

	t.Run("blank name returns error", func(t *testing.T) {
		// Arrange
		category, err := model.NewCategoryModel("数码", nil, 0)
		require.NoError(t, err)

		// Act
		err = category.Update("", 0)

		// Assert
		require.ErrorIs(t, err, errs.ErrCategoryNameRequired)
		require.Equal(t, "数码", category.Name())
	})
}
