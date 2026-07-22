package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RED metrics for HTTP traffic. These reuse the already-present
// prometheus/client_golang dependency (no new dependencies added).
var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests handled, labelled by method, route and status.",
	}, []string{"method", "route", "status"})

	httpRequestErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_errors_total",
		Help: "Total number of HTTP requests resulting in a server error (>= 500) or that were rate-limited (429).",
	}, []string{"method", "route"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Latency of HTTP requests in seconds, labelled by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

// statusWriter captures the HTTP status code written by the handler so the
// Metrics middleware can record it without altering the response.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// Metrics records the RED metrics (request count, error count, duration) for
// every request. It is placed after routing so the matched route pattern is
// available as a stable label.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = routeLabel(r)
		}

		dur := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(sw.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route).Observe(dur)
		if sw.status >= http.StatusInternalServerError {
			httpRequestErrorsTotal.WithLabelValues(r.Method, route).Inc()
		}
	})
}
