package nats

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const healthCheckTimeout = 5 * time.Second

func New(cfg Config) (*natsgo.Conn, jetstream.JetStream, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, nil, err
	}

	options := buildOptions(cfg)

	conn, err := natsgo.Connect(cfg.URL, options...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()

	if _, healthErr := js.AccountInfo(ctx); healthErr != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("jetstream health check failed: %w", healthErr)
	}

	return conn, js, nil
}

func buildOptions(cfg Config) []natsgo.Option {
	options := []natsgo.Option{
		natsgo.MaxReconnects(cfg.MaxReconnects),
	}

	if strings.TrimSpace(cfg.Name) != "" {
		options = append(options, natsgo.Name(cfg.Name))
	}

	if cfg.ReconnectWait > 0 {
		options = append(options, natsgo.ReconnectWait(cfg.ReconnectWait))
	}

	if strings.TrimSpace(cfg.Creds) != "" {
		options = append(options, natsgo.UserCredentials(cfg.Creds))
	}

	if cfg.TLS {
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		options = append(options, natsgo.Secure(tlsConfig))
	}

	return options
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("nats URL cannot be empty")
	}

	if cfg.DedupeWindow < 0 {
		return errors.New("nats dedupe window cannot be negative")
	}

	return nil
}
