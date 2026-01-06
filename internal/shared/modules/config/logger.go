package config

type Log struct {
	IsEnabled bool   `mapstructure:"LOG_ENABLED"`
	LogLevel  string `mapstructure:"LOG_LEVEL"`
}
