package command

import (
	"context"

	"auction/internal/modules/notification/ports"
)

type MarkAllNotificationsReadCommandInput struct {
	UserID uint64
}

type MarkAllNotificationsReadCommandOutput struct {
	UpdatedCount uint64
}

type MarkAllNotificationsReadCommand struct {
	notifications ports.NotificationRepository
}

func NewMarkAllNotificationsReadCommand(notifications ports.NotificationRepository) *MarkAllNotificationsReadCommand {
	return &MarkAllNotificationsReadCommand{notifications: notifications}
}

// Execute marks every unread notification of the user read and returns how many
// rows were affected.
func (command *MarkAllNotificationsReadCommand) Execute(
	ctx context.Context,
	input MarkAllNotificationsReadCommandInput,
) (MarkAllNotificationsReadCommandOutput, error) {
	updated, err := command.notifications.MarkAllRead(ctx, input.UserID)
	if err != nil {
		return MarkAllNotificationsReadCommandOutput{}, err
	}

	return MarkAllNotificationsReadCommandOutput{UpdatedCount: updated}, nil
}
