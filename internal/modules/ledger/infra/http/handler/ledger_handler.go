package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"auction/internal/modules/ledger/domain/model"
	"auction/internal/modules/ledger/infra/http/dto"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/httperrs"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

const defaultCurrency = "CNY"

const idempotencyHeader = "Idempotency-Key"

type LedgerHandler struct {
	service    ledgerports.LedgerAccountService
	httpServer *httpserver.Server
	logger     logger.Logger
}

func NewLedgerHandler(
	service ledgerports.LedgerAccountService,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *LedgerHandler {
	return &LedgerHandler{
		service:    service,
		httpServer: httpServer,
		logger:     logger,
	}
}

func (ledgerHandler *LedgerHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAccountRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	owner := req.Owner
	if owner == "" {
		owner = strconv.FormatUint(claims.UserID, 10)
	}

	currency := req.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	account, err := ledgerHandler.service.CreateAccount(r.Context(), owner, currency)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusCreated, toAccountResponse(account), nil)
}

func (ledgerHandler *LedgerHandler) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	accountID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	account, err := ledgerHandler.service.GetAccountByID(r.Context(), accountID)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toAccountResponse(account), nil)
}

func (ledgerHandler *LedgerHandler) GetAccountByOwner(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	owner := r.URL.Query().Get("owner")
	if owner == "" {
		owner = strconv.FormatUint(claims.UserID, 10)
	}

	currency := r.URL.Query().Get("currency")
	if currency == "" {
		currency = defaultCurrency
	}

	account, err := ledgerHandler.service.GetAccountByOwner(r.Context(), owner, currency)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toAccountResponse(account), nil)
}

func (ledgerHandler *LedgerHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req dto.TransferRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	input := ledgerports.TransferInput{
		FromAccountID:  req.FromAccountID,
		ToAccountID:    req.ToAccountID,
		Amount:         req.Amount,
		IdempotencyKey: resolveIdempotencyKey(r, req.IdempotencyKey),
		Reference:      req.Reference,
		Description:    req.Description,
	}

	result, err := ledgerHandler.service.Transfer(r.Context(), input)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toTransferResponse(result), nil)
}

func (ledgerHandler *LedgerHandler) Freeze(w http.ResponseWriter, r *http.Request) {
	var req dto.FreezeRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	input := ledgerports.FreezeInput{
		AccountID:      req.AccountID,
		Amount:         req.Amount,
		IdempotencyKey: resolveIdempotencyKey(r, req.IdempotencyKey),
		Reference:      req.Reference,
		Description:    req.Description,
	}

	result, err := ledgerHandler.service.Freeze(r.Context(), input)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toOperationResponse(result), nil)
}

func (ledgerHandler *LedgerHandler) Unfreeze(w http.ResponseWriter, r *http.Request) {
	var req dto.UnfreezeRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	input := ledgerports.UnfreezeInput{
		AccountID:      req.AccountID,
		Amount:         req.Amount,
		IdempotencyKey: resolveIdempotencyKey(r, req.IdempotencyKey),
		Reference:      req.Reference,
		Description:    req.Description,
	}

	result, err := ledgerHandler.service.Unfreeze(r.Context(), input)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toOperationResponse(result), nil)
}

func (ledgerHandler *LedgerHandler) WithdrawFromFrozen(w http.ResponseWriter, r *http.Request) {
	var req dto.WithdrawFromFrozenRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrDepositInvalidRequest)

		return
	}

	input := ledgerports.WithdrawFromFrozenInput{
		AccountID:             req.AccountID,
		CounterpartyAccountID: req.CounterpartyAccountID,
		Amount:                req.Amount,
		IdempotencyKey:        resolveIdempotencyKey(r, req.IdempotencyKey),
		Reference:             req.Reference,
		Description:           req.Description,
	}

	result, err := ledgerHandler.service.WithdrawFromFrozen(r.Context(), input)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))

		return
	}

	_ = response.JSON(w, http.StatusOK, toOperationResponse(result), nil)
}

func resolveIdempotencyKey(r *http.Request, bodyKey string) string {
	if headerKey := r.Header.Get(idempotencyHeader); headerKey != "" {
		return headerKey
	}

	if bodyKey != "" {
		return bodyKey
	}

	return uuid.NewString()
}

func parseIDParam(r *http.Request) (uint64, error) {
	idString := request.Param(r, "id")

	return strconv.ParseUint(idString, 10, 64)
}

func toAccountResponse(account model.AccountModel) dto.AccountResponse {
	return dto.AccountResponse{
		AccountID:     account.ID(),
		Owner:         account.Owner(),
		Balance:       account.Balance(),
		FrozenBalance: account.FrozenBalance(),
		Currency:      account.Currency(),
		Version:       account.Version(),
		CreatedAt:     account.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     account.UpdatedAt().Format(time.RFC3339),
	}
}

func toOperationResponse(operation model.OperationModel) dto.OperationResponse {
	operationType := operation.OperationType()
	status := operation.Status()

	return dto.OperationResponse{
		OperationID:           operation.ID(),
		AccountID:             operation.AccountID(),
		CounterpartyAccountID: operation.CounterpartyAccountID(),
		OperationType:         operationType.String(),
		Amount:                operation.Amount(),
		IdempotencyKey:        operation.IdempotencyKey(),
		Status:                status.String(),
		Reference:             operation.Reference(),
		Description:           operation.Description(),
		CreatedAt:             operation.CreatedAt().Format(time.RFC3339),
		UpdatedAt:             operation.UpdatedAt().Format(time.RFC3339),
	}
}

func toTransferResponse(transfer model.TransferModel) dto.TransferResponse {
	return dto.TransferResponse{
		TransferID:     transfer.ID(),
		FromAccountID:  transfer.FromAccountID(),
		ToAccountID:    transfer.ToAccountID(),
		Amount:         transfer.Amount(),
		IdempotencyKey: transfer.IdempotencyKey(),
		CreatedAt:      transfer.CreatedAt().Format(time.RFC3339),
	}
}
