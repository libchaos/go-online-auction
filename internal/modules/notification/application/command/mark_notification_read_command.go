package command

import (
	"context"

	"auction/internal/modules/notification/domain/errs"
	"auction/internal/modules/notification/ports"
)

type MarkNotificationReadCommandInput struct {
	NotificationID uint64
	UserID         uint64
}

type MarkNotificationReadCommand struct {
	notifications ports.NotificationRepository
}

func NewMarkNotificationReadCommand(notifications ports.NotificationRepository) *MarkNotificationReadCommand {
	return &MarkNotificationReadCommand{notifications: notifications}
}

// Execute marks a single notification read. It is scoped to the owning user so a
// caller can never mark another user's notification read. A missing or
// already-read notification yields ErrNotificationNotFound.
func (command *MarkNotificationReadCommand) Execute(
	ctx context.Context,
	input MarkNotificationReadCommandInput,
) error {
	updated, err := command.notifications.MarkRead(ctx, input.NotificationID, input.UserID)
	if err != nil {
		return err
	}

	if !updated {
		return errs.ErrNotificationNotFound
	}

	return nil
}
