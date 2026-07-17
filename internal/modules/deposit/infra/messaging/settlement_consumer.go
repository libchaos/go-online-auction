package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/modules/deposit/application/command"
	"auction/internal/modules/deposit/application/query"
	"auction/internal/modules/deposit/domain/errs"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/logger"
	sharednats "auction/internal/shared/modules/nats"
)

const auctionEndedEventType = "auction_ended"

type auctionEndedEnvelope struct {
	EventType string `json:"event_type"`
	AuctionID uint64 `json:"auction_id"`
}

type SettlementConsumer struct {
	js             jetstream.JetStream
	releaseCommand *command.ReleaseDepositCommand
	applyCommand   *command.ApplyDepositCommand
	heldQuery      *query.ListHeldDepositsByAuctionQuery
	winnerReader   ports.AuctionWinnerPort
	logger         logger.Logger
	consumeContext jetstream.ConsumeContext
}

func NewSettlementConsumer(
	js jetstream.JetStream,
	releaseCommand *command.ReleaseDepositCommand,
	applyCommand *command.ApplyDepositCommand,
	heldQuery *query.ListHeldDepositsByAuctionQuery,
	winnerReader ports.AuctionWinnerPort,
	logger logger.Logger,
) *SettlementConsumer {
	return &SettlementConsumer{
		js:             js,
		releaseCommand: releaseCommand,
		applyCommand:   applyCommand,
		heldQuery:      heldQuery,
		winnerReader:   winnerReader,
		logger:         logger,
	}
}

func (consumer *SettlementConsumer) Start(ctx context.Context) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(ctx, sharednats.StreamEvents, jetstream.ConsumerConfig{
		FilterSubject: sharednats.SubjectEvents,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckNonePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create settlement consumer: %w", err)
	}

	consumeContext, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		consumer.handle(ctx, msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start settlement consumer: %w", err)
	}

	consumer.consumeContext = consumeContext

	return nil
}

func (consumer *SettlementConsumer) Stop() {
	if consumer.consumeContext != nil {
		consumer.consumeContext.Drain()
	}
}

func (consumer *SettlementConsumer) handle(ctx context.Context, data []byte) {
	var envelope auctionEndedEnvelope
	if unmarshalErr := json.Unmarshal(data, &envelope); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode auction event envelope")

		return
	}

	if envelope.EventType != auctionEndedEventType {
		return
	}

	if settleErr := consumer.settle(ctx, envelope.AuctionID); settleErr != nil {
		consumer.logger.Error().Err(settleErr).
			Uint64("auction_id", envelope.AuctionID).
			Msg("failed to settle deposits after auction ended")
	}
}

func (consumer *SettlementConsumer) settle(ctx context.Context, auctionID uint64) error {
	winnerUserID, err := consumer.winnerReader.GetWinnerUserID(ctx, auctionID)
	if err != nil {
		if errors.Is(err, errs.ErrAuctionWinnerNotFound) {
			winnerUserID = nil
		} else {
			return err
		}
	}

	output, err := consumer.heldQuery.Execute(ctx, query.ListHeldDepositsByAuctionQueryInput{AuctionID: auctionID})
	if err != nil {
		return err
	}

	for _, deposit := range output.Deposits {
		if winnerUserID != nil && deposit.UserID == *winnerUserID {
			_, applyErr := consumer.applyCommand.Execute(ctx, command.ApplyDepositCommandInput{
				DepositID: deposit.DepositID,
			})
			if applyErr != nil && !errors.Is(applyErr, errs.ErrInvalidDepositTransition) {
				consumer.logger.Error().Err(applyErr).
					Uint64("deposit_id", deposit.DepositID).
					Msg("failed to apply winning deposit")

				return applyErr
			}

			continue
		}

		_, releaseErr := consumer.releaseCommand.Execute(ctx, command.ReleaseDepositCommandInput{
			DepositID: deposit.DepositID,
		})
		if releaseErr != nil && !errors.Is(releaseErr, errs.ErrInvalidDepositTransition) {
			consumer.logger.Error().Err(releaseErr).
				Uint64("deposit_id", deposit.DepositID).
				Msg("failed to release non-winning deposit")

			return releaseErr
		}
	}

	return nil
}
