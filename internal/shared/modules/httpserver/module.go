package httpserver

import (
	"context"

	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"shared/httpserver",
	fx.Provide(New),
	fx.Invoke(startHTTPServer),
)

func startHTTPServer(
	lc fx.Lifecycle,
	server *httpserver.Server,
	log logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info().Str("addr", server.Addr()).Msg("starting HTTP server")
			return server.Start()
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
