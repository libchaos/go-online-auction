package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/database"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/httpserver"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/redis"
	pkghttpserver "github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
	pkglogger "github.com/cristiano-pacheco/go-online-auction/pkg/logger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	Long:  `Start the auction HTTP server with WebSocket support for real-time updates.`,
	Run: func(_ *cobra.Command, _ []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServer() {
	fx.New(
		config.Module,
		logger.Module,
		database.Module,
		redis.Module,
		httpserver.Module,

		auction.Module,

		fx.Invoke(startHTTPServer),
	).Run()
}

func startHTTPServer(
	lc fx.Lifecycle,
	server *pkghttpserver.Server,
	log pkglogger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info().Str("addr", server.Addr()).Msg("starting HTTP server")
			go func() {
				if err := server.Start(); err != nil {
					log.Error().Err(err).Msg("HTTP server error")
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
