package cmd

import (
	"context"
	"errors"
	"net/http"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/database"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/httpserver"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/redis"
	pkghttpserver "github.com/cristiano-pacheco/go-online-auction/pkg/httpserver"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Start the HTTP server",
	Long:  `Start the auction HTTP server with WebSocket support for real-time updates.`,
	Run: func(_ *cobra.Command, _ []string) {
		runServer()
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
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
	log logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info().Str("addr", server.Addr()).Msg("starting HTTP server")
			go func() {
				if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
