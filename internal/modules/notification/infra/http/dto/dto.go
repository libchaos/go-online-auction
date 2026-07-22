package dto

import (
	"encoding/json"

	"auction/internal/modules/notification/domain/model"
)

type NotificationResponse struct {
	NotificationID uint64          `json:"notification_id"`
	UserID         uint64          `json:"user_id"`
	Category       string          `json:"category"`
	Type           string          `json:"type"`
	Title          string          `json:"title"`
	Body           string          `json:"body"`
	Payload        json.RawMessage `json:"payload"`
	Channels       []string        `json:"channels"`
	ReadAt         string          `json:"read_at,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

type NotificationListResponse struct {
	Items  []NotificationResponse `json:"items"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

type UnreadCountResponse struct {
	UnreadCount uint64 `json:"unread_count"`
}

type MarkAllReadResponse struct {
	UpdatedCount uint64 `json:"updated_count"`
}

type PreferencesResponse struct {
	UserID      uint64                   `json:"user_id"`
	Preferences model.PreferenceSettings `json:"preferences"`
	UpdatedAt   string                   `json:"updated_at"`
}

type UpdatePreferencesRequest struct {
	Preferences model.PreferenceSettings `json:"preferences"`
}

type WatchResponse struct {
	WatchID   uint64 `json:"watch_id"`
	UserID    uint64 `json:"user_id"`
	SpuID     uint64 `json:"spu_id"`
	CreatedAt string `json:"created_at"`
}

type WatchListResponse struct {
	Items  []WatchResponse `json:"items"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type CreateWatchRequest struct {
	SpuID uint64 `json:"spu_id"`
}
