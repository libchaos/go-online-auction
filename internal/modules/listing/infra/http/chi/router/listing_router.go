package router

import (
	"github.com/go-chi/chi/v5"

	"auction/internal/modules/listing/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"
)

func RegisterListingRoutes(
	server *httpserver.Server,
	categoryHandler *handler.CategoryHandler,
	spuHandler *handler.SpuHandler,
	skuHandler *handler.SkuHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/categories", func(r chi.Router) {
		admin := authn.RequireRole(authn.RoleAdmin)

		r.With(middleware.RequireAuth, admin).Post("/", categoryHandler.Create)
		r.Get("/", categoryHandler.List)
		r.Get("/{id}", categoryHandler.GetByID)
		r.With(middleware.RequireAuth, admin).Put("/{id}", categoryHandler.Update)
		r.With(middleware.RequireAuth, admin).Delete("/{id}", categoryHandler.Delete)
	})

	router.Route("/api/v1/spus", func(r chi.Router) {
		sellerOrAdmin := authn.RequireRole(authn.RoleSeller, authn.RoleAdmin)

		r.With(middleware.RequireAuth, sellerOrAdmin).Post("/", spuHandler.Create)
		r.Get("/", spuHandler.List)
		r.Get("/{id}", spuHandler.GetByID)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}", spuHandler.Update)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}/publish", spuHandler.Publish)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}/off-shelf", spuHandler.OffShelf)
		r.With(middleware.RequireAuth, sellerOrAdmin).Post("/{id}/skus", skuHandler.Create)
	})

	router.Route("/api/v1/skus", func(r chi.Router) {
		sellerOrAdmin := authn.RequireRole(authn.RoleSeller, authn.RoleAdmin)

		r.Get("/{id}", skuHandler.GetByID)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}", skuHandler.Update)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}/publish", skuHandler.Publish)
		r.With(middleware.RequireAuth, sellerOrAdmin).Put("/{id}/off-shelf", skuHandler.OffShelf)
	})
}
