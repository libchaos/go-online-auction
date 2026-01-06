package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string     `mapstructure:"ENVIRONMENT"`
	HTTPPort    uint       `mapstructure:"HTTP_PORT"`
	CORS        CORS       `mapstructure:",squash"`
	DB          DB         `mapstructure:",squash"`
	App         App        `mapstructure:",squash"`
	Log         Log        `mapstructure:",squash"`
	Redis       Redis      `mapstructure:",squash"`
	HTTPServer  HTTPServer `mapstructure:",squash"`
}

const EnvProduction = "production"
const EnvDevelopment = "development"
const EnvStaging = "staging"
const EnvIntegration = "integration"

var _global Config

func Init() {
	v := viper.New()

	// Allow environment variables to override config file settings
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Determine config file based on environment
	configFile := ".env"
	if isIntegrationEnvironment() {
		configFile = ".env.integration"
	}

	// Configure to read from config file
	v.SetConfigName(configFile)
	v.SetConfigType("env")
	v.AddConfigPath(".")

	// Read the config file (must exist)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	//nolint:sloglint // this is a module
	slog.Info("Using config file", "file", v.ConfigFileUsed())

	// Unmarshal the config into our struct
	if err := v.Unmarshal(&_global); err != nil {
		//nolint:sloglint // this is a module
		slog.Error("Failed to unmarshal config", "error", err)
		panic(err)
	}
}

func GetConfig() Config {
	return _global
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsIntegration() bool {
	return c.Environment == "integration"
}

// isIntegrationEnvironment checks if we're running in integration test mode
// This checks for the TEST_ENV environment variable or if we're running with integration build tags
func isIntegrationEnvironment() bool {
	// Check if ENVIRONMENT is explicitly set to integration
	if env := os.Getenv("ENVIRONMENT"); env == "integration" {
		return true
	}

	return false
}
