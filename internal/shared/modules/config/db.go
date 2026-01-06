package config

type DB struct {
	Host               string `mapstructure:"DB_HOST"`
	Name               string `mapstructure:"DB_NAME"`
	User               string `mapstructure:"DB_USER"`
	Password           string `mapstructure:"DB_PASS"`
	Port               uint   `mapstructure:"DB_PORT"`
	MaxOpenConnections int    `mapstructure:"DB_MAX_OPEN_CONNECTIONS"`
	MaxIdleConnections int    `mapstructure:"DB_MAX_IDLE_CONNECTIONS"`
	SSLMode            bool   `mapstructure:"DB_SSL_MODE"`
	PrepareSTMT        bool   `mapstructure:"DB_PREPARE_STMT"`
	EnableLogs         bool   `mapstructure:"DB_ENABLE_LOGS"`
}
