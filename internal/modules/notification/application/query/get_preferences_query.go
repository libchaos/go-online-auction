package query

import (
	"context"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/ports"
)

type GetPreferencesQueryInput struct {
	UserID uint64
}

type GetPreferencesQueryOutput struct {
	Preference model.NotificationPreferenceModel
}

type GetPreferencesQuery struct {
	preferences ports.PreferenceRepository
}

func NewGetPreferencesQuery(preferences ports.PreferenceRepository) *GetPreferencesQuery {
	return &GetPreferencesQuery{preferences: preferences}
}

// Execute returns the user's notification preferences. When the user has never
// customized them the repository yields the in-app defaults so the response is
// always well-formed.
func (query *GetPreferencesQuery) Execute(
	ctx context.Context,
	input GetPreferencesQueryInput,
) (GetPreferencesQueryOutput, error) {
	preference, err := query.preferences.Get(ctx, input.UserID)
	if err != nil {
		return GetPreferencesQueryOutput{}, err
	}

	return GetPreferencesQueryOutput{Preference: preference}, nil
}
