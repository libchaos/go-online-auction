package alipay

import (
	"auction/internal/modules/payment/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
)

// NewAlipayAdapter selects the AlipayPort implementation from configuration.
// When Provider is "alipay" the real smartwalle SDK adapter is used; otherwise
// the in-memory mock adapter is returned so the service runs without
// credentials.
func NewAlipayAdapter(cfg config.Alipay, logger logger.Logger) (ports.AlipayPort, error) {
	if cfg.UseRealClient() {
		return NewSDKAlipayAdapter(cfg, logger)
	}

	return NewMockAlipayAdapter(logger), nil
}
