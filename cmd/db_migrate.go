package cmd

import (
	"errors"
	"log/slog"
	"os"

	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
	"github.com/cristiano-pacheco/go-online-auction/pkg/database"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // needed for postgres migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"       // needed for file source migrations
	"github.com/spf13/cobra"
)

var dbMigrateCmd = &cobra.Command{
	Use:   "db:migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations. This command will run all the migrations that have not been run yet.`,
	Run: func(_ *cobra.Command, _ []string) {
		config.Init()
		cfg := config.GetConfig()
		dbConfig := database.Config{
			Host:               cfg.DB.Host,
			User:               cfg.DB.User,
			Password:           cfg.DB.Password,
			Name:               cfg.DB.Name,
			Port:               cfg.DB.Port,
			MaxOpenConnections: cfg.DB.MaxOpenConnections,
			MaxIdleConnections: cfg.DB.MaxIdleConnections,
			SSLMode:            cfg.DB.SSLMode,
			PrepareSTMT:        cfg.DB.PrepareSTMT,
			EnableLogs:         cfg.DB.EnableLogs,
		}
		dsn := database.GeneratePostgresDatabaseDSN(dbConfig)

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		m, err := migrate.New("file://migrations", dsn)
		if err != nil {
			logger.Error("failed to create migrate instance", "err", err)
			os.Exit(1)
		}

		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Error("failed to run migrations", "err", err)
			os.Exit(1)
		}

		logger.Info("Migrations executed successfully")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(dbMigrateCmd)
}
