package model

import (
	"time"

	"auction/internal/modules/notification/domain/enum"
	"auction/internal/modules/notification/domain/errs"
)

type PreferenceSettings map[string]map[string]bool

type NotificationPreferenceModel struct {
	userID    uint64
	settings  PreferenceSettings
	updatedAt time.Time
}

func NewNotificationPreference(userID uint64, settings PreferenceSettings) (NotificationPreferenceModel, error) {
	if userID == 0 {
		return NotificationPreferenceModel{}, errs.ErrPreferencesUserRequired
	}

	if settings == nil {
		settings = DefaultPreferences()
	}

	return NotificationPreferenceModel{
		userID:    userID,
		settings:  settings,
		updatedAt: time.Now().UTC(),
	}, nil
}

func RestoreNotificationPreference(
	userID uint64,
	settings PreferenceSettings,
	updatedAt time.Time,
) (NotificationPreferenceModel, error) {
	if userID == 0 {
		return NotificationPreferenceModel{}, errs.ErrPreferencesUserRequired
	}

	if settings == nil {
		settings = DefaultPreferences()
	}

	return NotificationPreferenceModel{
		userID:    userID,
		settings:  settings,
		updatedAt: updatedAt,
	}, nil
}

func DefaultPreferences() PreferenceSettings {
	// email is enabled by default only for the high-signal categories
	// (payment, deposit, system); auction/bid/listing stay in-app only to avoid
	// email noise. in_app is always on so the notification center is usable.
	defaultEmail := map[string]bool{
		enum.EnumNotificationCategoryPayment: true,
		enum.EnumNotificationCategoryDeposit: true,
		enum.EnumNotificationCategorySystem:  true,
	}

	categories := []string{
		enum.EnumNotificationCategoryPayment,
		enum.EnumNotificationCategoryDeposit,
		enum.EnumNotificationCategoryAuction,
		enum.EnumNotificationCategoryListing,
		enum.EnumNotificationCategorySystem,
	}

	settings := make(PreferenceSettings, len(categories))
	for _, category := range categories {
		settings[category] = map[string]bool{
			enum.EnumNotificationChannelInApp: true,
			enum.EnumNotificationChannelEmail: defaultEmail[category],
		}
	}

	return settings
}

func (preference *NotificationPreferenceModel) IsChannelEnabled(category string, channel string) bool {
	channels, ok := preference.settings[category]
	if !ok {
		defaults := DefaultPreferences()
		defaultChannels, defaultOK := defaults[category]
		if !defaultOK {
			return false
		}

		return defaultChannels[channel]
	}

	enabled, ok := channels[channel]
	if !ok {
		return false
	}

	return enabled
}

func (preference *NotificationPreferenceModel) UserID() uint64 {
	return preference.userID
}

func (preference *NotificationPreferenceModel) Settings() PreferenceSettings {
	return preference.settings
}

func (preference *NotificationPreferenceModel) UpdatedAt() time.Time {
	return preference.updatedAt
}
