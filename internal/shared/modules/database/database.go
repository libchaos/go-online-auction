package database

import (
	"auction/internal/shared/modules/config"
	"auction/pkg/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config) (*pgxpool.Pool, error) {
	dbConfig := database.Config{
		Host:               cfg.DB.Host,
		Port:               cfg.DB.Port,
		User:               cfg.DB.User,
		Password:           cfg.DB.Password,
		Name:               cfg.DB.Name,
		MaxOpenConnections: cfg.DB.MaxOpenConnections,
		MaxIdleConnections: cfg.DB.MaxIdleConnections,
		SSLMode:            cfg.DB.SSLMode,
		PrepareSTMT:        cfg.DB.PrepareSTMT,
		EnableLogs:         cfg.DB.EnableLogs,
	}
	return database.OpenConnection(dbConfig)
}
