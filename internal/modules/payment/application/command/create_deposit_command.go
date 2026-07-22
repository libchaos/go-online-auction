package command

import (
	"context"

	"github.com/google/uuid"

	"auction/internal/modules/payment/domain/errs"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
)

const defaultCurrency = "CNY"

type CreateDepositCommandInput struct {
	UserID         uint64
	AmountInCents  uint64
	Currency       string
	IdempotencyKey string
}

type CreateDepositCommandOutput struct {
	PaymentID  uint64
	OutTradeNo string
	QRCodeURL  string
	Status     string
}

type CreateDepositCommand struct {
	alipayPort ports.AlipayPort
	payments   ports.PaymentRepository
	alipayCfg  config.Alipay
	logger     logger.Logger
}

func NewCreateDepositCommand(
	alipayPort ports.AlipayPort,
	payments ports.PaymentRepository,
	alipayCfg config.Alipay,
	logger logger.Logger,
) *CreateDepositCommand {
	return &CreateDepositCommand{
		alipayPort: alipayPort,
		payments:   payments,
		alipayCfg:  alipayCfg,
		logger:     logger,
	}
}

// Execute opens an Alipay face-to-face prepay order and persists the recharge
// (user -> platform) order in the CREATED state. The QR code is returned so
// the client can render it for the user to scan.
func (command *CreateDepositCommand) Execute(
	ctx context.Context,
	input CreateDepositCommandInput,
) (CreateDepositCommandOutput, error) {
	currency := input.Currency
	if currency == "" {
		currency = defaultCurrency
	}

	if input.UserID == 0 {
		return CreateDepositCommandOutput{}, errs.ErrPaymentUserRequired
	}
	if input.AmountInCents == 0 {
		return CreateDepositCommandOutput{}, errs.ErrPaymentAmountRequired
	}

	outTradeNo := input.IdempotencyKey
	if outTradeNo == "" {
		outTradeNo = uuid.NewString()
	}

	notifyURL := command.alipayCfg.NotifyBaseURL + "/api/v1/payment/alipay/notify"

	precreate, precreateErr := command.alipayPort.CreateFaceToFacePayment(ctx, ports.FaceToFaceInput{
		OutTradeNo:    outTradeNo,
		Subject:       "账户充值",
		NotifyURL:     notifyURL,
		AmountInCents: input.AmountInCents,
		Currency:      currency,
	})
	if precreateErr != nil {
		command.logger.Error().Err(precreateErr).Uint64("user_id", input.UserID).
			Msg("failed to precreate alipay face-to-face payment")

		return CreateDepositCommandOutput{}, precreateErr
	}

	payment, buildErr := model.NewPayment(input.UserID, input.AmountInCents, currency, outTradeNo, precreate.QRCodeURL)
	if buildErr != nil {
		return CreateDepositCommandOutput{}, buildErr
	}

	persisted, saveErr := command.payments.Save(ctx, payment)
	if saveErr != nil {
		command.logger.Error().Err(saveErr).Uint64("user_id", input.UserID).
			Msg("failed to persist recharge order")

		return CreateDepositCommandOutput{}, saveErr
	}

	command.logger.Info().Uint64("payment_id", persisted.ID()).Uint64("user_id", input.UserID).
		Msg("recharge order created")

	return CreateDepositCommandOutput{
		PaymentID:  persisted.ID(),
		OutTradeNo: persisted.OutTradeNo(),
		QRCodeURL:  persisted.QRCodeURL(),
		Status:     string(persisted.Status()),
	}, nil
}
