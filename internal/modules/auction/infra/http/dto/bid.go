package dto

import "time"

type PlaceBidRequest struct {
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
}

type BidResponse struct {
	ID            uint64    `json:"id"`
	AuctionID     uint64    `json:"auction_id"`
	UserID        uint64    `json:"user_id"`
	AmountInCents uint64    `json:"amount_in_cents"`
	CreatedAt     time.Time `json:"created_at"`
}
