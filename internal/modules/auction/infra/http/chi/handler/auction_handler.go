package handler

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/application/command"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/application/query"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/http/dto"
	httperrs "github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/http/errs"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/websocket"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/sdk/http/request"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/sdk/http/response"
	"github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
)

type AuctionHandler struct {
	createAuctionCommand *command.CreateAuctionCommand
	startAuctionCommand  *command.StartAuctionCommand
	cancelAuctionCommand *command.CancelAuctionCommand
	placeBidCommand      *command.PlaceBidCommand
	getAuctionByIDQuery  *query.GetAuctionByIDQuery
	listAuctionsQuery    *query.ListAuctionsQuery
	websocketHub         *websocket.Hub
	httpServer           *httpserver.Server
	logger               logger.Logger
}

func NewAuctionHandler(
	createAuctionCommand *command.CreateAuctionCommand,
	startAuctionCommand *command.StartAuctionCommand,
	cancelAuctionCommand *command.CancelAuctionCommand,
	placeBidCommand *command.PlaceBidCommand,
	getAuctionByIDQuery *query.GetAuctionByIDQuery,
	listAuctionsQuery *query.ListAuctionsQuery,
	websocketHub *websocket.Hub,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *AuctionHandler {
	return &AuctionHandler{
		createAuctionCommand: createAuctionCommand,
		startAuctionCommand:  startAuctionCommand,
		cancelAuctionCommand: cancelAuctionCommand,
		placeBidCommand:      placeBidCommand,
		getAuctionByIDQuery:  getAuctionByIDQuery,
		listAuctionsQuery:    listAuctionsQuery,
		websocketHub:         websocketHub,
		httpServer:           httpServer,
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
		ListingID: req.ListingID,
		EndTime:   req.EndTime,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.AuctionResponse{
		ID:        output.ID,
		ListingID: output.ListingID,
		State:     output.State,
		EndTime:   output.EndTime,
		CreatedAt: output.CreatedAt,
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
			StartTime:               auction.StartTime,
			EndTime:                 auction.EndTime,
			HighestBidAmountInCents: auction.HighestBidAmountInCents,
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
			ID:            bid.ID,
			AuctionID:     auctionID,
			UserID:        bid.UserID,
			AmountInCents: bid.AmountInCents,
			CreatedAt:     bid.CreatedAt,
		})
	}

	_ = response.JSON(w, http.StatusOK, dto.AuctionDetailResponse{
		Auction: dto.AuctionResponse{
			ID:                      output.Auction.ID,
			ListingID:               output.Auction.ListingID,
			State:                   output.Auction.State,
			StartTime:               output.Auction.StartTime,
			EndTime:                 output.Auction.EndTime,
			HighestBidAmountInCents: output.Auction.HighestBidAmountInCents,
			CreatedAt:               output.Auction.CreatedAt,
		},
		Bids: bids,
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

	// random user id (constrained to int64 max for PostgreSQL BIGINT compatibility)
	// Using crypto/rand to generate a secure random value in range [1, 100]
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(100))
	userID := uint64(randomNum.Int64() + 1)

	_, err = h.placeBidCommand.Execute(r.Context(), command.PlaceBidCommandInput{
		AuctionID:     auctionID,
		UserID:        userID,
		AmountInCents: req.AmountInCents,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}
