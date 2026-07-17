package handler

import (
	"net/http"
	"strconv"

	httperrs "auction/internal/modules/auction/infra/http/errs"
	"auction/internal/modules/auction/infra/websocket"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/sdk/http/request"
	"auction/internal/shared/sdk/http/response"
	"auction/pkg/httpserver"
)

type WebsocketHandler struct {
	websocketHub *websocket.Hub
	httpServer   *httpserver.Server
	logger       logger.Logger
}

func NewWebsocketHandler(
	websocketHub *websocket.Hub,
	httpServer *httpserver.Server,
	logger logger.Logger,
) *WebsocketHandler {
	return &WebsocketHandler{
		websocketHub: websocketHub,
		httpServer:   httpServer,
		logger:       logger,
	}
}

func (h *WebsocketHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	idString := request.Param(r, "id")
	auctionID, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.Error(w, httperrs.ErrInvalidAuctionID)
		return
	}

	conn, err := h.httpServer.Upgrader().Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	h.websocketHub.RegisterClient(conn, auctionID)
}
