package model

import (
	"time"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
	"auction/internal/modules/auction/domain/strategy"
)

type AuctionModel struct {
	id                 uint64
	listingID          uint64
	tradingMode        enum.TradingModeEnum
	startingPrice      *uint64
	priceStep          *uint64
	reservePrice       *uint64
	currentPrice       *uint64
	highestBidAmount   *uint64
	winnerUserID       *uint64
	winningBidID       *uint64
	winningBidAmount   *uint64
	antiSnipeEnabled   bool
	extensionWindowSec int64
	startTime          *time.Time
	endTime            time.Time
	state              enum.AuctionStateEnum
	version            uint64
	createdAt          time.Time
	updatedAt          time.Time
}

func NewAuctionModel(listingID uint64, endTime time.Time) (AuctionModel, error) {
	if err := validateNewAuction(listingID, endTime); err != nil {
		return AuctionModel{}, err
	}

	draftState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	if err != nil {
		return AuctionModel{}, err
	}

	tradingMode, err := enum.NewTradingModeEnum(enum.EnumTradingModeEnglish)
	if err != nil {
		return AuctionModel{}, err
	}

	now := time.Now().UTC()
	return AuctionModel{
		listingID:        listingID,
		endTime:          endTime.UTC(),
		state:            draftState,
		tradingMode:      tradingMode,
		highestBidAmount: nil,
		version:          1,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

func NewAuctionModelWithMode(
	listingID uint64,
	endTime time.Time,
	tradingMode enum.TradingModeEnum,
	startingPrice *uint64,
	priceStep *uint64,
	reservePrice *uint64,
	antiSnipeEnabled bool,
	extensionWindowSec int64,
	scheduledStartTime *time.Time,
) (AuctionModel, error) {
	if err := validateNewAuction(listingID, endTime); err != nil {
		return AuctionModel{}, err
	}

	if err := validateScheduledStartTime(scheduledStartTime, endTime); err != nil {
		return AuctionModel{}, err
	}

	resolvedMode, err := enum.NewTradingModeEnum(tradingMode.String())
	if err != nil {
		return AuctionModel{}, err
	}

	if (resolvedMode.String() == enum.EnumTradingModeDutch ||
		resolvedMode.String() == enum.EnumTradingModeFixedPrice) && startingPrice == nil {
		return AuctionModel{}, errs.ErrStartingPriceRequired
	}

	draftState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	if err != nil {
		return AuctionModel{}, err
	}

	var startTime *time.Time
	if scheduledStartTime != nil {
		utc := scheduledStartTime.UTC()
		startTime = &utc
	}

	now := time.Now().UTC()
	return AuctionModel{
		listingID:          listingID,
		startTime:          startTime,
		endTime:            endTime.UTC(),
		state:              draftState,
		tradingMode:        resolvedMode,
		startingPrice:      startingPrice,
		priceStep:          priceStep,
		reservePrice:       reservePrice,
		antiSnipeEnabled:   antiSnipeEnabled,
		extensionWindowSec: extensionWindowSec,
		highestBidAmount:   nil,
		version:            1,
		createdAt:          now,
		updatedAt:          now,
	}, nil
}

func RestoreAuctionModel(
	id, listingID uint64,
	startTime *time.Time,
	endTime time.Time,
	state enum.AuctionStateEnum,
	highestBidAmount *uint64,
	version uint64,
	createdAt, updatedAt time.Time,
) (AuctionModel, error) {
	if id == 0 {
		return AuctionModel{}, errs.ErrAuctionIDRequired
	}

	if err := validateRestoreAuction(listingID, endTime); err != nil {
		return AuctionModel{}, err
	}

	tradingMode, err := enum.NewTradingModeEnum(enum.EnumTradingModeEnglish)
	if err != nil {
		return AuctionModel{}, err
	}

	var normalizedStartTime *time.Time
	if startTime != nil {
		utc := startTime.UTC()
		normalizedStartTime = &utc
	}

	return AuctionModel{
		id:               id,
		listingID:        listingID,
		startTime:        normalizedStartTime,
		endTime:          endTime.UTC(),
		state:            state,
		tradingMode:      tradingMode,
		highestBidAmount: highestBidAmount,
		version:          version,
		createdAt:        createdAt.UTC(),
		updatedAt:        updatedAt.UTC(),
	}, nil
}

func RestoreAuctionModelWithMode(
	id, listingID uint64,
	startTime *time.Time,
	endTime time.Time,
	state enum.AuctionStateEnum,
	tradingMode enum.TradingModeEnum,
	highestBidAmount *uint64,
	startingPrice *uint64,
	priceStep *uint64,
	reservePrice *uint64,
	currentPrice *uint64,
	winnerUserID *uint64,
	winningBidID *uint64,
	winningBidAmount *uint64,
	antiSnipeEnabled bool,
	extensionWindowSec int64,
	version uint64,
	createdAt, updatedAt time.Time,
) (AuctionModel, error) {
	if id == 0 {
		return AuctionModel{}, errs.ErrAuctionIDRequired
	}

	if err := validateRestoreAuction(listingID, endTime); err != nil {
		return AuctionModel{}, err
	}

	resolvedMode, err := enum.NewTradingModeEnum(tradingMode.String())
	if err != nil {
		return AuctionModel{}, err
	}

	var normalizedStartTime *time.Time
	if startTime != nil {
		utc := startTime.UTC()
		normalizedStartTime = &utc
	}

	return AuctionModel{
		id:                 id,
		listingID:          listingID,
		startTime:          normalizedStartTime,
		endTime:            endTime.UTC(),
		state:              state,
		tradingMode:        resolvedMode,
		highestBidAmount:   highestBidAmount,
		startingPrice:      startingPrice,
		priceStep:          priceStep,
		reservePrice:       reservePrice,
		currentPrice:       currentPrice,
		winnerUserID:       winnerUserID,
		winningBidID:       winningBidID,
		winningBidAmount:   winningBidAmount,
		antiSnipeEnabled:   antiSnipeEnabled,
		extensionWindowSec: extensionWindowSec,
		version:            version,
		createdAt:          createdAt.UTC(),
		updatedAt:          updatedAt.UTC(),
	}, nil
}

func (a *AuctionModel) ID() uint64 {
	return a.id
}

func (a *AuctionModel) ListingID() uint64 {
	return a.listingID
}

func (a *AuctionModel) StartTime() *time.Time {
	return a.startTime
}

func (a *AuctionModel) EndTime() time.Time {
	return a.endTime
}

func (a *AuctionModel) State() enum.AuctionStateEnum {
	return a.state
}

func (a *AuctionModel) TradingMode() enum.TradingModeEnum {
	return a.tradingMode
}

func (a *AuctionModel) HighestBidAmount() *uint64 {
	return a.highestBidAmount
}

func (a *AuctionModel) StartingPrice() *uint64 {
	return a.startingPrice
}

func (a *AuctionModel) PriceStep() *uint64 {
	return a.priceStep
}

func (a *AuctionModel) ReservePrice() *uint64 {
	return a.reservePrice
}

func (a *AuctionModel) CurrentPrice() *uint64 {
	return a.currentPrice
}

func (a *AuctionModel) WinnerUserID() *uint64 {
	return a.winnerUserID
}

func (a *AuctionModel) WinningBidID() *uint64 {
	return a.winningBidID
}

func (a *AuctionModel) WinningBidAmount() *uint64 {
	return a.winningBidAmount
}

func (a *AuctionModel) AntiSnipeEnabled() bool {
	return a.antiSnipeEnabled
}

func (a *AuctionModel) ExtensionWindowSec() int64 {
	return a.extensionWindowSec
}

func (a *AuctionModel) Version() uint64 {
	return a.version
}

func (a *AuctionModel) CreatedAt() time.Time {
	return a.createdAt
}

func (a *AuctionModel) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *AuctionModel) Start() error {
	if a.state.String() != enum.EnumAuctionStateDraft {
		return errs.ErrAuctionCanOnlyStartFromDraft
	}

	activeState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	a.startTime = &now
	if a.currentPrice == nil && a.startingPrice != nil {
		currentPrice := *a.startingPrice
		a.currentPrice = &currentPrice
	}
	a.state = activeState
	a.version++
	a.updatedAt = now

	return nil
}

func (a *AuctionModel) PlaceBid(amount MoneyModel) error {
	if a.state.String() != enum.EnumAuctionStateActive {
		return errs.ErrBidsOnlyOnActiveAuctions
	}

	if time.Now().UTC().After(a.endTime) {
		return errs.ErrAuctionExpired
	}

	resolver := strategy.GetResolver()
	selected, err := resolver.ForMode(a.tradingMode)
	if err != nil {
		return err
	}

	if validateErr := selected.ValidateBid(a, amount); validateErr != nil {
		return validateErr
	}

	amountInCents := amount.AmountInCents()
	a.highestBidAmount = &amountInCents
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

func (a *AuctionModel) RecordHighestBidAmount(amount uint64) {
	a.highestBidAmount = &amount
	a.version++
	a.updatedAt = time.Now().UTC()
}

func (a *AuctionModel) Close(bids []BidModel) error {
	if a.state.String() != enum.EnumAuctionStateActive {
		return errs.ErrAuctionCanOnlyCloseFromActive
	}

	resolver := strategy.GetResolver()
	selected, err := resolver.ForMode(a.tradingMode)
	if err != nil {
		return err
	}

	views := ToBidViews(bids)
	winner, err := selected.DetermineWinner(a, views)
	if err != nil {
		return err
	}

	if winner.UserID != 0 {
		a.winnerUserID = &winner.UserID
		a.winningBidID = winner.BidID
		payAmount := winner.PayAmount.AmountInCents()
		a.winningBidAmount = &payAmount
	}

	closedState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateClosed)
	if err != nil {
		return err
	}

	a.state = closedState
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

func (a *AuctionModel) Cancel() error {
	currentState := a.state.String()

	if currentState != enum.EnumAuctionStateDraft && currentState != enum.EnumAuctionStateActive {
		return errs.ErrAuctionCanOnlyCancelFromDraftOrActive
	}

	cancelledState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateCancelled)
	if err != nil {
		return err
	}

	a.state = cancelledState
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

func (a *AuctionModel) CheckAndCloseIfExpired(bids []BidModel) (bool, error) {
	if a.state.String() != enum.EnumAuctionStateActive {
		return false, nil
	}

	if time.Now().UTC().Before(a.endTime) {
		return false, nil
	}

	if err := a.Close(bids); err != nil {
		return false, err
	}

	return true, nil
}

func (a *AuctionModel) MaybeExtendEndTime(now time.Time) {
	if !a.antiSnipeEnabled {
		return
	}

	window := time.Duration(a.extensionWindowSec) * time.Second
	if now.Before(a.endTime) && now.Add(window).After(a.endTime) {
		a.endTime = now.Add(window)
		a.version++
		a.updatedAt = now
	}
}

func (a *AuctionModel) DecrementDutchPrice(now time.Time) {
	if a.tradingMode.String() != enum.EnumTradingModeDutch {
		return
	}

	if a.startingPrice == nil || a.startTime == nil {
		return
	}

	step := uint64(1)
	if a.priceStep != nil && *a.priceStep > 0 {
		step = *a.priceStep
	}

	reserve := uint64(0)
	if a.reservePrice != nil {
		reserve = *a.reservePrice
	}

	elapsed := now.UTC().Sub(a.startTime.UTC())
	if elapsed < 0 {
		return
	}

	drops := uint64(elapsed.Minutes())
	dropped := drops * step

	price := uint64(0)
	if *a.startingPrice > dropped {
		price = *a.startingPrice - dropped
	}

	if price < reserve {
		price = reserve
	}

	a.currentPrice = &price
	a.version++
	a.updatedAt = now.UTC()
}

func ToBidViews(bids []BidModel) []strategy.BidView {
	views := make([]strategy.BidView, 0, len(bids))
	for index := range bids {
		views = append(views, bidViewAdapter{bid: &bids[index]})
	}

	return views
}

type bidViewAdapter struct {
	bid *BidModel
}

func (adapter bidViewAdapter) ID() uint64 {
	return adapter.bid.ID()
}

func (adapter bidViewAdapter) UserID() uint64 {
	return adapter.bid.UserID()
}

func (adapter bidViewAdapter) Amount() strategy.MoneyView {
	return adapter.bid.Amount()
}

func (adapter bidViewAdapter) MaxAmount() strategy.MoneyView {
	if adapter.bid.MaxAmount() == nil {
		return nil
	}

	return *adapter.bid.MaxAmount()
}

func validateNewAuction(listingID uint64, endTime time.Time) error {
	if listingID == 0 {
		return errs.ErrListingIDRequired
	}

	if endTime.IsZero() {
		return errs.ErrEndTimeRequired
	}

	if endTime.Before(time.Now().UTC()) || endTime.Equal(time.Now().UTC()) {
		return errs.ErrEndTimeMustBeInFuture
	}

	return nil
}

// validateScheduledStartTime validates an optional scheduled start time used by
// the automatic auction scheduler to activate draft auctions.
func validateScheduledStartTime(startTime *time.Time, endTime time.Time) error {
	if startTime == nil {
		return nil
	}

	if startTime.UTC().Before(time.Now().UTC()) {
		return errs.ErrStartTimeMustBeInFuture
	}

	if !startTime.UTC().Before(endTime.UTC()) {
		return errs.ErrStartTimeMustBeBeforeEndTime
	}

	return nil
}

func validateRestoreAuction(listingID uint64, endTime time.Time) error {
	if listingID == 0 {
		return errs.ErrListingIDRequired
	}

	if endTime.IsZero() {
		return errs.ErrEndTimeRequired
	}

	return nil
}
