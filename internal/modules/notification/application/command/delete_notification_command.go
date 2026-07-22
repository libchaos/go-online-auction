package command

import (
	"context"

	"auction/internal/modules/notification/domain/errs"
	"auction/internal/modules/notification/ports"
)

type DeleteNotificationCommandInput struct {
	NotificationID uint64
	UserID         uint64
}

type DeleteNotificationCommand struct {
	notifications ports.NotificationRepository
}

func NewDeleteNotificationCommand(notifications ports.NotificationRepository) *DeleteNotificationCommand {
	return &DeleteNotificationCommand{notifications: notifications}
}

// Execute deletes a single notification scoped to its owning user. A missing
// notification yields ErrNotificationNotFound.
func (command *DeleteNotificationCommand) Execute(
	ctx context.Context,
	input DeleteNotificationCommandInput,
) error {
	deleted, err := command.notifications.Delete(ctx, input.NotificationID, input.UserID)
	if err != nil {
		return err
	}

	if !deleted {
		return errs.ErrNotificationNotFound
	}

	return nil
}
