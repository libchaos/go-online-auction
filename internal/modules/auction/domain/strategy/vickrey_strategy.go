package strategy

import (
	"sort"

	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

type VickreyAuctionStrategy struct{}

func NewVickreyAuctionStrategy() VickreyAuctionStrategy {
	return VickreyAuctionStrategy{}
}

func (strategy VickreyAuctionStrategy) Mode() enum.TradingModeEnum {
	return mode(enum.EnumTradingModeVickrey)
}

func (strategy VickreyAuctionStrategy) ValidateBid(_ AuctionView, amount MoneyView) error {
	if amount.AmountInCents() == 0 {
		return errs.ErrFirstBidMustBePositive
	}

	return nil
}

func (strategy VickreyAuctionStrategy) SuggestNextPrice(auction AuctionView) (MoneyView, error) {
	if auction.StartingPrice() != nil {
		return NewMoney(*auction.StartingPrice()), nil
	}

	return NewMoney(0), nil
}

func (strategy VickreyAuctionStrategy) DetermineWinner(auction AuctionView, bids []BidView) (Winner, error) {
	const minimumBiddersForSecondPrice = 2

	if len(bids) == 0 {
		return Winner{}, nil
	}

	sorted := make([]BidView, len(bids))
	copy(sorted, bids)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Amount().AmountInCents() > sorted[j].Amount().AmountInCents()
	})

	highest := sorted[0]
	if auction.ReservePrice() != nil && highest.Amount().AmountInCents() < *auction.ReservePrice() {
		return Winner{}, errs.ErrReserveNotMet
	}

	var pay uint64
	if len(sorted) >= minimumBiddersForSecondPrice {
		pay = sorted[1].Amount().AmountInCents()
	} else {
		pay = highest.Amount().AmountInCents()
	}

	if auction.ReservePrice() != nil && pay < *auction.ReservePrice() {
		reserve := *auction.ReservePrice()
		if reserve < highest.Amount().AmountInCents() {
			pay = reserve
		}
	}

	bidID := highest.ID()

	return Winner{UserID: highest.UserID(), PayAmount: NewMoney(pay), BidID: &bidID}, nil
}

func (strategy VickreyAuctionStrategy) ShouldCloseOnAccept() bool {
	return false
}
