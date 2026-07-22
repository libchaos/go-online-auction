package payment

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"auction/internal/modules/payment/application/command"
	"auction/internal/modules/payment/application/query"
	"auction/internal/modules/payment/application/service"
	"auction/internal/modules/payment/infra/alipay"
	"auction/internal/modules/payment/infra/http/chi/handler"
	"auction/internal/modules/payment/infra/http/chi/router"
	"auction/internal/modules/payment/infra/mapper"
	paymentoutbox "auction/internal/modules/payment/infra/outbox"
	"auction/internal/modules/payment/infra/repository"
	"auction/internal/modules/payment/infra/sqlcgen"
	"auction/internal/modules/payment/infra/uow"
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	"auction/pkg/httpserver"

	"go.uber.org/fx"
)

// Module wires the payment bounded context: recharge (user -> platform) and
// withdrawal (platform -> user Alipay) with the transactional outbox, the
// ledger (final fund arbiter), and the Alipay adapter.
var Module = fx.Module(
	"payment",

	fx.Provide(func(pool *pgxpool.Pool) sqlcgen.DBTX { return pool }),

	fx.Provide(mapper.NewPaymentMapper),

	fx.Provide(
		fx.Annotate(
			repository.NewPostgresPaymentRepository,
			fx.As(new(ports.PaymentRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			repository.NewPostgresWithdrawalRepository,
			fx.As(new(ports.WithdrawalRepository)),
		),
	),
	fx.Provide(
		fx.Annotate(
			uow.NewPaymentUnitOfWorkFactory,
			fx.As(new(ports.PaymentUnitOfWorkFactory)),
		),
	),

	fx.Provide(
		fx.Annotate(
			paymentoutbox.NewPostgresPaymentOutboxRepository,
			fx.As(new(ports.PaymentOutboxRepository)),
		),
	),

	// Alipay adapter: real smartwalle SDK when Provider == "alipay", otherwise
	// the in-memory mock so the service runs without credentials.
	fx.Provide(
		fx.Annotate(
			alipay.NewAlipayAdapter,
			fx.As(new(ports.AlipayPort)),
		),
	),

	fx.Provide(command.NewCreateDepositCommand),
	fx.Provide(command.NewCreateWithdrawalCommand),
	fx.Provide(command.NewAlipayNotifyCommand),

	fx.Provide(query.NewGetDepositQuery),
	fx.Provide(query.NewGetWithdrawalQuery),

	fx.Provide(handler.NewPaymentHandler),
	fx.Provide(handler.NewAlipayNotifyHandler),

	fx.Provide(service.NewDepositSuccessConsumer),
	fx.Provide(service.NewWithdrawalConsumer),

	fx.Provide(func(cfg config.Config) config.Alipay {
		return cfg.Alipay
	}),
	fx.Provide(func(cfg config.Config) paymentoutbox.Config {
		return paymentoutbox.Config{
			Interval:  cfg.Outbox.Interval,
			BatchSize: cfg.Outbox.BatchSize,
		}
	}),
	fx.Provide(paymentoutbox.NewRelay),
)

// RegisterPaymentRoutes wires the authenticated payment HTTP endpoints.
func RegisterPaymentRoutes(
	server *httpserver.Server,
	paymentHandler *handler.PaymentHandler,
	middleware *authn.Middleware,
) {
	router.RegisterPaymentRoutes(server, paymentHandler, middleware)
}

// RegisterAlipayNotifyRoute wires the public Alipay callback endpoint.
func RegisterAlipayNotifyRoute(
	server *httpserver.Server,
	alipayNotifyHandler *handler.AlipayNotifyHandler,
) {
	router.RegisterAlipayNotifyRoute(server, alipayNotifyHandler)
}

// RegisterPaymentConsumers starts the deposit-success and withdrawal Saga
// consumers (NATS subscriptions) and stops them with the application lifecycle.
func RegisterPaymentConsumers(
	lc fx.Lifecycle,
	depositSuccess *service.DepositSuccessConsumer,
	withdrawal *service.WithdrawalConsumer,
	logger logger.Logger,
) {
	consumerContext, consumerCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting payment consumers")

			if startErr := depositSuccess.Start(consumerContext); startErr != nil {
				logger.Error().Err(startErr).Msg("failed to start deposit-success consumer")

				return startErr
			}

			if startErr := withdrawal.Start(consumerContext); startErr != nil {
				logger.Error().Err(startErr).Msg("failed to start withdrawal consumer")

				return startErr
			}

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping payment consumers")
			consumerCancel()
			depositSuccess.Stop()
			withdrawal.Stop()

			return nil
		},
	})
}

// RegisterOutboxRelay wires the payment module's transactional outbox relay
// into the fx lifecycle. It runs in every process that writes payment events.
func RegisterOutboxRelay(
	lc fx.Lifecycle,
	relay *paymentoutbox.Relay,
	logger logger.Logger,
) {
	relayCtx, relayCancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info().Msg("starting payment outbox relay")
			relay.Start(relayCtx)

			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info().Msg("stopping payment outbox relay")
			relayCancel()
			relay.Stop()

			return nil
		},
	})
}
