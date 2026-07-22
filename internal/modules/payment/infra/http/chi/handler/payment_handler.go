package handler

import (
	"net/http"
	"strconv"
	"time"

	"auction/internal/modules/payment/application/command"
	"auction/internal/modules/payment/application/query"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/infra/http/dto"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/httperrs"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

type PaymentHandler struct {
	createDepositCommand    *command.CreateDepositCommand
	getDepositQuery         *query.GetDepositQuery
	createWithdrawalCommand *command.CreateWithdrawalCommand
	getWithdrawalQuery      *query.GetWithdrawalQuery
	httpServer              *httpserver.Server
	logger                  logger.Logger
}

func NewPaymentHandler(
	createDepositCommand *command.CreateDepositCommand,
	getDepositQuery *query.GetDepositQuery,
	createWithdrawalCommand *command.CreateWithdrawalCommand,
	getWithdrawalQuery *query.GetWithdrawalQuery,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		createDepositCommand:    createDepositCommand,
		getDepositQuery:         getDepositQuery,
		createWithdrawalCommand: createWithdrawalCommand,
		getWithdrawalQuery:      getWithdrawalQuery,
		httpServer:              httpServer,
		logger:                  logger,
	}
}

func (paymentHandler *PaymentHandler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDepositRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrPaymentInvalidRequest)
		return
	}

	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	output, err := paymentHandler.createDepositCommand.Execute(r.Context(), command.CreateDepositCommandInput{
		UserID:        claims.UserID,
		AmountInCents: req.AmountInCents,
		Currency:      req.Currency,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.CreateDepositResponse{
		PaymentID:  output.PaymentID,
		OutTradeNo: output.OutTradeNo,
		QRCodeURL:  output.QRCodeURL,
		Status:     output.Status,
	}, nil)
}

func (paymentHandler *PaymentHandler) GetDeposit(w http.ResponseWriter, r *http.Request) {
	paymentID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrPaymentInvalidRequest)
		return
	}

	output, err := paymentHandler.getDepositQuery.Execute(r.Context(), query.GetDepositQueryInput{PaymentID: paymentID})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, toDepositResponse(output.Payment), nil)
}

func (paymentHandler *PaymentHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWithdrawalRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrPaymentInvalidRequest)
		return
	}

	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	output, err := paymentHandler.createWithdrawalCommand.Execute(r.Context(), command.CreateWithdrawalCommandInput{
		UserID:         claims.UserID,
		AlipayAccount:  req.AlipayAccount,
		AlipayRealName: req.AlipayRealName,
		AmountInCents:  req.AmountInCents,
		Currency:       req.Currency,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusCreated, dto.CreateWithdrawalResponse{
		WithdrawalID: output.WithdrawalID,
		OutBizNo:     output.OutBizNo,
		Status:       output.Status,
	}, nil)
}

func (paymentHandler *PaymentHandler) GetWithdrawal(w http.ResponseWriter, r *http.Request) {
	withdrawalID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrPaymentInvalidRequest)
		return
	}

	output, err := paymentHandler.getWithdrawalQuery.Execute(
		r.Context(),
		query.GetWithdrawalQueryInput{WithdrawalID: withdrawalID},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, toWithdrawalResponse(output.Withdrawal), nil)
}

func parseIDParam(r *http.Request) (uint64, error) {
	idString := request.Param(r, "id")

	return strconv.ParseUint(idString, 10, 64)
}

func toDepositResponse(payment model.PaymentModel) dto.DepositResponse {
	return dto.DepositResponse{
		PaymentID:     payment.ID(),
		UserID:        payment.UserID(),
		AmountInCents: payment.AmountInCents(),
		Currency:      payment.Currency(),
		Status:        string(payment.Status()),
		OutTradeNo:    payment.OutTradeNo(),
		QRCodeURL:     payment.QRCodeURL(),
		AlipayTradeNo: payment.AlipayTradeNo(),
		Version:       payment.Version(),
		CreatedAt:     payment.CreatedAt().Format(time.RFC3339),
		UpdatedAt:     payment.UpdatedAt().Format(time.RFC3339),
	}
}

func toWithdrawalResponse(withdrawal model.WithdrawalModel) dto.WithdrawalResponse {
	return dto.WithdrawalResponse{
		WithdrawalID:    withdrawal.ID(),
		UserID:          withdrawal.UserID(),
		LedgerAccountID: withdrawal.LedgerAccountID(),
		AlipayAccount:   withdrawal.AlipayAccount(),
		AlipayRealName:  withdrawal.AlipayRealName(),
		AmountInCents:   withdrawal.AmountInCents(),
		Currency:        withdrawal.Currency(),
		Status:          string(withdrawal.Status()),
		OutBizNo:        withdrawal.OutBizNo(),
		AlipayOrderID:   withdrawal.AlipayOrderID(),
		FailReason:      withdrawal.FailReason(),
		Version:         withdrawal.Version(),
		CreatedAt:       withdrawal.CreatedAt().Format(time.RFC3339),
		UpdatedAt:       withdrawal.UpdatedAt().Format(time.RFC3339),
	}
}
