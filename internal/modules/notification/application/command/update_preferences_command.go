package command

import (
	"context"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/ports"
)

type UpdatePreferencesCommandInput struct {
	UserID   uint64
	Settings model.PreferenceSettings
}

type UpdatePreferencesCommandOutput struct {
	Preference model.NotificationPreferenceModel
}

type UpdatePreferencesCommand struct {
	preferences ports.PreferenceRepository
}

func NewUpdatePreferencesCommand(preferences ports.PreferenceRepository) *UpdatePreferencesCommand {
	return &UpdatePreferencesCommand{preferences: preferences}
}

// Execute upserts the user's notification preferences. A nil settings map falls
// back to the in-app defaults so the persisted row is always well-formed.
func (command *UpdatePreferencesCommand) Execute(
	ctx context.Context,
	input UpdatePreferencesCommandInput,
) (UpdatePreferencesCommandOutput, error) {
	preference, buildErr := model.NewNotificationPreference(input.UserID, input.Settings)
	if buildErr != nil {
		return UpdatePreferencesCommandOutput{}, buildErr
	}

	persisted, saveErr := command.preferences.Upsert(ctx, preference)
	if saveErr != nil {
		return UpdatePreferencesCommandOutput{}, saveErr
	}

	return UpdatePreferencesCommandOutput{Preference: persisted}, nil
}
