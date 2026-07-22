package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/shared/modules/logger"
	sharednats "auction/internal/shared/modules/nats"
)

const (
	subjectPaymentDepositSuccess      = "payment.evt.deposit.success"
	subjectPaymentWithdrawalCompleted = "payment.evt.withdrawal.completed"
	subjectPaymentWithdrawalFailed    = "payment.evt.withdrawal.failed"

	eventTypeBidPlaced    = "bid_placed"
	eventTypeAuctionEnded = "auction_ended"
)

type paymentSuccessMessage struct {
	EventID       string `json:"event_id"`
	PaymentID     uint64 `json:"payment_id"`
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
}

type withdrawalEventMessage struct {
	EventID       string `json:"event_id"`
	WithdrawalID  uint64 `json:"withdrawal_id"`
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	FailReason    string `json:"fail_reason"`
}

type depositEventMessage struct {
	EventType string          `json:"event_type"`
	EventID   string          `json:"event_id"`
	UserID    uint64          `json:"user_id"`
	AuctionID uint64          `json:"auction_id"`
	Data      json.RawMessage `json:"data"`
}

type depositEventData struct {
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
}

type auctionEventMessage struct {
	EventType string          `json:"event_type"`
	EventID   string          `json:"event_id"`
	AuctionID uint64          `json:"auction_id"`
	Data      json.RawMessage `json:"data"`
}

type moneyMessage struct {
	AmountInCents uint64 `json:"amount_in_cents"`
}

type bidPlacedData struct {
	BidID  uint64       `json:"bid_id"`
	UserID uint64       `json:"user_id"`
	Amount moneyMessage `json:"amount"`
}

type auctionEndedData struct {
	WinningBidID *uint64       `json:"winning_bid_id"`
	FinalAmount  *moneyMessage `json:"final_amount"`
}

type listingEnvelope struct {
	EventType     string          `json:"event_type"`
	EventID       string          `json:"event_id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uint64          `json:"aggregate_id"`
	Data          json.RawMessage `json:"data"`
}

type listingSkuData struct {
	SkuID        uint64 `json:"sku_id"`
	SpuID        uint64 `json:"spu_id"`
	PriceInCents uint64 `json:"price_in_cents"`
	Quantity     uint64 `json:"quantity"`
}

// SourceEventConsumer subscribes read-only to the payment, deposit and auction
// event streams and translates each relevant event into an in-app notification
// through the application service. It uses ephemeral, no-ack consumers because a
// notification missed while the service is down is recoverable from the
// notification-center REST endpoints; duplicate deliveries are absorbed by the
// idempotent create-notification write.
type SourceEventConsumer struct {
	js              jetstream.JetStream
	service         *NotificationApplicationService
	logger          logger.Logger
	consumeContexts []jetstream.ConsumeContext
}

func NewSourceEventConsumer(
	js jetstream.JetStream,
	service *NotificationApplicationService,
	logger logger.Logger,
) *SourceEventConsumer {
	return &SourceEventConsumer{
		js:      js,
		service: service,
		logger:  logger,
	}
}

func (consumer *SourceEventConsumer) Start(ctx context.Context) error {
	if startErr := consumer.subscribe(
		ctx,
		sharednats.StreamPaymentEvents,
		subjectPaymentDepositSuccess,
		consumer.handlePayment,
	); startErr != nil {
		return startErr
	}

	if startErr := consumer.subscribe(
		ctx,
		sharednats.StreamPaymentEvents,
		subjectPaymentWithdrawalCompleted,
		consumer.handleWithdrawalCompleted,
	); startErr != nil {
		return startErr
	}

	if startErr := consumer.subscribe(
		ctx,
		sharednats.StreamPaymentEvents,
		subjectPaymentWithdrawalFailed,
		consumer.handleWithdrawalFailed,
	); startErr != nil {
		return startErr
	}

	if startErr := consumer.subscribe(
		ctx,
		sharednats.StreamDepositEvents,
		sharednats.SubjectDepositEvents,
		consumer.handleDeposit,
	); startErr != nil {
		return startErr
	}

	if startErr := consumer.subscribe(
		ctx,
		sharednats.StreamEvents,
		sharednats.SubjectEvents,
		consumer.handleAuction,
	); startErr != nil {
		return startErr
	}

	if startErr := consumer.subscribe(
		ctx,
		sharednats.StreamListingEvents,
		sharednats.SubjectListingEvents,
		consumer.handleListing,
	); startErr != nil {
		return startErr
	}

	return nil
}

func (consumer *SourceEventConsumer) Stop() {
	for _, consumeContext := range consumer.consumeContexts {
		if consumeContext != nil {
			consumeContext.Drain()
		}
	}
}

func (consumer *SourceEventConsumer) subscribe(
	ctx context.Context,
	stream string,
	filterSubject string,
	handler func(ctx context.Context, data []byte),
) error {
	eventConsumer, err := consumer.js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		FilterSubject: filterSubject,
		DeliverPolicy: jetstream.DeliverNewPolicy,
		AckPolicy:     jetstream.AckNonePolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create notification source consumer for %s: %w", stream, err)
	}

	consumeContext, err := eventConsumer.Consume(func(msg jetstream.Msg) {
		handler(ctx, msg.Data())
	})
	if err != nil {
		return fmt.Errorf("failed to start notification source consumer for %s: %w", stream, err)
	}

	consumer.consumeContexts = append(consumer.consumeContexts, consumeContext)

	return nil
}

func (consumer *SourceEventConsumer) handlePayment(ctx context.Context, data []byte) {
	var message paymentSuccessMessage
	if unmarshalErr := json.Unmarshal(data, &message); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode payment success event")

		return
	}

	err := consumer.service.HandleRechargeSuccess(ctx, RechargeSuccessInput{
		SourceEventID: message.EventID,
		UserID:        message.UserID,
		PaymentID:     message.PaymentID,
		AmountInCents: message.AmountInCents,
		Currency:      message.Currency,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("user_id", message.UserID).
			Msg("failed to create recharge-success notification")
	}
}

func (consumer *SourceEventConsumer) handleWithdrawalCompleted(ctx context.Context, data []byte) {
	var message withdrawalEventMessage
	if unmarshalErr := json.Unmarshal(data, &message); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode withdrawal completed event")

		return
	}

	err := consumer.service.HandleWithdrawalCompleted(ctx, WithdrawalCompletedInput{
		SourceEventID: message.EventID,
		UserID:        message.UserID,
		WithdrawalID:  message.WithdrawalID,
		AmountInCents: message.AmountInCents,
		Currency:      message.Currency,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("user_id", message.UserID).
			Msg("failed to create withdrawal-completed notification")
	}
}

func (consumer *SourceEventConsumer) handleWithdrawalFailed(ctx context.Context, data []byte) {
	var message withdrawalEventMessage
	if unmarshalErr := json.Unmarshal(data, &message); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode withdrawal failed event")

		return
	}

	err := consumer.service.HandleWithdrawalFailed(ctx, WithdrawalFailedInput{
		SourceEventID: message.EventID,
		UserID:        message.UserID,
		WithdrawalID:  message.WithdrawalID,
		AmountInCents: message.AmountInCents,
		Currency:      message.Currency,
		FailReason:    message.FailReason,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("user_id", message.UserID).
			Msg("failed to create withdrawal-failed notification")
	}
}

func (consumer *SourceEventConsumer) handleDeposit(ctx context.Context, data []byte) {
	var message depositEventMessage
	if unmarshalErr := json.Unmarshal(data, &message); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode deposit event envelope")

		return
	}

	var eventData depositEventData
	if len(message.Data) > 0 {
		if unmarshalErr := json.Unmarshal(message.Data, &eventData); unmarshalErr != nil {
			consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode deposit event data")

			return
		}
	}

	err := consumer.service.HandleDepositEvent(ctx, DepositEventInput{
		SourceEventID: message.EventID,
		EventType:     message.EventType,
		UserID:        message.UserID,
		AuctionID:     message.AuctionID,
		AmountInCents: eventData.AmountInCents,
		Currency:      eventData.Currency,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("user_id", message.UserID).
			Str("event_type", message.EventType).
			Msg("failed to create deposit notification")
	}
}

func (consumer *SourceEventConsumer) handleAuction(ctx context.Context, data []byte) {
	var message auctionEventMessage
	if unmarshalErr := json.Unmarshal(data, &message); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode auction event envelope")

		return
	}

	switch message.EventType {
	case eventTypeBidPlaced:
		consumer.handleBidPlaced(ctx, message)
	case eventTypeAuctionEnded:
		consumer.handleAuctionEnded(ctx, message)
	default:
	}
}

func (consumer *SourceEventConsumer) handleBidPlaced(ctx context.Context, message auctionEventMessage) {
	var data bidPlacedData
	if unmarshalErr := json.Unmarshal(message.Data, &data); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode bid-placed event data")

		return
	}

	err := consumer.service.HandleBidPlaced(ctx, BidPlacedInput{
		SourceEventID: message.EventID,
		AuctionID:     message.AuctionID,
		NewBidderID:   data.UserID,
		AmountInCents: data.Amount.AmountInCents,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("auction_id", message.AuctionID).
			Msg("failed to create outbid notification")
	}
}

func (consumer *SourceEventConsumer) handleAuctionEnded(ctx context.Context, message auctionEventMessage) {
	var data auctionEndedData
	if unmarshalErr := json.Unmarshal(message.Data, &data); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode auction-ended event data")

		return
	}

	var finalAmount *uint64
	if data.FinalAmount != nil {
		amount := data.FinalAmount.AmountInCents
		finalAmount = &amount
	}

	err := consumer.service.HandleAuctionEnded(ctx, AuctionEndedInput{
		SourceEventID:      message.EventID,
		AuctionID:          message.AuctionID,
		WinningBidID:       data.WinningBidID,
		FinalAmountInCents: finalAmount,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("auction_id", message.AuctionID).
			Msg("failed to create auction-ended notification")
	}
}

func (consumer *SourceEventConsumer) handleListing(ctx context.Context, data []byte) {
	var envelope listingEnvelope
	if unmarshalErr := json.Unmarshal(data, &envelope); unmarshalErr != nil {
		consumer.logger.Error().Err(unmarshalErr).Msg("failed to decode listing event envelope")

		return
	}

	spuID, ok := resolveListingSpuID(envelope)
	if !ok {
		return
	}

	err := consumer.service.HandleListingEvent(ctx, ListingEventInput{
		SourceEventID: envelope.EventID,
		SpuID:         spuID,
		EventType:     envelope.EventType,
	})
	if err != nil {
		consumer.logger.Error().Err(err).
			Uint64("spu_id", spuID).
			Str("event_type", envelope.EventType).
			Msg("failed to create listing notification")
	}
}

// resolveListingSpuID maps a listing event envelope to the SPU a user may be
// watching. SPU events carry the id in AggregateID; SKU events nest it in Data.
func resolveListingSpuID(envelope listingEnvelope) (uint64, bool) {
	if envelope.AggregateType == "spu" {
		return envelope.AggregateID, true
	}

	if envelope.AggregateType == "sku" {
		var data listingSkuData
		if unmarshalErr := json.Unmarshal(envelope.Data, &data); unmarshalErr != nil {
			return 0, false
		}

		return data.SpuID, true
	}

	return 0, false
}
