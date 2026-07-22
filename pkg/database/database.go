package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

const defaultPingTimeout = 5 * time.Second

// OpenConnection creates and returns a new pgxpool connection pool
func OpenConnection(cfg Config) (*pgxpool.Pool, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}
	dsn, err := generatePgxDatabaseDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database DSN: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConnections > 0 && cfg.MaxOpenConnections <= math.MaxInt32 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 && cfg.MaxIdleConnections <= math.MaxInt32 {
		poolConfig.MinConns = int32(cfg.MaxIdleConnections)
	}

	// Configure logging if enabled
	if cfg.EnableLogs {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   newPgxLogger(logger),
			LogLevel: mapLogLevel(cfg.LogLevel),
		}
	}

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	if pingErr := pool.Ping(ctx); pingErr != nil {
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return pool, nil
}

func generatePgxDatabaseDSN(cfg Config) (string, error) {
	sslMode := "require"
	if !cfg.SSLMode {
		sslMode = "disable"
	}

	if cfg.Port > math.MaxInt {
		return "", fmt.Errorf("port value %d exceeds maximum int value", cfg.Port)
	}

	// Build DSN with connection pool settings
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s&timezone=Asia/Shanghai",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Name,
		sslMode,
	)

	// Add prepared statement setting if enabled
	if cfg.PrepareSTMT {
		dsn += "&default_query_exec_mode=cache_statement"
	}

	return dsn, nil
}

// GeneratePostgresDatabaseDSN generates a standard PostgreSQL DSN for migrations
func GeneratePostgresDatabaseDSN(cfg Config) (string, error) {
	if cfg.Port > math.MaxInt {
		return "", fmt.Errorf("port value %d exceeds maximum int value", cfg.Port)
	}
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&TimeZone=Asia/Shanghai",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Name,
	), nil
}

// pgxLogger adapts slog.Logger to pgx's Logger interface
type pgxLogger struct {
	logger *slog.Logger
}

func newPgxLogger(logger *slog.Logger) *pgxLogger {
	return &pgxLogger{logger: logger}
}

func (l *pgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	attrs := make([]slog.Attr, 0, len(data))
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}

	switch level {
	case tracelog.LogLevelNone:
		// No logging
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	case tracelog.LogLevelInfo:
		l.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	case tracelog.LogLevelWarn:
		l.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	case tracelog.LogLevelError:
		l.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	}
}

// mapLogLevel maps our LogLevel to pgx tracelog.LogLevel
func mapLogLevel(level LogLevel) tracelog.LogLevel {
	switch level {
	case LogLevelSilent:
		return tracelog.LogLevelNone
	case LogLevelError:
		return tracelog.LogLevelError
	case LogLevelWarn:
		return tracelog.LogLevelWarn
	case LogLevelInfo:
		return tracelog.LogLevelInfo
	default:
		return tracelog.LogLevelNone
	}
}

func validateConfig(cfg Config) error {
	if cfg.Host == "" {
		return errors.New("database host is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return errors.New("database port must be between 1 and 65535")
	}
	if cfg.User == "" {
		return errors.New("database user is required")
	}
	if cfg.Name == "" {
		return errors.New("database name is required")
	}
	return nil
}
