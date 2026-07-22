package model

import (
	"time"

	"auction/internal/modules/deposit/domain/enum"
	"auction/internal/modules/deposit/domain/errs"
)

type DepositModel struct {
	id                uint64
	userID            uint64
	auctionID         uint64
	amount            MoneyModel
	currency          string
	status            enum.DepositStatusEnum
	externalReference string
	reference         string
	version           uint64
	createdAt         time.Time
	updatedAt         time.Time
}

func NewDeposit(
	userID uint64,
	auctionID uint64,
	amount MoneyModel,
	currency string,
	reference string,
) (DepositModel, error) {
	if userID == 0 {
		return DepositModel{}, errs.ErrDepositUserRequired
	}

	if auctionID == 0 {
		return DepositModel{}, errs.ErrDepositAuctionRequired
	}

	if amount.IsZero() {
		return DepositModel{}, errs.ErrDepositAmountRequired
	}

	if currency == "" {
		return DepositModel{}, errs.ErrDepositCurrencyRequired
	}

	if reference == "" {
		return DepositModel{}, errs.ErrDepositReferenceRequired
	}

	status, err := enum.NewDepositStatusEnum(enum.EnumDepositStatusPending)
	if err != nil {
		return DepositModel{}, err
	}

	now := time.Now().UTC()

	return DepositModel{
		userID:    userID,
		auctionID: auctionID,
		amount:    amount,
		currency:  currency,
		status:    status,
		reference: reference,
		version:   1,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreDepositModel(
	id uint64,
	userID uint64,
	auctionID uint64,
	amountInCents uint64,
	currency string,
	status string,
	externalReference string,
	reference string,
	version uint64,
	createdAt time.Time,
	updatedAt time.Time,
) (DepositModel, error) {
	parsedStatus, err := enum.NewDepositStatusEnum(status)
	if err != nil {
		return DepositModel{}, err
	}

	return DepositModel{
		id:                id,
		userID:            userID,
		auctionID:         auctionID,
		amount:            NewMoneyModel(amountInCents),
		currency:          currency,
		status:            parsedStatus,
		externalReference: externalReference,
		reference:         reference,
		version:           version,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}, nil
}

func (deposit *DepositModel) ConfirmHold(externalReference string) error {
	if deposit.status.String() != enum.EnumDepositStatusPending {
		return errs.ErrInvalidDepositTransition
	}

	deposit.status, _ = enum.NewDepositStatusEnum(enum.EnumDepositStatusHeld)
	deposit.externalReference = externalReference
	deposit.version++
	deposit.updatedAt = time.Now().UTC()

	return nil
}

func (deposit *DepositModel) Release() error {
	if deposit.status.String() != enum.EnumDepositStatusHeld {
		return errs.ErrInvalidDepositTransition
	}

	deposit.status, _ = enum.NewDepositStatusEnum(enum.EnumDepositStatusReleased)
	deposit.version++
	deposit.updatedAt = time.Now().UTC()

	return nil
}

func (deposit *DepositModel) ApplyToWinning() error {
	if deposit.status.String() != enum.EnumDepositStatusHeld {
		return errs.ErrInvalidDepositTransition
	}

	deposit.status, _ = enum.NewDepositStatusEnum(enum.EnumDepositStatusApplied)
	deposit.version++
	deposit.updatedAt = time.Now().UTC()

	return nil
}

func (deposit *DepositModel) Forfeit() error {
	if deposit.status.String() != enum.EnumDepositStatusHeld {
		return errs.ErrInvalidDepositTransition
	}

	deposit.status, _ = enum.NewDepositStatusEnum(enum.EnumDepositStatusForfeited)
	deposit.version++
	deposit.updatedAt = time.Now().UTC()

	return nil
}

func (deposit *DepositModel) Cancel() error {
	if deposit.status.String() != enum.EnumDepositStatusPending &&
		deposit.status.String() != enum.EnumDepositStatusHeld {
		return errs.ErrInvalidDepositTransition
	}

	deposit.status, _ = enum.NewDepositStatusEnum(enum.EnumDepositStatusReleased)
	deposit.version++
	deposit.updatedAt = time.Now().UTC()

	return nil
}

func (deposit *DepositModel) IsEligible(required MoneyModel) bool {
	return deposit.status.String() == enum.EnumDepositStatusHeld && deposit.amount.IsGreaterThanOrEqual(required)
}

func (deposit *DepositModel) ID() uint64 {
	return deposit.id
}

func (deposit *DepositModel) UserID() uint64 {
	return deposit.userID
}

func (deposit *DepositModel) AuctionID() uint64 {
	return deposit.auctionID
}

func (deposit *DepositModel) Amount() MoneyModel {
	return deposit.amount
}

func (deposit *DepositModel) Currency() string {
	return deposit.currency
}

func (deposit *DepositModel) Status() enum.DepositStatusEnum {
	return deposit.status
}

func (deposit *DepositModel) ExternalReference() string {
	return deposit.externalReference
}

func (deposit *DepositModel) Reference() string {
	return deposit.reference
}

func (deposit *DepositModel) Version() uint64 {
	return deposit.version
}

func (deposit *DepositModel) CreatedAt() time.Time {
	return deposit.createdAt
}

func (deposit *DepositModel) UpdatedAt() time.Time {
	return deposit.updatedAt
}
