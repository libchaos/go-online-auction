package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"auction/internal/modules/auction"
	"auction/internal/modules/deposit"
	"auction/internal/modules/listing"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/database"
	"auction/internal/shared/modules/httpserver"
	"auction/internal/shared/modules/logger"
	"auction/internal/shared/modules/nats"
)

var bidProcessorCmd = &cobra.Command{
	Use:   "bid-processor",
	Short: "Start the bid processor",
	Long:  `Start the event-driven bid processor that consumes bid commands from JetStream.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		app := fx.New(
			config.Module,
			logger.Module,
			database.Module,
			nats.Module,
			httpserver.Module,
			auction.Module,
			deposit.Module,
			fx.Invoke(
				auction.RegisterBidProcessor,
				auction.RegisterOutboxRelay,
				deposit.RegisterOutboxRelay,
				listing.RegisterOutboxRelay,
			),
		)
		app.Run()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(bidProcessorCmd)
}
