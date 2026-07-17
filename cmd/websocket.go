package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"auction/internal/modules/auction"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/database"
	"auction/internal/shared/modules/httpserver"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/modules/nats"
)

var websocketCmd = &cobra.Command{
	Use:   "websocket",
	Short: "Start the websocket module",
	Long:  `Start the auction websocket server for real-time updates.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		app := fx.New(
			config.Module,
			logger.Module,
			database.Module,
			nats.Module,
			httpserver.Module,
			auction.Module,
			fx.Invoke(
				auction.RegisterWebsocketRoutes,
			),
		)
		app.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(websocketCmd)
}
