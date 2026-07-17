package outbox

import (
	"context"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

const (
	defaultInterval  = 500 * time.Millisecond
	defaultBatchSize = 200
)

// Config controls the outbox relay polling behavior
type Config struct {
	Interval  time.Duration
	BatchSize int
}

// Relay drains the transactional outbox and publishes pending events to
// JetStream. Delivery is at-least-once: an event is only marked published
// after JetStream acknowledges it, and redeliveries are deduplicated by the
// stream's duplicate window keyed on Nats-Msg-Id (the domain event ID).
// Multiple relay instances are safe: MarkPublished is guarded by
// `published_at IS NULL`, so duplicates are bounded and deduplicated anyway.
type Relay struct {
	outboxRepository ports.OutboxRepository
	js               jetstream.JetStream
	logger           logger.Logger
	interval         time.Duration
	batchSize        int

	stopOnce sync.Once
	done     chan struct{}
	stopped  chan struct{}
}

func NewRelay(
	outboxRepository ports.OutboxRepository,
	js jetstream.JetStream,
	logger logger.Logger,
	cfg Config,
) *Relay {
	interval := cfg.Interval
	if interval <= 0 {
		interval = defaultInterval
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	return &Relay{
		outboxRepository: outboxRepository,
		js:               js,
		logger:           logger,
		interval:         interval,
		batchSize:        batchSize,
		done:             make(chan struct{}),
		stopped:          make(chan struct{}),
	}
}

// Start launches the polling loop in a background goroutine
func (r *Relay) Start(ctx context.Context) {
	go r.run(ctx)
}

// Stop signals the polling loop to exit and waits for it to finish
func (r *Relay) Stop() {
	r.stopOnce.Do(func() {
		close(r.done)
	})
	<-r.stopped
}

func (r *Relay) run(ctx context.Context) {
	defer close(r.stopped)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.logger.Info().
		Dur("interval", r.interval).
		Int("batch_size", r.batchSize).
		Msg("outbox relay started")

	for {
		select {
		case <-ctx.Done():
			r.logger.Info().Msg("outbox relay stopped: context cancelled")
			return
		case <-r.done:
			r.logger.Info().Msg("outbox relay stopped")
			return
		case <-ticker.C:
			r.Drain(ctx)
		}
	}
}

// Drain publishes one batch of pending outbox events. It keeps draining until
// the outbox has fewer pending events than the batch size, so bursts are
// flushed without waiting for the next tick.
func (r *Relay) Drain(ctx context.Context) {
	for {
		published, pending := r.drainBatch(ctx)
		if pending < r.batchSize || published == 0 {
			return
		}
	}
}

func (r *Relay) drainBatch(ctx context.Context) (int, int) {
	events, err := r.outboxRepository.ListUnpublished(ctx, r.batchSize)
	if err != nil {
		r.logger.Error().Err(err).Msg("outbox relay: failed to list unpublished events")
		return 0, 0
	}

	published := 0

	for _, evt := range events {
		if ctx.Err() != nil {
			return published, len(events)
		}

		// Nats-Msg-Id enables server-side deduplication within the stream's
		// duplicate window, making redelivery after a crash idempotent.
		_, pubErr := r.js.Publish(ctx, evt.Subject, evt.Payload, jetstream.WithMsgID(evt.EventID))
		if pubErr != nil {
			// Stop the batch: publishing in order matters more than progress,
			// and a broker outage would fail the remaining events anyway.
			r.logger.Error().Err(pubErr).
				Str("event_id", evt.EventID).
				Str("subject", evt.Subject).
				Msg("outbox relay: failed to publish event; will retry")
			return published, len(events)
		}

		if _, markErr := r.outboxRepository.MarkPublished(ctx, evt.ID); markErr != nil {
			// The event reached JetStream but is still pending in the outbox; the
			// next pass republishes it and the duplicate window absorbs it.
			r.logger.Error().Err(markErr).
				Str("event_id", evt.EventID).
				Msg("outbox relay: failed to mark event as published; duplicate delivery possible")
			return published, len(events)
		}

		published++
		eventsRelayedTotal.Inc()
	}

	return published, len(events)
}
