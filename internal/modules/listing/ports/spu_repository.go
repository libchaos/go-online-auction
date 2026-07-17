package ports

import (
	"context"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/domain/model"
)

// ListSpusFilter narrows the SPU listing; nil fields are ignored
type ListSpusFilter struct {
	Status     *enum.ListingStatusEnum
	CategoryID *uint64
	Limit      int
	Offset     int
}

type SpuRepository interface {
	Create(ctx context.Context, spu model.SpuModel) (model.SpuModel, error)
	FindByID(ctx context.Context, id uint64) (model.SpuModel, error)
	FindByIDForUpdate(ctx context.Context, id uint64) (model.SpuModel, error)
	Update(ctx context.Context, spu model.SpuModel) error
	List(ctx context.Context, filter ListSpusFilter) ([]model.SpuModel, error)
	Count(ctx context.Context, filter ListSpusFilter) (uint64, error)
}
