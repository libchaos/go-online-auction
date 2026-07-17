package strategy

import (
	"auction/internal/modules/auction/domain/enum"
	"auction/internal/modules/auction/domain/errs"
)

type EbayProxyAuctionStrategy struct{}

func NewEbayProxyAuctionStrategy() EbayProxyAuctionStrategy {
	return EbayProxyAuctionStrategy{}
}

func (strategy EbayProxyAuctionStrategy) Mode() enum.TradingModeEnum {
	return mode(enum.EnumTradingModeEbayProxy)
}

func (strategy EbayProxyAuctionStrategy) ValidateBid(auction AuctionView, amount MoneyView) error {
	currentPublic := uint64(0)
	if auction.HighestBidAmount() != nil {
		currentPublic = *auction.HighestBidAmount()
	}

	if amount.AmountInCents() <= currentPublic {
		return errs.ErrProxyMaxTooLow
	}

	return nil
}

func (strategy EbayProxyAuctionStrategy) SuggestNextPrice(auction AuctionView) (MoneyView, error) {
	step := uint64(1)
	if auction.PriceStep() != nil && *auction.PriceStep() > 0 {
		step = *auction.PriceStep()
	}

	currentPublic := uint64(0)
	if auction.HighestBidAmount() != nil {
		currentPublic = *auction.HighestBidAmount()
	}

	return NewMoney(currentPublic + step), nil
}

type proxyBid struct {
	userID uint64
	max    uint64
	bidID  uint64
}

func aggregateProxyBids(bids []BidView) map[uint64]*proxyBid {
	aggregates := map[uint64]*proxyBid{}
	for _, bid := range bids {
		maxCents := bid.Amount().AmountInCents()
		if bid.MaxAmount() != nil {
			maxCents = bid.MaxAmount().AmountInCents()
		}

		existing, ok := aggregates[bid.UserID()]
		if !ok {
			aggregates[bid.UserID()] = &proxyBid{userID: bid.UserID(), max: maxCents, bidID: bid.ID()}
			continue
		}
		if maxCents > existing.max {
			existing.max = maxCents
			existing.bidID = bid.ID()
		}
	}

	return aggregates
}

func pickProxyWinner(aggregates map[uint64]*proxyBid, reserve *uint64, step uint64) (Winner, error) {
	var top1, top2 *proxyBid
	for _, proxy := range aggregates {
		if top1 == nil || proxy.max > top1.max {
			top2 = top1
			top1 = proxy
			continue
		}
		if top2 == nil || proxy.max > top2.max {
			top2 = proxy
		}
	}

	if top1 == nil {
		return Winner{}, nil
	}

	if reserve != nil && top1.max < *reserve {
		return Winner{}, errs.ErrReserveNotMet
	}

	pay := proxyPay(top1, top2, reserve, step)
	bidID := top1.bidID

	return Winner{UserID: top1.userID, PayAmount: NewMoney(pay), BidID: &bidID}, nil
}

func proxyPay(top1, top2 *proxyBid, reserve *uint64, step uint64) uint64 {
	var pay uint64
	if top2 != nil {
		pay = top2.max + step
	} else {
		pay = top1.max
	}

	if pay > top1.max {
		pay = top1.max
	}

	if reserve != nil && pay < *reserve {
		reserveValue := *reserve
		if reserveValue < top1.max {
			pay = reserveValue
		}
	}

	return pay
}

func (strategy EbayProxyAuctionStrategy) DetermineWinner(auction AuctionView, bids []BidView) (Winner, error) {
	if len(bids) == 0 {
		return Winner{}, nil
	}

	aggregates := aggregateProxyBids(bids)
	step := uint64(1)
	if auction.PriceStep() != nil && *auction.PriceStep() > 0 {
		step = *auction.PriceStep()
	}

	return pickProxyWinner(aggregates, auction.ReservePrice(), step)
}

func (strategy EbayProxyAuctionStrategy) ShouldCloseOnAccept() bool {
	return false
}

func computeProxyPublicPrice(proxyMax map[uint64]uint64, base, step uint64) uint64 {
	var top1Max uint64
	var top2Max uint64
	for _, max := range proxyMax {
		if max > top1Max {
			top2Max = top1Max
			top1Max = max
			continue
		}
		if max > top2Max {
			top2Max = max
		}
	}

	var publicPrice uint64
	if top2Max > 0 {
		publicPrice = top2Max + step
	} else {
		publicPrice = base + step
	}
	if publicPrice > top1Max {
		publicPrice = top1Max
	}
	if publicPrice < base {
		publicPrice = base
	}

	return publicPrice
}

func (strategy EbayProxyAuctionStrategy) ResolveProxyBids(
	auction AuctionView,
	existing []BidView,
	newBidder uint64,
	newMax MoneyView,
) ([]ProxyAction, error) {
	step := uint64(1)
	if auction.PriceStep() != nil && *auction.PriceStep() > 0 {
		step = *auction.PriceStep()
	}

	base := uint64(0)
	if auction.StartingPrice() != nil {
		base = *auction.StartingPrice()
	}

	proxyMax := map[uint64]uint64{}
	for _, bid := range existing {
		maxCents := bid.Amount().AmountInCents()
		if bid.MaxAmount() != nil {
			maxCents = bid.MaxAmount().AmountInCents()
		}
		if maxCents > proxyMax[bid.UserID()] {
			proxyMax[bid.UserID()] = maxCents
		}
	}

	newMaxCents := newMax.AmountInCents()
	if _, exists := proxyMax[newBidder]; !exists {
		proxyMax[newBidder] = 0
	}
	if newMaxCents > proxyMax[newBidder] {
		proxyMax[newBidder] = newMaxCents
	}

	currentPublic := base
	if auction.HighestBidAmount() != nil && *auction.HighestBidAmount() > currentPublic {
		currentPublic = *auction.HighestBidAmount()
	}

	if proxyMax[newBidder] <= currentPublic {
		return nil, nil
	}

	top1User, top1Max := topProxyBidder(proxyMax)
	publicPrice := computeProxyPublicPrice(proxyMax, base, step)

	return []ProxyAction{
		{
			UserID:    top1User,
			Amount:    NewMoney(publicPrice),
			MaxAmount: NewMoney(top1Max),
		},
	}, nil
}

func topProxyBidder(proxyMax map[uint64]uint64) (uint64, uint64) {
	var top1User uint64
	var top1Max uint64
	for user, max := range proxyMax {
		if max > top1Max {
			top1Max = max
			top1User = user
		}
	}

	return top1User, top1Max
}
