package handler

import (
	"net/http"

	alipaysdk "github.com/smartwalle/alipay/v3"

	"auction/internal/modules/payment/application/command"
	"auction/internal/shared/modules/logger"
)

// AlipayNotifyHandler exposes the public Alipay asynchronous-notification
// endpoint. It is intentionally unauthenticated: Alipay calls it server-to-
// server, and the signature is verified inside the command (not by a JWT).
type AlipayNotifyHandler struct {
	alipayNotifyCommand *command.AlipayNotifyCommand
	logger              logger.Logger
}

func NewAlipayNotifyHandler(
	alipayNotifyCommand *command.AlipayNotifyCommand,
	logger logger.Logger,
) *AlipayNotifyHandler {
	return &AlipayNotifyHandler{
		alipayNotifyCommand: alipayNotifyCommand,
		logger:              logger,
	}
}

func (notifyHandler *AlipayNotifyHandler) Notify(w http.ResponseWriter, r *http.Request) {
	if parseErr := r.ParseForm(); parseErr != nil {
		notifyHandler.logger.Error().Err(parseErr).Msg("failed to parse alipay notify form")
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	params := make(map[string]string, len(r.Form))
	for key, values := range r.Form {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	if _, err := notifyHandler.alipayNotifyCommand.Execute(r.Context(), command.AlipayNotifyCommandInput{
		Params: params,
	}); err != nil {
		notifyHandler.logger.Error().Err(err).Msg("failed to process alipay notify")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Acknowledge receipt to Alipay so it stops retrying. The body must be the
	// literal "success".
	alipaysdk.ACKNotification(w)
}
