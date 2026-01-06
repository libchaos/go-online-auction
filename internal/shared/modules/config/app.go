package config

type App struct {
	Name    string `mapstructure:"APP_NAME"`
	BaseURL string `mapstructure:"APP_BASE_URL"`
}
