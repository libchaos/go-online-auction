package errs

import (
	"net/http"

	"github.com/cristiano-pacheco/go-online-auction/pkg/errs"
)

var (
	ErrAuctionNotFound       = errs.New("AUCTION_01", "Auction not found", http.StatusNotFound, nil)
	ErrAuctionNotActive      = errs.New("AUCTION_02", "Auction is not active", http.StatusBadRequest, nil)
	ErrAuctionAlreadyStarted = errs.New("AUCTION_03", "Auction already started", http.StatusBadRequest, nil)
	ErrAuctionAlreadyClosed  = errs.New("AUCTION_04", "Auction already closed", http.StatusBadRequest, nil)
	ErrAuctionCancelled      = errs.New("AUCTION_05", "Auction has been cancelled", http.StatusBadRequest, nil)
	ErrInvalidEndTime        = errs.New("AUCTION_06", "End time must be after start time", http.StatusBadRequest, nil)
	ErrBidNotFound           = errs.New("AUCTION_07", "Bid not found", http.StatusNotFound, nil)
	ErrBidTooLow             = errs.New("AUCTION_08", "Bid amount must exceed current highest bid", http.StatusBadRequest, nil)
	ErrBidAmountInvalid      = errs.New("AUCTION_09", "Bid amount must be positive", http.StatusBadRequest, nil)
	ErrOptimisticLockFailed  = errs.New("AUCTION_10", "Resource was modified by another transaction", http.StatusConflict, nil)
	ErrInvalidRequest        = errs.New("AUCTION_11", "Invalid request body", http.StatusBadRequest, nil)
	ErrInvalidAuctionID      = errs.New("AUCTION_12", "Invalid auction ID", http.StatusBadRequest, nil)
	ErrAuctionExpired        = errs.New("AUCTION_13", "Auction has expired", http.StatusBadRequest, nil)
)

