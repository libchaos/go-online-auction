package ports

import "context"

// ListingValidator checks whether a listing (a SKU in the listing module) can
// back a new auction. Implemented outside the auction module so auctions stay
// decoupled from the listing catalog.
type ListingValidator interface {
	// IsAuctionable reports whether the listing exists and is in a state that
	// allows creating an auction for it (published with quantity available).
	IsAuctionable(ctx context.Context, listingID uint64) (bool, error)
}
