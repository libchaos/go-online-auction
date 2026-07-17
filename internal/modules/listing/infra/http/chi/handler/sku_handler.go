package handler

import (
	"net/http"

	"auction/internal/modules/listing/application/command"
	"auction/internal/modules/listing/application/query"
	"auction/internal/modules/listing/infra/http/dto"
	httperrs "auction/internal/modules/listing/infra/http/errs"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
)

type SkuHandler struct {
	createSkuCommand   *command.CreateSkuCommand
	updateSkuCommand   *command.UpdateSkuCommand
	publishSkuCommand  *command.PublishSkuCommand
	offShelfSkuCommand *command.OffShelfSkuCommand
	getSkuByIDQuery    *query.GetSkuByIDQuery
}

func NewSkuHandler(
	createSkuCommand *command.CreateSkuCommand,
	updateSkuCommand *command.UpdateSkuCommand,
	publishSkuCommand *command.PublishSkuCommand,
	offShelfSkuCommand *command.OffShelfSkuCommand,
	getSkuByIDQuery *query.GetSkuByIDQuery,
) *SkuHandler {
	return &SkuHandler{
		createSkuCommand:   createSkuCommand,
		updateSkuCommand:   updateSkuCommand,
		publishSkuCommand:  publishSkuCommand,
		offShelfSkuCommand: offShelfSkuCommand,
		getSkuByIDQuery:    getSkuByIDQuery,
	}
}

// Create handles POST /api/v1/spus/{id}/skus; the SPU ID comes from the path.
func (h *SkuHandler) Create(w http.ResponseWriter, r *http.Request) {
	spuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	var req dto.CreateSkuRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.createSkuCommand.Execute(r.Context(), command.CreateSkuCommandInput{
		SpuID:        spuID,
		SpecValues:   req.SpecValues,
		PriceInCents: req.PriceInCents,
		Quantity:     req.Quantity,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.SkuResponse{
		ID:           output.ID,
		SpuID:        output.SpuID,
		SpecValues:   output.SpecValues,
		PriceInCents: output.PriceInCents,
		Quantity:     output.Quantity,
		Status:       output.Status,
		CreatedAt:    output.CreatedAt,
	}, nil)
}

func (h *SkuHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	skuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.getSkuByIDQuery.Execute(r.Context(), query.GetSkuByIDQueryInput{ID: skuID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SkuResponse{
		ID:           output.Sku.ID,
		SpuID:        output.Sku.SpuID,
		SpecValues:   output.Sku.SpecValues,
		PriceInCents: output.Sku.PriceInCents,
		Quantity:     output.Sku.Quantity,
		Status:       output.Sku.Status,
		CreatedAt:    output.Sku.CreatedAt,
		UpdatedAt:    output.Sku.UpdatedAt,
	}, nil)
}

func (h *SkuHandler) Update(w http.ResponseWriter, r *http.Request) {
	skuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	var req dto.UpdateSkuRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.updateSkuCommand.Execute(r.Context(), command.UpdateSkuCommandInput{
		ID:           skuID,
		SpecValues:   req.SpecValues,
		PriceInCents: req.PriceInCents,
		Quantity:     req.Quantity,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SkuResponse{
		ID:           output.ID,
		SpuID:        output.SpuID,
		SpecValues:   output.SpecValues,
		PriceInCents: output.PriceInCents,
		Quantity:     output.Quantity,
		Status:       output.Status,
		UpdatedAt:    output.UpdatedAt,
	}, nil)
}

func (h *SkuHandler) Publish(w http.ResponseWriter, r *http.Request) {
	skuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.publishSkuCommand.Execute(r.Context(), command.PublishSkuCommandInput{ID: skuID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SkuResponse{
		ID:        output.ID,
		SpuID:     output.SpuID,
		Status:    output.Status,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}

func (h *SkuHandler) OffShelf(w http.ResponseWriter, r *http.Request) {
	skuID, err := parseIDParam(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidID)
		return
	}

	output, err := h.offShelfSkuCommand.Execute(r.Context(), command.OffShelfSkuCommandInput{ID: skuID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.SkuResponse{
		ID:        output.ID,
		SpuID:     output.SpuID,
		Status:    output.Status,
		UpdatedAt: output.UpdatedAt,
	}, nil)
}
