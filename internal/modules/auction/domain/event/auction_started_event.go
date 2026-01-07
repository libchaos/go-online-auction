package event

import "time"

const (
	AuctionStartedEventType = "auction_started"
)

type AuctionStartedEvent struct {
	DomainEvent
	auctionID uint64
	listingID uint64
	startTime time.Time
	endTime   time.Time
}

func NewAuctionStartedEvent(auctionID, listingID uint64, startTime *time.Time, endTime time.Time) AuctionStartedEvent {
	var st time.Time
	if startTime != nil {
		st = *startTime
	}
	return AuctionStartedEvent{
		DomainEvent: newDomainEvent(),
		auctionID:   auctionID,
		listingID:   listingID,
		startTime:   st,
		endTime:     endTime,
	}
}

func (e AuctionStartedEvent) AuctionID() uint64 {
	return e.auctionID
}

func (e AuctionStartedEvent) ListingID() uint64 {
	return e.listingID
}

func (e AuctionStartedEvent) StartTime() time.Time {
	return e.startTime
}

func (e AuctionStartedEvent) EndTime() time.Time {
	return e.endTime
}
