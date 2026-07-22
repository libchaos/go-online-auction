package enum

import "errors"

const (
	EnumOperationTypeFreeze             string = "freeze"
	EnumOperationTypeUnfreeze           string = "unfreeze"
	EnumOperationTypeWithdrawFromFrozen string = "withdraw_from_frozen"
)

type OperationTypeEnum struct {
	value string
}

func NewOperationTypeEnum(value string) (OperationTypeEnum, error) {
	if err := validateOperationTypeEnum(value); err != nil {
		return OperationTypeEnum{}, err
	}

	return OperationTypeEnum{value: value}, nil
}

func (enum *OperationTypeEnum) String() string {
	return enum.value
}

func validateOperationTypeEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumOperationTypeFreeze:             {},
		EnumOperationTypeUnfreeze:           {},
		EnumOperationTypeWithdrawFromFrozen: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid ledger operation type: " + value)
	}

	return nil
}
