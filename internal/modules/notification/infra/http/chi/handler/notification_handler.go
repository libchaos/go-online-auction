package handler

import (
	"net/http"
	"strconv"
	"time"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/application/query"
	"auction/internal/modules/notification/domain/model"
	"auction/internal/modules/notification/infra/http/dto"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/httperrs"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

type NotificationHandler struct {
	listNotificationsQuery      *query.ListNotificationsQuery
	getUnreadCountQuery         *query.GetUnreadCountQuery
	getPreferencesQuery         *query.GetPreferencesQuery
	markNotificationReadCommand *command.MarkNotificationReadCommand
	markAllNotificationsReadCmd *command.MarkAllNotificationsReadCommand
	deleteNotificationCommand   *command.DeleteNotificationCommand
	updatePreferencesCommand    *command.UpdatePreferencesCommand
	httpServer                  *httpserver.Server
	logger                      logger.Logger
}

func NewNotificationHandler(
	listNotificationsQuery *query.ListNotificationsQuery,
	getUnreadCountQuery *query.GetUnreadCountQuery,
	getPreferencesQuery *query.GetPreferencesQuery,
	markNotificationReadCommand *command.MarkNotificationReadCommand,
	markAllNotificationsReadCmd *command.MarkAllNotificationsReadCommand,
	deleteNotificationCommand *command.DeleteNotificationCommand,
	updatePreferencesCommand *command.UpdatePreferencesCommand,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		listNotificationsQuery:      listNotificationsQuery,
		getUnreadCountQuery:         getUnreadCountQuery,
		getPreferencesQuery:         getPreferencesQuery,
		markNotificationReadCommand: markNotificationReadCommand,
		markAllNotificationsReadCmd: markAllNotificationsReadCmd,
		deleteNotificationCommand:   deleteNotificationCommand,
		updatePreferencesCommand:    updatePreferencesCommand,
		httpServer:                  httpServer,
		logger:                      logger,
	}
}

func (notificationHandler *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	unreadOnly := r.URL.Query().Get("unread") == "true"
	limit := parseIntQuery(r, "limit")
	offset := parseIntQuery(r, "offset")

	output, err := notificationHandler.listNotificationsQuery.Execute(r.Context(), query.ListNotificationsQueryInput{
		UserID:     claims.UserID,
		UnreadOnly: unreadOnly,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	items := make([]dto.NotificationResponse, 0, len(output.Notifications))
	for index := range output.Notifications {
		items = append(items, toNotificationResponse(output.Notifications[index]))
	}

	_ = response.JSON(w, http.StatusOK, dto.NotificationListResponse{
		Items:  items,
		Limit:  limit,
		Offset: offset,
	}, nil)
}

func (notificationHandler *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	output, err := notificationHandler.getUnreadCountQuery.Execute(r.Context(), query.GetUnreadCountQueryInput{
		UserID: claims.UserID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.UnreadCountResponse{UnreadCount: output.UnreadCount}, nil)
}

func (notificationHandler *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	notificationID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrNotificationInvalidRequest)
		return
	}

	err := notificationHandler.markNotificationReadCommand.Execute(
		r.Context(),
		command.MarkNotificationReadCommandInput{
			NotificationID: notificationID,
			UserID:         claims.UserID,
		},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (notificationHandler *NotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	output, err := notificationHandler.markAllNotificationsReadCmd.Execute(
		r.Context(),
		command.MarkAllNotificationsReadCommandInput{
			UserID: claims.UserID,
		},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, dto.MarkAllReadResponse{UpdatedCount: output.UpdatedCount}, nil)
}

func (notificationHandler *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	notificationID, parseErr := parseIDParam(r)
	if parseErr != nil {
		response.Error(w, httperrs.ErrNotificationInvalidRequest)
		return
	}

	err := notificationHandler.deleteNotificationCommand.Execute(r.Context(), command.DeleteNotificationCommandInput{
		NotificationID: notificationID,
		UserID:         claims.UserID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	response.NoContent(w)
}

func (notificationHandler *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	output, err := notificationHandler.getPreferencesQuery.Execute(r.Context(), query.GetPreferencesQueryInput{
		UserID: claims.UserID,
	})
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, toPreferencesResponse(output.Preference), nil)
}

func (notificationHandler *NotificationHandler) PutPreferences(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)
		return
	}

	var req dto.UpdatePreferencesRequest
	if readErr := request.ReadJSON(w, r, &req); readErr != nil {
		response.Error(w, httperrs.ErrNotificationInvalidRequest)
		return
	}

	output, err := notificationHandler.updatePreferencesCommand.Execute(
		r.Context(),
		command.UpdatePreferencesCommandInput{
			UserID:   claims.UserID,
			Settings: req.Preferences,
		},
	)
	if err != nil {
		response.Error(w, httperrs.MapDomainError(err))
		return
	}

	_ = response.JSON(w, http.StatusOK, toPreferencesResponse(output.Preference), nil)
}

func parseIDParam(r *http.Request) (uint64, error) {
	idString := request.Param(r, "id")

	return strconv.ParseUint(idString, 10, 64)
}

func parseIntQuery(r *http.Request, name string) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return 0
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}

	return value
}

func toNotificationResponse(notification model.NotificationModel) dto.NotificationResponse {
	category := notification.Category()
	notificationType := notification.Type()

	channels := make([]string, 0, len(notification.Channels()))
	for _, channel := range notification.Channels() {
		channelValue := channel
		channels = append(channels, channelValue.String())
	}

	readAt := ""
	if notification.ReadAt() != nil {
		readAt = notification.ReadAt().Format(time.RFC3339)
	}

	return dto.NotificationResponse{
		NotificationID: notification.ID(),
		UserID:         notification.UserID(),
		Category:       category.String(),
		Type:           notificationType.String(),
		Title:          notification.Title(),
		Body:           notification.Body(),
		Payload:        notification.Payload(),
		Channels:       channels,
		ReadAt:         readAt,
		CreatedAt:      notification.CreatedAt().Format(time.RFC3339),
	}
}

func toPreferencesResponse(preference model.NotificationPreferenceModel) dto.PreferencesResponse {
	return dto.PreferencesResponse{
		UserID:      preference.UserID(),
		Preferences: preference.Settings(),
		UpdatedAt:   preference.UpdatedAt().Format(time.RFC3339),
	}
}
