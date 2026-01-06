package logger

const (
	LogLevelDebug    = "debug"
	LogLevelInfo     = "info"
	LogLevelWarn     = "warning"
	LogLevelError    = "error"
	LogLevelFatal    = "fatal"
	LogLevelPanic    = "panic"
	LogLevelNoLevel  = "nolevel"
	LogLevelDisabled = "disabled"
)

type Config struct {
	LogLevel LogLevel
}

type LogLevel struct {
	value string
}

func MustLogLevel(level string) LogLevel {
	if isValidLogLevel(level) {
		return LogLevel{value: level}
	}
	panic("invalid log level")
}

func (l LogLevel) String() string {
	return l.value
}

func isValidLogLevel(level string) bool {
	switch level {
	case LogLevelDebug:
		return true
	case LogLevelInfo:
		return true
	case LogLevelWarn:
		return true
	case LogLevelError:
		return true
	case LogLevelFatal:
		return true
	case LogLevelPanic:
		return true
	case LogLevelNoLevel:
		return true
	case LogLevelDisabled:
		return true
	default:
		return false
	}
}
