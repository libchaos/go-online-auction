package model

type MoneyModel struct {
	amountInCents uint64
}

func NewMoneyModel(amountInCents uint64) MoneyModel {
	return MoneyModel{
		amountInCents: amountInCents,
	}
}

func (m MoneyModel) AmountInCents() uint64 {
	return m.amountInCents
}

func (m MoneyModel) IsGreaterThan(other MoneyModel) bool {
	return m.amountInCents > other.amountInCents
}

func (m MoneyModel) IsGreaterThanOrEqual(other MoneyModel) bool {
	return m.amountInCents >= other.amountInCents
}
