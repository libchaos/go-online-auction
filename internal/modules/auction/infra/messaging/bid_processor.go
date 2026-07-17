package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"

	"auction/internal/modules/auction/domain/errs"
	domainevent "auction/internal/modules/auction/domain/event"
	"auction/internal/modules/auction/domain/model"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/event/envelope"
	"auction/internal/modules/auction/ports"
	depositerrs "auction/internal/modules/deposit/domain/errs"
	depositports "auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/logger"
	sharednats "auction/internal/shared/modules/nats"
)

const (
	bidProcessorDurable    = "AUCTION_BID_PROCESSOR"
	bidProcessorMaxAck     = 1
	bidProcessorMaxDeliver = 5
	bidProcessorNakDelay   = 2 * time.Second
)

// IsPermanentBidError reports whether a domain error is permanent (invalid bid
// that will never succeed) and must be routed to the DLQ rather than retried.
func IsPermanentBidError(err error) bool {
	switch {
	case errors.Is(err, errs.ErrBidMustExceedHighest),
		errors.Is(err, errs.ErrFirstBidMustBePositive),
		errors.Is(err, errs.ErrBidsOnlyOnActiveAuctions),
		errors.Is(err, errs.ErrAuctionExpired),
		errors.Is(err, errs.ErrAuctionNotFound),
		errors.Is(err, errs.ErrInvalidAuctionState),
		errors.Is(err, errs.ErrDutchBidMustMatchPrice),
		errors.Is(err, errs.ErrDutchPriceNotAvailable),
		errors.Is(err, errs.ErrFixedPriceMismatch),
		errors.Is(err, errs.ErrFixedPriceNotConfigured),
		errors.Is(err, errs.ErrProxyMaxTooLow),
		errors.Is(err, errs.ErrStartingPriceRequired),
		errors.Is(err, depositerrs.ErrDepositRequired),
		errors.Is(err, depositerrs.ErrDepositInsufficient),
		errors.Is(err, depositerrs.ErrDepositNotHeld):
		return true
	default:
		return false
	}
}

type BidProcessor struct {
	js             jetstream.JetStream
	uowFactory     ports.AuctionUnitOfWorkFactory
	resolver       strategy.Resolver
	depositGuard   depositports.DepositGuard
	logger         logger.Logger
	consumeContext jetstream.ConsumeContext
}

func NewBidProcessor(
	js jetstream.JetStream,
	uowFactory ports.AuctionUnitOfWorkFactory,
	resolver strategy.Resolver,
	depositGuard depositports.DepositGuard,
	logger logger.Logger,
) *BidProcessor {
	return &BidProcessor{
		js:           js,
		uowFactory:   uowFactory,
		resolver:     resolver,
		depositGuard: depositGuard,
		logger:       logger,
	}
}

func (p *BidProcessor) Start(ctx context.Context) error {
	consumer, err := p.js.CreateOrUpdateConsumer(ctx, sharednats.StreamCommands, jetstream.ConsumerConfig{
		Durable:       bidProcessorDurable,
		FilterSubject: sharednats.SubjectCommands,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxAckPending: bidProcessorMaxAck,
		MaxDeliver:    bidProcessorMaxDeliver,
	})
	if err != nil {
		return err
	}

	consumeContext, err := consumer.Consume(func(msg jetstream.Msg) {
		p.handleMessage(ctx, msg)
	})
	if err != nil {
		return err
	}

	p.consumeContext = consumeContext
	return nil
}

func (p *BidProcessor) Stop() {
	if p.consumeContext != nil {
		p.consumeContext.Drain()
	}
}

func (p *BidProcessor) handleMessage(ctx context.Context, msg jetstream.Msg) {
	var cmd ports.BidCommand
	if err := json.Unmarshal(msg.Data(), &cmd); err != nil {
		p.logger.Error().Err(err).Msg("failed to unmarshal bid command; routing to DLQ")
		p.sendToDLQ(ctx, 0, msg)
		return
	}

	err := p.ProcessBid(ctx, cmd)
	switch {
	case err == nil:
		bidCommandsProcessedTotal.WithLabelValues("ok").Inc()
		p.ackAndLog(msg, cmd, "ok")
	case errors.Is(err, errs.ErrBidDuplicateIdempotencyKey):
		bidCommandsProcessedTotal.WithLabelValues("dup").Inc()
		p.ackAndLog(msg, cmd, "dup")
	case IsPermanentBidError(err):
		p.logger.Error().Err(err).
			Uint64("auction_id", cmd.AuctionID).
			Str("idempotency_key", cmd.IdempotencyKey).
			Msg("permanent bid error; routing to DLQ")
		p.sendToDLQ(ctx, cmd.AuctionID, msg)
	default:
		p.handleTransientError(ctx, cmd, msg, err)
	}
}

func (p *BidProcessor) handleTransientError(
	ctx context.Context,
	cmd ports.BidCommand,
	msg jetstream.Msg,
	cause error,
) {
	metadata, metaErr := msg.Metadata()
	if metaErr == nil && metadata.NumDelivered >= bidProcessorMaxDeliver {
		p.logger.Error().Err(cause).
			Uint64("auction_id", cmd.AuctionID).
			Str("idempotency_key", cmd.IdempotencyKey).
			Uint64("num_delivered", metadata.NumDelivered).
			Msg("max delivery attempts reached; routing to DLQ")
		p.sendToDLQ(ctx, cmd.AuctionID, msg)
		return
	}

	p.logger.Warn().Err(cause).
		Uint64("auction_id", cmd.AuctionID).
		Str("idempotency_key", cmd.IdempotencyKey).
		Msg("transient bid error; scheduling redelivery")
	_ = msg.NakWithDelay(bidProcessorNakDelay)
}

func (p *BidProcessor) ProcessBid(ctx context.Context, cmd ports.BidCommand) error {
	if guardErr := p.depositGuard.EnsureEligible(ctx, cmd.UserID, cmd.AuctionID); guardErr != nil {
		return guardErr
	}

	uow, err := p.uowFactory.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = uow.Rollback(ctx) }()

	auction, err := uow.AuctionRepository().FindByID(ctx, cmd.AuctionID)
	if err != nil {
		return err
	}

	selected, err := p.resolver.ForMode(auction.TradingMode())
	if err != nil {
		return err
	}

	bid, err := p.buildBid(ctx, uow, auction, cmd, selected)
	if err != nil {
		return err
	}

	if placeErr := auction.PlaceBid(bid.Amount()); placeErr != nil {
		return placeErr
	}

	persistedBid, err := uow.BidRepository().Create(ctx, bid, cmd.IdempotencyKey)
	if err != nil {
		return err
	}

	// Anti-snipe: a bid landing inside the extension window pushes the close time out.
	auction.MaybeExtendEndTime(time.Now().UTC())

	if proxyErr := p.applyProxyBid(ctx, uow, &auction, cmd, selected); proxyErr != nil {
		return proxyErr
	}

	if selected.ShouldCloseOnAccept() {
		if closeErr := p.closeAuctionOnAccept(ctx, uow, &auction); closeErr != nil {
			return closeErr
		}
	}

	return p.commitAndDispatch(ctx, uow, auction, persistedBid, cmd, selected)
}

func (p *BidProcessor) commitAndDispatch(
	ctx context.Context,
	uow ports.AuctionUnitOfWork,
	auction model.AuctionModel,
	persistedBid model.BidModel,
	cmd ports.BidCommand,
	selected strategy.TradingStrategy,
) error {
	if updateErr := uow.AuctionRepository().Update(ctx, auction); updateErr != nil {
		return updateErr
	}

	// Record events in the transactional outbox before committing so the state
	// change and its events are atomic; the outbox relay delivers them to JetStream.
	bidPlacedEvent := domainevent.NewBidPlacedEvent(
		persistedBid.ID(),
		auction.ID(),
		cmd.UserID,
		persistedBid.Amount(),
	)
	bidPlacedOutbox, err := envelope.FromBidPlaced(bidPlacedEvent)
	if err != nil {
		return err
	}
	if saveErr := uow.OutboxRepository().Save(ctx, bidPlacedOutbox); saveErr != nil {
		return saveErr
	}

	if selected.ShouldCloseOnAccept() {
		endedEvent := p.buildAuctionEndedEvent(auction)
		endedOutbox, envErr := envelope.FromAuctionEnded(endedEvent)
		if envErr != nil {
			return envErr
		}
		if saveErr := uow.OutboxRepository().Save(ctx, endedOutbox); saveErr != nil {
			return saveErr
		}
	}

	return uow.Complete(ctx)
}

func (p *BidProcessor) applyProxyBid(
	ctx context.Context,
	uow ports.AuctionUnitOfWork,
	auction *model.AuctionModel,
	cmd ports.BidCommand,
	selected strategy.TradingStrategy,
) error {
	proxy, ok := selected.(strategy.ProxyResolvable)
	if !ok || cmd.MaxAmountInCents == nil {
		return nil
	}

	existing, ferr := uow.BidRepository().FindByAuctionID(ctx, auction.ID())
	if ferr != nil {
		return ferr
	}

	actions, rerr := proxy.ResolveProxyBids(
		auction,
		model.ToBidViews(existing),
		cmd.UserID,
		strategy.NewMoney(*cmd.MaxAmountInCents),
	)
	if rerr != nil {
		return rerr
	}

	if len(actions) == 0 {
		return nil
	}

	publicPrice := actions[0].Amount.AmountInCents()
	if auction.HighestBidAmount() == nil || publicPrice > *auction.HighestBidAmount() {
		auction.RecordHighestBidAmount(publicPrice)
	}

	return nil
}

func (p *BidProcessor) closeAuctionOnAccept(
	ctx context.Context,
	uow ports.AuctionUnitOfWork,
	auction *model.AuctionModel,
) error {
	existing, ferr := uow.BidRepository().FindByAuctionID(ctx, auction.ID())
	if ferr != nil {
		return ferr
	}

	return auction.Close(existing)
}

func (p *BidProcessor) buildBid(
	ctx context.Context,
	uow ports.AuctionUnitOfWork,
	auction model.AuctionModel,
	cmd ports.BidCommand,
	selected strategy.TradingStrategy,
) (model.BidModel, error) {
	if cmd.MaxAmountInCents == nil {
		money := model.NewMoneyModel(cmd.AmountInCents)
		return model.NewBidModel(cmd.AuctionID, cmd.UserID, money)
	}

	maxMoney := model.NewMoneyModel(*cmd.MaxAmountInCents)
	money := model.NewMoneyModel(cmd.AmountInCents)

	proxy, ok := selected.(strategy.ProxyResolvable)
	if !ok {
		return model.NewBidModelWithMax(cmd.AuctionID, cmd.UserID, money, &maxMoney)
	}

	existing, err := uow.BidRepository().FindByAuctionID(ctx, auction.ID())
	if err != nil {
		return model.BidModel{}, err
	}

	actions, err := proxy.ResolveProxyBids(
		&auction,
		model.ToBidViews(existing),
		cmd.UserID,
		strategy.NewMoney(*cmd.MaxAmountInCents),
	)
	if err != nil {
		return model.BidModel{}, err
	}

	publicAmount := cmd.AmountInCents
	for _, action := range actions {
		if action.UserID == cmd.UserID {
			publicAmount = action.Amount.AmountInCents()
		}
	}

	return model.NewBidModelWithMax(cmd.AuctionID, cmd.UserID, model.NewMoneyModel(publicAmount), &maxMoney)
}

func (p *BidProcessor) buildAuctionEndedEvent(auction model.AuctionModel) domainevent.AuctionEndedEvent {
	var winningBidID *uint64
	var finalAmount *model.MoneyModel
	if auction.WinnerUserID() != nil && auction.WinningBidAmount() != nil {
		winningBidID = auction.WinningBidID()
		amount := model.NewMoneyModel(*auction.WinningBidAmount())
		finalAmount = &amount
	}

	return domainevent.NewAuctionEndedEvent(auction.ID(), winningBidID, finalAmount)
}

func (p *BidProcessor) sendToDLQ(ctx context.Context, auctionID uint64, msg jetstream.Msg) {
	subject := BuildDLQSubject(auctionID)
	if _, err := p.js.Publish(ctx, subject, msg.Data()); err != nil {
		p.logger.Error().Err(err).
			Str("subject", subject).
			Uint64("auction_id", auctionID).
			Msg("failed to publish message to DLQ; leaving for redelivery")
		_ = msg.NakWithDelay(bidProcessorNakDelay)
		return
	}
	bidCommandsProcessedTotal.WithLabelValues("dlq").Inc()
	_ = msg.Term()
}

func (p *BidProcessor) ackAndLog(msg jetstream.Msg, cmd ports.BidCommand, result string) {
	p.logger.Debug().
		Uint64("auction_id", cmd.AuctionID).
		Str("idempotency_key", cmd.IdempotencyKey).
		Str("result", result).
		Msg("bid command processed")
	_ = msg.Ack()
}
