package database

type Config struct {
	Host               string
	Name               string
	User               string
	Password           string
	Port               uint
	MaxOpenConnections int
	MaxIdleConnections int
	SSLMode            bool
	PrepareSTMT        bool
	EnableLogs         bool
	LogLevel           LogLevel
}

// LogLevel represents the logging level for database operations
type LogLevel int

const (
	// LogLevelSilent no logging
	LogLevelSilent LogLevel = iota
	// LogLevelError only error logs
	LogLevelError
	// LogLevelWarn warning and error logs
	LogLevelWarn
	// LogLevelInfo info, warning and error logs
	LogLevelInfo
)
