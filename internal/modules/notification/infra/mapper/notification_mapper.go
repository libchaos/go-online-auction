package mapper

import (
	"encoding/json"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/sqlcgen"
)

// NotificationMapper converts between the notification domain models and the
// sqlc generated persistence structs.
type NotificationMapper struct{}

func NewNotificationMapper() *NotificationMapper {
	return &NotificationMapper{}
}

func (mapper *NotificationMapper) ToNotificationDomain(row sqlcgen.Notification) (model.NotificationModel, error) {
	return model.RestoreNotificationModel(
		uint64(row.ID),
		uint64(row.UserID),
		row.Category,
		row.Type,
		row.Title,
		row.Body,
		row.Payload,
		row.Channels,
		row.IdempotencyKey,
		row.ReadAt,
		row.CreatedAt,
	)
}

func (mapper *NotificationMapper) ToInsertNotificationParams(
	notification model.NotificationModel,
) sqlcgen.InsertNotificationParams {
	domainChannels := notification.Channels()
	channels := make([]string, 0, len(domainChannels))
	for index := range domainChannels {
		channel := domainChannels[index]
		channels = append(channels, channel.String())
	}

	category := notification.Category()
	notificationType := notification.Type()

	return sqlcgen.InsertNotificationParams{
		UserID:         int64(notification.UserID()),
		Category:       category.String(),
		Type:           notificationType.String(),
		Title:          notification.Title(),
		Body:           notification.Body(),
		Payload:        notification.Payload(),
		Channels:       channels,
		IdempotencyKey: notification.IdempotencyKey(),
	}
}

func (mapper *NotificationMapper) ToPreferenceDomain(
	row sqlcgen.NotificationPreference,
) (model.NotificationPreferenceModel, error) {
	settings := model.PreferenceSettings{}
	if len(row.Preferences) > 0 {
		if err := json.Unmarshal(row.Preferences, &settings); err != nil {
			return model.NotificationPreferenceModel{}, err
		}
	}

	return model.RestoreNotificationPreference(uint64(row.UserID), settings, row.UpdatedAt)
}

func (mapper *NotificationMapper) ToUpsertPreferenceParams(
	preference model.NotificationPreferenceModel,
) (sqlcgen.UpsertNotificationPreferencesParams, error) {
	encoded, err := json.Marshal(preference.Settings())
	if err != nil {
		return sqlcgen.UpsertNotificationPreferencesParams{}, err
	}

	return sqlcgen.UpsertNotificationPreferencesParams{
		UserID:      int64(preference.UserID()),
		Preferences: encoded,
	}, nil
}
