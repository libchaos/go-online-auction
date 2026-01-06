package config

import "github.com/cristiano-pacheco/go-online-auction/pkg/redis"

type Redis struct {
	URL        string           `mapstructure:"REDIS_URL"`
	DB         int              `mapstructure:"REDIS_DB"`
	Password   string           `mapstructure:"REDIS_PASSWORD"`
	ClientType redis.ClientType `mapstructure:"REDIS_CLIENT_TYPE"`
}
