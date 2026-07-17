package dto

import "time"

type CreateSkuRequest struct {
	SpecValues   map[string]string `json:"spec_values"`
	PriceInCents uint64            `json:"price_in_cents"`
	Quantity     uint64            `json:"quantity"`
}

type UpdateSkuRequest struct {
	SpecValues   map[string]string `json:"spec_values"`
	PriceInCents uint64            `json:"price_in_cents"`
	Quantity     uint64            `json:"quantity"`
}

type SkuResponse struct {
	ID           uint64            `json:"id"`
	SpuID        uint64            `json:"spu_id"`
	SpecValues   map[string]string `json:"spec_values"`
	PriceInCents uint64            `json:"price_in_cents"`
	Quantity     uint64            `json:"quantity"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}
