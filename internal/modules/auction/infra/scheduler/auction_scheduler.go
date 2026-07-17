package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/logger"
)

const (
	defaultInterval  = 5 * time.Second
	defaultBatchSize = 100
)

// Config controls the scheduler polling behavior
type Config struct {
	Interval  time.Duration
	BatchSize int
}

// AuctionScheduler periodically activates draft auctions whose scheduled start
// time has passed and closes active auctions whose end time has passed. It reuses
// the Start/Close commands, so all state transitions go through the same
// concurrency-controlled path as manual operations (FOR UPDATE NOWAIT + optimistic
// locking), which makes concurrent scheduler instances safe: conflicts are skipped.
type AuctionScheduler struct {
	auctionRepository   ports.AuctionRepository
	startAuctionCommand *command.StartAuctionCommand
	closeAuctionCommand *command.CloseAuctionCommand
	logger              logger.Logger
	interval            time.Duration
	batchSize           int

	stopOnce sync.Once
	done     chan struct{}
	stopped  chan struct{}
}

func NewAuctionScheduler(
	auctionRepository ports.AuctionRepository,
	startAuctionCommand *command.StartAuctionCommand,
	closeAuctionCommand *command.CloseAuctionCommand,
	logger logger.Logger,
	cfg Config,
) *AuctionScheduler {
	interval := cfg.Interval
	if interval <= 0 {
		interval = defaultInterval
	}
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	return &AuctionScheduler{
		auctionRepository:   auctionRepository,
		startAuctionCommand: startAuctionCommand,
		closeAuctionCommand: closeAuctionCommand,
		logger:              logger,
		interval:            interval,
		batchSize:           batchSize,
		done:                make(chan struct{}),
		stopped:             make(chan struct{}),
	}
}

// Start launches the polling loop in a background goroutine
func (s *AuctionScheduler) Start(ctx context.Context) {
	go s.run(ctx)
}

// Stop signals the polling loop to exit and waits for it to finish
func (s *AuctionScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.done)
	})
	<-s.stopped
}

func (s *AuctionScheduler) run(ctx context.Context) {
	defer close(s.stopped)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info().
		Dur("interval", s.interval).
		Int("batch_size", s.batchSize).
		Msg("auction scheduler started")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("auction scheduler stopped: context cancelled")
			return
		case <-s.done:
			s.logger.Info().Msg("auction scheduler stopped")
			return
		case <-ticker.C:
			s.Tick(ctx)
		}
	}
}

// Tick runs a single scheduling pass: start due auctions, then close expired ones
func (s *AuctionScheduler) Tick(ctx context.Context) {
	s.startDueAuctions(ctx)
	s.closeExpiredAuctions(ctx)
}

func (s *AuctionScheduler) startDueAuctions(ctx context.Context) {
	ids, err := s.auctionRepository.FindIDsDueToStart(ctx, s.batchSize)
	if err != nil {
		s.logger.Error().Err(err).Msg("scheduler: failed to find auctions due to start")
		return
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}

		_, err = s.startAuctionCommand.Execute(ctx, command.StartAuctionCommandInput{AuctionID: id})
		if err != nil {
			if isBenignSchedulingError(err) {
				continue // another instance or an operator got there first
			}
			s.logger.Error().Err(err).Uint64("auction_id", id).Msg("scheduler: failed to start auction")
			continue
		}

		s.logger.Info().Uint64("auction_id", id).Msg("scheduler: auction started")
	}
}

func (s *AuctionScheduler) closeExpiredAuctions(ctx context.Context) {
	ids, err := s.auctionRepository.FindIDsDueToClose(ctx, s.batchSize)
	if err != nil {
		s.logger.Error().Err(err).Msg("scheduler: failed to find auctions due to close")
		return
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}

		_, err = s.closeAuctionCommand.Execute(ctx, command.CloseAuctionCommandInput{AuctionID: id})
		if err != nil {
			if isBenignSchedulingError(err) {
				continue // another instance or an operator got there first
			}
			s.logger.Error().Err(err).Uint64("auction_id", id).Msg("scheduler: failed to close auction")
			continue
		}

		s.logger.Info().Uint64("auction_id", id).Msg("scheduler: auction closed")
	}
}

// isBenignSchedulingError reports whether the error indicates the auction was
// already transitioned by a concurrent actor, which is expected under multiple
// scheduler instances and requires no logging or retry.
func isBenignSchedulingError(err error) bool {
	return errors.Is(err, errs.ErrConcurrencyConflict) ||
		errors.Is(err, errs.ErrAuctionCanOnlyStartFromDraft) ||
		errors.Is(err, errs.ErrAuctionCanOnlyCloseFromActive) ||
		errors.Is(err, errs.ErrAuctionNotFound)
}
