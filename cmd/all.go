package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"auction/internal/modules/auction"
	"auction/internal/modules/deposit"
	"auction/internal/modules/ledger"
	"auction/internal/modules/listing"
	"auction/internal/modules/notification"
	"auction/internal/modules/payment"
	"auction/internal/modules/users"
	"auction/internal/shared/modules/authn"
	"auction/internal/shared/modules/authz"
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
			authz.Module,
			users.Module,
			listing.Module,
			auction.Module,
			deposit.Module,
			ledger.Module,
			payment.Module,
			notification.Module,
			fx.Invoke(
				users.RegisterUserRoutes,
				users.RegisterRBACRoutes,
				listing.RegisterListingRoutes,
				auction.RegisterAuctionRoutes,
				auction.RegisterWebsocketRoutes,
				auction.RegisterBidProcessor,
				auction.RegisterOutboxRelay,
				deposit.RegisterOutboxRelay,
				listing.RegisterOutboxRelay,
				payment.RegisterOutboxRelay,
				auction.RegisterAuctionScheduler,
				auction.RegisterMetricsRoute,
				deposit.RegisterDepositRoutes,
				deposit.RegisterDepositWebsocketRoutes,
				ledger.RegisterLedgerRoutes,
				deposit.RegisterDepositHub,
				deposit.RegisterDepositEventConsumer,
				payment.RegisterPaymentRoutes,
				payment.RegisterAlipayNotifyRoute,
				payment.RegisterPaymentConsumers,
				notification.RegisterOutboxRelay,
				notification.RegisterNotificationRoutes,
				notification.RegisterNotificationStreamRoute,
				notification.RegisterNotificationHub,
				notification.RegisterNotificationSourceConsumer,
				notification.RegisterNotificationEmailConsumer,
				notification.RegisterWatchlistRoutes,
			),
		)
		app.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(allCmd)
}
