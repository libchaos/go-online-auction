package handler

import (
	"net/http"
	"strconv"
	"time"

	"auction/internal/modules/deposit/application/command"
	"auction/internal/modules/deposit/application/query"
	"auction/internal/modules/deposit/infra/http/dto"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/httperrs"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

const defaultCurrency = "CNY"

type DepositHandler struct {
	createCommand    *command.CreateDepositCommand
	getQuery         *query.GetDepositQuery
	listQuery        *query.ListDepositsByUserQuery
	eligibilityQuery *query.GetEligibilityQuery
	releaseCommand   *command.ReleaseDepositCommand
	applyCommand     *command.ApplyDepositCommand
	forfeitCommand   *command.ForfeitDepositCommand
	cancelCommand    *command.CancelDepositCommand
	httpServer       *httpserver.Server
	logger           logger.Logger
}

func NewDepositHandler(
	createCommand *command.CreateDepositCommand,
	getQuery *query.GetDepositQuery,
	listQuery *query.ListDepositsByUserQuery,
	eligibilityQuery *query.GetEligibilityQuery,
	releaseCommand *command.ReleaseDepositCommand,
	applyCommand *command.ApplyDepositCommand,
	forfeitCommand *command.ForfeitDepositCommand,
	cancelCommand *command.CancelDepositCommand,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *DepositHandler {
	return &DepositHandler{
		createCommand:    createCommand,
		getQuery:         getQuery,
		listQuery:        listQuery,
		eligibilityQuery: eligibilityQuery,
		releaseCommand:   releaseCommand,
		applyCommand:     applyCommand,
		forfeitCommand:   forfeitCommand,
		cancelCommand:    cancelCommand,
		httpServer:       httpServer,
		logger:           logger,
	}
}

func (depositHandler *DepositHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDepositRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	currency := req.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	output, err := depositHandler.createCommand.Execute(r.Context(), command.CreateDepositCommandInput{
		UserID:        claims.UserID,
		AuctionID:     req.AuctionID,
		AmountInCents: req.AmountInCents,
		Currency:      currency,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.CreateDepositResponse{
		DepositID: output.DepositID,
		Status:    output.Status,
		AccountID: output.AccountID,
	}, nil)
}

func (depositHandler *DepositHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	depositID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	output, err := depositHandler.getQuery.Execute(r.Context(), query.GetDepositQueryInput{DepositID: depositID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toDepositResponse(output.Deposit), nil)
}

func (depositHandler *DepositHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	output, err := depositHandler.listQuery.Execute(
		r.Context(),
		query.ListDepositsByUserQueryInput{UserID: claims.UserID},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	deposits := make([]dto.DepositResponse, 0, len(output.Deposits))
	for _, view := range output.Deposits {
		deposits = append(deposits, toDepositResponse(view))
	}

	_ = response.JSON(w, http.StatusOK, dto.ListDepositsResponse{Deposits: deposits}, nil)
}

func (depositHandler *DepositHandler) GetEligibility(w http.ResponseWriter, r *http.Request) {
	auctionID, parseErr := strconv.ParseUint(r.URL.Query().Get("auction_id"), 10, 64)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	output, err := depositHandler.eligibilityQuery.Execute(r.Context(), query.GetEligibilityQueryInput{
		UserID:    claims.UserID,
		AuctionID: auctionID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, dto.EligibilityResponse{Eligible: output.Eligible}, nil)
}

func (depositHandler *DepositHandler) Release(w http.ResponseWriter, r *http.Request) {
	depositID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	output, err := depositHandler.releaseCommand.Execute(
		r.Context(),
		command.ReleaseDepositCommandInput{DepositID: depositID},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, dto.ActionDepositResponse{
		DepositID: output.DepositID,
		Status:    output.Status,
	}, nil)
}

func (depositHandler *DepositHandler) Forfeit(w http.ResponseWriter, r *http.Request) {
	depositID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	output, err := depositHandler.forfeitCommand.Execute(
		r.Context(),
		command.ForfeitDepositCommandInput{DepositID: depositID},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, dto.ActionDepositResponse{
		DepositID: output.DepositID,
		Status:    output.Status,
	}, nil)
}

func (depositHandler *DepositHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	depositID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	output, err := depositHandler.cancelCommand.Execute(
		r.Context(),
		command.CancelDepositCommandInput{DepositID: depositID},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, dto.ActionDepositResponse{
		DepositID: output.DepositID,
		Status:    output.Status,
	}, nil)
}

func (depositHandler *DepositHandler) Apply(w http.ResponseWriter, r *http.Request) {
	depositID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	var req dto.ApplyDepositRequest
	_ = request.ReadJSON(w, r, &req)

	output, err := depositHandler.applyCommand.Execute(r.Context(), command.ApplyDepositCommandInput{
		DepositID:            depositID,
		CaptureAmountInCents: req.CaptureAmountInCents,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, dto.ActionDepositResponse{
		DepositID: output.DepositID,
		Status:    output.Status,
	}, nil)
}

func parseIDParam(r *http.Request) (uint64, error) {
	idString := request.Param(r, "id")

	return strconv.ParseUint(idString, 10, 64)
}

func toDepositResponse(view query.DepositView) dto.DepositResponse {
	return dto.DepositResponse{
		DepositID:         view.DepositID,
		UserID:            view.UserID,
		AuctionID:         view.AuctionID,
		AmountInCents:     view.AmountInCents,
		Currency:          view.Currency,
		Status:            view.Status,
		ExternalReference: view.ExternalReference,
		Reference:         view.Reference,
		Version:           view.Version,
		CreatedAt:         view.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         view.UpdatedAt.Format(time.RFC3339),
	}
}
