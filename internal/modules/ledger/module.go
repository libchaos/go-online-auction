package ledger

import (
	"auction/internal/modules/ledger/application/service"
	ledgerrouter "auction/internal/modules/ledger/infra/http/chi/router"
	ledgerhandler "auction/internal/modules/ledger/infra/http/handler"
	ledgermapper "auction/internal/modules/ledger/infra/mapper"
	ledgeruow "auction/internal/modules/ledger/infra/uow"
	ledgerports "auction/internal/modules/ledger/ports"
	"auction/internal/shared/modules/authn"
	"auction/pkg/httpserver"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"ledger",

	fx.Provide(ledgermapper.NewLedgerMapper),

	fx.Provide(
		fx.Annotate(
			ledgeruow.NewLedgerUnitOfWorkFactory,
			fx.As(new(ledgerports.LedgerUnitOfWorkFactory)),
		),
	),

	fx.Provide(
		fx.Annotate(
			service.NewLedgerAccountService,
			fx.As(new(ledgerports.LedgerAccountService)),
		),
	),

	fx.Provide(ledgerhandler.NewLedgerHandler),
)

func RegisterLedgerRoutes(
	server *httpserver.Server,
	ledgerHandlerInstance *ledgerhandler.LedgerHandler,
	middleware *authn.Middleware,
) {
	ledgerrouter.RegisterLedgerRoutes(server, ledgerHandlerInstance, middleware)
}
