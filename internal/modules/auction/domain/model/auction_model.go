package model

import (
	"time"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/enum"
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/domain/errs"
)

type AuctionModel struct {
	id               uint64
	listingID        uint64
	startTime        *time.Time // nil for draft auctions
	endTime          time.Time
	state            enum.AuctionStateEnum
	highestBidAmount *uint64 // nil if no bids yet
	version          uint64  // for optimistic locking
	createdAt        time.Time
	updatedAt        time.Time
}

func NewAuctionModel(listingID uint64, endTime time.Time) (AuctionModel, error) {
	if err := validateNewAuction(listingID, endTime); err != nil {
		return AuctionModel{}, err
	}

	draftState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateDraft)
	if err != nil {
		return AuctionModel{}, err
	}

	now := time.Now().UTC()
	return AuctionModel{
		listingID:        listingID,
		endTime:          endTime.UTC(),
		state:            draftState,
		highestBidAmount: nil,
		version:          1,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

func RestoreAuctionModel(
	id, listingID uint64,
	startTime *time.Time,
	endTime time.Time,
	state enum.AuctionStateEnum,
	highestBidAmount *uint64,
	version uint64,
	createdAt, updatedAt time.Time,
) (AuctionModel, error) {
	if id == 0 {
		return AuctionModel{}, errs.ErrAuctionIDRequired
	}

	if err := validateRestoreAuction(listingID, endTime); err != nil {
		return AuctionModel{}, err
	}

	var normalizedStartTime *time.Time
	if startTime != nil {
		utc := startTime.UTC()
		normalizedStartTime = &utc
	}

	return AuctionModel{
		id:               id,
		listingID:        listingID,
		startTime:        normalizedStartTime,
		endTime:          endTime.UTC(),
		state:            state,
		highestBidAmount: highestBidAmount,
		version:          version,
		createdAt:        createdAt.UTC(),
		updatedAt:        updatedAt.UTC(),
	}, nil
}

func (a *AuctionModel) ID() uint64 {
	return a.id
}

func (a *AuctionModel) ListingID() uint64 {
	return a.listingID
}

func (a *AuctionModel) StartTime() *time.Time {
	return a.startTime
}

func (a *AuctionModel) EndTime() time.Time {
	return a.endTime
}

func (a *AuctionModel) State() enum.AuctionStateEnum {
	return a.state
}

func (a *AuctionModel) HighestBidAmount() *uint64 {
	return a.highestBidAmount
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
		return errs.ErrAuctionCanOnlyStartFromDraft
	}

	activeState, err := enum.NewAuctionStateEnum(enum.EnumAuctionStateActive)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	a.startTime = &now
	a.state = activeState
	a.version++
	a.updatedAt = now

	return nil
}

func (a *AuctionModel) PlaceBid(amount MoneyModel) error {
	if a.state.String() != enum.EnumAuctionStateActive {
		return errs.ErrBidsOnlyOnActiveAuctions
	}

	if time.Now().UTC().After(a.endTime) {
		return errs.ErrAuctionExpired
	}

	// Validate bid amount against current highest
	if a.highestBidAmount == nil {
		if amount.AmountInCents() == 0 {
			return errs.ErrFirstBidMustBePositive
		}
	} else {
		currentHighest := NewMoneyModel(*a.highestBidAmount)
		if !amount.IsGreaterThan(currentHighest) {
			return errs.ErrBidMustExceedHighest
		}
	}

	amountInCents := amount.AmountInCents()
	a.highestBidAmount = &amountInCents
	a.version++
	a.updatedAt = time.Now().UTC()

	return nil
}

func (a *AuctionModel) Close() error {
	if a.state.String() != enum.EnumAuctionStateActive {
		return errs.ErrAuctionCanOnlyCloseFromActive
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

func (a *AuctionModel) Cancel() error {
	currentState := a.state.String()

	if currentState != enum.EnumAuctionStateDraft && currentState != enum.EnumAuctionStateActive {
		return errs.ErrAuctionCanOnlyCancelFromDraftOrActive
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

func (a *AuctionModel) CheckAndCloseIfExpired() (bool, error) {
	if a.state.String() != enum.EnumAuctionStateActive {
		return false, nil
	}

	if time.Now().UTC().Before(a.endTime) {
		return false, nil
	}

	if err := a.Close(); err != nil {
		return false, err
	}

	return true, nil
}

func validateNewAuction(listingID uint64, endTime time.Time) error {
	if listingID == 0 {
		return errs.ErrListingIDRequired
	}

	if endTime.IsZero() {
		return errs.ErrEndTimeRequired
	}

	return nil
}

func validateRestoreAuction(listingID uint64, endTime time.Time) error {
	if listingID == 0 {
		return errs.ErrListingIDRequired
	}

	if endTime.IsZero() {
		return errs.ErrEndTimeRequired
	}

	return nil
}
