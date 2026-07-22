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

const paymentOutTradeNoIndexName = "uq_payments_out_trade_no"

var _ ports.PaymentRepository = (*PostgresPaymentRepository)(nil)

type PostgresPaymentRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.PaymentMapper
}

func NewPostgresPaymentRepository(db sqlcgen.DBTX, paymentMapper *mapper.PaymentMapper) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{
		q:      sqlcgen.New(db),
		mapper: paymentMapper,
	}
}

func (repository *PostgresPaymentRepository) Save(
	ctx context.Context,
	payment model.PaymentModel,
) (model.PaymentModel, error) {
	row, err := repository.q.CreatePayment(ctx, repository.mapper.ToCreatePaymentParams(payment))
	if err != nil {
		if isUniqueViolation(err, paymentOutTradeNoIndexName) {
			return model.PaymentModel{}, errs.ErrPaymentAlreadyExists
		}

		return model.PaymentModel{}, err
	}

	return repository.mapper.ToPaymentDomain(row)
}

func (repository *PostgresPaymentRepository) FindByID(
	ctx context.Context,
	id uint64,
) (model.PaymentModel, error) {
	row, err := repository.q.GetPaymentByID(ctx, int64(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PaymentModel{}, errs.ErrPaymentNotFound
		}

		return model.PaymentModel{}, err
	}

	return repository.mapper.ToPaymentDomain(row)
}

func (repository *PostgresPaymentRepository) FindByOutTradeNo(
	ctx context.Context,
	outTradeNo string,
) (model.PaymentModel, error) {
	row, err := repository.q.GetPaymentByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PaymentModel{}, errs.ErrPaymentNotFound
		}

		return model.PaymentModel{}, err
	}

	return repository.mapper.ToPaymentDomain(row)
}

func (repository *PostgresPaymentRepository) Update(
	ctx context.Context,
	payment model.PaymentModel,
) (model.PaymentModel, error) {
	params := repository.mapper.ToUpdatePaymentParams(payment)

	row, err := repository.q.UpdatePayment(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PaymentModel{}, errs.ErrPaymentConcurrencyConflict
		}

		return model.PaymentModel{}, err
	}

	return repository.mapper.ToPaymentDomain(row)
}
