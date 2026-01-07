package entity

import "time"

type AuctionEntity struct {
	ID                      uint64
	ListingID               uint64
	StartTime               *time.Time
	EndTime                 time.Time
	State                   string
	HighestBidAmountInCents *uint64
	Version                 uint64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
