package database

import (
	"context"
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
func OpenConnection(cfg Config) *pgxpool.Pool {
	dsn := generatePgxDatabaseDSN(cfg)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Errorf("failed to parse database config: %w", err))
	}

	// Set connection pool settings
	if cfg.MaxOpenConnections > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConnections)
	}
	if cfg.MaxIdleConnections > 0 {
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
		panic(fmt.Errorf("failed to create connection pool: %w", err))
	}

	// Verify the connection
	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	return pool
}

func generatePgxDatabaseDSN(cfg Config) string {
	sslMode := "require"
	if !cfg.SSLMode {
		sslMode = "disable"
	}

	if cfg.Port > math.MaxInt {
		panic(fmt.Sprintf("port value %d exceeds maximum int value", cfg.Port))
	}

	// Build DSN with connection pool settings
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s&timezone=UTC",
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

	return dsn
}

// GeneratePostgresDatabaseDSN generates a standard PostgreSQL DSN for migrations
func GeneratePostgresDatabaseDSN(cfg Config) string {
	if cfg.Port > math.MaxInt {
		panic(fmt.Sprintf("port value %d exceeds maximum int value", cfg.Port))
	}
	hostPort := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	return fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&TimeZone=UTC",
		cfg.User,
		cfg.Password,
		hostPort,
		cfg.Name,
	)
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
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	case tracelog.LogLevelInfo:
		l.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
	case tracelog.LogLevelWarn:
		l.logger.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
	case tracelog.LogLevelError:
		l.logger.LogAttrs(ctx, slog.LevelError, msg, attrs...)
	default:
		l.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
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
