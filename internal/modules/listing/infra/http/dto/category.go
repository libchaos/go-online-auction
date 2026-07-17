package dto

import "time"

type CreateCategoryRequest struct {
	Name      string  `json:"name"`
	ParentID  *uint64 `json:"parent_id,omitempty"`
	SortOrder int32   `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      string `json:"name"`
	SortOrder int32  `json:"sort_order"`
}

type CategoryResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *uint64   `json:"parent_id,omitempty"`
	Depth     int32     `json:"depth"`
	Path      string    `json:"path,omitempty"`
	SortOrder int32     `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// CategoryTreeNode is a category plus its nested children, forming a multi-level
// tree. Roots is the list of top-level nodes returned by the tree endpoints.
type CategoryTreeNode struct {
	ID        uint64              `json:"id"`
	Name      string              `json:"name"`
	ParentID  *uint64             `json:"parent_id,omitempty"`
	Depth     int32               `json:"depth"`
	Path      string              `json:"path,omitempty"`
	SortOrder int32               `json:"sort_order"`
	Children  []*CategoryTreeNode `json:"children,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type CategoryTreeResponse struct {
	Roots []*CategoryTreeNode `json:"roots"`
}
