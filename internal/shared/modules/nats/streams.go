package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	StreamCommands      = "AUCTION_COMMANDS"
	StreamEvents        = "AUCTION_EVENTS"
	StreamDLQ           = "AUCTION_DLQ"
	StreamListingEvents = "LISTING_EVENTS"

	SubjectCommands = "auction.cmd.bid.*"
	SubjectEvents   = "auction.evt.*"
	SubjectDLQ      = "auction.dlq.*"
	// SubjectListingEvents carries two tokens after the prefix (aggregate type
	// + id), e.g. listing.evt.sku.42, so the wildcard is ">" rather than "*".
	SubjectListingEvents = "listing.evt.>"

	StreamDepositEvents  = "DEPOSIT_EVENTS"
	SubjectDepositEvents = "deposit.evt.*"

	StreamPaymentEvents  = "PAYMENT_EVENTS"
	SubjectPaymentEvents = "payment.>"

	StreamNotificationEvents = "NOTIFICATION_EVENTS"
	// SubjectNotificationEvents is the wildcard the NOTIFICATION_EVENTS stream
	// accepts. It must cover both notification.evt.{userID} (SSE fan-out) and
	// notification.evt.email.requested (email dispatch), so it uses the multi-
	// token ">" form rather than the single-token "*".
	SubjectNotificationEvents = "notification.evt.>"
	// SubjectNotificationUserEvents is the per-user subject the SSE realtime hub
	// filters on; using the single-token "*" keeps email requests out of the
	// hub's delivery path.
	SubjectNotificationUserEvents = "notification.evt.*"
	// SubjectNotificationEmailRequested is published by the outbox relay for
	// each email the EmailDispatchConsumer should send.
	SubjectNotificationEmailRequested = "notification.evt.email.requested"

	defaultDedupeWindow = 2 * time.Minute
)

func CreateOrUpdateStreams(ctx context.Context, js jetstream.JetStream, dedupeWindow time.Duration) error {
	if dedupeWindow <= 0 {
		dedupeWindow = defaultDedupeWindow
	}

	configs := []jetstream.StreamConfig{
		{
			Name:       StreamCommands,
			Subjects:   []string{SubjectCommands},
			Retention:  jetstream.WorkQueuePolicy,
			Storage:    jetstream.FileStorage,
			Duplicates: dedupeWindow,
		},
		{
			// AUCTION_EVENTS is the event store: LimitsPolicy with no limits keeps
			// every event indefinitely on file storage, enabling replay and temporal
			// queries. The duplicate window lets the outbox relay redeliver safely
			// (messages carry the domain event ID as Nats-Msg-Id).
			Name:       StreamEvents,
			Subjects:   []string{SubjectEvents},
			Retention:  jetstream.LimitsPolicy,
			Storage:    jetstream.FileStorage,
			Duplicates: dedupeWindow,
		},
		{
			Name:      StreamDLQ,
			Subjects:  []string{SubjectDLQ},
			Retention: jetstream.LimitsPolicy,
			Storage:   jetstream.FileStorage,
		},
		{
			// LISTING_EVENTS mirrors AUCTION_EVENTS: an event store fed by the
			// shared transactional outbox relay, deduplicated by Nats-Msg-Id.
			Name:       StreamListingEvents,
			Subjects:   []string{SubjectListingEvents},
			Retention:  jetstream.LimitsPolicy,
			Storage:    jetstream.FileStorage,
			Duplicates: dedupeWindow,
		},
		{
			// DEPOSIT_EVENTS is the event store for deposit lifecycle events,
			// fed by the shared transactional outbox relay and deduplicated by
			// Nats-Msg-Id. Consumers push status changes to the owning user.
			Name:       StreamDepositEvents,
			Subjects:   []string{SubjectDepositEvents},
			Retention:  jetstream.LimitsPolicy,
			Storage:    jetstream.FileStorage,
			Duplicates: dedupeWindow,
		},
		{
			// PAYMENT_EVENTS is the event store for payment lifecycle events
			// (recharge success, withdrawal requested/completed/failed), fed by
			// the payment module's own transactional outbox relay and
			// deduplicated by Nats-Msg-Id.
			Name:       StreamPaymentEvents,
			Subjects:   []string{SubjectPaymentEvents},
			Retention:  jetstream.LimitsPolicy,
			Storage:    jetstream.FileStorage,
			Duplicates: dedupeWindow,
		},
		{
			// NOTIFICATION_EVENTS is the event store for notification.evt.created,
			// fed by the notification module's own transactional outbox relay and
			// deduplicated by Nats-Msg-Id. The SSE realtime hub consumes it and
			// fans each event out to the recipient user's connected clients.
			Name:       StreamNotificationEvents,
			Subjects:   []string{SubjectNotificationEvents},
			Retention:  jetstream.LimitsPolicy,
			Storage:    jetstream.FileStorage,
			Duplicates: dedupeWindow,
		},
	}

	for _, cfg := range configs {
		if _, err := js.CreateOrUpdateStream(ctx, cfg); err != nil {
			return fmt.Errorf("failed to create or update stream %s: %w", cfg.Name, err)
		}
	}

	return nil
}
