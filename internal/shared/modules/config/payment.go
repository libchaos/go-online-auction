package config

import "time"

type Payment struct {
	Provider    string        `mapstructure:"PAYMENT_PROVIDER"`
	BaseURL     string        `mapstructure:"PAYMENT_BASE_URL"`
	APIKey      string        `mapstructure:"PAYMENT_API_KEY"`
	AuthHeader  string        `mapstructure:"PAYMENT_AUTH_HEADER"`
	Timeout     time.Duration `mapstructure:"PAYMENT_TIMEOUT"`
	HoldPath    string        `mapstructure:"PAYMENT_HOLD_PATH"`
	ReleasePath string        `mapstructure:"PAYMENT_RELEASE_PATH"`
	CapturePath string        `mapstructure:"PAYMENT_CAPTURE_PATH"`
	ForfeitPath string        `mapstructure:"PAYMENT_FORFEIT_PATH"`
}
