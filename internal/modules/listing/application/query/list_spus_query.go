package query

import (
	"context"
	"time"

	"auction/internal/modules/listing/domain/enum"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type ListSpusQueryInput struct {
	Status     *string
	CategoryID *uint64
	Limit      int
	Offset     int
}

type ListSpusQueryOutput struct {
	Spus       []SpuSummaryOutput
	TotalCount uint64
	Limit      int
	Offset     int
}

type SpuSummaryOutput struct {
	ID          uint64
	Title       string
	Description string
	CategoryID  uint64
	Brand       *string
	Images      []string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListSpusQuery struct {
	spuRepository ports.SpuRepository
	logger        logger.Logger
}

func NewListSpusQuery(
	spuRepository ports.SpuRepository,
	logger logger.Logger,
) *ListSpusQuery {
	return &ListSpusQuery{
		spuRepository: spuRepository,
		logger:        logger,
	}
}

func (q *ListSpusQuery) Execute(
	ctx context.Context,
	input ListSpusQueryInput,
) (ListSpusQueryOutput, error) {
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

	var status *enum.ListingStatusEnum
	if input.Status != nil {
		s, err := enum.NewListingStatusEnum(*input.Status)
		if err != nil {
			q.logger.Error().Err(err).Msg("invalid listing status")
			return ListSpusQueryOutput{}, err
		}
		status = &s
	}

	filter := ports.ListSpusFilter{
		Status:     status,
		CategoryID: input.CategoryID,
		Limit:      limit,
		Offset:     offset,
	}

	spus, err := q.spuRepository.List(ctx, filter)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to fetch spus")
		return ListSpusQueryOutput{}, err
	}

	totalCount, err := q.spuRepository.Count(ctx, filter)
	if err != nil {
		q.logger.Error().Err(err).Msg("failed to count spus")
		return ListSpusQueryOutput{}, err
	}

	outputs := make([]SpuSummaryOutput, 0, len(spus))
	for _, spu := range spus {
		status := spu.Status()
		outputs = append(outputs, SpuSummaryOutput{
			ID:          spu.ID(),
			Title:       spu.Title(),
			Description: spu.Description(),
			CategoryID:  spu.CategoryID(),
			Brand:       spu.Brand(),
			Images:      spu.Images(),
			Status:      status.String(),
			CreatedAt:   spu.CreatedAt(),
			UpdatedAt:   spu.UpdatedAt(),
		})
	}

	return ListSpusQueryOutput{
		Spus:       outputs,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}
