package ports

import (
	"context"

	"auction/internal/modules/listing/domain/model"
)

type SkuRepository interface {
	Create(ctx context.Context, sku model.SkuModel) (model.SkuModel, error)
	FindByID(ctx context.Context, id uint64) (model.SkuModel, error)
	FindByIDForUpdate(ctx context.Context, id uint64) (model.SkuModel, error)
	Update(ctx context.Context, sku model.SkuModel) error
	FindBySpuID(ctx context.Context, spuID uint64) ([]model.SkuModel, error)
	// FindPublishedBySpuIDForUpdate locks the published SKUs of an SPU for the
	// off-shelf cascade
	FindPublishedBySpuIDForUpdate(ctx context.Context, spuID uint64) ([]model.SkuModel, error)
}
