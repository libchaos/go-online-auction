package httpserver

import (
	"context"

	"auction/internal/shared/modules/logger"
	"auction/pkg/httpserver"

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
			go func() {
				if err := server.Start(); err != nil {
					log.Error().Err(err).Msg("HTTP server stopped")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
