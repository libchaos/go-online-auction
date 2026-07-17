package dto

import "time"

type CreateSpuRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CategoryID  uint64   `json:"category_id"`
	Brand       *string  `json:"brand,omitempty"`
	Images      []string `json:"images,omitempty"`
}

type UpdateSpuRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	CategoryID  uint64   `json:"category_id"`
	Brand       *string  `json:"brand,omitempty"`
	Images      []string `json:"images,omitempty"`
}

type SpuResponse struct {
	ID          uint64        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	CategoryID  uint64        `json:"category_id"`
	Brand       *string       `json:"brand,omitempty"`
	Images      []string      `json:"images"`
	Status      string        `json:"status"`
	Skus        []SkuResponse `json:"skus,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type SpuListResponse struct {
	Spus       []SpuResponse `json:"spus"`
	TotalCount uint64        `json:"total_count"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}
