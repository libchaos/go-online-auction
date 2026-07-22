package deposit

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/modules/deposit/application/command"
	"auction/internal/modules/deposit/application/guard"
	"auction/internal/modules/deposit/application/query"
	"auction/internal/modules/deposit/infra/http/chi/handler"
	"auction/internal/modules/deposit/infra/http/chi/router"
	"auction/internal/modules/deposit/infra/mapper"
	"auction/internal/modules/deposit/infra/messaging"
	depositoutbox "auction/internal/modules/deposit/infra/outbox"
	"auction/internal/modules/deposit/infra/payment"
	"auction/internal/modules/deposit/infra/repository"
	"auction/internal/modules/deposit/infra/sqlcgen"
	"auction/internal/modules/deposit/infra/uow"
	"auction/internal/modules/deposit/infra/websocket"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	"auction/pkg/httpserver"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"deposit",

	fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX { return pool }),

	fx.Provide(mapper.NewDepositMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresDepositRepository,
			fx.As(new(ports.DepositRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresAuctionConfigRepository,
			fx.As(new(ports.AuctionConfigPort)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresAuctionConfigRepository,
			fx.As(new(ports.AuctionWinnerPort)),
		),
	),

	fx.Provide(
		fx.Annotate(
			uow.NewDepositUnitOfWorkFactory,
			fx.As(new(ports.DepositUnitOfWorkFactory)),
		),
	),

	fx.Provide(
		fx.Annotate(
			payment.NewPaymentPort,
			fx.As(new(ports.PaymentPort)),
		),
	),

	fx.Provide(
		fx.Annotate(
			guard.NewDepositGuard,
			fx.As(new(ports.DepositGuard)),
		),
	),

	fx.Provide(command.NewCreateDepositCommand),
	fx.Provide(command.NewReleaseDepositCommand),
	fx.Provide(command.NewApplyDepositCommand),
	fx.Provide(command.NewForfeitDepositCommand),
	fx.Provide(command.NewCancelDepositCommand),

	fx.Provide(query.NewGetDepositQuery),
	fx.Provide(query.NewListDepositsByUserQuery),
	fx.Provide(query.NewListHeldDepositsByAuctionQuery),
	fx.Provide(query.NewGetEligibilityQuery),

	fx.Provide(handler.NewDepositHandler),
	fx.Provide(handler.NewDepositWebSocketHandler),

	fx.Provide(websocket.NewUserSubscriberRegistry),
	fx.Provide(websocket.NewHub),

	fx.Provide(
		fx.Annotate(
			messaging.NewJetStreamDepositEventConsumer,
			fx.As(new(websocket.EventConsumer)),
		),
	),
	fx.Provide(messaging.NewSettlementConsumer),

	// Transactional outbox: the deposit module owns its own outbox table
	// (deposit_outbox) and its own relay, so it can be deployed independently
	// of the auction module.
	fx.Provide(
		fx.Annotate(
			depositoutbox.NewPostgresOutboxRepository,
			fx.As(new(ports.DepositOutboxRepository)),
		),
	),
	fx.Provide(func(cfg config.Config) depositoutbox.Config {
		return depositoutbox.Config{
			Interval:  cfg.Outbox.Interval,
			BatchSize: cfg.Outbox.BatchSize,
		}
	}),
	fx.Provide(depositoutbox.NewRelay),
)

func RegisterDepositRoutes(
	server *httpserver.Server,
	depositHandler *handler.DepositHandler,
	middleware *authn.Middleware,
	authzMiddleware *authz.Middleware,
) {
	router.RegisterDepositRoutes(server, depositHandler, middleware, authzMiddleware)
}

func RegisterDepositWebsocketRoutes(
	server *httpserver.Server,
	depositWebSocketHandler *handler.DepositWebSocketHandler,
	middleware *authn.Middleware,
) {
	router.RegisterDepositWebsocketRoutes(server, depositWebSocketHandler, middleware)
}

func RegisterDepositHub(
	lc fx.Lifecycle,
	hub *websocket.Hub,
	logger logger.Logger,
) {
	hubContext, hubCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting deposit websocket hub")
			go hub.Run(hubContext)

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping deposit websocket hub")
			hubCancel()
			if shutdownErr := hub.Shutdown(); shutdownErr != nil {
				logger.Error().Err(shutdownErr).Msg("failed to shutdown deposit websocket hub")
			}

			return nil
		},
	})
}

func RegisterDepositEventConsumer(
	lc fx.Lifecycle,
	settlementConsumer *messaging.SettlementConsumer,
	logger logger.Logger,
) {
	consumerContext, consumerCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting deposit settlement consumer")
			startErr := settlementConsumer.Start(consumerContext)
			if startErr != nil {
				logger.Error().Err(startErr).Msg("failed to start deposit settlement consumer")
			}

			return startErr
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping deposit settlement consumer")
			consumerCancel()
			settlementConsumer.Stop()

			return nil
		},
	})
}

// RegisterOutboxRelay wires the deposit module's transactional outbox relay
// into the fx lifecycle. It runs in every process that writes deposit events.
func RegisterOutboxRelay(
	lc fx.Lifecycle,
	relay *depositoutbox.Relay,
	logger logger.Logger,
) {
	relayCtx, relayCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting deposit outbox relay")
			relay.Start(relayCtx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping deposit outbox relay")
			relayCancel()
			relay.Stop()
			return nil
		},
	})
}
