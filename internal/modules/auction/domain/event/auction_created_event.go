package event

import "time"

const (
	AuctionCreatedEventType = "auction_created"
)

// AuctionCreatedEvent is emitted when an auction is persisted. It completes
// the auction lifecycle event stream so downstream consumers (search index,
// audit, category linkage, notifications) can react to creation just like they
// do to start/close/place_bid.
type AuctionCreatedEvent struct {
	DomainEvent
	auctionID          uint64
	listingID          uint64
	tradingMode        string
	startingPrice      *uint64
	priceStep          *uint64
	reservePrice       *uint64
	antiSnipeEnabled   bool
	extensionWindowSec int64
	startTime          *time.Time
	endTime            time.Time
}

func NewAuctionCreatedEvent(
	auctionID, listingID uint64,
	tradingMode string,
	startingPrice, priceStep, reservePrice *uint64,
	antiSnipeEnabled bool,
	extensionWindowSec int64,
	startTime *time.Time,
	endTime time.Time,
) AuctionCreatedEvent {
	return AuctionCreatedEvent{
		DomainEvent:        newDomainEvent(),
		auctionID:          auctionID,
		listingID:          listingID,
		tradingMode:        tradingMode,
		startingPrice:      startingPrice,
		priceStep:          priceStep,
		reservePrice:       reservePrice,
		antiSnipeEnabled:   antiSnipeEnabled,
		extensionWindowSec: extensionWindowSec,
		startTime:          startTime,
		endTime:            endTime,
	}
}

func (e AuctionCreatedEvent) AuctionID() uint64 {
	return e.auctionID
}

func (e AuctionCreatedEvent) ListingID() uint64 {
	return e.listingID
}

func (e AuctionCreatedEvent) TradingMode() string {
	return e.tradingMode
}

func (e AuctionCreatedEvent) StartingPrice() *uint64 {
	return e.startingPrice
}

func (e AuctionCreatedEvent) PriceStep() *uint64 {
	return e.priceStep
}

func (e AuctionCreatedEvent) ReservePrice() *uint64 {
	return e.reservePrice
}

func (e AuctionCreatedEvent) AntiSnipeEnabled() bool {
	return e.antiSnipeEnabled
}

func (e AuctionCreatedEvent) ExtensionWindowSec() int64 {
	return e.extensionWindowSec
}

func (e AuctionCreatedEvent) StartTime() *time.Time {
	return e.startTime
}

func (e AuctionCreatedEvent) EndTime() time.Time {
	return e.endTime
}
