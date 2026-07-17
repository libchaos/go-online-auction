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

type SpuHandler struct {
	createSpuCommand   *command.CreateSpuCommand
	updateSpuCommand   *command.UpdateSpuCommand
	publishSpuCommand  *command.PublishSpuCommand
	offShelfSpuCommand *command.OffShelfSpuCommand
	listSpusQuery      *query.ListSpusQuery
	getSpuByIDQuery    *query.GetSpuByIDQuery
}

func NewSpuHandler(
	createSpuCommand *command.CreateSpuCommand,
	updateSpuCommand *command.UpdateSpuCommand,
	publishSpuCommand *command.PublishSpuCommand,
	offShelfSpuCommand *command.OffShelfSpuCommand,
	listSpusQuery *query.ListSpusQuery,
	getSpuByIDQuery *query.GetSpuByIDQuery,
) *SpuHandler {
	return &SpuHandler{
		createSpuCommand:   createSpuCommand,
		updateSpuCommand:   updateSpuCommand,
		publishSpuCommand:  publishSpuCommand,
		offShelfSpuCommand: offShelfSpuCommand,
		listSpusQuery:      listSpusQuery,
		getSpuByIDQuery:    getSpuByIDQuery,
	}
}

func (h *SpuHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSpuRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.createSpuCommand.Execute(r.Context(), command.CreateSpuCommandInput{
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Brand:       req.Brand,
		Images:      req.Images,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.SpuResponse{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		CategoryID:  output.CategoryID,
		Brand:       output.Brand,
		Images:      output.Images,
		Status:      output.Status,
		CreatedAt:   output.CreatedAt,
	}, nil)
}

func (h *SpuHandler) List(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var status *string
	if statusParam := queryParams.Get("status"); statusParam != "" {
		status = &statusParam
	}

	var categoryID *uint64
	if categoryParam := queryParams.Get("category_id"); categoryParam != "" {
		id, err := strconv.ParseUint(categoryParam, 10, 64)
		if err != nil {
			response.Error(w, httperrs.ErrInvalidID)
			return
		}
		categoryID = &id
	}

	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	offset, _ := strconv.Atoi(queryParams.Get("offset"))

	output, err := h.listSpusQuery.Execute(r.Context(), query.ListSpusQueryInput{
		Status:     status,
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	spus := make([]dto.SpuResponse, 0, len(output.Spus))
	for _, spu := range output.Spus {
		spus = append(spus, dto.SpuResponse{
			ID:          spu.ID,
			Title:       spu.Title,
			Description: spu.Description,
			CategoryID:  spu.CategoryID,
			Brand:       spu.Brand,
			Images:      spu.Images,
			Status:      spu.Status,
			CreatedAt:   spu.CreatedAt,
			UpdatedAt:   spu.UpdatedAt,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.SpuListResponse{
		Spus:       spus,
		TotalCount: output.TotalCount,
		Limit:      output.Limit,
		Offset:     output.Offset,
	}, nil)
}

func (h *SpuHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	spuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.getSpuByIDQuery.Execute(r.Context(), query.GetSpuByIDQueryInput{ID: spuID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	skus := make([]dto.SkuResponse, 0, len(output.Skus))
	for _, sku := range output.Skus {
		skus = append(skus, dto.SkuResponse{
			ID:           sku.ID,
			SpuID:        sku.SpuID,
			SpecValues:   sku.SpecValues,
			PriceInCents: sku.PriceInCents,
			Quantity:     sku.Quantity,
			Status:       sku.Status,
			CreatedAt:    sku.CreatedAt,
			UpdatedAt:    sku.UpdatedAt,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.SpuResponse{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		CategoryID:  output.CategoryID,
		Brand:       output.Brand,
		Images:      output.Images,
		Status:      output.Status,
		Skus:        skus,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}, nil)
}

func (h *SpuHandler) Update(w http.ResponseWriter, r *http.Request) {
	spuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	var req dto.UpdateSpuRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.updateSpuCommand.Execute(r.Context(), command.UpdateSpuCommandInput{
		ID:          spuID,
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		Brand:       req.Brand,
		Images:      req.Images,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SpuResponse{
		ID:          output.ID,
		Title:       output.Title,
		Description: output.Description,
		CategoryID:  output.CategoryID,
		Brand:       output.Brand,
		Images:      output.Images,
		Status:      output.Status,
		UpdatedAt:   output.UpdatedAt,
	}, nil)
}

func (h *SpuHandler) Publish(w http.ResponseWriter, r *http.Request) {
	spuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.publishSpuCommand.Execute(r.Context(), command.PublishSpuCommandInput{ID: spuID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SpuResponse{
		ID:        output.ID,
		Status:    output.Status,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}

func (h *SpuHandler) OffShelf(w http.ResponseWriter, r *http.Request) {
	spuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.offShelfSpuCommand.Execute(r.Context(), command.OffShelfSpuCommandInput{ID: spuID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SpuResponse{
		ID:        output.ID,
		Status:    output.Status,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}
