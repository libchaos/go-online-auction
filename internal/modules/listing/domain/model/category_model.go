package model

import (
	"strings"
	"time"

	"auction/internal/modules/listing/domain/errs"
)

// CategoryModel is a node in the category adjacency-list tree; a nil parentID
// marks a root category.
type CategoryModel struct {
	id        uint64
	name      string
	parentID  *uint64
	sortOrder int32
	version   uint64
	createdAt time.Time
	updatedAt time.Time
}

func NewCategoryModel(name string, parentID *uint64, sortOrder int32) (CategoryModel, error) {
	if err := validateCategoryName(name); err != nil {
		return CategoryModel{}, err
	}

	now := time.Now().UTC()
	return CategoryModel{
		name:      strings.TrimSpace(name),
		parentID:  parentID,
		sortOrder: sortOrder,
		version:   1,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreCategoryModel(
	id uint64,
	name string,
	parentID *uint64,
	sortOrder int32,
	version uint64,
	createdAt, updatedAt time.Time,
) (CategoryModel, error) {
	if id == 0 {
		return CategoryModel{}, errs.ErrCategoryIDRequired
	}

	if err := validateCategoryName(name); err != nil {
		return CategoryModel{}, err
	}

	return CategoryModel{
		id:        id,
		name:      name,
		parentID:  parentID,
		sortOrder: sortOrder,
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

func (c *CategoryModel) ID() uint64 {
	return c.id
}

func (c *CategoryModel) Name() string {
	return c.name
}

func (c *CategoryModel) ParentID() *uint64 {
	return c.parentID
}

func (c *CategoryModel) SortOrder() int32 {
	return c.sortOrder
}

func (c *CategoryModel) Version() uint64 {
	return c.version
}

func (c *CategoryModel) CreatedAt() time.Time {
	return c.createdAt
}

func (c *CategoryModel) UpdatedAt() time.Time {
	return c.updatedAt
}

// Update changes the category name and sort order
func (c *CategoryModel) Update(name string, sortOrder int32) error {
	if err := validateCategoryName(name); err != nil {
		return err
	}

	c.name = strings.TrimSpace(name)
	c.sortOrder = sortOrder
	c.touch()
	return nil
}

func (c *CategoryModel) touch() {
	c.version++
	c.updatedAt = time.Now().UTC()
}

func validateCategoryName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errs.ErrCategoryNameRequired
	}

	return nil
}
