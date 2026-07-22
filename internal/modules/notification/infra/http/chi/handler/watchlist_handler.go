package handler

import (
	"net/http"
	"strconv"
	"time"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/application/query"
	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/http/dto"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/httperrs"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
)

type WatchlistHandler struct {
	createWatchCommand *command.CreateWatchCommand
	deleteWatchCommand *command.DeleteWatchCommand
	listMyWatchesQuery *query.ListMyWatchesQuery
	logger             logger.Logger
}

func NewWatchlistHandler(
	createWatchCommand *command.CreateWatchCommand,
	deleteWatchCommand *command.DeleteWatchCommand,
	listMyWatchesQuery *query.ListMyWatchesQuery,
	logger logger.Logger,
) *WatchlistHandler {
	return &WatchlistHandler{
		createWatchCommand: createWatchCommand,
		deleteWatchCommand: deleteWatchCommand,
		listMyWatchesQuery: listMyWatchesQuery,
		logger:             logger,
	}
}

func (watchlistHandler *WatchlistHandler) CreateWatch(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	var req dto.CreateWatchRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrWatchInvalidRequest)
		return
	}

	if req.SpuID == 0 {
		response.Error(w, httperrs.ErrWatchInvalidRequest)
		return
	}

	output, err := watchlistHandler.createWatchCommand.Execute(r.Context(), command.CreateWatchCommandInput{
		UserID: claims.UserID,
		SpuID:  req.SpuID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, toWatchResponse(output.Watchlist), nil)
}

func (watchlistHandler *WatchlistHandler) DeleteWatch(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	spuID, parseErr := parseSpuIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrWatchInvalidRequest)
		return
	}

	err := watchlistHandler.deleteWatchCommand.Execute(r.Context(), command.DeleteWatchCommandInput{
		UserID: claims.UserID,
		SpuID:  spuID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (watchlistHandler *WatchlistHandler) ListMyWatches(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	limit := parseIntQuery(r, "limit")
	offset := parseIntQuery(r, "offset")

	output, err := watchlistHandler.listMyWatchesQuery.Execute(r.Context(), query.ListMyWatchesQueryInput{
		UserID: claims.UserID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	items := make([]dto.WatchResponse, 0, len(output.Watches))
	for index := range output.Watches {
		items = append(items, toWatchResponse(output.Watches[index]))
	}

	_ = response.JSON(w, http.StatusOK, dto.WatchListResponse{
		Items:  items,
		Limit:  limit,
		Offset: offset,
	}, nil)
}

func parseSpuIDParam(r *http.Request) (uint64, error) {
	idString := request.Param(r, "spuId")

	return strconv.ParseUint(idString, 10, 64)
}

func toWatchResponse(watchlist model.Watchlist) dto.WatchResponse {
	return dto.WatchResponse{
		WatchID:   watchlist.ID(),
		UserID:    watchlist.UserID(),
		SpuID:     watchlist.SpuID(),
		CreatedAt: watchlist.CreatedAt().Format(time.RFC3339),
	}
}
