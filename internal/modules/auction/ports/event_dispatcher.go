package ports

import (
	"context"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/event"
)

// AuctionStartedEventDispatcher defines the interface for dispatching auction started events.
type AuctionStartedEventDispatcher interface {
	Dispatch(ctx context.Context, event event.AuctionStartedEvent) error
}

// BidPlacedEventDispatcher defines the interface for dispatching bid placed events.
type BidPlacedEventDispatcher interface {
	Dispatch(ctx context.Context, event event.BidPlacedEvent) error
}

// AuctionEndedEventDispatcher defines the interface for dispatching auction ended events.
type AuctionEndedEventDispatcher interface {
	Dispatch(ctx context.Context, event event.AuctionEndedEvent) error
}
