package enum

import "errors"

const (
	EnumNotificationTypeRechargeSuccess     string = "recharge_success"
	EnumNotificationTypeWithdrawalCompleted string = "withdrawal_completed"
	EnumNotificationTypeWithdrawalFailed    string = "withdrawal_failed"
	EnumNotificationTypeDepositHeld         string = "deposit_held"
	EnumNotificationTypeDepositReleased     string = "deposit_released"
	EnumNotificationTypeDepositApplied      string = "deposit_applied"
	EnumNotificationTypeDepositForfeited    string = "deposit_forfeited"
	EnumNotificationTypeOutbid              string = "outbid"
	EnumNotificationTypeAuctionStarted      string = "auction_started"
	EnumNotificationTypeAuctionEnded        string = "auction_ended"
	EnumNotificationTypeListingPublished    string = "listing_published"
	EnumNotificationTypeListingOffShelf     string = "listing_off_shelf"
)

type NotificationTypeEnum struct {
	value string
}

func NewNotificationTypeEnum(value string) (NotificationTypeEnum, error) {
	if err := validateNotificationTypeEnum(value); err != nil {
		return NotificationTypeEnum{}, err
	}

	return NotificationTypeEnum{value: value}, nil
}

func (e *NotificationTypeEnum) String() string {
	return e.value
}

func (e *NotificationTypeEnum) Category() NotificationCategoryEnum {
	return notificationTypeToCategory[e.value]
}

var notificationTypeToCategory = map[string]NotificationCategoryEnum{
	EnumNotificationTypeRechargeSuccess:     {value: EnumNotificationCategoryPayment},
	EnumNotificationTypeWithdrawalCompleted: {value: EnumNotificationCategoryPayment},
	EnumNotificationTypeWithdrawalFailed:    {value: EnumNotificationCategoryPayment},
	EnumNotificationTypeDepositHeld:         {value: EnumNotificationCategoryDeposit},
	EnumNotificationTypeDepositReleased:     {value: EnumNotificationCategoryDeposit},
	EnumNotificationTypeDepositApplied:      {value: EnumNotificationCategoryDeposit},
	EnumNotificationTypeDepositForfeited:    {value: EnumNotificationCategoryDeposit},
	EnumNotificationTypeOutbid:              {value: EnumNotificationCategoryAuction},
	EnumNotificationTypeAuctionStarted:      {value: EnumNotificationCategoryAuction},
	EnumNotificationTypeAuctionEnded:        {value: EnumNotificationCategoryAuction},
	EnumNotificationTypeListingPublished:    {value: EnumNotificationCategoryListing},
	EnumNotificationTypeListingOffShelf:     {value: EnumNotificationCategoryListing},
}

func validateNotificationTypeEnum(value string) error {
	if _, ok := notificationTypeToCategory[value]; !ok {
		return errors.New("invalid notification type: " + value)
	}

	return nil
}
