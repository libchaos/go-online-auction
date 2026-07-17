package query

import (
	"context"
	"time"

	"auction/internal/modules/users/domain/errs"
	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

type GetUserByIDQueryInput struct {
	UserID uint64
}

type GetUserByIDQueryOutput struct {
	ID        uint64
	Name      string
	Email     string
	Role      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetUserByIDQuery struct {
	userRepository ports.UserRepository
	logger         logger.Logger
}

func NewGetUserByIDQuery(
	userRepository ports.UserRepository,
	logger logger.Logger,
) *GetUserByIDQuery {
	return &GetUserByIDQuery{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (q *GetUserByIDQuery) Execute(
	ctx context.Context,
	input GetUserByIDQueryInput,
) (GetUserByIDQueryOutput, error) {
	if input.UserID == 0 {
		return GetUserByIDQueryOutput{}, errs.ErrUserIDRequired
	}

	user, err := q.userRepository.FindByID(ctx, input.UserID)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to find user")
		return GetUserByIDQueryOutput{}, err
	}

	role := user.Role()
	status := user.Status()
	return GetUserByIDQueryOutput{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		Role:      role.String(),
		Status:    status.String(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}, nil
}
