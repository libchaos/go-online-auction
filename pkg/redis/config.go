package redis

import "errors"

type ClientType string

const (
	ClientTypeSingleNode ClientType = "single_node"
	ClientTypeCluster    ClientType = "cluster"
)

type Config struct {
	URL        string
	ClientType ClientType
}

func (cfg *Config) validate() error {
	if cfg.URL == "" {
		return errors.New("url is required")
	}

	switch cfg.ClientType {
	case ClientTypeSingleNode, ClientTypeCluster:
		return nil
	default:
		return errors.New("client_type must be 'single_node' or 'cluster'")
	}
}
