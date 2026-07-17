package query

import (
	"context"
	"time"

	"auction/internal/modules/users/ports"
	"auction/internal/shared/modules/logger"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type ListUsersQueryInput struct {
	Limit  int
	Offset int
}

type ListUsersQueryOutput struct {
	Users      []UserSummaryOutput
	TotalCount uint64
	Limit      int
	Offset     int
}

type UserSummaryOutput struct {
	ID        uint64
	Name      string
	Email     string
	Role      string
	Status    string
	CreatedAt time.Time
}

type ListUsersQuery struct {
	userRepository ports.UserRepository
	logger         logger.Logger
}

func NewListUsersQuery(
	userRepository ports.UserRepository,
	logger logger.Logger,
) *ListUsersQuery {
	return &ListUsersQuery{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (q *ListUsersQuery) Execute(
	ctx context.Context,
	input ListUsersQueryInput,
) (ListUsersQueryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	users, err := q.userRepository.FindAllPaginated(ctx, limit, offset)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to fetch users")
		return ListUsersQueryOutput{}, err
	}

	totalCount, err := q.userRepository.Count(ctx)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to count users")
		return ListUsersQueryOutput{}, err
	}

	userOutputs := make([]UserSummaryOutput, 0, len(users))
	for _, user := range users {
		role := user.Role()
		status := user.Status()
		userOutputs = append(userOutputs, UserSummaryOutput{
			ID:        user.ID(),
			Name:      user.Name(),
			Email:     user.Email(),
			Role:      role.String(),
			Status:    status.String(),
			CreatedAt: user.CreatedAt(),
		})
	}

	return ListUsersQueryOutput{
		Users:      userOutputs,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
