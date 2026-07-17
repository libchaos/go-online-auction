package dto

import (
	"encoding/json"
	"time"
)

type CreateAuctionRequest struct {
	ListingID          uint64     `json:"listing_id"`
	EndTime            time.Time  `json:"end_time"`
	TradingMode        string     `json:"trading_mode"`
	StartingPrice      *uint64    `json:"starting_price,omitempty"`
	PriceStep          *uint64    `json:"price_step,omitempty"`
	ReservePrice       *uint64    `json:"reserve_price,omitempty"`
	AntiSnipeEnabled   bool       `json:"anti_snipe_enabled"`
	ExtensionWindowSec int64      `json:"extension_window_sec"`
	StartTime          *time.Time `json:"start_time,omitempty"`
}

type AuctionResponse struct {
	ID                      uint64     `json:"id"`
	ListingID               uint64     `json:"listing_id"`
	State                   string     `json:"state"`
	TradingMode             string     `json:"trading_mode"`
	StartTime               *time.Time `json:"start_time,omitempty"`
	EndTime                 time.Time  `json:"end_time"`
	StartingPrice           *uint64    `json:"starting_price,omitempty"`
	PriceStep               *uint64    `json:"price_step,omitempty"`
	ReservePrice            *uint64    `json:"reserve_price,omitempty"`
	CurrentPrice            *uint64    `json:"current_price,omitempty"`
	HighestBidAmountInCents *uint64    `json:"highest_bid_amount_in_cents,omitempty"`
	WinnerUserID            *uint64    `json:"winner_user_id,omitempty"`
	WinningBidID            *uint64    `json:"winning_bid_id,omitempty"`
	WinningBidAmountInCents *uint64    `json:"winning_bid_amount_in_cents,omitempty"`
	AntiSnipeEnabled        bool       `json:"anti_snipe_enabled"`
	ExtensionWindowSec      int64      `json:"extension_window_sec"`
	CreatedAt               time.Time  `json:"created_at"`
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

// AuctionEventResponse is one event replayed from the event store.
type AuctionEventResponse struct {
	EventType     string          `json:"event_type"`
	EventID       string          `json:"event_id"`
	SchemaVersion int             `json:"schema_version"`
	Timestamp     time.Time       `json:"timestamp"`
	AuctionID     uint64          `json:"auction_id"`
	Data          json.RawMessage `json:"data"`
}

type AuctionEventListResponse struct {
	Events     []AuctionEventResponse `json:"events"`
	TotalCount int                    `json:"total_count"`
}
