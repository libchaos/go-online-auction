package events

import "github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/model"

type AuctionEndedEvent struct {
	DomainEvent
	auctionID    uint64
	winningBidID *uint64
	finalAmount  *model.MoneyModel
}

// winningBidID and finalAmount can be nil if the auction ended without any bids
func NewAuctionEndedEvent(auctionID uint64, winningBidID *uint64, finalAmount *model.MoneyModel) AuctionEndedEvent {
	return AuctionEndedEvent{
		DomainEvent:  newDomainEvent(),
		auctionID:    auctionID,
		winningBidID: winningBidID,
		finalAmount:  finalAmount,
	}
}

func (e AuctionEndedEvent) AuctionID() uint64 {
	return e.auctionID
}

func (e AuctionEndedEvent) WinningBidID() *uint64 {
	return e.winningBidID
}

func (e AuctionEndedEvent) FinalAmount() *model.MoneyModel {
	return e.finalAmount
}
