package query

import (
	"context"

	"auction/internal/modules/notification/ports"
)

type GetUnreadCountQueryInput struct {
	UserID uint64
}

type GetUnreadCountQueryOutput struct {
	UnreadCount uint64
}

type GetUnreadCountQuery struct {
	notifications ports.NotificationRepository
}

func NewGetUnreadCountQuery(notifications ports.NotificationRepository) *GetUnreadCountQuery {
	return &GetUnreadCountQuery{notifications: notifications}
}

// Execute returns the number of unread notifications for the user, used to
// render the notification-center badge.
func (query *GetUnreadCountQuery) Execute(
	ctx context.Context,
	input GetUnreadCountQueryInput,
) (GetUnreadCountQueryOutput, error) {
	count, err := query.notifications.CountUnread(ctx, input.UserID)
	if err != nil {
		return GetUnreadCountQueryOutput{}, err
	}

	return GetUnreadCountQueryOutput{UnreadCount: count}, nil
}
