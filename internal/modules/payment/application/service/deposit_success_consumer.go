package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/nats-io/nats.go/jetstream"

	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/modules/payment/domain/event"
	"auction/internal/modules/payment/infra/event/envelope"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	sharednats "auction/internal/shared/modules/nats"
)

// DepositSuccessConsumer credits the user's ledger account when a recharge is
// confirmed paid. The Transfer uses the out_trade_no as its ledger
// idempotency key, so duplicate deliveries (relay redelivery) never double
// credit.
type DepositSuccessConsumer struct {
	js         jetstream.JetStream
	uowFactory ports.PaymentUnitOfWorkFactory
	alipayCfg  config.Alipay
	logger     logger.Logger
	consumeCtx jetstream.ConsumeContext
}

func NewDepositSuccessConsumer(
	js jetstream.JetStream,
	uowFactory ports.PaymentUnitOfWorkFactory,
	alipayCfg config.Alipay,
	logger logger.Logger,
) *DepositSuccessConsumer {
	return &DepositSuccessConsumer{
		js:         js,
		uowFactory: uowFactory,
		alipayCfg:  alipayCfg,
		logger:     logger,
	}
}

func (consumer *DepositSuccessConsumer) Start(ctx context.Context) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(
		ctx,
		sharednats.StreamPaymentEvents,
		jetstream.ConsumerConfig{
			FilterSubject: event.SubjectDepositSuccess,
			DeliverPolicy: jetstream.DeliverNewPolicy,
			AckPolicy:     jetstream.AckNonePolicy,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create deposit-success consumer: %w", err)
	}

	consumeCtx, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		consumer.handle(ctx, msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start deposit-success consumer: %w", err)
	}

	consumer.consumeCtx = consumeCtx

	return nil
}

func (consumer *DepositSuccessConsumer) Stop() {
	if consumer.consumeCtx != nil {
		consumer.consumeCtx.Drain()
	}
}

func (consumer *DepositSuccessConsumer) handle(ctx context.Context, data []byte) {
	var payload envelope.PaymentSuccessPayload
	if unmarshalErr := json.Unmarshal(data, &payload); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode payment success payload")
		return
	}

	unitOfWork, beginErr := consumer.uowFactory.Begin(ctx)
	if beginErr != nil {
		consumer.logger.Error().Err(beginErr).Msg("failed to begin uow for deposit credit")
		return
	}
	defer func() { _ = unitOfWork.Rollback(ctx) }()

	ledger := unitOfWork.LedgerRepository()

	userOwner := strconv.FormatUint(payload.UserID, 10)
	userAccount, userErr := ledger.GetOrCreateAccountByOwner(ctx, userOwner, payload.Currency)
	if userErr != nil {
		consumer.logger.Error().Err(userErr).Msg("failed to resolve user ledger account")
		return
	}

	platformAccount, platformErr := ledger.GetOrCreateAccountByOwner(
		ctx,
		consumer.alipayCfg.PlatformAccountOwner,
		payload.Currency,
	)
	if platformErr != nil {
		consumer.logger.Error().Err(platformErr).Msg("failed to resolve platform ledger account")
		return
	}

	if _, transferErr := ledger.Transfer(ctx, ledgerports.TransferInput{
		FromAccountID:  platformAccount.ID(),
		ToAccountID:    userAccount.ID(),
		Amount:         payload.AmountInCents,
		IdempotencyKey: payload.OutTradeNo,
		Reference:      payload.OutTradeNo,
		Description:    "recharge credit for payment " + strconv.FormatUint(payload.PaymentID, 10),
	}); transferErr != nil {
		consumer.logger.Error().Err(transferErr).
			Uint64("payment_id", payload.PaymentID).
			Msg("failed to credit user ledger on recharge")
		return
	}

	if completeErr := unitOfWork.Complete(ctx); completeErr != nil {
		consumer.logger.Error().Err(completeErr).Msg("failed to commit deposit credit")
		return
	}

	consumer.logger.Info().Uint64("payment_id", payload.PaymentID).Uint64("user_id", payload.UserID).
		Msg("recharge credited to user ledger")
}
