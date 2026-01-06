package entity

import "time"

type BidEntity struct {
	ID            uint64
	AuctionID     uint64
	UserID        uint64
	AmountInCents uint64
	Currency      string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
