package router

import (
	"auction/internal/modules/auction/infra/http/chi/handler"
	"auction/pkg/httpserver"
)

func RegisterWebsocketRoutes(server *httpserver.Server, websocketHandler *handler.WebsocketHandler) {
	router := server.Router()

	router.Get("/ws/v1/auctions/{id}", websocketHandler.WebSocket)
}
