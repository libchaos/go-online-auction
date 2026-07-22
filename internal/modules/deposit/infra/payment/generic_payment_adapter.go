package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
)

const defaultPaymentTimeout = 15 * time.Second

var _ ports.PaymentPort = (*GenericPaymentAdapter)(nil)

type RequestSigner interface {
	Sign(request *http.Request)
}

type bearerSigner struct {
	header string
	token  string
}

func (signer bearerSigner) Sign(request *http.Request) {
	if signer.header == "" || signer.token == "" {
		return
	}

	request.Header.Set(signer.header, signer.token)
}

type holdRequest struct {
	UserID        uint64 `json:"user_id"`
	AmountInCents uint64 `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	Reference     string `json:"reference"`
}

type referenceRequest struct {
	ExternalReference string `json:"external_reference"`
}

type captureRequest struct {
	ExternalReference string `json:"external_reference"`
	AmountInCents     uint64 `json:"amount_in_cents"`
}

type paymentResponse struct {
	ExternalReference string `json:"external_reference"`
	Reference         string `json:"reference"`
	ID                string `json:"id"`
}

type GenericPaymentAdapter struct {
	paymentConfig config.Payment
	client        *http.Client
	signer        RequestSigner
	logger        logger.Logger
}

func NewGenericPaymentAdapter(paymentConfig config.Payment, appLogger logger.Logger) *GenericPaymentAdapter {
	timeout := paymentConfig.Timeout
	if timeout <= 0 {
		timeout = defaultPaymentTimeout
	}

	authHeader := paymentConfig.AuthHeader
	if authHeader == "" {
		authHeader = "Authorization"
	}

	return &GenericPaymentAdapter{
		paymentConfig: paymentConfig,
		client:        &http.Client{Timeout: timeout},
		signer:        bearerSigner{header: authHeader, token: paymentConfig.APIKey},
		logger:        appLogger,
	}
}

func (adapter *GenericPaymentAdapter) do(ctx context.Context, path string, body any) (string, error) {
	url := strings.TrimRight(adapter.paymentConfig.BaseURL, "/") + path

	var payload bytes.Buffer
	if body != nil {
		if encodeErr := json.NewEncoder(&payload).Encode(body); encodeErr != nil {
			return "", fmt.Errorf("failed to encode payment request: %w", encodeErr)
		}
	}

	request, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, url, &payload)
	if requestErr != nil {
		return "", fmt.Errorf("failed to build payment request: %w", requestErr)
	}

	request.Header.Set("Content-Type", "application/json")
	adapter.signer.Sign(request)

	response, responseErr := adapter.client.Do(request)
	if responseErr != nil {
		return "", fmt.Errorf("payment provider request failed: %w", responseErr)
	}
	if response == nil {
		return "", errors.New("payment provider returned an empty response")
	}
	defer func() { _ = response.Body.Close() }()

	raw, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read payment response: %w", readErr)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("payment provider returned status %d: %s", response.StatusCode, string(raw))
	}

	var parsed paymentResponse
	if unmarshalErr := json.Unmarshal(raw, &parsed); unmarshalErr != nil {
		return "", fmt.Errorf("failed to decode payment response: %w", unmarshalErr)
	}

	if parsed.ExternalReference != "" {
		return parsed.ExternalReference, nil
	}

	return parsed.ID, nil
}

func (adapter *GenericPaymentAdapter) Hold(
	ctx context.Context,
	userID uint64,
	amount model.MoneyModel,
	currency string,
	reference string,
) (string, error) {
	externalReference, holdErr := adapter.do(ctx, adapter.paymentConfig.HoldPath, holdRequest{
		UserID:        userID,
		AmountInCents: amount.AmountInCents(),
		Currency:      currency,
		Reference:     reference,
	})
	if holdErr != nil {
		return "", holdErr
	}

	if externalReference == "" {
		return reference, nil
	}

	return externalReference, nil
}

func (adapter *GenericPaymentAdapter) Release(ctx context.Context, externalReference string) error {
	_, releaseErr := adapter.do(ctx, adapter.paymentConfig.ReleasePath, referenceRequest{
		ExternalReference: externalReference,
	})

	return releaseErr
}

func (adapter *GenericPaymentAdapter) Capture(
	ctx context.Context,
	externalReference string,
	amount model.MoneyModel,
) error {
	_, captureErr := adapter.do(ctx, adapter.paymentConfig.CapturePath, captureRequest{
		ExternalReference: externalReference,
		AmountInCents:     amount.AmountInCents(),
	})

	return captureErr
}

func (adapter *GenericPaymentAdapter) Forfeit(ctx context.Context, externalReference string) error {
	_, forfeitErr := adapter.do(ctx, adapter.paymentConfig.ForfeitPath, referenceRequest{
		ExternalReference: externalReference,
	})

	return forfeitErr
}
