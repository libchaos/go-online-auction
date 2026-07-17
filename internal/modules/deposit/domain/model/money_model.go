package model

import "auction/internal/modules/deposit/domain/errs"

type MoneyModel struct {
	amountInCents uint64
}

func NewMoneyModel(amountInCents uint64) MoneyModel {
	return MoneyModel{
		amountInCents: amountInCents,
	}
}

func (money MoneyModel) AmountInCents() uint64 {
	return money.amountInCents
}

func (money MoneyModel) IsZero() bool {
	return money.amountInCents == 0
}

func (money MoneyModel) IsGreaterThan(other MoneyModel) bool {
	return money.amountInCents > other.amountInCents
}

func (money MoneyModel) IsGreaterThanOrEqual(other MoneyModel) bool {
	return money.amountInCents >= other.amountInCents
}

func (money MoneyModel) IsLessThan(other MoneyModel) bool {
	return money.amountInCents < other.amountInCents
}

func (money MoneyModel) Add(other MoneyModel) MoneyModel {
	return MoneyModel{amountInCents: money.amountInCents + other.amountInCents}
}

func (money MoneyModel) Subtract(other MoneyModel) (MoneyModel, error) {
	if money.amountInCents < other.amountInCents {
		return MoneyModel{}, errs.ErrInsufficientBalance
	}

	return MoneyModel{amountInCents: money.amountInCents - other.amountInCents}, nil
}
