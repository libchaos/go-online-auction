package router

import (
	"auction/internal/modules/auction/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"
	"github.com/go-chi/chi/v5"
)

func RegisterAuctionRoutes(
	server *httpserver.Server,
	auctionHandler *handler.AuctionHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/auctions", func(r chi.Router) {
		sellerOrAdmin := authn.RequireRole(authn.RoleSeller, authn.RoleAdmin)

		r.With(middleware.RequireAuth, sellerOrAdmin).Post("/", auctionHandler.Create)
		r.Get("/", auctionHandler.List)
		r.Get("/{id}", auctionHandler.GetByID)
		r.Get("/{id}/events", auctionHandler.Events)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}/start", auctionHandler.Start)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}/cancel", auctionHandler.Cancel)
		r.With(middleware.RequireAuth).Post("/{id}/bids", auctionHandler.PlaceBid)
	})
}
