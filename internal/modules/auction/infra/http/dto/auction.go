package dto

import "time"

type CreateAuctionRequest struct {
	ListingID uint64    `json:"listing_id"`
	EndTime   time.Time `json:"end_time"`
}

type AuctionResponse struct {
	ID                      uint64    `json:"id"`
	ListingID               uint64    `json:"listing_id"`
	State                   string    `json:"state"`
	StartTime               time.Time `json:"start_time"`
	EndTime                 time.Time `json:"end_time"`
	HighestBidID            *uint64   `json:"highest_bid_id,omitempty"`
	HighestBidAmountInCents *uint64   `json:"highest_bid_amount_in_cents,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
}

type AuctionDetailResponse struct {
	Auction AuctionResponse `json:"auction"`
	Bids    []BidResponse   `json:"bids"`
}

type AuctionListResponse struct {
	Auctions   []AuctionResponse `json:"auctions"`
	TotalCount uint64            `json:"total_count"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}
