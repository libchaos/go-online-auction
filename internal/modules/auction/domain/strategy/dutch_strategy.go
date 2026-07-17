package strategy

import (
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

type DutchAuctionStrategy struct{}

func NewDutchAuctionStrategy() DutchAuctionStrategy {
	return DutchAuctionStrategy{}
}

func (strategy DutchAuctionStrategy) Mode() enum.TradingModeEnum {
	return mode(enum.EnumTradingModeDutch)
}

func (strategy DutchAuctionStrategy) ValidateBid(auction AuctionView, amount MoneyView) error {
	if auction.CurrentPrice() == nil {
		return errs.ErrDutchPriceNotAvailable
	}

	if amount.AmountInCents() != *auction.CurrentPrice() {
		return errs.ErrDutchBidMustMatchPrice
	}

	return nil
}

func (strategy DutchAuctionStrategy) SuggestNextPrice(auction AuctionView) (MoneyView, error) {
	if auction.CurrentPrice() == nil {
		return nil, errs.ErrDutchPriceNotAvailable
	}

	return NewMoney(*auction.CurrentPrice()), nil
}

func (strategy DutchAuctionStrategy) DetermineWinner(_ AuctionView, bids []BidView) (Winner, error) {
	if len(bids) == 0 {
		return Winner{}, nil
	}

	first := bids[0]
	amount := NewMoney(first.Amount().AmountInCents())
	bidID := first.ID()

	return Winner{UserID: first.UserID(), PayAmount: amount, BidID: &bidID}, nil
}

func (strategy DutchAuctionStrategy) ShouldCloseOnAccept() bool {
	return true
}
