package enum

import "errors"

const (
	EnumEntryTypeDeposit                  string = "deposit"
	EnumEntryTypeWithdraw                 string = "withdraw"
	EnumEntryTypeTransferOut              string = "transfer_out"
	EnumEntryTypeTransferIn               string = "transfer_in"
	EnumEntryTypeFreeze                   string = "freeze"
	EnumEntryTypeUnfreeze                 string = "unfreeze"
	EnumEntryTypeWithdrawFromFrozen       string = "withdraw_from_frozen"
	EnumEntryTypeWithdrawFromFrozenCredit string = "withdraw_from_frozen_credit"
)

type EntryTypeEnum struct {
	value string
}

func NewEntryTypeEnum(value string) (EntryTypeEnum, error) {
	if err := validateEntryTypeEnum(value); err != nil {
		return EntryTypeEnum{}, err
	}

	return EntryTypeEnum{value: value}, nil
}

func (enum *EntryTypeEnum) String() string {
	return enum.value
}

func validateEntryTypeEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumEntryTypeDeposit:                  {},
		EnumEntryTypeWithdraw:                 {},
		EnumEntryTypeTransferOut:              {},
		EnumEntryTypeTransferIn:               {},
		EnumEntryTypeFreeze:                   {},
		EnumEntryTypeUnfreeze:                 {},
		EnumEntryTypeWithdrawFromFrozen:       {},
		EnumEntryTypeWithdrawFromFrozenCredit: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid ledger entry type: " + value)
	}

	return nil
}
