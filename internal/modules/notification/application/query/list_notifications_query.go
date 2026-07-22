package query

import (
	"context"

	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/ports"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type ListNotificationsQueryInput struct {
	UserID     uint64
	UnreadOnly bool
	Limit      int
	Offset     int
}

type ListNotificationsQueryOutput struct {
	Notifications []model.NotificationModel
}

type ListNotificationsQuery struct {
	notifications ports.NotificationRepository
}

func NewListNotificationsQuery(notifications ports.NotificationRepository) *ListNotificationsQuery {
	return &ListNotificationsQuery{notifications: notifications}
}

// Execute returns a page of the user's notifications, newest first. When
// UnreadOnly is set only unread notifications are returned. The page size is
// clamped to a sane range so a client can never request an unbounded scan.
func (query *ListNotificationsQuery) Execute(
	ctx context.Context,
	input ListNotificationsQueryInput,
) (ListNotificationsQueryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	if input.UnreadOnly {
		unread, err := query.notifications.ListUnreadByUser(ctx, input.UserID, limit, offset)
		if err != nil {
			return ListNotificationsQueryOutput{}, err
		}

		return ListNotificationsQueryOutput{Notifications: unread}, nil
	}

	all, err := query.notifications.ListByUser(ctx, input.UserID, limit, offset)
	if err != nil {
		return ListNotificationsQueryOutput{}, err
	}

	return ListNotificationsQueryOutput{Notifications: all}, nil
}
