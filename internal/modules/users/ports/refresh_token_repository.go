package ports

import (
	"context"

	"auction/internal/modules/users/domain/model"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token model.RefreshTokenModel) (model.RefreshTokenModel, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (model.RefreshTokenModel, error)
	Update(ctx context.Context, token model.RefreshTokenModel) error
	RevokeAllForUser(ctx context.Context, userID uint64) error
}
