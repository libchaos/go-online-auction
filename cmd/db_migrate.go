package cmd

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx database/sql driver used by goose
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"

	"auction/internal/shared/modules/config"
	"auction/migrations"
	"auction/pkg/database"
)

var dbMigrateCmd = &cobra.Command{
	Use:   "db:migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations. This command will run all the migrations that have not been run yet.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runGoose(func(db *sql.DB) error {
			return goose.Up(db, ".")
		}, "Migrations executed successfully")
	},
}

var dbMigrateDownCmd = &cobra.Command{
	Use:   "db:migrate:down",
	Short: "Rollback the last database migration",
	Long:  `Rollback the last database migration. This command reverts the most recently applied migration.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runGoose(func(db *sql.DB) error {
			return goose.Down(db, ".")
		}, "Migration rolled back successfully")
	},
}

var dbMigrateStatusCmd = &cobra.Command{
	Use:   "db:migrate:status",
	Short: "Show database migration status",
	Long:  `Show database migration status. This command lists all migrations and whether they have been applied.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runGoose(func(db *sql.DB) error {
			return goose.Status(db, ".")
		}, "")
	},
}

// runGoose opens a database connection, executes the given goose operation
// against the embedded migrations and logs successMsg when it is not empty.
func runGoose(op func(db *sql.DB) error, successMsg string) error {
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

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err = op(db); err != nil {
		return fmt.Errorf("failed to run goose operation: %w", err)
	}

	if successMsg != "" {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		logger.Info(successMsg)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(dbMigrateCmd)
	rootCmd.AddCommand(dbMigrateDownCmd)
	rootCmd.AddCommand(dbMigrateStatusCmd)
}
