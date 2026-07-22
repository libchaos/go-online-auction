package ports

import "context"

type ListingReadPort interface {
	FindSpuTitleByID(ctx context.Context, spuID uint64) (string, bool, error)
}
