package config

type Redis struct {
	Addr     string `mapstructure:"REDIS_ADDR"     default:"localhost:6379"`
	Password string `mapstructure:"REDIS_PASSWORD" default:""`
	DB       int    `mapstructure:"REDIS_DB"       default:"0"`
}
