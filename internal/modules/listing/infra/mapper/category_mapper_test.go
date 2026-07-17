package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/mapper"
	"auction/internal/modules/listing/infra/sqlcgen"
)

func TestCategoryMapper_RoundTrip(t *testing.T) {
	sut := mapper.NewCategoryMapper()
	now := time.Now().UTC()

	t.Run("root category round trip", func(t *testing.T) {
		// Arrange
		category, err := model.NewCategoryModel("数码", nil, 10)
		require.NoError(t, err)

		// Act
		params := sut.ToCreateParams(category)
		restored, err := sut.ToDomain(sqlcgen.Category{
			ID:        1,
			Name:      params.Name,
			ParentID:  params.ParentID,
			SortOrder: params.SortOrder,
			Version:   params.Version,
			CreatedAt: now,
			UpdatedAt: now,
		})

		// Assert
		require.NoError(t, err)
		require.Equal(t, "数码", restored.Name())
		require.Nil(t, restored.ParentID())
		require.Equal(t, int32(10), restored.SortOrder())
	})

	t.Run("child category keeps parent id", func(t *testing.T) {
		// Arrange
		parentID := uint64(1)
		category, err := model.NewCategoryModel("手机", &parentID, 0)
		require.NoError(t, err)

		// Act
		params := sut.ToCreateParams(category)

		// Assert
		require.NotNil(t, params.ParentID)
		require.Equal(t, int64(1), *params.ParentID)
	})
}
