package handler

import (
	"net/http"
	"strconv"

	"auction/internal/modules/listing/application/command"
	"auction/internal/modules/listing/application/query"
	"auction/internal/modules/listing/infra/http/dto"
	httperrs "auction/internal/modules/listing/infra/http/errs"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
)

type CategoryHandler struct {
	createCategoryCommand *command.CreateCategoryCommand
	updateCategoryCommand *command.UpdateCategoryCommand
	deleteCategoryCommand *command.DeleteCategoryCommand
	listCategoriesQuery   *query.ListCategoriesQuery
	getCategoryByIDQuery  *query.GetCategoryByIDQuery
	getCategoryTreeQuery  *query.GetCategoryTreeQuery
}

func NewCategoryHandler(
	createCategoryCommand *command.CreateCategoryCommand,
	updateCategoryCommand *command.UpdateCategoryCommand,
	deleteCategoryCommand *command.DeleteCategoryCommand,
	listCategoriesQuery *query.ListCategoriesQuery,
	getCategoryByIDQuery *query.GetCategoryByIDQuery,
	getCategoryTreeQuery *query.GetCategoryTreeQuery,
) *CategoryHandler {
	return &CategoryHandler{
		createCategoryCommand: createCategoryCommand,
		updateCategoryCommand: updateCategoryCommand,
		deleteCategoryCommand: deleteCategoryCommand,
		listCategoriesQuery:   listCategoriesQuery,
		getCategoryByIDQuery:  getCategoryByIDQuery,
		getCategoryTreeQuery:  getCategoryTreeQuery,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCategoryRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.createCategoryCommand.Execute(r.Context(), command.CreateCategoryCommandInput{
		Name:      req.Name,
		ParentID:  req.ParentID,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.CategoryResponse{
		ID:        output.ID,
		Name:      output.Name,
		ParentID:  output.ParentID,
		Depth:     output.Depth,
		Path:      output.Path,
		SortOrder: output.SortOrder,
		CreatedAt: output.CreatedAt,
	}, nil)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	var parentID *uint64
	if parentParam := r.URL.Query().Get("parent_id"); parentParam != "" {
		id, err := strconv.ParseUint(parentParam, 10, 64)
		if err != nil {
			response.Error(w, httperrs.ErrInvalidID)
			return
		}
		parentID = &id
	}

	output, err := h.listCategoriesQuery.Execute(r.Context(), query.ListCategoriesQueryInput{
		ParentID: parentID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	categories := make([]dto.CategoryResponse, 0, len(output.Categories))
	for _, category := range output.Categories {
		categories = append(categories, dto.CategoryResponse{
			ID:        category.ID,
			Name:      category.Name,
			ParentID:  category.ParentID,
			Depth:     category.Depth,
			Path:      category.Path,
			SortOrder: category.SortOrder,
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.CategoryListResponse{Categories: categories}, nil)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.getCategoryByIDQuery.Execute(r.Context(), query.GetCategoryByIDQueryInput{
		ID: categoryID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.CategoryResponse{
		ID:        output.Category.ID,
		Name:      output.Category.Name,
		ParentID:  output.Category.ParentID,
		Depth:     output.Category.Depth,
		Path:      output.Category.Path,
		SortOrder: output.Category.SortOrder,
		CreatedAt: output.Category.CreatedAt,
		UpdatedAt: output.Category.UpdatedAt,
	}, nil)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	var req dto.UpdateCategoryRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.updateCategoryCommand.Execute(r.Context(), command.UpdateCategoryCommandInput{
		ID:        categoryID,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.CategoryResponse{
		ID:        output.ID,
		Name:      output.Name,
		ParentID:  output.ParentID,
		Depth:     output.Depth,
		Path:      output.Path,
		SortOrder: output.SortOrder,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	if err = h.deleteCategoryCommand.Execute(r.Context(), command.DeleteCategoryCommandInput{
		ID: categoryID,
	}); err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) Tree(w http.ResponseWriter, r *http.Request) {
	output, err := h.getCategoryTreeQuery.Execute(r.Context(), query.GetCategoryTreeQueryInput{RootID: nil})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	roots := make([]*dto.CategoryTreeNode, 0, len(output.Roots))
	for _, node := range output.Roots {
		roots = append(roots, toCategoryTreeNodeDTO(node))
	}

	_ = response.JSON(w, http.StatusOK, dto.CategoryTreeResponse{Roots: roots}, nil)
}

func (h *CategoryHandler) Subtree(w http.ResponseWriter, r *http.Request) {
	categoryID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.getCategoryTreeQuery.Execute(r.Context(), query.GetCategoryTreeQueryInput{RootID: &categoryID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	roots := make([]*dto.CategoryTreeNode, 0, len(output.Roots))
	for _, node := range output.Roots {
		roots = append(roots, toCategoryTreeNodeDTO(node))
	}

	_ = response.JSON(w, http.StatusOK, dto.CategoryTreeResponse{Roots: roots}, nil)
}

func toCategoryTreeNodeDTO(node *query.CategoryTreeNode) *dto.CategoryTreeNode {
	children := make([]*dto.CategoryTreeNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, toCategoryTreeNodeDTO(child))
	}

	return &dto.CategoryTreeNode{
		ID:        node.ID,
		ParentID:  node.ParentID,
		Name:      node.Name,
		Depth:     node.Depth,
		Path:      node.Path,
		SortOrder: node.SortOrder,
		Children:  children,
		CreatedAt: node.CreatedAt,
		UpdatedAt: node.UpdatedAt,
	}
}

func parseIDParam(r *http.Request) (uint64, error) {
	return strconv.ParseUint(request.Param(r, "id"), 10, 64)
}
