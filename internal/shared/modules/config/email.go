package config

// Email holds the SMTP server configuration used by the notification module's
// SMTPEmailAdapter. When SMTPHost or FromAddress is empty the adapter is
// considered disabled and email sends become no-ops, which keeps local dev and
// tests healthy without a reachable mail server.
type Email struct {
	SMTPHost     string `mapstructure:"EMAIL_SMTP_HOST"`
	SMTPPort     int    `mapstructure:"EMAIL_SMTP_PORT"`
	SMTPUsername string `mapstructure:"EMAIL_SMTP_USERNAME"`
	SMTPPassword string `mapstructure:"EMAIL_SMTP_PASSWORD"`
	FromAddress  string `mapstructure:"EMAIL_FROM"`
	FromName     string `mapstructure:"EMAIL_FROM_NAME"`
	// UseTLS selects implicit TLS (SMTPS on port 465) instead of plaintext with
	// a STARTTLS upgrade (port 587).
	UseTLS bool `mapstructure:"EMAIL_SMTP_USE_TLS"`
}

// DefaultEmail returns a disabled-by-default email configuration. Operators opt
// in by setting EMAIL_SMTP_HOST and EMAIL_FROM.
func DefaultEmail() Email {
	return Email{}
}
