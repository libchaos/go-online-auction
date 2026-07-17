package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/application/query"
	"auction/internal/modules/auction/infra/http/dto"
	httperrs "auction/internal/modules/auction/infra/http/errs"
	"auction/internal/modules/auction/infra/messaging"
	"auction/internal/modules/auction/infra/websocket"
	depositports "auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

type AuctionHandler struct {
	createAuctionCommand *command.CreateAuctionCommand
	startAuctionCommand  *command.StartAuctionCommand
	cancelAuctionCommand *command.CancelAuctionCommand
	placeBidCommand      *command.PlaceBidCommand
	getAuctionByIDQuery  *query.GetAuctionByIDQuery
	listAuctionsQuery    *query.ListAuctionsQuery
	eventReplayer        messaging.EventReplayer
	websocketHub         *websocket.Hub
	httpServer           *httpserver.Server
	depositGuard         depositports.DepositGuard
	logger               logger.Logger
}

func NewAuctionHandler(
	createAuctionCommand *command.CreateAuctionCommand,
	startAuctionCommand *command.StartAuctionCommand,
	cancelAuctionCommand *command.CancelAuctionCommand,
	placeBidCommand *command.PlaceBidCommand,
	getAuctionByIDQuery *query.GetAuctionByIDQuery,
	listAuctionsQuery *query.ListAuctionsQuery,
	eventReplayer messaging.EventReplayer,
	websocketHub *websocket.Hub,
	httpServer *httpserver.Server,
	depositGuard depositports.DepositGuard,
	logger logger.Logger,
) *AuctionHandler {
	return &AuctionHandler{
		createAuctionCommand: createAuctionCommand,
		startAuctionCommand:  startAuctionCommand,
		cancelAuctionCommand: cancelAuctionCommand,
		placeBidCommand:      placeBidCommand,
		getAuctionByIDQuery:  getAuctionByIDQuery,
		listAuctionsQuery:    listAuctionsQuery,
		eventReplayer:        eventReplayer,
		websocketHub:         websocketHub,
		httpServer:           httpServer,
		depositGuard:         depositGuard,
		logger:               logger,
	}
}

func (h *AuctionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAuctionRequest
	if err := request.ReadJSON(w, r, &req); err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	output, err := h.createAuctionCommand.Execute(r.Context(), command.CreateAuctionCommandInput{
		ListingID:          req.ListingID,
		EndTime:            req.EndTime,
		TradingMode:        req.TradingMode,
		StartingPrice:      req.StartingPrice,
		PriceStep:          req.PriceStep,
		ReservePrice:       req.ReservePrice,
		AntiSnipeEnabled:   req.AntiSnipeEnabled,
		ExtensionWindowSec: req.ExtensionWindowSec,
		StartTime:          req.StartTime,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.AuctionResponse{
		ID:                 output.ID,
		ListingID:          output.ListingID,
		State:              output.State,
		TradingMode:        output.TradingMode,
		StartingPrice:      output.StartingPrice,
		PriceStep:          output.PriceStep,
		ReservePrice:       output.ReservePrice,
		AntiSnipeEnabled:   output.AntiSnipeEnabled,
		ExtensionWindowSec: output.ExtensionWindowSec,
		StartTime:          output.StartTime,
		EndTime:            output.EndTime,
		CreatedAt:          output.CreatedAt,
	}, nil)
}

func (h *AuctionHandler) List(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	var state *string
	if stateParam := queryParams.Get("state"); stateParam != "" {
		state = &stateParam
	}

	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	offset, _ := strconv.Atoi(queryParams.Get("offset"))

	output, err := h.listAuctionsQuery.Execute(r.Context(), query.ListAuctionsQueryInput{
		State:  state,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	auctions := make([]dto.AuctionResponse, 0, len(output.Auctions))
	for _, auction := range output.Auctions {
		auctions = append(auctions, dto.AuctionResponse{
			ID:                      auction.ID,
			ListingID:               auction.ListingID,
			State:                   auction.State,
			TradingMode:             auction.TradingMode,
			StartTime:               auction.StartTime,
			EndTime:                 auction.EndTime,
			StartingPrice:           auction.StartingPrice,
			ReservePrice:            auction.ReservePrice,
			CurrentPrice:            auction.CurrentPrice,
			HighestBidAmountInCents: auction.HighestBidAmountInCents,
			WinnerUserID:            auction.WinnerUserID,
			WinningBidAmountInCents: auction.WinningBidAmountInCents,
			AntiSnipeEnabled:        auction.AntiSnipeEnabled,
			ExtensionWindowSec:      auction.ExtensionWindowSec,
			CreatedAt:               auction.CreatedAt,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.AuctionListResponse{
		Auctions:   auctions,
		TotalCount: output.TotalCount,
		Limit:      output.Limit,
		Offset:     output.Offset,
	}, nil)
}

func (h *AuctionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idString := request.Param(r, "id")
	auctionID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidAuctionID)
		return
	}

	output, err := h.getAuctionByIDQuery.Execute(r.Context(), query.GetAuctionByIDQueryInput{
		ID: auctionID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	bids := make([]dto.BidResponse, 0, len(output.Bids))
	for _, bid := range output.Bids {
		bids = append(bids, dto.BidResponse{
			ID:               bid.ID,
			AuctionID:        auctionID,
			UserID:           bid.UserID,
			AmountInCents:    bid.AmountInCents,
			MaxAmountInCents: bid.MaxAmountInCents,
			CreatedAt:        bid.CreatedAt,
		})
	}

	detail := dto.AuctionResponse{
		ID:                      output.Auction.ID,
		ListingID:               output.Auction.ListingID,
		State:                   output.Auction.State,
		TradingMode:             output.Auction.TradingMode,
		StartTime:               output.Auction.StartTime,
		EndTime:                 output.Auction.EndTime,
		StartingPrice:           output.Auction.StartingPrice,
		PriceStep:               output.Auction.PriceStep,
		ReservePrice:            output.Auction.ReservePrice,
		CurrentPrice:            output.Auction.CurrentPrice,
		HighestBidAmountInCents: output.Auction.HighestBidAmountInCents,
		WinnerUserID:            output.Auction.WinnerUserID,
		WinningBidID:            output.Auction.WinningBidID,
		WinningBidAmountInCents: output.Auction.WinningBidAmountInCents,
		AntiSnipeEnabled:        output.Auction.AntiSnipeEnabled,
		ExtensionWindowSec:      output.Auction.ExtensionWindowSec,
		CreatedAt:               output.Auction.CreatedAt,
	}

	_ = response.JSON(w, http.StatusOK, dto.AuctionDetailResponse{
		Auction: detail,
		Bids:    bids,
	}, nil)
}

func (h *AuctionHandler) Start(w http.ResponseWriter, r *http.Request) {
	idString := request.Param(r, "id")
	auctionID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidAuctionID)
		return
	}

	output, err := h.startAuctionCommand.Execute(r.Context(), command.StartAuctionCommandInput{
		AuctionID: auctionID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.AuctionResponse{
		ID:        output.ID,
		ListingID: output.ListingID,
		State:     output.State,
		StartTime: output.StartTime,
		EndTime:   output.EndTime,
	}, nil)
}

func (h *AuctionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	idString := request.Param(r, "id")
	auctionID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidAuctionID)
		return
	}

	output, err := h.cancelAuctionCommand.Execute(r.Context(), command.CancelAuctionCommandInput{
		AuctionID: auctionID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.AuctionResponse{
		ID:        output.ID,
		ListingID: output.ListingID,
		State:     output.State,
		StartTime: output.StartTime,
		EndTime:   output.EndTime,
	}, nil)
}

func (h *AuctionHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	idString := request.Param(r, "id")
	auctionID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidAuctionID)
		return
	}

	var req dto.PlaceBidRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	if guardErr := h.depositGuard.EnsureEligible(r.Context(), claims.UserID, auctionID); guardErr != nil {
		response.Error(w, httperrs.MapDomainError(guardErr))
		return
	}

	output, err := h.placeBidCommand.Execute(r.Context(), command.PlaceBidCommandInput{
		AuctionID:        auctionID,
		UserID:           claims.UserID,
		AmountInCents:    req.AmountInCents,
		MaxAmountInCents: req.MaxAmountInCents,
		IdempotencyKey:   r.Header.Get("Idempotency-Key"),
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusAccepted, dto.PlaceBidAcceptedResponse{
		IdempotencyKey: output.IdempotencyKey,
		Status:         output.Status,
	}, nil)
}

// Events replays the persisted event history of an auction from the event
// store. Optional `from`/`until` (RFC3339) narrow the window (temporal query);
// `limit` caps the number of returned events.
func (h *AuctionHandler) Events(w http.ResponseWriter, r *http.Request) {
	idString := request.Param(r, "id")
	auctionID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidAuctionID)
		return
	}

	filter, err := parseReplayFilter(r)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidRequest)
		return
	}

	events, err := h.eventReplayer.ReplayAuction(r.Context(), auctionID, filter)
	if err != nil {
		h.logger.Error().Err(err).Uint64("auction_id", auctionID).Msg("failed to replay auction events")
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	items := make([]dto.AuctionEventResponse, 0, len(events))
	for _, env := range events {
		items = append(items, dto.AuctionEventResponse{
			EventType:     env.EventType,
			EventID:       env.EventID,
			SchemaVersion: env.SchemaVersion,
			Timestamp:     env.Timestamp,
			AuctionID:     env.AuctionID,
			Data:          env.Data,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.AuctionEventListResponse{
		Events:     items,
		TotalCount: len(items),
	}, nil)
}

func parseReplayFilter(r *http.Request) (messaging.ReplayFilter, error) {
	var filter messaging.ReplayFilter

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return messaging.ReplayFilter{}, err
		}
		filter.From = from
	}

	if untilStr := r.URL.Query().Get("until"); untilStr != "" {
		until, err := time.Parse(time.RFC3339, untilStr)
		if err != nil {
			return messaging.ReplayFilter{}, err
		}
		filter.Until = until
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return messaging.ReplayFilter{}, fmt.Errorf("invalid limit %q", limitStr)
		}
		filter.Limit = limit
	}

	return filter, nil
}
