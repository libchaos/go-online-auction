package ports

import "context"

// NotificationUnitOfWork bundles the repositories that must commit atomically
// inside a single PostgreSQL transaction: the notification store, the
// preference store, and the transactional outbox store.
type NotificationUnitOfWork interface {
	NotificationRepository() NotificationRepository
	PreferenceRepository() PreferenceRepository
	OutboxRepository() NotificationOutboxRepository
	Complete(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// NotificationUnitOfWorkFactory opens a new unit of work on a pooled connection.
type NotificationUnitOfWorkFactory interface {
	Begin(ctx context.Context) (NotificationUnitOfWork, error)
}
