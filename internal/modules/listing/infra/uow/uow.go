package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/listing/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.ListingUnitOfWork = (*ListingUnitOfWork)(nil)

// ListingUnitOfWork coordinates persistence of SPU and SKU aggregates within a
// single transaction boundary
type ListingUnitOfWork struct {
	tx               pgx.Tx
	spuRepository    ports.SpuRepository
	skuRepository    ports.SkuRepository
	outboxRepository ports.ListingOutboxRepository
	completed        bool
}

// SpuRepository returns the SPU repository bound to this unit of work
func (uow *ListingUnitOfWork) SpuRepository() ports.SpuRepository {
	return uow.spuRepository
}

// SkuRepository returns the SKU repository bound to this unit of work
func (uow *ListingUnitOfWork) SkuRepository() ports.SkuRepository {
	return uow.skuRepository
}

// OutboxRepository returns the outbox repository bound to this unit of work
func (uow *ListingUnitOfWork) OutboxRepository() ports.ListingOutboxRepository {
	return uow.outboxRepository
}

// Complete commits the transaction
func (uow *ListingUnitOfWork) Complete(ctx context.Context) error {
	if uow.completed {
		return nil
	}
	uow.completed = true

	if err := uow.tx.Commit(ctx); err != nil {
		return shareduow.ErrTransactionFailed
	}
	return nil
}

// Rollback aborts the transaction
func (uow *ListingUnitOfWork) Rollback(ctx context.Context) error {
	if uow.completed {
		return nil
	}
	uow.completed = true
	return uow.tx.Rollback(ctx)
}
