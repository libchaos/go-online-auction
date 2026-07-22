package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/modules/payment/domain/enum"
	"auction/internal/modules/payment/domain/event"
	"auction/internal/modules/payment/infra/event/envelope"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	sharednats "auction/internal/shared/modules/nats"
)

// WithdrawalConsumer runs the payout Saga. On a successful Alipay transfer it
// permanently withdraws the frozen balance; on failure it compensates by
// unfreezing the reserved funds. Both ledger actions carry derived idempotency
// keys (out_biz_no + ":withdraw" / ":unfreeze") so replays are safe.
type WithdrawalConsumer struct {
	js         jetstream.JetStream
	alipayPort ports.AlipayPort
	uowFactory ports.PaymentUnitOfWorkFactory
	alipayCfg  config.Alipay
	logger     logger.Logger
	consumeCtx jetstream.ConsumeContext
}

func NewWithdrawalConsumer(
	js jetstream.JetStream,
	alipayPort ports.AlipayPort,
	uowFactory ports.PaymentUnitOfWorkFactory,
	alipayCfg config.Alipay,
	logger logger.Logger,
) *WithdrawalConsumer {
	return &WithdrawalConsumer{
		js:         js,
		alipayPort: alipayPort,
		uowFactory: uowFactory,
		alipayCfg:  alipayCfg,
		logger:     logger,
	}
}

func (consumer *WithdrawalConsumer) Start(ctx context.Context) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(
		ctx,
		sharednats.StreamPaymentEvents,
		jetstream.ConsumerConfig{
			FilterSubject: event.SubjectWithdrawalRequested,
			DeliverPolicy: jetstream.DeliverNewPolicy,
			AckPolicy:     jetstream.AckNonePolicy,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal consumer: %w", err)
	}

	consumeCtx, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		consumer.handle(ctx, msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start withdrawal consumer: %w", err)
	}

	consumer.consumeCtx = consumeCtx

	return nil
}

func (consumer *WithdrawalConsumer) Stop() {
	if consumer.consumeCtx != nil {
		consumer.consumeCtx.Drain()
	}
}

func (consumer *WithdrawalConsumer) handle(ctx context.Context, data []byte) {
	var payload envelope.WithdrawalRequestedPayload
	if unmarshalErr := json.Unmarshal(data, &payload); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode withdrawal requested payload")
		return
	}

	// Phase 1: short read to skip orders that are already terminal (idempotent
	// replay) before performing any external call.
	readUoW, beginErr := consumer.uowFactory.Begin(ctx)
	if beginErr != nil {
		consumer.logger.Error().Err(beginErr).Msg("failed to begin read uow for withdrawal")
		return
	}
	withdrawal, findErr := readUoW.WithdrawalRepository().FindByOutBizNo(ctx, payload.OutBizNo)
	if rollbackErr := readUoW.Rollback(ctx); rollbackErr != nil {
		consumer.logger.Error().Err(rollbackErr).Msg("failed to rollback read uow")
	}
	if findErr != nil {
		consumer.logger.Error().Err(findErr).Str("out_biz_no", payload.OutBizNo).
			Msg("failed to find withdrawal order")
		return
	}
	if withdrawal.Status() != enum.WithdrawalStatusFrozen {
		consumer.logger.Info().Str("out_biz_no", payload.OutBizNo).Str("status", string(withdrawal.Status())).
			Msg("withdrawal already terminal; skipping")
		return
	}

	// Phase 2: call Alipay (idempotent via OutBizNo) outside the DB transaction.
	transfer, transferErr := consumer.alipayPort.TransferToAlipayAccount(ctx, ports.TransferInput{
		OutBizNo:      payload.OutBizNo,
		PayeeAccount:  payload.AlipayAccount,
		PayeeRealName: payload.AlipayRealName,
		AmountInCents: payload.AmountInCents,
		Currency:      payload.Currency,
		Remark:        "账户提现",
	})

	// Phase 3: apply the Saga outcome atomically.
	if sagaErr := consumer.applyPayoutSaga(ctx, payload, transfer, transferErr); sagaErr != nil {
		return
	}

	consumer.logger.Info().Str("out_biz_no", payload.OutBizNo).Bool("success", transferErr == nil).
		Msg("withdrawal saga completed")
}

// applyPayoutSaga runs the write phase of the payout Saga: it opens a unit of
// work, resolves the platform ledger account, then either permanently withdraws
// the frozen balance (successful Alipay transfer) or compensates by unfreezing
// it (failed transfer), and finally marks the order and commits.
func (consumer *WithdrawalConsumer) applyPayoutSaga(
	ctx context.Context,
	payload envelope.WithdrawalRequestedPayload,
	transfer ports.TransferOutput,
	transferErr error,
) error {
	writeUoW, beginErr := consumer.uowFactory.Begin(ctx)
	if beginErr != nil {
		consumer.logger.Error().Err(beginErr).Msg("failed to begin write uow for withdrawal")
		return beginErr
	}
	defer func() { _ = writeUoW.Rollback(ctx) }()

	ledger := writeUoW.LedgerRepository()

	platformAccount, platformErr := ledger.GetOrCreateAccountByOwner(
		ctx,
		consumer.alipayCfg.PlatformAccountOwner,
		payload.Currency,
	)
	if platformErr != nil {
		consumer.logger.Error().Err(platformErr).Msg("failed to resolve platform ledger account")
		return platformErr
	}

	if transferErr == nil {
		if _, withdrawErr := ledger.WithdrawFromFrozen(ctx, ledgerports.WithdrawFromFrozenInput{
			AccountID:             payload.LedgerAccountID,
			CounterpartyAccountID: platformAccount.ID(),
			Amount:                payload.AmountInCents,
			IdempotencyKey:        payload.OutBizNo + ":withdraw",
			Reference:             payload.OutBizNo,
			Description:           "withdrawal payout " + payload.OutBizNo,
		}); withdrawErr != nil {
			consumer.logger.Error().Err(withdrawErr).Str("out_biz_no", payload.OutBizNo).
				Msg("failed to withdraw frozen balance")
			return withdrawErr
		}
	} else {
		consumer.logger.Error().Err(transferErr).Str("out_biz_no", payload.OutBizNo).
			Msg("alipay payout failed; compensating by unfreezing")

		if _, unfreezeErr := ledger.Unfreeze(ctx, ledgerports.UnfreezeInput{
			AccountID:      payload.LedgerAccountID,
			Amount:         payload.AmountInCents,
			IdempotencyKey: payload.OutBizNo + ":unfreeze",
			Reference:      payload.OutBizNo,
			Description:    "withdrawal compensation " + payload.OutBizNo,
		}); unfreezeErr != nil {
			consumer.logger.Error().Err(unfreezeErr).Str("out_biz_no", payload.OutBizNo).
				Msg("failed to unfreeze withdrawn balance")
			return unfreezeErr
		}
	}

	if markErr := consumer.markWithdrawal(ctx, writeUoW, payload.OutBizNo, transfer, transferErr); markErr != nil {
		return markErr
	}

	if eventErr := consumer.writeOutcomeEvent(ctx, writeUoW, payload, transfer, transferErr); eventErr != nil {
		return eventErr
	}

	if completeErr := writeUoW.Complete(ctx); completeErr != nil {
		consumer.logger.Error().Err(completeErr).Msg("failed to commit withdrawal saga")
		return completeErr
	}

	return nil
}

func (consumer *WithdrawalConsumer) markWithdrawal(
	ctx context.Context,
	unitOfWork ports.PaymentUnitOfWork,
	outBizNo string,
	transfer ports.TransferOutput,
	transferErr error,
) error {
	reload, reloadErr := unitOfWork.WithdrawalRepository().FindByOutBizNo(ctx, outBizNo)
	if reloadErr != nil {
		consumer.logger.Error().Err(reloadErr).Msg("failed to reload withdrawal for status mark")
		return reloadErr
	}

	if transferErr == nil {
		if markErr := reload.MarkSuccess(transfer.AlipayOrderID); markErr != nil {
			consumer.logger.Error().Err(markErr).Msg("failed to mark withdrawal success")
			return markErr
		}
	} else {
		if markErr := reload.MarkFailed(transferErr.Error()); markErr != nil {
			consumer.logger.Error().Err(markErr).Msg("failed to mark withdrawal failed")
			return markErr
		}
	}

	if _, updateErr := unitOfWork.WithdrawalRepository().Update(ctx, reload); updateErr != nil {
		consumer.logger.Error().Err(updateErr).Msg("failed to persist withdrawal status")
		return updateErr
	}

	return nil
}

// writeOutcomeEvent enqueues the terminal withdrawal event (completed or
// failed) into the payment outbox within the Saga transaction, so downstream
// consumers such as notifications are driven reliably once the outcome commits.
func (consumer *WithdrawalConsumer) writeOutcomeEvent(
	ctx context.Context,
	unitOfWork ports.PaymentUnitOfWork,
	payload envelope.WithdrawalRequestedPayload,
	transfer ports.TransferOutput,
	transferErr error,
) error {
	var outboxEvent ports.OutboxEvent
	var conversionErr error

	if transferErr == nil {
		outboxEvent, conversionErr = envelope.ToWithdrawalCompletedOutboxEvent(event.NewWithdrawalCompletedEvent(
			payload.WithdrawalID,
			payload.UserID,
			payload.AmountInCents,
			payload.Currency,
			payload.OutBizNo,
			transfer.AlipayOrderID,
		))
	} else {
		outboxEvent, conversionErr = envelope.ToWithdrawalFailedOutboxEvent(event.NewWithdrawalFailedEvent(
			payload.WithdrawalID,
			payload.UserID,
			payload.AmountInCents,
			payload.Currency,
			payload.OutBizNo,
			transferErr.Error(),
		))
	}
	if conversionErr != nil {
		consumer.logger.Error().Err(conversionErr).Str("out_biz_no", payload.OutBizNo).
			Msg("failed to build withdrawal outcome event")
		return conversionErr
	}

	if saveErr := unitOfWork.OutboxRepository().Save(ctx, outboxEvent); saveErr != nil {
		consumer.logger.Error().Err(saveErr).Str("out_biz_no", payload.OutBizNo).
			Msg("failed to enqueue withdrawal outcome event")
		return saveErr
	}

	return nil
}
