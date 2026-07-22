package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/notification/domain/errs"
	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.WatchlistRepository = (*PostgresWatchlistRepository)(nil)

// PostgresWatchlistRepository persists a user's explicit interest in a catalogue
// product (SPU). The unique (user_id, spu_id) constraint makes Save idempotent.
type PostgresWatchlistRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresWatchlistRepository(db sqlcgen.DBTX) *PostgresWatchlistRepository {
	return &PostgresWatchlistRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresWatchlistRepository) Save(
	ctx context.Context,
	watchlist model.Watchlist,
) (model.Watchlist, error) {
	row, err := repository.q.UpsertWatchlist(ctx, sqlcgen.UpsertWatchlistParams{
		UserID: int64(watchlist.UserID()),
		SpuID:  int64(watchlist.SpuID()),
	})
	if err != nil {
		return model.Watchlist{}, err
	}

	return model.ReconstructWatchlist(
		uint64(row.ID),
		uint64(row.UserID),
		uint64(row.SpuID),
		row.CreatedAt,
	), nil
}

func (repository *PostgresWatchlistRepository) Remove(ctx context.Context, userID, spuID uint64) error {
	_, err := repository.q.DeleteWatchlist(ctx, sqlcgen.DeleteWatchlistParams{
		UserID: int64(userID),
		SpuID:  int64(spuID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrWatchNotFound
		}

		return err
	}

	return nil
}

func (repository *PostgresWatchlistRepository) ListByUser(
	ctx context.Context,
	userID uint64,
	limit, offset int,
) ([]model.Watchlist, error) {
	rows, err := repository.q.ListWatchlistsByUser(ctx, sqlcgen.ListWatchlistsByUserParams{
		UserID: int64(userID),
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	watches := make([]model.Watchlist, 0, len(rows))
	for _, row := range rows {
		watches = append(watches, model.ReconstructWatchlist(
			uint64(row.ID),
			uint64(row.UserID),
			uint64(row.SpuID),
			row.CreatedAt,
		))
	}

	return watches, nil
}

func (repository *PostgresWatchlistRepository) FindWatcherIDsBySpuID(
	ctx context.Context,
	spuID uint64,
) ([]uint64, error) {
	rows, err := repository.q.FindWatcherIDsBySpuID(ctx, int64(spuID))
	if err != nil {
		return nil, err
	}

	watcherIDs := make([]uint64, 0, len(rows))
	for _, userID := range rows {
		watcherIDs = append(watcherIDs, uint64(userID))
	}

	return watcherIDs, nil
}
