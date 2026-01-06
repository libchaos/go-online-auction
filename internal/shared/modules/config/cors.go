package config

import "strings"

type CORS struct {
	// AllowedOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	AllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`

	// AllowedMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (HEAD, GET and POST).
	AllowedMethods string `mapstructure:"CORS_ALLOWED_METHODS"`

	// AllowedHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	AllowedHeaders string `mapstructure:"CORS_ALLOWED_HEADERS"`

	// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposedHeaders string `mapstructure:"CORS_EXPOSED_HEADERS"`

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool `mapstructure:"CORS_ALLOW_CREDENTIALS"`

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached
	MaxAge int `mapstructure:"CORS_MAX_AGE"`

	// OptionsPassthrough instructs preflight to let other potential next handlers to
	// process the OPTIONS method. Turn this on if your application handles OPTIONS.
	OptionsPassthrough bool `mapstructure:"CORS_OPTIONS_PASSTHROUGH"`

	// Debug flag adds additional output to debug server side CORS issues
	Debug bool `mapstructure:"CORS_DEBUG"`
}

// GetAllowedOrigins returns the allowed origins as a string slice.
func (c *CORS) GetAllowedOrigins() []string {
	if c.AllowedOrigins == "" {
		return []string{"*"}
	}
	return splitAndTrim(c.AllowedOrigins)
}

// GetAllowedMethods returns the allowed methods as a string slice.
func (c *CORS) GetAllowedMethods() []string {
	if c.AllowedMethods == "" {
		return []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	return splitAndTrim(c.AllowedMethods)
}

// GetAllowedHeaders returns the allowed headers as a string slice.
func (c *CORS) GetAllowedHeaders() []string {
	if c.AllowedHeaders == "" {
		return []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}
	}
	return splitAndTrim(c.AllowedHeaders)
}

// GetExposedHeaders returns the exposed headers as a string slice.
func (c *CORS) GetExposedHeaders() []string {
	if c.ExposedHeaders == "" {
		return []string{"Link"}
	}
	return splitAndTrim(c.ExposedHeaders)
}

// Helper function to split comma-separated values and trim spaces.
func splitAndTrim(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
