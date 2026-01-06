package errs

import "errors"

var (
	ErrAuctionNotFound                       = errors.New("auction not found")
	ErrBidNotFound                           = errors.New("bid not found")
	ErrConcurrencyConflict                   = errors.New("concurrency conflict: resource was modified")
	ErrTransactionFailed                     = errors.New("transaction failed")
	ErrAuctionIDRequired                     = errors.New("auction id must be greater than zero")
	ErrListingIDRequired                     = errors.New("listing id must be greater than zero")
	ErrEndTimeRequired                       = errors.New("end time is required")
	ErrAuctionCanOnlyStartFromDraft          = errors.New("auction can only be started from draft state")
	ErrBidsOnlyOnActiveAuctions              = errors.New("bids can only be placed on active auctions")
	ErrAuctionExpired                        = errors.New("auction has expired")
	ErrAuctionCanOnlyCloseFromActive         = errors.New("auction can only be closed from active state")
	ErrAuctionCanOnlyCancelFromDraftOrActive = errors.New("auction can only be cancelled from draft or active state")
	ErrFirstBidMustBePositive                = errors.New("first bid amount must be greater than zero")
	ErrBidMustExceedHighest                  = errors.New("bid amount must exceed current highest bid")
	ErrBidIDRequired                         = errors.New("bid id must be greater than zero")
	ErrUserIDRequired                        = errors.New("user id must be greater than zero")
	ErrEndTimeMustBeInFuture                 = errors.New("end time must be in the future")
)
