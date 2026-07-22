package router

import (
	"auction/internal/modules/ledger/infra/http/handler"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"

	"github.com/go-chi/chi/v5"
)

func RegisterLedgerRoutes(
	server *httpserver.Server,
	ledgerHandler *handler.LedgerHandler,
	middleware *authn.Middleware,
) {
	router := server.Router()

	router.Route("/api/v1/ledger", func(r chi.Router) {
		r.Use(middleware.RequireAuth)

		r.Post("/accounts", ledgerHandler.CreateAccount)
		r.Get("/accounts", ledgerHandler.GetAccountByOwner)
		r.Get("/accounts/{id}", ledgerHandler.GetAccountByID)
		r.Post("/transfer", ledgerHandler.Transfer)
		r.Post("/freeze", ledgerHandler.Freeze)
		r.Post("/unfreeze", ledgerHandler.Unfreeze)
		r.Post("/withdraw-from-frozen", ledgerHandler.WithdrawFromFrozen)
	})
}
