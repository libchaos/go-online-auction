package command

import (
	"context"

	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/event"
	"auction/internal/modules/payment/domain/model"
	"auction/internal/modules/payment/infra/event/envelope"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/logger"
)

const (
	tradeStatusSuccess  = "TRADE_SUCCESS"
	tradeStatusFinished = "TRADE_FINISHED"
)

type AlipayNotifyCommandInput struct {
	Params map[string]string
}

type AlipayNotifyCommandOutput struct {
	PaymentID uint64
	Status    string
}

type AlipayNotifyCommand struct {
	alipayPort ports.AlipayPort
	uowFactory ports.PaymentUnitOfWorkFactory
	logger     logger.Logger
}

func NewAlipayNotifyCommand(
	alipayPort ports.AlipayPort,
	uowFactory ports.PaymentUnitOfWorkFactory,
	logger logger.Logger,
) *AlipayNotifyCommand {
	return &AlipayNotifyCommand{
		alipayPort: alipayPort,
		uowFactory: uowFactory,
		logger:     logger,
	}
}

// Execute verifies the asynchronous Alipay notification, advances the recharge
// order to SUCCESS/FAILED, and (on success) writes a payment-success outbox
// event. The deposit-success consumer credits the user's ledger from that
// event. Replays of an already-terminal payment are ignored to stay idempotent.
func (command *AlipayNotifyCommand) Execute(
	ctx context.Context,
	input AlipayNotifyCommandInput,
) (AlipayNotifyCommandOutput, error) {
	result, verifyErr := command.alipayPort.VerifyNotify(ctx, input.Params)
	if verifyErr != nil {
		return AlipayNotifyCommandOutput{}, verifyErr
	}

	unitOfWork, beginErr := command.uowFactory.Begin(ctx)
	if beginErr != nil {
		return AlipayNotifyCommandOutput{}, beginErr
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	payment, findErr := unitOfWork.PaymentRepository().FindByOutTradeNo(ctx, result.OutTradeNo)
	if findErr != nil {
		return AlipayNotifyCommandOutput{}, findErr
	}

	if payment.Status() != enum.PaymentStatusCreated {
		return AlipayNotifyCommandOutput{
			PaymentID: payment.ID(),
			Status:    string(payment.Status()),
		}, nil
	}

	paid := result.TradeStatus == tradeStatusSuccess || result.TradeStatus == tradeStatusFinished
	if markErr := command.applyOutcome(payment, paid, result.TradeNo); markErr != nil {
		return AlipayNotifyCommandOutput{}, markErr
	}

	persisted, commitErr := command.commitPayment(ctx, unitOfWork, payment, paid)
	if commitErr != nil {
		return AlipayNotifyCommandOutput{}, commitErr
	}

	command.logger.Info().Uint64("payment_id", persisted.ID()).Str("status", string(persisted.Status())).
		Msg("alipay notify processed")

	return AlipayNotifyCommandOutput{
		PaymentID: persisted.ID(),
		Status:    string(persisted.Status()),
	}, nil
}

// applyOutcome advances the payment to SUCCESS or FAILED based on the Alipay
// trade status. It fails if the payment is no longer in the CREATED state.
func (command *AlipayNotifyCommand) applyOutcome(payment model.PaymentModel, paid bool, tradeNo string) error {
	if paid {
		return payment.MarkSuccess(tradeNo)
	}

	return payment.MarkFailed()
}

// commitPayment persists the updated payment, and (on a successful trade)
// writes the payment-success outbox event, then commits the unit of work.
func (command *AlipayNotifyCommand) commitPayment(
	ctx context.Context,
	unitOfWork ports.PaymentUnitOfWork,
	payment model.PaymentModel,
	paid bool,
) (model.PaymentModel, error) {
	persisted, updateErr := unitOfWork.PaymentRepository().Update(ctx, payment)
	if updateErr != nil {
		return model.PaymentModel{}, updateErr
	}

	if paid {
		successEvent := event.NewPaymentSuccessEvent(
			persisted.ID(),
			persisted.UserID(),
			persisted.AmountInCents(),
			persisted.Currency(),
			persisted.OutTradeNo(),
		)
		outboxEvent, envelopeErr := envelope.ToDepositSuccessOutboxEvent(successEvent)
		if envelopeErr != nil {
			return model.PaymentModel{}, envelopeErr
		}
		if saveErr := unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
			return model.PaymentModel{}, saveErr
		}
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		return model.PaymentModel{}, completeErr
	}

	return persisted, nil
}
