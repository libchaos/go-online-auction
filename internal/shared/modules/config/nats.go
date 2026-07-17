package config

import "time"

type NATS struct {
	URL           string        `mapstructure:"NATS_URL"`
	Name          string        `mapstructure:"NATS_NAME"`
	Creds         string        `mapstructure:"NATS_CREDS"`
	TLS           bool          `mapstructure:"NATS_TLS"`
	MaxReconnects int           `mapstructure:"NATS_MAX_RECONNECTS"`
	ReconnectWait time.Duration `mapstructure:"NATS_RECONNECT_WAIT"`
	DedupeWindow  time.Duration `mapstructure:"NATS_DEDUPE_WINDOW"`
}
