package ports

import "context"

type DepositGuard interface {
	EnsureEligible(ctx context.Context, userID uint64, auctionID uint64) error
}
