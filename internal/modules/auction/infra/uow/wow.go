package uow

import (
	"context"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/auction/ports"
	shareduow "auction/internal/shared/modules/uow"
)

var _ ports.AuctionUnitOfWork = (*AuctionUnitOfWork)(nil)

// AuctionUnitOfWork coordinates persistence of auction and bid aggregates within a single transaction boundary
type AuctionUnitOfWork struct {
	tx                pgx.Tx
	auctionRepository ports.AuctionRepository
	bidRepository     ports.BidRepository
	outboxRepository  ports.OutboxRepository
	completed         bool
}

// AuctionRepository returns the auction repository bound to this unit of work
func (uow *AuctionUnitOfWork) AuctionRepository() ports.AuctionRepository {
	return uow.auctionRepository
}

// BidRepository returns the bid repository bound to this unit of work
func (uow *AuctionUnitOfWork) BidRepository() ports.BidRepository {
	return uow.bidRepository
}

// OutboxRepository returns the outbox repository bound to this unit of work
func (uow *AuctionUnitOfWork) OutboxRepository() ports.OutboxRepository {
	return uow.outboxRepository
}

// Complete commits the transaction
func (uow *AuctionUnitOfWork) Complete(ctx context.Context) error {
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
func (uow *AuctionUnitOfWork) Rollback(ctx context.Context) error {
	if uow.completed {
		return nil
	}
	uow.completed = true
	return uow.tx.Rollback(ctx)
}
