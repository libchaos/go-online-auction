package ports

import "context"

// ListingUnitOfWork coordinates persistence of SPU and SKU aggregates within a
// single transaction boundary
type ListingUnitOfWork interface {
	SpuRepository() SpuRepository
	SkuRepository() SkuRepository
	// OutboxRepository records domain events in the transactional outbox so they
	// commit atomically with the state change that produced them
	OutboxRepository() ListingOutboxRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// ListingUnitOfWorkFactory creates new ListingUnitOfWork instances
type ListingUnitOfWorkFactory interface {
	Begin(ctx context.Context) (ListingUnitOfWork, error)
}
