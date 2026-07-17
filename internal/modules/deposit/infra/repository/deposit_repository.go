package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/mapper"
	"auction/internal/modules/deposit/infra/sqlcgen"
	"auction/internal/modules/deposit/ports"
)

const (
	pgUniqueViolationCode       = "23505"
	depositUserAuctionIndexName = "uq_deposits_user_auction"
)

var _ ports.DepositRepository = (*PostgresDepositRepository)(nil)

type PostgresDepositRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.DepositMapper
}

func NewPostgresDepositRepository(db sqlcgen.DBTX, depositMapper *mapper.DepositMapper) *PostgresDepositRepository {
	return &PostgresDepositRepository{
		q:      sqlcgen.New(db),
		mapper: depositMapper,
	}
}

func (repository *PostgresDepositRepository) Save(
	ctx context.Context,
	deposit model.DepositModel,
) (model.DepositModel, error) {
	row, err := repository.q.CreateDeposit(ctx, repository.mapper.ToCreateParams(deposit))
	if err != nil {
		if isUniqueViolation(err, depositUserAuctionIndexName) {
			return model.DepositModel{}, errs.ErrDepositAlreadyExists
		}

		return model.DepositModel{}, err
	}

	return repository.mapper.ToDomain(row)
}

func (repository *PostgresDepositRepository) FindByID(
	ctx context.Context,
	id uint64,
) (model.DepositModel, error) {
	row, err := repository.q.GetDepositByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.DepositModel{}, errs.ErrDepositNotFound
		}

		return model.DepositModel{}, err
	}

	return repository.mapper.ToDomain(row)
}

func (repository *PostgresDepositRepository) FindByUserAndAuction(
	ctx context.Context,
	userID uint64,
	auctionID uint64,
) (model.DepositModel, error) {
	row, err := repository.q.GetDepositByUserAndAuction(ctx, sqlcgen.GetDepositByUserAndAuctionParams{
		UserID:    int64(userID),
		AuctionID: int64(auctionID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.DepositModel{}, errs.ErrDepositNotFound
		}

		return model.DepositModel{}, err
	}

	return repository.mapper.ToDomain(row)
}

func (repository *PostgresDepositRepository) ListByUser(
	ctx context.Context,
	userID uint64,
) ([]model.DepositModel, error) {
	rows, err := repository.q.ListDepositsByUser(ctx, sqlcgen.ListDepositsByUserParams{
		UserID: int64(userID),
		Limit:  maxDepositListLimit,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	return repository.mapRows(rows)
}

func (repository *PostgresDepositRepository) ListHeldByAuction(
	ctx context.Context,
	auctionID uint64,
) ([]model.DepositModel, error) {
	rows, err := repository.q.ListHeldDepositsByAuction(ctx, int64(auctionID))
	if err != nil {
		return nil, err
	}

	return repository.mapRows(rows)
}

func (repository *PostgresDepositRepository) Update(
	ctx context.Context,
	deposit model.DepositModel,
) (model.DepositModel, error) {
	params := repository.mapper.ToUpdateParams(deposit)
	params.Version_2 = params.Version - 1

	row, err := repository.q.UpdateDeposit(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.DepositModel{}, errs.ErrDepositConcurrencyConflict
		}

		return model.DepositModel{}, err
	}

	return repository.mapper.ToDomain(row)
}

func (repository *PostgresDepositRepository) mapRows(rows []sqlcgen.Deposit) ([]model.DepositModel, error) {
	deposits := make([]model.DepositModel, 0, len(rows))
	for _, row := range rows {
		deposit, mapErr := repository.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}

		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgUniqueViolationCode && pgErr.ConstraintName == constraintName
	}

	return false
}

const maxDepositListLimit = 100
