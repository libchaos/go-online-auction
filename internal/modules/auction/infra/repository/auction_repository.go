package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/sqlcgen"
	"auction/internal/modules/auction/ports"
)

var _ ports.AuctionRepository = (*PostgresAuctionRepository)(nil)

type PostgresAuctionRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.AuctionMapper
}

func NewPostgresAuctionRepository(db sqlcgen.DBTX, mapper *mapper.AuctionMapper) *PostgresAuctionRepository {
	return &PostgresAuctionRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresAuctionRepository) Create(
	ctx context.Context,
	auction model.AuctionModel,
) (model.AuctionModel, error) {
	row, err := r.q.CreateAuction(ctx, r.mapper.ToCreateParams(auction))
	if err != nil {
		return model.AuctionModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresAuctionRepository) FindByID(ctx context.Context, id uint64) (model.AuctionModel, error) {
	row, err := r.q.GetAuctionByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AuctionModel{}, errs.ErrAuctionNotFound
		}
		return model.AuctionModel{}, err
	}

	return r.mapper.ToDomain(row)
}

// FindByIDForUpdate retrieves an auction with row-level lock for update
// Uses NOWAIT to fail fast under contention
func (r *PostgresAuctionRepository) FindByIDForUpdate(ctx context.Context, id uint64) (model.AuctionModel, error) {
	row, err := r.q.GetAuctionByIDForUpdate(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AuctionModel{}, errs.ErrAuctionNotFound
		}
		// Check for lock_not_available error (NOWAIT)
		if isPgLockError(err) {
			return model.AuctionModel{}, errs.ErrConcurrencyConflict
		}
		return model.AuctionModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresAuctionRepository) Update(ctx context.Context, auction model.AuctionModel) error {
	params := r.mapper.ToUpdateParams(auction)
	params.PreviousVersion = params.Version - 1 // Domain increments version before calling Update

	rowsAffected, err := r.q.UpdateAuction(ctx, params)
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrConcurrencyConflict
	}

	return nil
}

func (r *PostgresAuctionRepository) FindAllPaginated(
	ctx context.Context,
	state *enum.AuctionStateEnum,
	limit, offset int,
) ([]model.AuctionModel, error) {
	var rows []sqlcgen.Auction
	var err error

	if state != nil {
		rows, err = r.q.ListAuctionsByState(ctx, sqlcgen.ListAuctionsByStateParams{
			State:  sqlcgen.AuctionState(state.String()),
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	} else {
		rows, err = r.q.ListAuctions(ctx, sqlcgen.ListAuctionsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}
	if err != nil {
		return nil, err
	}

	auctions := []model.AuctionModel{}
	for _, row := range rows {
		auction, mapErr := r.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}
		auctions = append(auctions, auction)
	}

	return auctions, nil
}

func (r *PostgresAuctionRepository) Count(ctx context.Context, state *enum.AuctionStateEnum) (uint64, error) {
	var count int64
	var err error

	if state != nil {
		count, err = r.q.CountAuctionsByState(ctx, sqlcgen.AuctionState(state.String()))
	} else {
		count, err = r.q.CountAuctions(ctx)
	}
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

// FindIDsDueToStart returns IDs of draft auctions whose scheduled start time has passed
func (r *PostgresAuctionRepository) FindIDsDueToStart(ctx context.Context, limit int) ([]uint64, error) {
	ids, err := r.q.ListAuctionIDsDueToStart(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	return toUint64IDs(ids), nil
}

// FindIDsDueToClose returns IDs of active auctions whose end time has passed
func (r *PostgresAuctionRepository) FindIDsDueToClose(ctx context.Context, limit int) ([]uint64, error) {
	ids, err := r.q.ListAuctionIDsDueToClose(ctx, int32(limit))
	if err != nil {
		return nil, err
	}

	return toUint64IDs(ids), nil
}

func toUint64IDs(ids []int64) []uint64 {
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		result = append(result, uint64(id))
	}
	return result
}

func isPgLockError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 55P03 = lock_not_available
		return pgErr.Code == "55P03"
	}
	return false
}
