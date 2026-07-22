package model

import (
	"time"

	"auction/internal/modules/notification/domain/errs"
)

type Watchlist struct {
	id        uint64
	userID    uint64
	spuID     uint64
	createdAt time.Time
}

func NewWatchlist(userID, spuID uint64) (Watchlist, error) {
	if userID == 0 {
		return Watchlist{}, errs.ErrWatchUserRequired
	}

	if spuID == 0 {
		return Watchlist{}, errs.ErrWatchSpuRequired
	}

	return Watchlist{userID: userID, spuID: spuID}, nil
}

func ReconstructWatchlist(id, userID, spuID uint64, createdAt time.Time) Watchlist {
	return Watchlist{id: id, userID: userID, spuID: spuID, createdAt: createdAt}
}

func (watchlist Watchlist) ID() uint64 {
	return watchlist.id
}

func (watchlist Watchlist) UserID() uint64 {
	return watchlist.userID
}

func (watchlist Watchlist) SpuID() uint64 {
	return watchlist.spuID
}

func (watchlist Watchlist) CreatedAt() time.Time {
	return watchlist.createdAt
}
