package ports

import "context"

// UserEmailResolver is a read-only adapter that resolves a user id to the email
// address the notification module should send mail to. It reaches into the
// shared users table without mutating it, mirroring the other read-only
// resolvers (AuctionReadPort, ListingReadPort, WatchlistRepository) the
// notification module already uses.
type UserEmailResolver interface {
	ResolveEmail(ctx context.Context, userID uint64) (string, bool, error)
}
