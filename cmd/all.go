package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"auction/internal/modules/auction"
	"auction/internal/modules/deposit"
	"auction/internal/modules/listing"
	"auction/internal/modules/users"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/database"
	"auction/internal/shared/modules/httpserver"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/modules/nats"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Start the HTTP server",
	Long:  `Start the auction HTTP server with WebSocket support for real-time updates.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		app := fx.New(
			config.Module,
			logger.Module,
			database.Module,
			nats.Module,
			httpserver.Module,
			authn.Module,
			users.Module,
			listing.Module,
			auction.Module,
			deposit.Module,
			fx.Invoke(
				users.RegisterUserRoutes,
				listing.RegisterListingRoutes,
				auction.RegisterAuctionRoutes,
				auction.RegisterWebsocketRoutes,
				auction.RegisterBidProcessor,
				auction.RegisterOutboxRelay,
				auction.RegisterAuctionScheduler,
				auction.RegisterMetricsRoute,
				deposit.RegisterDepositRoutes,
				deposit.RegisterDepositWebsocketRoutes,
				deposit.RegisterDepositHub,
				deposit.RegisterDepositEventConsumer,
			),
		)
		app.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}
