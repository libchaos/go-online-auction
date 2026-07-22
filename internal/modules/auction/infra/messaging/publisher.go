package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

type JetStreamBidCommandPublisher struct {
	js     jetstream.JetStream
	logger logger.Logger
}

func NewJetStreamBidCommandPublisher(
	js jetstream.JetStream,
	logger logger.Logger,
) *JetStreamBidCommandPublisher {
	return &JetStreamBidCommandPublisher{
		js:     js,
		logger: logger,
	}
}

var _ ports.BidCommandPublisher = (*JetStreamBidCommandPublisher)(nil)

func BuildBidCommandSubject(auctionID uint64) string {
	return fmt.Sprintf("auction.cmd.bid.%d", auctionID)
}

func BuildDLQSubject(auctionID uint64) string {
	return fmt.Sprintf("auction.dlq.%d", auctionID)
}

func (p *JetStreamBidCommandPublisher) Publish(
	ctx context.Context,
	cmd ports.BidCommand,
) (ports.BidCommandAck, error) {
	data, err := json.Marshal(cmd)
	if err != nil {
		p.logger.Error().
			Err(err).
			Uint64("auction_id", cmd.AuctionID).
			Str("idempotency_key", cmd.IdempotencyKey).
			Msg("failed to marshal bid command")
		return ports.BidCommandAck{}, err
	}

	subject := BuildBidCommandSubject(cmd.AuctionID)
	startedAt := time.Now()

	// Inject the active W3C trace context into the NATS headers so the trace
	// continues across the message boundary. With no active span the
	// propagator injects nothing, keeping this safe.
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	headers := nats.Header{}
	for k, v := range carrier {
		headers.Set(k, v)
	}

	_, err = p.js.PublishMsg(ctx, &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  headers,
	}, jetstream.WithMsgID(cmd.IdempotencyKey))
	bidCommandPublishDuration.Observe(time.Since(startedAt).Seconds())
	if err != nil {
		p.logger.Error().
			Err(err).
			Str("subject", subject).
			Uint64("auction_id", cmd.AuctionID).
			Str("idempotency_key", cmd.IdempotencyKey).
			Msg("failed to publish bid command")
		return ports.BidCommandAck{}, err
	}

	bidCommandsPublishedTotal.Inc()
	p.logger.Debug().
		Str("subject", subject).
		Uint64("auction_id", cmd.AuctionID).
		Str("idempotency_key", cmd.IdempotencyKey).
		Msg("bid command published successfully")

	return ports.BidCommandAck{IdempotencyKey: cmd.IdempotencyKey}, nil
}
