package model

import (
	"time"

	"auction/internal/modules/ledger/domain/errs"
)

type AccountModel struct {
	id            uint64
	owner         string
	balance       uint64
	frozenBalance uint64
	currency      string
	version       uint64
	createdAt     time.Time
	updatedAt     time.Time
}

func NewAccount(owner string, currency string) (AccountModel, error) {
	if owner == "" {
		return AccountModel{}, errs.ErrAccountOwnerRequired
	}

	if currency == "" {
		return AccountModel{}, errs.ErrAccountCurrencyRequired
	}

	now := time.Now().UTC()

	return AccountModel{
		owner:         owner,
		balance:       0,
		frozenBalance: 0,
		currency:      currency,
		version:       1,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func RestoreAccountModel(
	id uint64,
	owner string,
	balance uint64,
	frozenBalance uint64,
	currency string,
	version uint64,
	createdAt time.Time,
	updatedAt time.Time,
) (AccountModel, error) {
	return AccountModel{
		id:            id,
		owner:         owner,
		balance:       balance,
		frozenBalance: frozenBalance,
		currency:      currency,
		version:       version,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func (account *AccountModel) Freeze(amount uint64) error {
	if account.balance < amount {
		return errs.ErrInsufficientBalance
	}

	account.balance -= amount
	account.frozenBalance += amount
	account.version++
	account.updatedAt = time.Now().UTC()

	return nil
}

func (account *AccountModel) Unfreeze(amount uint64) error {
	if account.frozenBalance < amount {
		return errs.ErrInsufficientFrozenBalance
	}

	account.frozenBalance -= amount
	account.balance += amount
	account.version++
	account.updatedAt = time.Now().UTC()

	return nil
}

func (account *AccountModel) WithdrawFromFrozen(amount uint64) error {
	if account.frozenBalance < amount {
		return errs.ErrInsufficientFrozenBalance
	}

	account.frozenBalance -= amount
	account.version++
	account.updatedAt = time.Now().UTC()

	return nil
}

func (account *AccountModel) ID() uint64 {
	return account.id
}

func (account *AccountModel) Owner() string {
	return account.owner
}

func (account *AccountModel) Balance() uint64 {
	return account.balance
}

func (account *AccountModel) FrozenBalance() uint64 {
	return account.frozenBalance
}

func (account *AccountModel) Currency() string {
	return account.currency
}

func (account *AccountModel) Version() uint64 {
	return account.version
}

func (account *AccountModel) CreatedAt() time.Time {
	return account.createdAt
}

func (account *AccountModel) UpdatedAt() time.Time {
	return account.updatedAt
}
