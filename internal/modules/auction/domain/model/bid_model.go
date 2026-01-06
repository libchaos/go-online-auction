package model

import (
	"errors"
	"time"
)

type BidModel struct {
	id        uint64
	auctionID uint64
	userID    uint64
	amount    MoneyModel
	createdAt time.Time
	updatedAt time.Time
}

func NewBidModel(auctionID, userID uint64, amount MoneyModel) (BidModel, error) {
	if err := validateBid(auctionID, userID); err != nil {
		return BidModel{}, err
	}

	now := time.Now().UTC()
	return BidModel{
		auctionID: auctionID,
		userID:    userID,
		amount:    amount,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreBidModel(
	id, auctionID, userID uint64,
	amount MoneyModel,
	createdAt, updatedAt time.Time,
) (BidModel, error) {
	if id == 0 {
		return BidModel{}, errors.New("bid id must be greater than zero")
	}

	if err := validateBid(auctionID, userID); err != nil {
		return BidModel{}, err
	}

	return BidModel{
		id:        id,
		auctionID: auctionID,
		userID:    userID,
		amount:    amount,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

func (b *BidModel) ID() uint64 {
	return b.id
}

func (b *BidModel) AuctionID() uint64 {
	return b.auctionID
}

func (b *BidModel) UserID() uint64 {
	return b.userID
}

func (b *BidModel) Amount() MoneyModel {
	return b.amount
}

func (b *BidModel) CreatedAt() time.Time {
	return b.createdAt
}

func (b *BidModel) UpdatedAt() time.Time {
	return b.updatedAt
}

func validateBid(auctionID, userID uint64) error {
	if auctionID == 0 {
		return errors.New("auction id must be greater than zero")
	}

	if userID == 0 {
		return errors.New("user id must be greater than zero")
	}

	return nil
}
