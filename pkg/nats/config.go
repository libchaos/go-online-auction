package nats

import "time"

type Config struct {
	URL           string
	Name          string
	Creds         string
	TLS           bool
	MaxReconnects int
	ReconnectWait time.Duration
	DedupeWindow  time.Duration
}
