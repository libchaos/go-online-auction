package mapper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/mapper"
	"auction/internal/modules/listing/infra/sqlcgen"
)

func TestSpuMapper_RoundTrip(t *testing.T) {
	sut := mapper.NewSpuMapper()
	now := time.Now().UTC()

	t.Run("images survive create params round trip", func(t *testing.T) {
		// Arrange
		brand := "Apple"
		images := []string{"https://cdn.example.com/1.jpg", "https://cdn.example.com/2.jpg"}
		spu, err := model.NewSpuModel("iPhone 15", "旗舰手机", 1, &brand, images)
		require.NoError(t, err)

		// Act
		params, err := sut.ToCreateParams(spu)
		require.NoError(t, err)

		restored, err := sut.ToDomain(sqlcgen.Spu{
			ID:          1,
			Title:       params.Title,
			Description: params.Description,
			CategoryID:  params.CategoryID,
			Brand:       params.Brand,
			Images:      params.Images,
			Status:      params.Status,
			Version:     params.Version,
			CreatedAt:   now,
			UpdatedAt:   now,
		})

		// Assert
		require.NoError(t, err)
		require.Equal(t, images, restored.Images())
		require.Equal(t, "iPhone 15", restored.Title())
		require.Equal(t, "Apple", *restored.Brand())
		status := restored.Status()
		require.True(t, status.IsDraft())
	})

	t.Run("nil images marshal to empty json array", func(t *testing.T) {
		// Arrange
		spu, err := model.NewSpuModel("iPhone 15", "", 1, nil, nil)
		require.NoError(t, err)

		// Act
		params, err := sut.ToCreateParams(spu)

		// Assert
		require.NoError(t, err)
		require.JSONEq(t, "[]", string(params.Images))
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		// Act
		_, err := sut.ToDomain(sqlcgen.Spu{
			ID:         1,
			Title:      "iPhone 15",
			CategoryID: 1,
			Status:     sqlcgen.ListingStatus("archived"),
			Version:    1,
			CreatedAt:  now,
			UpdatedAt:  now,
		})

		// Assert
		require.Error(t, err)
	})
}

func TestSkuMapper_RoundTrip(t *testing.T) {
	sut := mapper.NewSkuMapper()
	now := time.Now().UTC()

	t.Run("chinese spec keys survive round trip", func(t *testing.T) {
		// Arrange
		specValues := map[string]string{"颜色": "红", "尺寸": "L"}
		sku, err := model.NewSkuModel(1, specValues, 19900, 5)
		require.NoError(t, err)

		// Act
		params, err := sut.ToCreateParams(sku)
		require.NoError(t, err)

		restored, err := sut.ToDomain(sqlcgen.Sku{
			ID:           10,
			SpuID:        params.SpuID,
			SpecValues:   params.SpecValues,
			PriceInCents: params.PriceInCents,
			Quantity:     params.Quantity,
			Status:       params.Status,
			Version:      params.Version,
			CreatedAt:    now,
			UpdatedAt:    now,
		})

		// Assert
		require.NoError(t, err)
		require.Equal(t, specValues, restored.SpecValues())
		require.Equal(t, uint64(19900), restored.PriceInCents())
		require.Equal(t, uint64(5), restored.Quantity())
	})

	t.Run("update params carry incremented version", func(t *testing.T) {
		// Arrange
		publishedStatus, _ := enum.NewListingStatusEnum(enum.EnumListingStatusPublished)
		sku, err := model.RestoreSkuModel(
			10, 1, map[string]string{"颜色": "红"}, 19900, 5, publishedStatus, 3, now, now,
		)
		require.NoError(t, err)

		// Act
		params, err := sut.ToUpdateParams(sku)

		// Assert
		require.NoError(t, err)
		require.Equal(t, int64(3), params.Version)
		require.Equal(t, int64(10), params.ID)
		require.Equal(t, sqlcgen.ListingStatus("published"), params.Status)
	})
}
