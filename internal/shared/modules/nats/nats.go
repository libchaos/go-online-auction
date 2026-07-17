package nats

import (
	"auction/internal/shared/modules/config"
	"auction/pkg/nats"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func New(cfg config.Config) (*natsgo.Conn, jetstream.JetStream, error) {
	return nats.New(nats.Config{
		URL:           cfg.NATS.URL,
		Name:          cfg.NATS.Name,
		Creds:         cfg.NATS.Creds,
		TLS:           cfg.NATS.TLS,
		MaxReconnects: cfg.NATS.MaxReconnects,
		ReconnectWait: cfg.NATS.ReconnectWait,
		DedupeWindow:  cfg.NATS.DedupeWindow,
	})
}
