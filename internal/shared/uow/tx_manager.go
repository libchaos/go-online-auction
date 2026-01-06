package uow

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager handles transaction lifecycle
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithTransaction executes fn within a transaction
// Automatically commits on success, rolls back on error
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return ErrTransactionFailed
	}

	if fnErr := fn(ctx, tx); fnErr != nil {
		_ = tx.Rollback(ctx)
		return fnErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return ErrTransactionFailed
	}

	return nil
}
