package enum

import "errors"

const (
	EnumNotificationChannelInApp string = "in_app"
	EnumNotificationChannelEmail string = "email"
)

type NotificationChannelEnum struct {
	value string
}

func NewNotificationChannelEnum(value string) (NotificationChannelEnum, error) {
	if err := validateNotificationChannelEnum(value); err != nil {
		return NotificationChannelEnum{}, err
	}

	return NotificationChannelEnum{value: value}, nil
}

func (e *NotificationChannelEnum) String() string {
	return e.value
}

func validateNotificationChannelEnum(value string) error {
	allowedValues := map[string]struct{}{
		EnumNotificationChannelInApp: {},
		EnumNotificationChannelEmail: {},
	}

	if _, ok := allowedValues[value]; !ok {
		return errors.New("invalid notification channel: " + value)
	}

	return nil
}
