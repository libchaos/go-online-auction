package router

import (
	"auction/internal/modules/deposit/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
	"auction/pkg/httpserver"

	"github.com/go-chi/chi/v5"
)

func RegisterDepositRoutes(
	server *httpserver.Server,
	depositHandler *handler.DepositHandler,
	middleware *authn.Middleware,
	authzMiddleware *authz.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/deposits", func(r chi.Router) {
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).Post("/", depositHandler.Create)
		r.With(middleware.RequireAuth).Get("/", depositHandler.ListByUser)
		r.With(middleware.RequireAuth).Get("/{id}", depositHandler.GetByID)
		r.With(middleware.RequireAuth, authzMiddleware.RequirePermission()).
			Get("/eligibility", depositHandler.GetEligibility)
		r.With(middleware.RequireAuth).Post("/{id}/release", depositHandler.Release)
		r.With(middleware.RequireAuth).Post("/{id}/apply", depositHandler.Apply)
		r.With(middleware.RequireAuth).Post("/{id}/forfeit", depositHandler.Forfeit)
		r.With(middleware.RequireAuth).Post("/{id}/cancel", depositHandler.Cancel)
	})
}
