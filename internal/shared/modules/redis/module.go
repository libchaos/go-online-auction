package redis

import (
	"context"

	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"shared/redis",
	fx.Provide(New),
	fx.Invoke(closeConnection),
)

func closeConnection(
	lc fx.Lifecycle,
	client UniversalClient,
	log logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			log.Info().Msg("closing Redis client connection")
			return client.Close()
		},
	})
}
