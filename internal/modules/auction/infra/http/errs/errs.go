package errs

import (
	"errors"
	"net/http"

	domainerrs "github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
	"github.com/cristiano-pacheco/go-online-auction/pkg/errs"
)

var (
	ErrAuctionNotFound       = errs.New("AUCTION_01", "Auction not found", http.StatusNotFound, nil)
	ErrAuctionNotActive      = errs.New("AUCTION_02", "Auction is not active", http.StatusBadRequest, nil)
	ErrAuctionAlreadyStarted = errs.New("AUCTION_03", "Auction already started", http.StatusBadRequest, nil)
	ErrAuctionAlreadyClosed  = errs.New("AUCTION_04", "Auction already closed", http.StatusBadRequest, nil)
	ErrAuctionCancelled      = errs.New("AUCTION_05", "Auction is already cancelled", http.StatusBadRequest, nil)
	ErrInvalidEndTime        = errs.New("AUCTION_06", "End time must be after start time", http.StatusBadRequest, nil)
	ErrBidNotFound           = errs.New("AUCTION_07", "Bid not found", http.StatusNotFound, nil)
	ErrBidTooLow             = errs.New(
		"AUCTION_08",
		"Bid amount must exceed current highest bid",
		http.StatusBadRequest,
		nil,
	)
	ErrBidAmountInvalid     = errs.New("AUCTION_09", "Bid amount must be positive", http.StatusBadRequest, nil)
	ErrOptimisticLockFailed = errs.New(
		"AUCTION_10",
		"Resource was modified by another transaction",
		http.StatusConflict,
		nil,
	)
	ErrInvalidRequest               = errs.New("AUCTION_11", "Invalid request body", http.StatusBadRequest, nil)
	ErrInvalidAuctionID             = errs.New("AUCTION_12", "Invalid auction ID", http.StatusBadRequest, nil)
	ErrAuctionExpired               = errs.New("AUCTION_13", "Auction has expired", http.StatusBadRequest, nil)
	ErrInvalidAuctionState          = errs.New("AUCTION_14", "Invalid auction state", http.StatusBadRequest, nil)
	ErrAuctionCanOnlyStartFromDraft = errs.New("AUCTION_15", "Auction can only start from draft state", http.StatusBadRequest, nil)
	ErrTransactionFailed            = errs.New("AUCTION_16", "Transaction failed", http.StatusInternalServerError, nil)
	ErrAuctionIDRequired            = errs.New("AUCTION_17", "Auction ID must be greater than zero", http.StatusBadRequest, nil)
	ErrListingIDRequired            = errs.New("AUCTION_18", "Listing ID must be greater than zero", http.StatusBadRequest, nil)
	ErrEndTimeRequired              = errs.New("AUCTION_19", "End time is required", http.StatusBadRequest, nil)
	ErrBidIDRequired                = errs.New("AUCTION_20", "Bid ID must be greater than zero", http.StatusBadRequest, nil)
	ErrUserIDRequired               = errs.New("AUCTION_21", "User ID must be greater than zero", http.StatusBadRequest, nil)
)

var domainToHTTPErrorMap = []struct {
	domainError error
	httpError   error
}{
	{domainerrs.ErrAuctionNotFound, ErrAuctionNotFound},
	{domainerrs.ErrBidNotFound, ErrBidNotFound},
	{domainerrs.ErrConcurrencyConflict, ErrOptimisticLockFailed},
	{domainerrs.ErrAuctionCanOnlyStartFromDraft, ErrAuctionCanOnlyStartFromDraft},
	{domainerrs.ErrBidsOnlyOnActiveAuctions, ErrAuctionNotActive},
	{domainerrs.ErrAuctionExpired, ErrAuctionExpired},
	{domainerrs.ErrAuctionCanOnlyCloseFromActive, ErrAuctionAlreadyClosed},
	{domainerrs.ErrAuctionCanOnlyCancelFromDraftOrActive, ErrAuctionCancelled},
	{domainerrs.ErrFirstBidMustBePositive, ErrBidAmountInvalid},
	{domainerrs.ErrBidMustExceedHighest, ErrBidTooLow},
	{domainerrs.ErrEndTimeMustBeInFuture, ErrInvalidEndTime},
	{domainerrs.ErrInvalidAuctionState, ErrInvalidAuctionState},
	{domainerrs.ErrTransactionFailed, ErrTransactionFailed},
	{domainerrs.ErrAuctionIDRequired, ErrAuctionIDRequired},
	{domainerrs.ErrListingIDRequired, ErrListingIDRequired},
	{domainerrs.ErrEndTimeRequired, ErrEndTimeRequired},
	{domainerrs.ErrBidIDRequired, ErrBidIDRequired},
	{domainerrs.ErrUserIDRequired, ErrUserIDRequired},
}

func MapDomainError(err error) error {
	for _, mapping := range domainToHTTPErrorMap {
		if errors.Is(err, mapping.domainError) {
			return mapping.httpError
		}
	}
	return err
}
