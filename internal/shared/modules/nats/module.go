package nats

import (
	"context"

	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"shared/nats",
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(
	lc fx.Lifecycle,
	conn *natsgo.Conn,
	js jetstream.JetStream,
	cfg config.Config,
	log logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info().Msg("creating or updating NATS JetStream streams")
			return CreateOrUpdateStreams(ctx, js, cfg.NATS.DedupeWindow)
		},
		OnStop: func(_ context.Context) error {
			log.Info().Msg("draining NATS connection")
			return conn.Drain()
		},
	})
}
