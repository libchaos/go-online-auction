package enum

import "errors"

const (
	EnumDepositStatusPending   string = "pending"
	EnumDepositStatusHeld      string = "held"
	EnumDepositStatusReleased  string = "released"
	EnumDepositStatusApplied   string = "applied"
	EnumDepositStatusForfeited string = "forfeited"
)

type DepositStatusEnum struct {
	value string
}

func NewDepositStatusEnum(value string) (DepositStatusEnum, error) {
	if err := validateDepositStatusEnum(value); err != nil {
		return DepositStatusEnum{}, err
	}

	return DepositStatusEnum{value: value}, nil
}

func (e *DepositStatusEnum) String() string {
	return e.value
}

func validateDepositStatusEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumDepositStatusPending:   {},
		EnumDepositStatusHeld:      {},
		EnumDepositStatusReleased:  {},
		EnumDepositStatusApplied:   {},
		EnumDepositStatusForfeited: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid deposit status: " + value)
	}

	return nil
}
