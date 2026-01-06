package ports

import "context"

// AuctionUnitOfWork coordinates persistence of auction and bid aggregates within a single transaction boundary
type AuctionUnitOfWork interface {
	AuctionRepository() AuctionRepository
	BidRepository() BidRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// AuctionUnitOfWorkFactory creates new AuctionUnitOfWork instances
type AuctionUnitOfWorkFactory interface {
	Begin(ctx context.Context) (AuctionUnitOfWork, error)
}

