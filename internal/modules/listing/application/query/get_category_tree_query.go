package query

import (
	"context"
	"errors"
	"time"

	"auction/internal/modules/listing/domain/errs"
	"auction/internal/modules/listing/domain/model"
	"auction/internal/modules/listing/ports"
	"auction/internal/shared/modules/logger"
)

type GetCategoryTreeQueryInput struct {
	// RootID, when set, returns the subtree rooted at that category (inclusive of
	// the root). When nil, the full category forest is returned.
	RootID *uint64
}

type GetCategoryTreeQueryOutput struct {
	Roots []*CategoryTreeNode
}

type CategoryTreeNode struct {
	ID        uint64
	ParentID  *uint64
	Name      string
	Depth     int32
	Path      string
	SortOrder int32
	Children  []*CategoryTreeNode
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GetCategoryTreeQuery struct {
	categoryRepository ports.CategoryRepository
	logger             logger.Logger
}

func NewGetCategoryTreeQuery(
	categoryRepository ports.CategoryRepository,
	logger logger.Logger,
) *GetCategoryTreeQuery {
	return &GetCategoryTreeQuery{
		categoryRepository: categoryRepository,
		logger:             logger,
	}
}

func (q *GetCategoryTreeQuery) Execute(
	ctx context.Context,
	input GetCategoryTreeQueryInput,
) (GetCategoryTreeQueryOutput, error) {
	if input.RootID == nil {
		categories, err := q.categoryRepository.ListAll(ctx)
		if err != nil {
			q.logger.Error().Err(err).Msg("failed to load categories for tree")
			return GetCategoryTreeQueryOutput{}, err
		}
		return GetCategoryTreeQueryOutput{Roots: buildCategoryForest(categories)}, nil
	}

	return q.loadSubtree(ctx, *input.RootID)
}

func (q *GetCategoryTreeQuery) loadSubtree(
	ctx context.Context,
	rootID uint64,
) (GetCategoryTreeQueryOutput, error) {
	anchor, err := q.categoryRepository.FindByID(ctx, rootID)
	if err != nil {
		if errors.Is(err, errs.ErrCategoryNotFound) {
			return GetCategoryTreeQueryOutput{}, errs.ErrCategoryNotFound
		}
		q.logger.Error().Err(err).Uint64("root_id", rootID).Msg("failed to find tree root category")
		return GetCategoryTreeQueryOutput{}, err
	}

	descendants, err := q.categoryRepository.ListDescendants(ctx, rootID)
	if err != nil {
		q.logger.Error().Err(err).Uint64("root_id", rootID).Msg("failed to load category descendants")
		return GetCategoryTreeQueryOutput{}, err
	}

	categories := make([]model.CategoryModel, 0, len(descendants)+1)
	categories = append(categories, anchor)
	categories = append(categories, descendants...)

	return GetCategoryTreeQueryOutput{Roots: buildCategoryForest(categories)}, nil
}

func buildCategoryForest(categories []model.CategoryModel) []*CategoryTreeNode {
	nodes := make(map[uint64]*CategoryTreeNode, len(categories))
	order := make([]uint64, 0, len(categories))

	for _, category := range categories {
		node := &CategoryTreeNode{
			ID:        category.ID(),
			ParentID:  category.ParentID(),
			Name:      category.Name(),
			Depth:     category.Depth(),
			Path:      category.Path(),
			SortOrder: category.SortOrder(),
			CreatedAt: category.CreatedAt(),
			UpdatedAt: category.UpdatedAt(),
		}
		nodes[category.ID()] = node
		order = append(order, category.ID())
	}

	roots := make([]*CategoryTreeNode, 0)
	for _, id := range order {
		node := nodes[id]
		if node.ParentID != nil {
			if parent, ok := nodes[*node.ParentID]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	return roots
}
