package enum

import "errors"

const (
	EnumOperationStatusPending   string = "pending"
	EnumOperationStatusCommitted string = "committed"
	EnumOperationStatusFailed    string = "failed"
)

type OperationStatusEnum struct {
	value string
}

func NewOperationStatusEnum(value string) (OperationStatusEnum, error) {
	if err := validateOperationStatusEnum(value); err != nil {
		return OperationStatusEnum{}, err
	}

	return OperationStatusEnum{value: value}, nil
}

func (enum *OperationStatusEnum) String() string {
	return enum.value
}

func validateOperationStatusEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumOperationStatusPending:   {},
		EnumOperationStatusCommitted: {},
		EnumOperationStatusFailed:    {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid ledger operation status: " + value)
	}

	return nil
}
