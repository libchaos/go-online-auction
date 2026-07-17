package ports

import (
	"context"

	"auction/internal/modules/listing/domain/model"
)

type CategoryRepository interface {
	Create(ctx context.Context, category model.CategoryModel) (model.CategoryModel, error)
	FindByID(ctx context.Context, id uint64) (model.CategoryModel, error)
	Update(ctx context.Context, category model.CategoryModel) error
	Delete(ctx context.Context, id uint64) error
	// List returns categories filtered by parent; a nil parentID returns root categories
	List(ctx context.Context, parentID *uint64) ([]model.CategoryModel, error)
	// CountChildren returns the number of direct child categories
	CountChildren(ctx context.Context, id uint64) (uint64, error)
	// CountSpusByCategory returns the number of SPUs referencing the category
	CountSpusByCategory(ctx context.Context, id uint64) (uint64, error)
}
