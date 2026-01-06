package model

import (
	"errors"
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
)

// AuctionModel represents the auction aggregate root
type AuctionModel struct {
	id           uint64
	listingID    uint64
	startTime    time.Time
	endTime      time.Time
	state        enum.AuctionStateEnum
	highestBidID *uint64 // nil if no bids yet
	version      uint64  // for optimistic locking
	createdAt    time.Time
	updatedAt    time.Time
}

// NewAuctionModel creates a new auction in Draft state with version 0
func NewAuctionModel(listingID uint64, startTime, endTime time.Time) (AuctionModel, error) {
	if err := validateAuction(listingID, startTime, endTime); err != nil {
		return AuctionModel{}, err
	}

	draftState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	if err != nil {
		return AuctionModel{}, err
	}

	now := time.Now().UTC()
	return AuctionModel{
		listingID:    listingID,
		startTime:    startTime.UTC(),
		endTime:      endTime.UTC(),
		state:        draftState,
		highestBidID: nil,
		version:      0,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// RestoreAuctionModel reconstitutes an existing auction from persistence
func RestoreAuctionModel(
	id, listingID uint64,
	startTime, endTime time.Time,
	state enum.AuctionStateEnum,
	highestBidID *uint64,
	version uint64,
	createdAt, updatedAt time.Time,
) (AuctionModel, error) {
	if id == 0 {
		return AuctionModel{}, errors.New("auction id must be greater than zero")
	}

	if err := validateAuction(listingID, startTime, endTime); err != nil {
		return AuctionModel{}, err
	}

	return AuctionModel{
		id:           id,
		listingID:    listingID,
		startTime:    startTime.UTC(),
		endTime:      endTime.UTC(),
		state:        state,
		highestBidID: highestBidID,
		version:      version,
		createdAt:    createdAt.UTC(),
		updatedAt:    updatedAt.UTC(),
	}, nil
}

func (a *AuctionModel) ID() uint64 {
	return a.id
}

func (a *AuctionModel) ListingID() uint64 {
	return a.listingID
}

func (a *AuctionModel) StartTime() time.Time {
	return a.startTime
}

func (a *AuctionModel) EndTime() time.Time {
	return a.endTime
}

func (a *AuctionModel) State() enum.AuctionStateEnum {
	return a.state
}

func (a *AuctionModel) HighestBidID() *uint64 {
	return a.highestBidID
}

func (a *AuctionModel) Version() uint64 {
	return a.version
}

func (a *AuctionModel) CreatedAt() time.Time {
	return a.createdAt
}

func (a *AuctionModel) UpdatedAt() time.Time {
	return a.updatedAt
}

func (a *AuctionModel) Start() error {
	if a.state.String() != enum.EnumAuctionStateDraft {
		return errors.New("auction can only be started from draft state")
	}

	activeState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	if err != nil {
		return err
	}

	a.state = activeState
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

// PlaceBid validates and accepts a bid on the auction
func (a *AuctionModel) PlaceBid(bidID uint64, amount MoneyModel, currentHighestBid *BidModel) error {
	// Validate auction is in Active state
	if a.state.String() != enum.EnumAuctionStateActive {
		return errors.New("bids can only be placed on active auctions")
	}

	// Validate auction has not expired
	if time.Now().UTC().After(a.endTime) {
		return errors.New("auction has expired")
	}

	// Validate bid amount
	if err := validateBidAmount(amount, currentHighestBid); err != nil {
		return err
	}

	// Update highest bid
	a.highestBidID = &bidID
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

// Close transitions the auction from Active to Closed state
func (a *AuctionModel) Close() error {
	if a.state.String() != enum.EnumAuctionStateActive {
		return errors.New("auction can only be closed from active state")
	}

	closedState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateClosed)
	if err != nil {
		return err
	}

	a.state = closedState
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

// Cancel transitions the auction to Cancelled state
func (a *AuctionModel) Cancel() error {
	currentState := a.state.String()

	// Can only cancel from Draft or Active state
	if currentState != enum.EnumAuctionStateDraft && currentState != enum.EnumAuctionStateActive {
		return errors.New("auction can only be cancelled from draft or active state")
	}

	cancelledState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateCancelled)
	if err != nil {
		return err
	}

	a.state = cancelledState
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

// CheckAndCloseIfExpired automatically closes the auction if it has expired
// Returns true if the auction was closed, false otherwise
func (a *AuctionModel) CheckAndCloseIfExpired() (bool, error) {
	// Only close if auction is Active and has expired
	if a.state.String() != enum.EnumAuctionStateActive {
		return false, nil
	}

	if time.Now().UTC().Before(a.endTime) {
		return false, nil
	}

	// Auction has expired, close it
	if err := a.Close(); err != nil {
		return false, err
	}

	return true, nil
}

func validateAuction(listingID uint64, startTime, endTime time.Time) error {
	if listingID == 0 {
		return errors.New("listing id must be greater than zero")
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return errors.New("end time must be after start time")
	}

	return nil
}

// validateBidAmount validates the bid amount against business rules
func validateBidAmount(amount MoneyModel, currentHighestBid *BidModel) error {
	// First bid: any positive amount is valid
	if currentHighestBid == nil {
		if amount.AmountInCents() == 0 {
			return errors.New("first bid amount must be greater than zero")
		}
		return nil
	}

	// Subsequent bids: must exceed current highest bid
	if !amount.IsGreaterThan(currentHighestBid.Amount()) {
		return errors.New("bid amount must exceed current highest bid")
	}

	return nil
}
