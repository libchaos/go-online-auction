package router

import (
	"github.com/go-chi/chi/v5"

	"auction/internal/modules/users/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
	"auction/pkg/httpserver"
)

func RegisterUserRoutes(
	server *httpserver.Server,
	userHandler *handler.UserHandler,
	middleware *authn.Middleware,
	authzMiddleware *authz.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", userHandler.Register)
		r.Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.Refresh)
		r.With(middleware.RequireAuth).Post("/logout", userHandler.Logout)
	})

	router.Route("/api/v1/users", func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Get("/me", userHandler.GetMe)
		r.Put("/me", userHandler.UpdateMe)
		r.Put("/me/password", userHandler.ChangePassword)

		r.With(authzMiddleware.RequirePermission()).Get("/", userHandler.List)
		r.With(authzMiddleware.RequirePermission()).Get("/{id}", userHandler.GetByID)
		r.With(authzMiddleware.RequirePermission()).Patch("/{id}/role", userHandler.UpdateRole)
	})
}
