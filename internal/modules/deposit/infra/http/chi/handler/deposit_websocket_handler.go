package handler

import (
	"net/http"

	"auction/internal/modules/deposit/infra/websocket"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

type DepositWebSocketHandler struct {
	hub        *websocket.Hub
	httpServer *httpserver.Server
	logger     logger.Logger
}

func NewDepositWebSocketHandler(
	hub *websocket.Hub,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *DepositWebSocketHandler {
	return &DepositWebSocketHandler{
		hub:        hub,
		httpServer: httpServer,
		logger:     logger,
	}
}

func (depositWebSocketHandler *DepositWebSocketHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	claims, ok := authn.ClaimsFromContext(r.Context())
	if !ok {
		response.Error(w, authn.ErrUnauthorized)

		return
	}

	conn, err := depositWebSocketHandler.httpServer.Upgrader().Upgrade(w, r, nil)
	if err != nil {
		depositWebSocketHandler.logger.Error().Err(err).Msg("deposit websocket upgrade failed")

		return
	}

	depositWebSocketHandler.hub.RegisterClient(conn, claims.UserID)
}
