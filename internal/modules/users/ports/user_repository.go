package ports

import (
	"context"

	"auction/internal/modules/users/domain/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.UserModel) (model.UserModel, error)
	FindByID(ctx context.Context, id uint64) (model.UserModel, error)
	FindByEmail(ctx context.Context, email string) (model.UserModel, error)
	Update(ctx context.Context, user model.UserModel) error
	FindAllPaginated(ctx context.Context, limit, offset int) ([]model.UserModel, error)
	Count(ctx context.Context) (uint64, error)
}
