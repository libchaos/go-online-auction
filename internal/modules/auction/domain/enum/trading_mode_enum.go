package enum

import "errors"

const (
	EnumTradingModeEnglish    string = "english"
	EnumTradingModeDutch      string = "dutch"
	EnumTradingModeSealedBid  string = "sealed_bid"
	EnumTradingModeVickrey    string = "vickrey"
	EnumTradingModeFixedPrice string = "fixed_price"
	EnumTradingModeEbayProxy  string = "ebay_proxy"
)

type TradingModeEnum struct {
	value string
}

func NewTradingModeEnum(value string) (TradingModeEnum, error) {
	if err := validateTradingModeEnum(value); err != nil {
		return TradingModeEnum{}, err
	}

	return TradingModeEnum{value: value}, nil
}

func (e *TradingModeEnum) String() string {
	return e.value
}

func validateTradingModeEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumTradingModeEnglish:    {},
		EnumTradingModeDutch:      {},
		EnumTradingModeSealedBid:  {},
		EnumTradingModeVickrey:    {},
		EnumTradingModeFixedPrice: {},
		EnumTradingModeEbayProxy:  {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid trading mode: " + value)
	}

	return nil
}
