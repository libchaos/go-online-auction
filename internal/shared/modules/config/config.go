package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string     `mapstructure:"ENVIRONMENT"`
	CORS        CORS       `mapstructure:",squash"`
	DB          DB         `mapstructure:",squash"`
	App         App        `mapstructure:",squash"`
	Log         Log        `mapstructure:",squash"`
	NATS        NATS       `mapstructure:",squash"`
	HTTPServer  HTTPServer `mapstructure:",squash"`
	JWT         JWT        `mapstructure:",squash"`
	Scheduler   Scheduler  `mapstructure:",squash"`
	Outbox      Outbox     `mapstructure:",squash"`
	Payment     Payment    `mapstructure:",squash"`
	Alipay      Alipay     `mapstructure:",squash"`
	Email       Email      `mapstructure:",squash"`
}

const EnvProduction = "production"
const EnvDevelopment = "development"
const EnvStaging = "staging"
const EnvIntegration = "integration"

// WeakJWTSecret is the placeholder shipped in .env.example. It must never be
// used in production. The minimum safe length guards against operators who set
// a short, custom secret.
const WeakJWTSecret = "change-me-in-production-min-32-chars"
const minJWTSecretLen = 32

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

	// Read the config file. A missing file is a common misconfiguration
	// (wrong ENVIRONMENT, forgotten .env.integration, etc.); fail loudly with
	// the expected file name and the valid environments instead of a bare panic.
	if err := v.ReadInConfig(); err != nil {
		//nolint:sloglint // this is a module
		slog.Error("failed to read config file",
			"expected_file", configFile,
			"environment", os.Getenv("ENVIRONMENT"),
			"valid_environments", []string{"development", "staging", "production", "integration"},
			"hint", "create the expected .env file or set the required values via environment variables",
			"error", err)
		os.Exit(2) //nolint:mnd // 2 is the conventional exit code for configuration/usage errors
	}

	//nolint:sloglint // this is a module
	slog.Info("Using config file", "file", v.ConfigFileUsed())

	// Unmarshal the config into our struct
	if err := v.Unmarshal(&_global); err != nil {
		//nolint:sloglint // this is a module
		slog.Error("Failed to unmarshal config", "error", err)
		panic(err)
	}

	// sqashed sub-structs are not covered by viper SetDefault, so apply the
	// safe Alipay defaults explicitly (keep any value the operator provided).
	_global.Alipay = applyAlipayDefaults(_global.Alipay)

	// Fail fast (and loudly) on an unsafe JWT secret rather than letting the
	// service run with a trivially forgeable token signing key.
	validateJWTSecret(_global)
}

// validateJWTSecret enforces a minimum-strength JWT signing secret. In
// production a weak or placeholder secret is a critical misconfiguration, so we
// refuse to start. In every other environment we only warn, which keeps
// dev/test (where the shipped default is intentionally used) working.
func validateJWTSecret(cfg Config) {
	secret := cfg.JWT.Secret
	weak := secret == WeakJWTSecret || len(secret) < minJWTSecretLen
	if !weak {
		return
	}

	if cfg.IsProduction() {
		//nolint:sloglint // this is a module
		slog.Error("refusing to start: JWT secret is weak or missing",
			"reason", "JWT_SECRET must not be the placeholder value and must be at least 32 characters",
			"min_length", minJWTSecretLen,
			"hint", "set a strong, unique JWT_SECRET before running in production")
		os.Exit(1)
	}

	//nolint:sloglint // this is a module
	slog.Warn("JWT secret is weak or shorter than 32 characters",
		"reason", "this is unsafe for production; set a strong, unique JWT_SECRET",
		"min_length", minJWTSecretLen)
}

func GetConfig() Config {
	return _global
}

// applyAlipayDefaults fills any empty Alipay field with its safe default,
// leaving operator-provided values untouched.
func applyAlipayDefaults(a Alipay) Alipay {
	def := DefaultAlipay()
	if a.Provider == "" {
		a.Provider = def.Provider
	}
	if a.Gateway == "" {
		a.Gateway = def.Gateway
	}
	if a.PlatformAccountOwner == "" {
		a.PlatformAccountOwner = def.PlatformAccountOwner
	}

	return a
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
