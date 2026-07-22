package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/mapper"
	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/ports"
)

var _ ports.PreferenceRepository = (*PostgresPreferenceRepository)(nil)

type PostgresPreferenceRepository struct {
	q      *sqlcgen.Queries
	mapper *mapper.NotificationMapper
}

func NewPostgresPreferenceRepository(
	db sqlcgen.DBTX,
	notificationMapper *mapper.NotificationMapper,
) *PostgresPreferenceRepository {
	return &PostgresPreferenceRepository{
		q:      sqlcgen.New(db),
		mapper: notificationMapper,
	}
}

// Get returns the user's stored preferences, or the in-app defaults when the
// user has never customized them, so every caller receives a well-formed value.
func (repository *PostgresPreferenceRepository) Get(
	ctx context.Context,
	userID uint64,
) (model.NotificationPreferenceModel, error) {
	row, err := repository.q.GetNotificationPreferences(ctx, int64(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.NewNotificationPreference(userID, model.DefaultPreferences())
		}

		return model.NotificationPreferenceModel{}, err
	}

	return repository.mapper.ToPreferenceDomain(row)
}

func (repository *PostgresPreferenceRepository) Upsert(
	ctx context.Context,
	preference model.NotificationPreferenceModel,
) (model.NotificationPreferenceModel, error) {
	params, mapErr := repository.mapper.ToUpsertPreferenceParams(preference)
	if mapErr != nil {
		return model.NotificationPreferenceModel{}, mapErr
	}

	row, err := repository.q.UpsertNotificationPreferences(ctx, params)
	if err != nil {
		return model.NotificationPreferenceModel{}, err
	}

	return repository.mapper.ToPreferenceDomain(row)
}
