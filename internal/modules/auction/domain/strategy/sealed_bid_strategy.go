package strategy

import (
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

type SealedBidAuctionStrategy struct{}

func NewSealedBidAuctionStrategy() SealedBidAuctionStrategy {
	return SealedBidAuctionStrategy{}
}

func (strategy SealedBidAuctionStrategy) Mode() enum.TradingModeEnum {
	return mode(enum.EnumTradingModeSealedBid)
}

func (strategy SealedBidAuctionStrategy) ValidateBid(_ AuctionView, amount MoneyView) error {
	if amount.AmountInCents() == 0 {
		return errs.ErrFirstBidMustBePositive
	}

	return nil
}

func (strategy SealedBidAuctionStrategy) SuggestNextPrice(auction AuctionView) (MoneyView, error) {
	if auction.StartingPrice() != nil {
		return NewMoney(*auction.StartingPrice()), nil
	}

	return NewMoney(0), nil
}

func (strategy SealedBidAuctionStrategy) DetermineWinner(auction AuctionView, bids []BidView) (Winner, error) {
	return highestBidWinner(auction, bids)
}

func (strategy SealedBidAuctionStrategy) ShouldCloseOnAccept() bool {
	return false
}
