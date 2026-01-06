package errs

import "errors"

var (
	// ErrAuctionNotFound is returned when an auction lookup yields no results
	ErrAuctionNotFound = errors.New("auction not found")

	// ErrBidNotFound is returned when a bid lookup yields no results
	ErrBidNotFound = errors.New("bid not found")

	// ErrConcurrencyConflict is returned when optimistic lock fails due to version mismatch
	ErrConcurrencyConflict = errors.New("concurrency conflict: resource was modified")

	// ErrTransactionFailed is returned when transaction commit fails
	ErrTransactionFailed = errors.New("transaction failed")
)
