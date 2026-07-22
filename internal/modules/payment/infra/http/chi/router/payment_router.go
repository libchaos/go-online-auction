package router

import (
	"auction/internal/modules/payment/infra/http/chi/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"

	"github.com/go-chi/chi/v5"
)

// RegisterPaymentRoutes wires the authenticated payment endpoints. The Alipay
// notify callback is registered separately (public, server-to-server).
func RegisterPaymentRoutes(
	server *httpserver.Server,
	paymentHandler *handler.PaymentHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/payment", func(r chi.Router) {
		r.With(middleware.RequireAuth).Post("/deposit", paymentHandler.CreateDeposit)
		r.With(middleware.RequireAuth).Get("/deposit/{id}", paymentHandler.GetDeposit)
		r.With(middleware.RequireAuth).Post("/withdraw", paymentHandler.CreateWithdrawal)
		r.With(middleware.RequireAuth).Get("/withdraw/{id}", paymentHandler.GetWithdrawal)
	})
}

// RegisterAlipayNotifyRoute wires the public Alipay asynchronous-notification
// callback. It must not require authentication.
func RegisterAlipayNotifyRoute(
	server *httpserver.Server,
	alipayNotifyHandler *handler.AlipayNotifyHandler,
) {
	router := server.Router()

	router.Post("/api/v1/payment/alipay/notify", alipayNotifyHandler.Notify)
}
