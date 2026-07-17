package deposit

import (
	"context"

	"go.uber.org/fx"

	"auction/internal/modules/deposit/application/command"
	"auction/internal/modules/deposit/application/guard"
	"auction/internal/modules/deposit/application/query"
	"auction/internal/modules/deposit/infra/http/chi/handler"
	"auction/internal/modules/deposit/infra/http/chi/router"
	"auction/internal/modules/deposit/infra/mapper"
	"auction/internal/modules/deposit/infra/messaging"
	"auction/internal/modules/deposit/infra/payment"
	"auction/internal/modules/deposit/infra/repository"
	"auction/internal/modules/deposit/infra/uow"
	"auction/internal/modules/deposit/infra/websocket"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/logger"
	"auction/pkg/httpserver"
)

var Module = fx.Module(
	"deposit",

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

	fx.Provide(messaging.NewJetStreamDepositEventConsumer),
	fx.Provide(messaging.NewSettlementConsumer),
)

func RegisterDepositRoutes(
	server *httpserver.Server,
	depositHandler *handler.DepositHandler,
	middleware *authn.Middleware,
) {
	router.RegisterDepositRoutes(server, depositHandler, middleware)
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
