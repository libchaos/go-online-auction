package config

type Redis struct {
	URL        string `mapstructure:"REDIS_URL"`
	ClientType string `mapstructure:"REDIS_CLIENT_TYPE"`
}
