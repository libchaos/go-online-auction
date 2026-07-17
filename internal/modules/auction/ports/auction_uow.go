package ports

import "context"

// AuctionUnitOfWork coordinates persistence of auction and bid aggregates within a single transaction boundary
type AuctionUnitOfWork interface {
	AuctionRepository() AuctionRepository
	BidRepository() BidRepository
	// OutboxRepository records domain events in the transactional outbox so they
	// commit atomically with the state change that produced them
	OutboxRepository() OutboxRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// AuctionUnitOfWorkFactory creates new AuctionUnitOfWork instances
type AuctionUnitOfWorkFactory interface {
	Begin(ctx context.Context) (AuctionUnitOfWork, error)
}
