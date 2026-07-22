package ports

import (
	"context"

	"auction/internal/modules/notification/domain/model"
)

// NotificationRepository persists and reads in-app notifications. Insertion is
// idempotent on the notification idempotency key so a redelivered source event
// never produces a duplicate notification.
type NotificationRepository interface {
	Save(ctx context.Context, notification model.NotificationModel) (model.NotificationModel, bool, error)
	FindByID(ctx context.Context, id uint64) (model.NotificationModel, error)
	ListByUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.NotificationModel, error)
	ListUnreadByUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.NotificationModel, error)
	CountUnread(ctx context.Context, userID uint64) (uint64, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) (bool, error)
	MarkAllRead(ctx context.Context, userID uint64) (uint64, error)
	Delete(ctx context.Context, id uint64, userID uint64) (bool, error)
}

// PreferenceRepository persists and reads per-user notification preferences.
type PreferenceRepository interface {
	Get(ctx context.Context, userID uint64) (model.NotificationPreferenceModel, error)
	Upsert(ctx context.Context, preference model.NotificationPreferenceModel) (model.NotificationPreferenceModel, error)
}
