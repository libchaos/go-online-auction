package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/mapper"
	"auction/internal/modules/users/infra/sqlcgen"
	"auction/internal/modules/users/ports"
)

var _ ports.RefreshTokenRepository = (*PostgresRefreshTokenRepository)(nil)

type PostgresRefreshTokenRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.RefreshTokenMapper
}

func NewPostgresRefreshTokenRepository(
	db sqlcgen.DBTX,
	mapper *mapper.RefreshTokenMapper,
) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresRefreshTokenRepository) Create(
	ctx context.Context,
	token model.RefreshTokenModel,
) (model.RefreshTokenModel, error) {
	row, err := r.q.CreateRefreshToken(ctx, r.mapper.ToCreateParams(token))
	if err != nil {
		return model.RefreshTokenModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresRefreshTokenRepository) FindByTokenHash(
	ctx context.Context,
	tokenHash string,
) (model.RefreshTokenModel, error) {
	row, err := r.q.GetRefreshTokenByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshTokenModel{}, errs.ErrRefreshTokenNotFound
		}
		return model.RefreshTokenModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresRefreshTokenRepository) Update(ctx context.Context, token model.RefreshTokenModel) error {
	rowsAffected, err := r.q.UpdateRefreshToken(ctx, r.mapper.ToUpdateParams(token))
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uint64) error {
	return r.q.RevokeAllRefreshTokensForUser(ctx, int64(userID))
}
