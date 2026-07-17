package event

import "auction/internal/modules/auction/domain/model"

const (
	BidPlacedEventType = "bid_placed"
)

type BidPlacedEvent struct {
	DomainEvent
	bidID     uint64
	auctionID uint64
	userID    uint64
	amount    model.MoneyModel
}

func NewBidPlacedEvent(bidID, auctionID, userID uint64, amount model.MoneyModel) BidPlacedEvent {
	return BidPlacedEvent{
		DomainEvent: newDomainEvent(),
		bidID:       bidID,
		auctionID:   auctionID,
		userID:      userID,
		amount:      amount,
	}
}

func (e BidPlacedEvent) BidID() uint64 {
	return e.bidID
}

func (e BidPlacedEvent) AuctionID() uint64 {
	return e.auctionID
}

func (e BidPlacedEvent) UserID() uint64 {
	return e.userID
}

func (e BidPlacedEvent) Amount() model.MoneyModel {
	return e.amount
}
