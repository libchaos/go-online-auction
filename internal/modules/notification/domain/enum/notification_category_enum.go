package enum

import "errors"

const (
	EnumNotificationCategoryPayment string = "payment"
	EnumNotificationCategoryDeposit string = "deposit"
	EnumNotificationCategoryAuction string = "auction"
	EnumNotificationCategoryListing string = "listing"
	EnumNotificationCategorySystem  string = "system"
)

type NotificationCategoryEnum struct {
	value string
}

func NewNotificationCategoryEnum(value string) (NotificationCategoryEnum, error) {
	if err := validateNotificationCategoryEnum(value); err != nil {
		return NotificationCategoryEnum{}, err
	}

	return NotificationCategoryEnum{value: value}, nil
}

func (e *NotificationCategoryEnum) String() string {
	return e.value
}

func validateNotificationCategoryEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumNotificationCategoryPayment: {},
		EnumNotificationCategoryDeposit: {},
		EnumNotificationCategoryAuction: {},
		EnumNotificationCategoryListing: {},
		EnumNotificationCategorySystem:  {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid notification category: " + value)
	}

	return nil
}
