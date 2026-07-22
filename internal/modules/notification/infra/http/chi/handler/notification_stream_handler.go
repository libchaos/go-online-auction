package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"auction/internal/modules/notification/infra/sse"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/response"
)

const (
	bearerPrefix      = "Bearer "
	heartbeatInterval = 25 * time.Second
	retryDirective    = "retry: 3000\n\n"
)

type NotificationStreamHandler struct {
	hub      *sse.RealtimeHub
	verifier authn.TokenVerifier
	logger   logger.Logger
}

func NewNotificationStreamHandler(
	hub *sse.RealtimeHub,
	verifier authn.TokenVerifier,
	logger logger.Logger,
) *NotificationStreamHandler {
	return &NotificationStreamHandler{
		hub:      hub,
		verifier: verifier,
		logger:   logger,
	}
}

// Stream opens a Server-Sent-Events channel that pushes the authenticated user's
// notifications in real time. Authentication is via the token query parameter
// because the browser EventSource API cannot set an Authorization header; a
// bearer header is still accepted for non-browser clients.
func (streamHandler *NotificationStreamHandler) Stream(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	claims, verifyErr := streamHandler.verifier.Verify(token)
	if verifyErr != nil {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		streamHandler.logger.Error().Msg("notification stream: response writer does not support flushing")
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	client := sse.NewClient(claims.UserID)
	streamHandler.hub.AddClient(client)
	defer streamHandler.hub.RemoveClient(client)

	if _, writeErr := fmt.Fprint(w, retryDirective); writeErr != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	requestContext := r.Context()

	for {
		select {
		case <-requestContext.Done():
			return
		case <-client.Done():
			return
		case message := <-client.Send():
			if _, writeErr := fmt.Fprintf(w, "data: %s\n\n", message); writeErr != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, writeErr := fmt.Fprint(w, ": ping\n\n"); writeErr != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func extractToken(r *http.Request) string {
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return strings.TrimPrefix(authHeader, bearerPrefix)
	}

	return ""
}
