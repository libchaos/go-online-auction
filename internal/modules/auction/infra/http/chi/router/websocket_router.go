package router

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/http/chi/handler"
	"github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
)

func RegisterWebsocketRoutes(server *httpserver.Server, websocketHandler *handler.WebsocketHandler) {
	router := server.Router()

	router.Get("/ws/v1/auctions/{id}", websocketHandler.WebSocket)
}
