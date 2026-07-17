package strategy

import (
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

type EnglishAuctionStrategy struct{}

func NewEnglishAuctionStrategy() EnglishAuctionStrategy {
	return EnglishAuctionStrategy{}
}

func (strategy EnglishAuctionStrategy) Mode() enum.TradingModeEnum {
	return mode(enum.EnumTradingModeEnglish)
}

func (strategy EnglishAuctionStrategy) ValidateBid(auction AuctionView, amount MoneyView) error {
	if auction.HighestBidAmount() == nil {
		if amount.AmountInCents() == 0 {
			return errs.ErrFirstBidMustBePositive
		}

		return nil
	}

	currentHighest := NewMoney(*auction.HighestBidAmount())
	if !NewMoney(amount.AmountInCents()).IsGreaterThan(currentHighest) {
		return errs.ErrBidMustExceedHighest
	}

	return nil
}

func (strategy EnglishAuctionStrategy) SuggestNextPrice(auction AuctionView) (MoneyView, error) {
	step := uint64(1)
	if auction.PriceStep() != nil {
		step = *auction.PriceStep()
	}

	if auction.HighestBidAmount() == nil {
		if auction.StartingPrice() != nil {
			return NewMoney(*auction.StartingPrice()), nil
		}

		return NewMoney(0), nil
	}

	return NewMoney(*auction.HighestBidAmount()).IncrementBy(step), nil
}

func (strategy EnglishAuctionStrategy) DetermineWinner(auction AuctionView, bids []BidView) (Winner, error) {
	return highestBidWinner(auction, bids)
}

func (strategy EnglishAuctionStrategy) ShouldCloseOnAccept() bool {
	return false
}
