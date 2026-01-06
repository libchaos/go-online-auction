package router

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction/infra/http/chi/handler"
	"github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(server *httpserver.Server, auctionHandler *handler.AuctionHandler) {
	router := server.Router()

	router.Route("/api/v1/auctions", func(r chi.Router) {
		r.Post("/", auctionHandler.Create)
		r.Get("/", auctionHandler.List)
		r.Get("/{id}", auctionHandler.GetByID)
		r.Put("/{id}/start", auctionHandler.Start)
		r.Put("/{id}/cancel", auctionHandler.Cancel)
		r.Post("/{id}/bids", auctionHandler.PlaceBid)
	})

	router.Get("/ws/v1/auctions/{id}", auctionHandler.WebSocket)
}
