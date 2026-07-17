package strategy

import (
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

type FixedPriceAuctionStrategy struct{}

func NewFixedPriceAuctionStrategy() FixedPriceAuctionStrategy {
	return FixedPriceAuctionStrategy{}
}

func (strategy FixedPriceAuctionStrategy) Mode() enum.TradingModeEnum {
	return mode(enum.EnumTradingModeFixedPrice)
}

func (strategy FixedPriceAuctionStrategy) ValidateBid(auction AuctionView, amount MoneyView) error {
	if auction.StartingPrice() == nil {
		return errs.ErrFixedPriceNotConfigured
	}

	if amount.AmountInCents() != *auction.StartingPrice() {
		return errs.ErrFixedPriceMismatch
	}

	return nil
}

func (strategy FixedPriceAuctionStrategy) SuggestNextPrice(auction AuctionView) (MoneyView, error) {
	if auction.StartingPrice() == nil {
		return nil, errs.ErrFixedPriceNotConfigured
	}

	return NewMoney(*auction.StartingPrice()), nil
}

func (strategy FixedPriceAuctionStrategy) DetermineWinner(_ AuctionView, bids []BidView) (Winner, error) {
	if len(bids) == 0 {
		return Winner{}, nil
	}

	first := bids[0]
	amount := NewMoney(first.Amount().AmountInCents())
	bidID := first.ID()

	return Winner{UserID: first.UserID(), PayAmount: amount, BidID: &bidID}, nil
}

func (strategy FixedPriceAuctionStrategy) ShouldCloseOnAccept() bool {
	return true
}
