package strategy

type Money struct {
	amountInCents uint64
}

func NewMoney(amountInCents uint64) Money {
	return Money{amountInCents: amountInCents}
}

func (money Money) AmountInCents() uint64 {
	return money.amountInCents
}

func (money Money) Add(other Money) Money {
	return Money{amountInCents: money.amountInCents + other.amountInCents}
}

func (money Money) IncrementBy(step uint64) Money {
	return Money{amountInCents: money.amountInCents + step}
}

func (money Money) IsGreaterThan(other Money) bool {
	return money.amountInCents > other.amountInCents
}

func (money Money) IsGreaterThanOrEqual(other Money) bool {
	return money.amountInCents >= other.amountInCents
}
