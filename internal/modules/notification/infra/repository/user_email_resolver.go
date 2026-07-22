package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.UserEmailResolver = (*PostgresUserEmailResolver)(nil)

// PostgresUserEmailResolver is a read-only adapter over the shared users table.
// It lets the notification module resolve a recipient's email address without
// reaching into the users module.
type PostgresUserEmailResolver struct {
	q *sqlcgen.Queries
}

func NewPostgresUserEmailResolver(db sqlcgen.DBTX) *PostgresUserEmailResolver {
	return &PostgresUserEmailResolver{q: sqlcgen.New(db)}
}

func (resolver *PostgresUserEmailResolver) ResolveEmail(
	ctx context.Context,
	userID uint64,
) (string, bool, error) {
	email, err := resolver.q.GetUserEmailByID(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}

		return "", false, err
	}

	return email, true, nil
}
