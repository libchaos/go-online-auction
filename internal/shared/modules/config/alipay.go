package config

// Alipay holds the configuration for the Alipay integration used by the
// payment module. The Provider selects which AlipayPort implementation is
// wired: "alipay" uses the real smartwalle SDK adapter, anything else (the
// default "mock") uses an in-memory adapter so the service compiles and runs
// without credentials (local dev, CI, unit tests).
type Alipay struct {
	Provider             string `mapstructure:"ALIPAY_PROVIDER"`
	AppID                string `mapstructure:"ALIPAY_APP_ID"`
	AppPrivateKey        string `mapstructure:"ALIPAY_APP_PRIVATE_KEY"`
	PublicKey            string `mapstructure:"ALIPAY_PUBLIC_KEY"`
	Gateway              string `mapstructure:"ALIPAY_GATEWAY"`
	NotifyBaseURL        string `mapstructure:"ALIPAY_NOTIFY_BASE_URL"`
	PlatformAccountOwner string `mapstructure:"PAYMENT_PLATFORM_ACCOUNT_OWNER"`
}

// DefaultAlipay returns the Alipay configuration with safe defaults applied.
func DefaultAlipay() Alipay {
	return Alipay{
		Provider:             "mock",
		Gateway:              "https://openapi.alipaydev.com/gateway.do",
		PlatformAccountOwner: "platform",
	}
}

// IsProductionGateway reports whether the configured gateway points at the
// real Alipay production endpoint (vs the sandbox openapi.alipaydev.com).
func (a Alipay) IsProductionGateway() bool {
	return a.Gateway == "https://openapi.alipay.com/gateway.do"
}

// UseRealClient reports whether the real Alipay SDK adapter should be used.
func (a Alipay) UseRealClient() bool {
	return a.Provider == "alipay"
}
