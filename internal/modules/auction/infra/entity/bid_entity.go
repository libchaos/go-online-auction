package entity

import "time"

type BidEntity struct {
	ID            uint64
	AuctionID     uint64
	UserID        uint64
	AmountInCents uint64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
