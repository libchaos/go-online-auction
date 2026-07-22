package ports

import "context"

// AuctionReadPort resolves auction facts that a source event alone does not
// carry. The bid-placed event only names the new highest bidder, so producing an
// "outbid" notification for the previous leader requires a read-only lookup of
// the prior highest bidder on that auction. The auction-ended event only carries
// the winning bid id, so notifying the winner requires resolving the bidder
// behind that bid.
type AuctionReadPort interface {
	FindPreviousHighestBidderID(
		ctx context.Context,
		auctionID uint64,
		currentBidderID uint64,
	) (uint64, bool, error)
	FindBidderIDByBidID(
		ctx context.Context,
		bidID uint64,
	) (uint64, bool, error)
}
