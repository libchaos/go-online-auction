package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/cristiano-pacheco/go-online-auction/internal/modules/auction"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/database"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/httpserver"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/logger"
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/redis"
)

var auctionCmd = &cobra.Command{
	Use:   "auction",
	Short: "Start the auction module",
	Long:  `Start the auction HTTP server with WebSocket support for real-time updates.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		app := fx.New(
			config.Module,
			logger.Module,
			database.Module,
			redis.Module,
			httpserver.Module,
			auction.Module,
			fx.Invoke(
				auction.RegisterAuctionRoutes,
			),
		)
		app.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(auctionCmd)
}
