package router

import (
	"github.com/go-chi/chi/v5"

	"auction/internal/modules/listing/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
	"auction/pkg/httpserver"
)

func RegisterListingRoutes(
	server *httpserver.Server,
	categoryHandler *handler.CategoryHandler,
	spuHandler *handler.SpuHandler,
	skuHandler *handler.SkuHandler,
	middleware *authn.Middleware,
	authzMiddleware *authz.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/categories", func(r chi.Router) {
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Post("/", categoryHandler.Create)
		r.Get("/", categoryHandler.List)
		r.Get("/tree", categoryHandler.Tree)
		r.Get("/{id}/tree", categoryHandler.Subtree)
		r.Get("/{id}", categoryHandler.GetByID)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}", categoryHandler.Update)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Delete("/{id}", categoryHandler.Delete)
	})

	router.Route("/api/v1/spus", func(r chi.Router) {
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Post("/", spuHandler.Create)
		r.Get("/", spuHandler.List)
		r.Get("/{id}", spuHandler.GetByID)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}", spuHandler.Update)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}/publish", spuHandler.Publish)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}/off-shelf", spuHandler.OffShelf)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Post("/{id}/skus", skuHandler.Create)
	})

	router.Route("/api/v1/skus", func(r chi.Router) {
		r.Get("/{id}", skuHandler.GetByID)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}", skuHandler.Update)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}/publish", skuHandler.Publish)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Put("/{id}/off-shelf", skuHandler.OffShelf)
	})
}
