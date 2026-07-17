package strategy

import "auction/internal/modules/auction/domain/errs"

func highestBidWinner(auction AuctionView, bids []BidView) (Winner, error) {
	var top BidView

	for _, bid := range bids {
		if top == nil {
			top = bid
			continue
		}
		if NewMoney(bid.Amount().AmountInCents()).IsGreaterThan(NewMoney(top.Amount().AmountInCents())) {
			top = bid
		}
	}

	if top == nil {
		return Winner{}, nil
	}

	if auction.ReservePrice() != nil && top.Amount().AmountInCents() < *auction.ReservePrice() {
		return Winner{}, errs.ErrReserveNotMet
	}

	amount := NewMoney(top.Amount().AmountInCents())
	bidID := top.ID()

	return Winner{UserID: top.UserID(), PayAmount: amount, BidID: &bidID}, nil
}
