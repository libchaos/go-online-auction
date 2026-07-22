package router

import (
	"auction/internal/modules/notification/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"

	"github.com/go-chi/chi/v5"
)

// RegisterNotificationRoutes wires the authenticated notification-center and
// preference endpoints. Every route is scoped to the authenticated user.
func RegisterNotificationRoutes(
	server *httpserver.Server,
	notificationHandler *handler.NotificationHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/notifications", func(r chi.Router) {
		r.With(middleware.RequireAuth).Get("/", notificationHandler.ListNotifications)
		r.With(middleware.RequireAuth).Get("/unread-count", notificationHandler.GetUnreadCount)
		r.With(middleware.RequireAuth).Post("/read-all", notificationHandler.MarkAllNotificationsRead)
		r.With(middleware.RequireAuth).Post("/{id}/read", notificationHandler.MarkNotificationRead)
		r.With(middleware.RequireAuth).Delete("/{id}", notificationHandler.DeleteNotification)
		r.With(middleware.RequireAuth).Get("/preferences", notificationHandler.GetPreferences)
		r.With(middleware.RequireAuth).Put("/preferences", notificationHandler.PutPreferences)
	})
}

// RegisterNotificationStreamRoute wires the Server-Sent Events endpoint. It is
// authenticated through a token query parameter rather than the standard
// Authorization header because the browser EventSource API cannot set headers,
// so it must not sit behind the header-only RequireAuth middleware.
func RegisterNotificationStreamRoute(
	server *httpserver.Server,
	streamHandler *handler.NotificationStreamHandler,
) {
	router := server.Router()

	router.Get("/api/v1/notifications/stream", streamHandler.Stream)
}

// RegisterWatchlistRoutes wires the authenticated watchlist (favourite a
// product) endpoints. Every route is scoped to the authenticated user.
func RegisterWatchlistRoutes(
	server *httpserver.Server,
	watchlistHandler *handler.WatchlistHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/watchlists", func(r chi.Router) {
		r.With(middleware.RequireAuth).Post("/", watchlistHandler.CreateWatch)
		r.With(middleware.RequireAuth).Get("/", watchlistHandler.ListMyWatches)
		r.With(middleware.RequireAuth).Delete("/{spuId}", watchlistHandler.DeleteWatch)
	})
}
