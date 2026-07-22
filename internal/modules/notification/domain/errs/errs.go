package errs

import "errors"

var (
	ErrNotificationNotFound      = errors.New("notification not found")
	ErrNotificationUserRequired  = errors.New("notification user id is required")
	ErrNotificationTitleRequired = errors.New("notification title is required")
	ErrNotificationBodyRequired  = errors.New("notification body is required")
	ErrNotificationChannelsEmpty = errors.New("notification must target at least one channel")
	ErrNotificationAlreadyRead   = errors.New("notification is already read")

	ErrPreferencesUserRequired = errors.New("notification preferences user id is required")
	ErrPreferencesInvalid      = errors.New("notification preferences payload is invalid")

	ErrWatchUserRequired = errors.New("watch user id is required")
	ErrWatchSpuRequired  = errors.New("watch spu id is required")
	ErrWatchNotFound     = errors.New("watch not found")
)
