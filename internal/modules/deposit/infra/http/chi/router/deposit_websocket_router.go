package router

import (
	"auction/internal/modules/deposit/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"
)

func RegisterDepositWebsocketRoutes(
	server *httpserver.Server,
	depositWebSocketHandler *handler.DepositWebSocketHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.With(middleware.RequireAuth).Get("/ws/v1/deposits", depositWebSocketHandler.WebSocket)
}
