package config

type HTTPServer struct {
	Host            string `mapstructure:"HTTP_SERVER_HOST"`
	Port            uint   `mapstructure:"HTTP_SERVER_PORT"`
	ReadTimeout     int64  `mapstructure:"HTTP_SERVER_READ_TIMEOUT"`
	WriteTimeout    int64  `mapstructure:"HTTP_SERVER_WRITE_TIMEOUT"`
	IdleTimeout     int64  `mapstructure:"HTTP_SERVER_IDLE_TIMEOUT"`
	ShutdownTimeout int64  `mapstructure:"HTTP_SERVER_SHUTDOWN_TIMEOUT"`
}
