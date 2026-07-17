package config

type JWT struct {
	Secret                string `mapstructure:"JWT_SECRET"`
	AccessTokenTTLMinutes int    `mapstructure:"JWT_ACCESS_TOKEN_TTL_MINUTES"`
	RefreshTokenTTLHours  int    `mapstructure:"JWT_REFRESH_TOKEN_TTL_HOURS"`
}
