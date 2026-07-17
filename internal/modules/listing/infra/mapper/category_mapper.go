package mapper

import (
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/infra/sqlcgen"
)

type CategoryMapper struct{}

func NewCategoryMapper() *CategoryMapper {
	return &CategoryMapper{}
}

func (m *CategoryMapper) ToDomain(c sqlcgen.Category) (model.CategoryModel, error) {
	return model.RestoreCategoryModel(
		uint64(c.ID),
		c.Name,
		toNullableUint64(c.ParentID),
		c.SortOrder,
		uint64(c.Version),
		c.CreatedAt,
		c.UpdatedAt,
	)
}

func (m *CategoryMapper) ToCreateParams(category model.CategoryModel) sqlcgen.CreateCategoryParams {
	return sqlcgen.CreateCategoryParams{
		Name:      category.Name(),
		ParentID:  toNullableInt64(category.ParentID()),
		SortOrder: category.SortOrder(),
		Version:   int64(category.Version()),
		CreatedAt: category.CreatedAt(),
		UpdatedAt: category.UpdatedAt(),
	}
}

func (m *CategoryMapper) ToUpdateParams(category model.CategoryModel) sqlcgen.UpdateCategoryParams {
	return sqlcgen.UpdateCategoryParams{
		Name:      category.Name(),
		SortOrder: category.SortOrder(),
		Version:   int64(category.Version()),
		UpdatedAt: category.UpdatedAt(),
		ID:        int64(category.ID()),
	}
}

func toNullableUint64(v *int64) *uint64 {
	if v == nil {
		return nil
	}
	u := uint64(*v)
	return &u
}

func toNullableInt64(v *uint64) *int64 {
	if v == nil {
		return nil
	}
	i := int64(*v)
	return &i
}
