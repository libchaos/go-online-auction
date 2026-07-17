package ports

import (
	"context"
	"time"
)

type BidCommand struct {
	IdempotencyKey   string    `json:"idempotency_key"`
	AuctionID        uint64    `json:"auction_id"`
	UserID           uint64    `json:"user_id"`
	AmountInCents    uint64    `json:"amount_in_cents"`
	MaxAmountInCents *uint64   `json:"max_amount_in_cents"`
	IssuedAt         time.Time `json:"issued_at"`
}

type BidCommandAck struct {
	IdempotencyKey string
}

// BidCommandPublisher publishes a bid command to the auction command stream.
type BidCommandPublisher interface {
	// Publish publishes the bid command; IdempotencyKey is used as Nats-Msg-Id
	// so JetStream deduplicates repeated submissions within the dedupe window.
	Publish(ctx context.Context, cmd BidCommand) (BidCommandAck, error)
}
