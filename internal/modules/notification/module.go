package notification

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"auction/internal/modules/notification/application/command"
	"auction/internal/modules/notification/application/query"
	"auction/internal/modules/notification/application/service"
	"auction/internal/modules/notification/infra/email"
	"auction/internal/modules/notification/infra/http/chi/handler"
	"auction/internal/modules/notification/infra/http/chi/router"
	"auction/internal/modules/notification/infra/mapper"
	"auction/internal/modules/notification/infra/messaging"
	notificationoutbox "auction/internal/modules/notification/infra/outbox"
	"auction/internal/modules/notification/infra/repository"
	"auction/internal/modules/notification/infra/sqlcgen"
	"auction/internal/modules/notification/infra/sse"
	"auction/internal/modules/notification/infra/uow"
	"auction/internal/modules/notification/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	"auction/pkg/httpserver"
)

// Module wires the notification bounded context: an in-app notification center
// (REST), per-user preferences, a transactional outbox that reliably publishes
// notification.evt.{userID}, an SSE realtime hub that fans those events out to
// connected clients, and a read-only source-event consumer that turns payment,
// deposit and auction events into notifications.
var Module = fx.Module(
	"notification",

	fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX {
		return pool
	}),

	fx.Provide(mapper.NewNotificationMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresNotificationRepository,
			fx.As(new(ports.NotificationRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresPreferenceRepository,
			fx.As(new(ports.PreferenceRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresAuctionReadRepository,
			fx.As(new(ports.AuctionReadPort)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresWatchlistRepository,
			fx.As(new(ports.WatchlistRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresListingReadRepository,
			fx.As(new(ports.ListingReadPort)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresUserEmailResolver,
			fx.As(new(ports.UserEmailResolver)),
		),
	),
	fx.Provide(
		fx.Annotate(
			email.NewSMTPEmailAdapter,
			fx.As(new(ports.EmailPort)),
		),
	),
	fx.Provide(
		fx.Annotate(
			uow.NewNotificationUnitOfWorkFactory,
			fx.As(new(ports.NotificationUnitOfWorkFactory)),
		),
	),

	fx.Provide(
		fx.Annotate(
			notificationoutbox.NewPostgresNotificationOutboxRepository,
			fx.As(new(ports.NotificationOutboxRepository)),
		),
	),

	fx.Provide(command.NewCreateNotificationCommand),
	fx.Provide(command.NewMarkNotificationReadCommand),
	fx.Provide(command.NewMarkAllNotificationsReadCommand),
	fx.Provide(command.NewDeleteNotificationCommand),
	fx.Provide(command.NewUpdatePreferencesCommand),
	fx.Provide(command.NewCreateWatchCommand),
	fx.Provide(command.NewDeleteWatchCommand),
	fx.Provide(command.NewCreateEmailRequestCommand),

	fx.Provide(query.NewListNotificationsQuery),
	fx.Provide(query.NewListMyWatchesQuery),
	fx.Provide(query.NewGetUnreadCountQuery),
	fx.Provide(query.NewGetPreferencesQuery),

	fx.Provide(service.NewNotificationApplicationService),
	fx.Provide(service.NewSourceEventConsumer),
	fx.Provide(service.NewEmailDispatchConsumer),

	fx.Provide(sse.NewSubscriberRegistry),
	fx.Provide(
		fx.Annotate(
			messaging.NewJetStreamRealtimeEventConsumer,
			fx.As(new(sse.EventConsumer)),
		),
	),
	fx.Provide(sse.NewRealtimeHub),

	fx.Provide(handler.NewNotificationHandler),
	fx.Provide(handler.NewNotificationStreamHandler),
	fx.Provide(handler.NewWatchlistHandler),

	fx.Provide(func(cfg config.Config) config.Email {
		return cfg.Email
	}),
	fx.Provide(func(cfg config.Config) notificationoutbox.Config {
		return notificationoutbox.Config{
			Interval:  cfg.Outbox.Interval,
			BatchSize: cfg.Outbox.BatchSize,
		}
	}),
	fx.Provide(notificationoutbox.NewRelay),
)

// RegisterNotificationRoutes wires the authenticated notification-center and
// preference HTTP endpoints.
func RegisterNotificationRoutes(
	server *httpserver.Server,
	notificationHandler *handler.NotificationHandler,
	middleware *authn.Middleware,
) {
	router.RegisterNotificationRoutes(server, notificationHandler, middleware)
}

// RegisterWatchlistRoutes wires the authenticated watchlist (favourite a
// product) endpoints into the fx lifecycle.
func RegisterWatchlistRoutes(
	server *httpserver.Server,
	watchlistHandler *handler.WatchlistHandler,
	middleware *authn.Middleware,
) {
	router.RegisterWatchlistRoutes(server, watchlistHandler, middleware)
}

// RegisterNotificationStreamRoute wires the SSE endpoint, which authenticates via
// a token query parameter rather than the header-only RequireAuth middleware.
func RegisterNotificationStreamRoute(
	server *httpserver.Server,
	streamHandler *handler.NotificationStreamHandler,
) {
	router.RegisterNotificationStreamRoute(server, streamHandler)
}

// RegisterNotificationHub runs the SSE realtime hub for the application lifetime.
func RegisterNotificationHub(
	lc fx.Lifecycle,
	hub *sse.RealtimeHub,
	logger logger.Logger,
) {
	hubContext, hubCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting notification realtime hub")
			go hub.Run(hubContext)

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping notification realtime hub")
			hubCancel()

			return nil
		},
	})
}

// RegisterNotificationSourceConsumer starts the read-only source-event consumer
// that translates payment, deposit and auction events into notifications.
func RegisterNotificationSourceConsumer(
	lc fx.Lifecycle,
	sourceConsumer *service.SourceEventConsumer,
	logger logger.Logger,
) {
	consumerContext, consumerCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting notification source consumer")
			startErr := sourceConsumer.Start(consumerContext)
			if startErr != nil {
				logger.Error().Err(startErr).Msg("failed to start notification source consumer")
			}

			return startErr
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping notification source consumer")
			consumerCancel()
			sourceConsumer.Stop()

			return nil
		},
	})
}

// RegisterOutboxRelay wires the notification module's transactional outbox relay
// into the fx lifecycle.
func RegisterOutboxRelay(
	lc fx.Lifecycle,
	relay *notificationoutbox.Relay,
	logger logger.Logger,
) {
	relayCtx, relayCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting notification outbox relay")
			relay.Start(relayCtx)

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping notification outbox relay")
			relayCancel()
			relay.Stop()

			return nil
		},
	})
}

// RegisterNotificationEmailConsumer starts the email dispatch consumer for the
// application lifetime. It is a durable, explicit-ack consumer so emails are
// sent at-least-once and never lost on restart.
func RegisterNotificationEmailConsumer(
	lc fx.Lifecycle,
	emailConsumer *service.EmailDispatchConsumer,
	logger logger.Logger,
) {
	consumerCtx, consumerCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting notification email consumer")
			startErr := emailConsumer.Start(consumerCtx)
			if startErr != nil {
				logger.Error().Err(startErr).Msg("failed to start notification email consumer")
			}

			return startErr
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping notification email consumer")
			consumerCancel()
			emailConsumer.Stop()

			return nil
		},
	})
}
