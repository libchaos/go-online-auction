package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.ListingReadPort = (*PostgresListingReadRepository)(nil)

// PostgresListingReadRepository is a read-only adapter over the shared spus
// table. It lets the notification module enrich a listing notification with the
// product title without reaching into the listing module.
type PostgresListingReadRepository struct {
	q *sqlcgen.Queries
}

func NewPostgresListingReadRepository(db sqlcgen.DBTX) *PostgresListingReadRepository {
	return &PostgresListingReadRepository{q: sqlcgen.New(db)}
}

func (repository *PostgresListingReadRepository) FindSpuTitleByID(
	ctx context.Context,
	spuID uint64,
) (string, bool, error) {
	title, err := repository.q.FindSpuTitleByID(ctx, int64(spuID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}

		return "", false, err
	}

	return title, true, nil
}
