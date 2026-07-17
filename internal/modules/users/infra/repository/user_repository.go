package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/mapper"
	"auction/internal/modules/users/infra/sqlcgen"
	"auction/internal/modules/users/ports"
)

const uniqueViolationCode = "23505"

var _ ports.UserRepository = (*PostgresUserRepository)(nil)

type PostgresUserRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.UserMapper
}

func NewPostgresUserRepository(db sqlcgen.DBTX, mapper *mapper.UserMapper) *PostgresUserRepository {
	return &PostgresUserRepository{
		q:      sqlcgen.New(db),
		mapper: mapper,
	}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user model.UserModel) (model.UserModel, error) {
	row, err := r.q.CreateUser(ctx, r.mapper.ToCreateParams(user))
	if err != nil {
		if isUniqueViolation(err) {
			return model.UserModel{}, errs.ErrEmailAlreadyExists
		}
		return model.UserModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uint64) (model.UserModel, error) {
	row, err := r.q.GetUserByID(ctx, int64(id))
	return r.mapUserResult(row, err)
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (model.UserModel, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	return r.mapUserResult(row, err)
}

func (r *PostgresUserRepository) Update(ctx context.Context, user model.UserModel) error {
	params := r.mapper.ToUpdateParams(user)
	params.PreviousVersion = params.Version - 1 // Domain increments version before calling Update

	rowsAffected, err := r.q.UpdateUser(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return errs.ErrEmailAlreadyExists
		}
		return err
	}

	if rowsAffected == 0 {
		return errs.ErrConcurrencyConflict
	}

	return nil
}

func (r *PostgresUserRepository) FindAllPaginated(
	ctx context.Context,
	limit, offset int,
) ([]model.UserModel, error) {
	rows, err := r.q.ListUsers(ctx, sqlcgen.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	users := []model.UserModel{}
	for _, row := range rows {
		user, mapErr := r.mapper.ToDomain(row)
		if mapErr != nil {
			return nil, mapErr
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *PostgresUserRepository) Count(ctx context.Context) (uint64, error) {
	count, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func (r *PostgresUserRepository) mapUserResult(row sqlcgen.User, err error) (model.UserModel, error) {
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserModel{}, errs.ErrUserNotFound
		}
		return model.UserModel{}, err
	}

	return r.mapper.ToDomain(row)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == uniqueViolationCode
	}
	return false
}
