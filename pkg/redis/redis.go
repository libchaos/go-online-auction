package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func New(cfg Config) (redis.UniversalClient, error) {
	var client redis.UniversalClient

	switch cfg.ClientType {
	case ClientTypeSingleNode:
		opts, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}
		opts.Password = cfg.Password
		opts.DB = cfg.DB
		if cfg.Password != "" {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		client = redis.NewClient(opts)

	case ClientTypeCluster:
		opts, err := redis.ParseClusterURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis cluster url: %w", err)
		}
		opts.Password = cfg.Password
		if cfg.Password != "" {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		client = redis.NewClusterClient(opts)

	default:
		return nil, fmt.Errorf("unsupported client type: %s", cfg.ClientType)
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}
