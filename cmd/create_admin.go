package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"auction/internal/modules/users/application/command"
	"auction/internal/modules/users/domain/enum"
	"auction/internal/modules/users/domain/model"
	"auction/internal/modules/users/infra/hasher"
	"auction/internal/modules/users/infra/mapper"
	"auction/internal/modules/users/infra/repository"
	"auction/internal/shared/modules/config"
	"auction/pkg/database"
)

var (
	createAdminName     string
	createAdminEmail    string
	createAdminPassword string
)

var createAdminCmd = &cobra.Command{
	Use:   "create-admin",
	Short: "Create an admin user",
	Long:  `Create an admin user. Admin accounts cannot be created via the public register endpoint.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if createAdminName == "" || createAdminEmail == "" || createAdminPassword == "" {
			return errors.New("--name, --email and --password are required")
		}

		if err := command.ValidatePassword(createAdminPassword); err != nil {
			return err
		}

		config.Init()
		cfg := config.GetConfig()
		pool, err := database.OpenConnection(database.Config{
			Host:               cfg.DB.Host,
			User:               cfg.DB.User,
			Password:           cfg.DB.Password,
			Name:               cfg.DB.Name,
			Port:               cfg.DB.Port,
			MaxOpenConnections: cfg.DB.MaxOpenConnections,
			MaxIdleConnections: cfg.DB.MaxIdleConnections,
			SSLMode:            cfg.DB.SSLMode,
			PrepareSTMT:        cfg.DB.PrepareSTMT,
			EnableLogs:         false,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()

		role, err := enum.NewRoleEnum(enum.EnumRoleAdmin)
		if err != nil {
			return err
		}

		passwordHasher := hasher.NewBcryptPasswordHasher()
		passwordHash, err := passwordHasher.Hash(createAdminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		user, err := model.NewUserModel(createAdminName, createAdminEmail, passwordHash, role)
		if err != nil {
			return err
		}

		userRepository := repository.NewPostgresUserRepository(pool, mapper.NewUserMapper())
		persistedUser, err := userRepository.Create(cmd.Context(), user)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		logger.Info("Admin user created", "id", persistedUser.ID(), "email", persistedUser.Email())
		return nil
	},
}

func init() {
	createAdminCmd.Flags().StringVar(&createAdminName, "name", "", "Admin user name")
	createAdminCmd.Flags().StringVar(&createAdminEmail, "email", "", "Admin user email")
	createAdminCmd.Flags().StringVar(&createAdminPassword, "password", "", "Admin user password")
	rootCmd.AddCommand(createAdminCmd)
}
