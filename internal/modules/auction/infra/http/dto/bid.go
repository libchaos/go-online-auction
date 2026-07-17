package dto

import "time"

type PlaceBidRequest struct {
	AmountInCents    uint64  `json:"amount_in_cents"`
	MaxAmountInCents *uint64 `json:"max_amount_in_cents,omitempty"`
}

type PlaceBidAcceptedResponse struct {
	IdempotencyKey string `json:"idempotency_key"`
	Status         string `json:"status"`
}

type BidResponse struct {
	ID               uint64    `json:"id"`
	AuctionID        uint64    `json:"auction_id"`
	UserID           uint64    `json:"user_id"`
	AmountInCents    uint64    `json:"amount_in_cents"`
	MaxAmountInCents *uint64   `json:"max_amount_in_cents,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
