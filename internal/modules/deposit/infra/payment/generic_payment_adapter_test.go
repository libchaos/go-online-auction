package payment_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"auction/internal/modules/deposit/domain/model"
	"auction/internal/modules/deposit/infra/payment"
	"auction/internal/modules/deposit/ports"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/logger"
	"github.com/stretchr/testify/require"
)

type capturedRequest struct {
	method string
	path   string
	auth   string
	body   map[string]any
}

type fixture struct {
	server   *httptest.Server
	adapter  *payment.GenericPaymentAdapter
	mu       sync.Mutex
	requests []capturedRequest
}

func newFixture(t *testing.T, holdReply map[string]any, status int) *fixture {
	t.Helper()

	f := &fixture{}

	handler := func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		_ = json.Unmarshal(raw, &parsed)

		f.mu.Lock()
		f.requests = append(f.requests, capturedRequest{
			method: r.Method,
			path:   r.URL.Path,
			auth:   r.Header.Get("Authorization"),
			body:   parsed,
		})
		f.mu.Unlock()

		if status != 0 {
			w.WriteHeader(status)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(holdReply)
	}

	f.server = httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(f.server.Close)

	cfg := config.Payment{
		BaseURL:      f.server.URL,
		APIKey:       "Bearer tok",
		AuthHeader:   "Authorization",
		Timeout:      5 * time.Second,
		HoldPath:     "/v1/holds",
		ReleasePath:  "/v1/holds/release",
		CapturePath:  "/v1/holds/capture",
		ForfeitPath:  "/v1/holds/forfeit",
	}
	f.adapter = payment.NewGenericPaymentAdapter(cfg, logger.New(config.Config{Log: config.Log{LogLevel: "info"}}))

	return f
}

func (f *fixture) last() capturedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.requests) == 0 {
		return capturedRequest{}
	}

	return f.requests[len(f.requests)-1]
}

func TestGenericPaymentAdapter_Hold_ReturnsProviderReference(t *testing.T) {
	f := newFixture(t, map[string]any{"external_reference": "ext_abc"}, 0)

	ref, err := f.adapter.Hold(context.Background(), 42, model.NewMoneyModel(1000), "CNY", "ref-1")

	require.NoError(t, err)
	require.Equal(t, "ext_abc", ref)

	last := f.last()
	require.Equal(t, http.MethodPost, last.method)
	require.Equal(t, "/v1/holds", last.path)
	require.Equal(t, "Bearer tok", last.auth)
	require.Equal(t, "ref-1", last.body["reference"])
	require.EqualValues(t, 1000, last.body["amount_in_cents"])
	require.Equal(t, "CNY", last.body["currency"])
}

func TestGenericPaymentAdapter_Hold_FallsBackToEchoedReference(t *testing.T) {
	f := newFixture(t, map[string]any{}, 0)

	ref, err := f.adapter.Hold(context.Background(), 42, model.NewMoneyModel(1000), "CNY", "ref-echo")

	require.NoError(t, err)
	require.Equal(t, "ref-echo", ref)
}

func TestGenericPaymentAdapter_Release(t *testing.T) {
	f := newFixture(t, map[string]any{}, 0)

	err := f.adapter.Release(context.Background(), "ext_abc")

	require.NoError(t, err)

	last := f.last()
	require.Equal(t, http.MethodPost, last.method)
	require.Equal(t, "/v1/holds/release", last.path)
	require.Equal(t, "ext_abc", last.body["external_reference"])
}

func TestGenericPaymentAdapter_Capture(t *testing.T) {
	f := newFixture(t, map[string]any{}, 0)

	err := f.adapter.Capture(context.Background(), "ext_abc", model.NewMoneyModel(750))

	require.NoError(t, err)

	last := f.last()
	require.Equal(t, "/v1/holds/capture", last.path)
	require.Equal(t, "ext_abc", last.body["external_reference"])
	require.EqualValues(t, 750, last.body["amount_in_cents"])
}

func TestGenericPaymentAdapter_Forfeit(t *testing.T) {
	f := newFixture(t, map[string]any{}, 0)

	err := f.adapter.Forfeit(context.Background(), "ext_abc")

	require.NoError(t, err)
	require.Equal(t, "/v1/holds/forfeit", f.last().path)
}

func TestGenericPaymentAdapter_NonSuccessStatus_ReturnsError(t *testing.T) {
	f := newFixture(t, nil, http.StatusInternalServerError)

	_, err := f.adapter.Hold(context.Background(), 42, model.NewMoneyModel(1000), "CNY", "ref-1")

	require.Error(t, err)
}

func TestNewPaymentPort_DefaultsToMock(t *testing.T) {
	port := payment.NewPaymentPort(logger.New(config.Config{Log: config.Log{LogLevel: "info"}}))

	require.NotNil(t, port)

	ref, err := port.Hold(context.Background(), 1, model.NewMoneyModel(100), "CNY", "ref-mock")
	require.NoError(t, err)
	require.Equal(t, "ref-mock", ref)

	var _ ports.PaymentPort = port
}
