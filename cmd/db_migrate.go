package cmd

import (
	"errors"
	"fmt"
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
	RunE: func(_ *cobra.Command, _ []string) error {
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
		dsn, err := database.GeneratePostgresDatabaseDSN(dbConfig)
		if err != nil {
			return err
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		m, err := migrate.New("file://migrations", dsn)
		if err != nil {
			return fmt.Errorf("failed to create migrate instance: %w", err)
		}

		if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		logger.Info("Migrations executed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dbMigrateCmd)
}
