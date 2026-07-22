package auction

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"auction/internal/modules/auction/application/command"
	"auction/internal/modules/auction/application/query"
	"auction/internal/modules/auction/domain/strategy"
	"auction/internal/modules/auction/infra/http/chi/handler"
	"auction/internal/modules/auction/infra/http/chi/router"
	"auction/internal/modules/auction/infra/mapper"
	"auction/internal/modules/auction/infra/messaging"
	"auction/internal/modules/auction/infra/outbox"
	"auction/internal/modules/auction/infra/repository"
	"auction/internal/modules/auction/infra/scheduler"
	"auction/internal/modules/auction/infra/sqlcgen"
	"auction/internal/modules/auction/infra/uow"
	"auction/internal/modules/auction/infra/websocket"
	"auction/internal/modules/auction/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	"auction/pkg/httpserver"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type tradingStrategyGroup struct {
	fx.In
	Strategies strategy.Strategies `group:"trading_strategy"`
}

var Module = fx.Module(
	"auction",

	fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX { return pool }),

	fx.Provide(mapper.NewAuctionMapper),
	fx.Provide(mapper.NewBidMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresAuctionRepository,
			fx.As(new(ports.AuctionRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresBidRepository,
			fx.As(new(ports.BidRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresOutboxRepository,
			fx.As(new(ports.OutboxRepository)),
		),
	),

	fx.Provide(
		fx.Annotate(
			uow.NewAuctionUnitOfWorkFactory,
			fx.As(new(ports.AuctionUnitOfWorkFactory)),
		),
	),

	fx.Provide(func(cfg config.Config) scheduler.Config {
		return scheduler.Config{
			Interval:  cfg.Scheduler.Interval,
			BatchSize: cfg.Scheduler.BatchSize,
		}
	}),
	fx.Provide(scheduler.NewAuctionScheduler),

	fx.Provide(func(cfg config.Config) outbox.Config {
		return outbox.Config{
			Interval:  cfg.Outbox.Interval,
			BatchSize: cfg.Outbox.BatchSize,
		}
	}),
	fx.Provide(outbox.NewRelay),

	fx.Provide(
		fx.Annotate(
			messaging.NewJetStreamEventPublisher,
			fx.As(new(messaging.EventPublisher)),
		),
	),
	fx.Provide(
		fx.Annotate(
			messaging.NewJetStreamEventConsumer,
			fx.As(new(messaging.EventConsumer)),
		),
	),
	fx.Provide(
		fx.Annotate(
			messaging.NewJetStreamBidCommandPublisher,
			fx.As(new(ports.BidCommandPublisher)),
		),
	),
	fx.Provide(
		fx.Annotate(
			messaging.NewJetStreamEventReplayer,
			fx.As(new(messaging.EventReplayer)),
		),
	),

	fx.Provide(websocket.NewAuctionSubscriberRegistry),
	fx.Provide(websocket.NewHub),

	fx.Provide(
		fx.Annotate(
			strategy.NewEnglishAuctionStrategy,
			fx.As(new(strategy.TradingStrategy)),
			fx.ResultTags(`group:"trading_strategy"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			strategy.NewDutchAuctionStrategy,
			fx.As(new(strategy.TradingStrategy)),
			fx.ResultTags(`group:"trading_strategy"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			strategy.NewSealedBidAuctionStrategy,
			fx.As(new(strategy.TradingStrategy)),
			fx.ResultTags(`group:"trading_strategy"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			strategy.NewVickreyAuctionStrategy,
			fx.As(new(strategy.TradingStrategy)),
			fx.ResultTags(`group:"trading_strategy"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			strategy.NewFixedPriceAuctionStrategy,
			fx.As(new(strategy.TradingStrategy)),
			fx.ResultTags(`group:"trading_strategy"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			strategy.NewEbayProxyAuctionStrategy,
			fx.As(new(strategy.TradingStrategy)),
			fx.ResultTags(`group:"trading_strategy"`),
		),
	),

	fx.Provide(
		fx.Annotate(
			func(group tradingStrategyGroup) strategy.Resolver {
				return strategy.NewResolver(group.Strategies)
			},
			fx.As(new(strategy.Resolver)),
		),
	),

	fx.Provide(messaging.NewBidProcessor),

	fx.Provide(command.NewCreateAuctionCommand),
	fx.Provide(command.NewStartAuctionCommand),
	fx.Provide(command.NewPlaceBidCommand),
	fx.Provide(command.NewCloseAuctionCommand),
	fx.Provide(command.NewCancelAuctionCommand),

	fx.Provide(query.NewGetAuctionByIDQuery),
	fx.Provide(query.NewListAuctionsQuery),

	fx.Provide(handler.NewAuctionHandler),
	fx.Provide(handler.NewWebsocketHandler),
)

func RegisterAuctionRoutes(
	server *httpserver.Server,
	auctionHandler *handler.AuctionHandler,
	middleware *authn.Middleware,
	authzMiddleware *authz.Middleware,
) {
	router.RegisterAuctionRoutes(server, auctionHandler, middleware, authzMiddleware)
}

func RegisterMetricsRoute(server *httpserver.Server) {
	server.Router().Handle("/metrics", promhttp.Handler())
}

func RegisterWebsocketRoutes(
	lc fx.Lifecycle,
	hub *websocket.Hub,
	server *httpserver.Server,
	websocketHandler *handler.WebsocketHandler,
	logger logger.Logger,
) {
	router.RegisterWebsocketRoutes(server, websocketHandler)

	hubCtx, hubCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting websocket hub")
			go hub.Run(hubCtx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping websocket hub")
			hubCancel()
			if shutdownErr := hub.Shutdown(); shutdownErr != nil {
				logger.Error().Err(shutdownErr).Msg("failed to shutdown websocket hub")
			}

			return nil
		},
	})
}

func RegisterBidProcessor(
	lc fx.Lifecycle,
	bidProcessor *messaging.BidProcessor,
	logger logger.Logger,
) {
	processorCtx, processorCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting bid processor")
			return bidProcessor.Start(processorCtx)
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping bid processor")
			processorCancel()
			bidProcessor.Stop()
			return nil
		},
	})
}

// RegisterOutboxRelay wires the transactional outbox relay into the fx
// lifecycle. It runs in every process that writes domain events.
func RegisterOutboxRelay(
	lc fx.Lifecycle,
	relay *outbox.Relay,
	logger logger.Logger,
) {
	relayCtx, relayCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting outbox relay")
			relay.Start(relayCtx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping outbox relay")
			relayCancel()
			relay.Stop()
			return nil
		},
	})
}

// RegisterAuctionScheduler wires the automatic auction scheduler into the fx
// lifecycle. When SCHEDULER_ENABLED is false, nothing is started.
func RegisterAuctionScheduler(
	lc fx.Lifecycle,
	cfg config.Config,
	auctionScheduler *scheduler.AuctionScheduler,
	logger logger.Logger,
) {
	if !cfg.Scheduler.Enabled {
		logger.Info().Msg("auction scheduler is disabled")
		return
	}

	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting auction scheduler")
			auctionScheduler.Start(schedulerCtx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping auction scheduler")
			schedulerCancel()
			auctionScheduler.Stop()
			return nil
		},
	})
}
