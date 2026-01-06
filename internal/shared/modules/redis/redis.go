package redis

import (
	"github.com/cristiano-pacheco/go-online-auction/internal/shared/modules/config"
	"github.com/cristiano-pacheco/go-online-auction/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
)

type UniversalClient = goredis.UniversalClient

func New(cfg config.Config) (UniversalClient, error) {
	return redis.New(redis.Config{
		URL:        cfg.Redis.URL,
		Password:   cfg.Redis.Password,
		DB:         cfg.Redis.DB,
		ClientType: cfg.Redis.ClientType,
	})
}
