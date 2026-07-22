package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/payment/domain/errs"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/infra/mapper"
	"auction/internal/modules/payment/infra/sqlcgen"
	"auction/internal/modules/payment/ports"
)

const withdrawalOutBizNoIndexName = "uq_withdrawals_out_biz_no"

var _ ports.WithdrawalRepository = (*PostgresWithdrawalRepository)(nil)

type PostgresWithdrawalRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.PaymentMapper
}

func NewPostgresWithdrawalRepository(
	db sqlcgen.DBTX,
	paymentMapper *mapper.PaymentMapper,
) *PostgresWithdrawalRepository {
	return &PostgresWithdrawalRepository{
		q:      sqlcgen.New(db),
		mapper: paymentMapper,
	}
}

func (repository *PostgresWithdrawalRepository) Save(
	ctx context.Context,
	withdrawal model.WithdrawalModel,
) (model.WithdrawalModel, error) {
	row, err := repository.q.CreateWithdrawal(ctx, repository.mapper.ToCreateWithdrawalParams(withdrawal))
	if err != nil {
		if isUniqueViolation(err, withdrawalOutBizNoIndexName) {
			return model.WithdrawalModel{}, errs.ErrWithdrawalAlreadyExists
		}

		return model.WithdrawalModel{}, err
	}

	return repository.mapper.ToWithdrawalDomain(row)
}

func (repository *PostgresWithdrawalRepository) FindByID(
	ctx context.Context,
	id uint64,
) (model.WithdrawalModel, error) {
	row, err := repository.q.GetWithdrawalByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.WithdrawalModel{}, errs.ErrWithdrawalNotFound
		}

		return model.WithdrawalModel{}, err
	}

	return repository.mapper.ToWithdrawalDomain(row)
}

func (repository *PostgresWithdrawalRepository) FindByOutBizNo(
	ctx context.Context,
	outBizNo string,
) (model.WithdrawalModel, error) {
	row, err := repository.q.GetWithdrawalByOutBizNo(ctx, outBizNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.WithdrawalModel{}, errs.ErrWithdrawalNotFound
		}

		return model.WithdrawalModel{}, err
	}

	return repository.mapper.ToWithdrawalDomain(row)
}

func (repository *PostgresWithdrawalRepository) Update(
	ctx context.Context,
	withdrawal model.WithdrawalModel,
) (model.WithdrawalModel, error) {
	params := repository.mapper.ToUpdateWithdrawalParams(withdrawal)

	row, err := repository.q.UpdateWithdrawal(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.WithdrawalModel{}, errs.ErrWithdrawalConcurrencyConflict
		}

		return model.WithdrawalModel{}, err
	}

	return repository.mapper.ToWithdrawalDomain(row)
}
