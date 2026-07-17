package payment

import (
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
)

func NewPaymentPort(appLogger logger.Logger) ports.PaymentPort {
	paymentConfig := config.GetConfig().Payment

	switch paymentConfig.Provider {
	case "generic":
		return NewGenericPaymentAdapter(paymentConfig, appLogger)
	case "mock", "":
		return NewMockPaymentAdapter()
	default:
		appLogger.Warn().
			Str("provider", paymentConfig.Provider).
			Msg("unknown payment provider configured, falling back to mock adapter")

		return NewMockPaymentAdapter()
	}
}
