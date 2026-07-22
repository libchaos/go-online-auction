package httpserver

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a minimal, dependency-free, in-memory token-bucket limiter
// keyed by client IP. It is safe for concurrent use. It is intended for a
// single-process baseline; a multi-instance deployment would need a shared
// store (e.g. Redis), which is out of scope here.
type RateLimiter struct {
	mu         sync.Mutex
	clients    map[string]*tokenBucket
	refillRate float64 // tokens added per second
	burst      float64 // maximum tokens (capacity)
}

type tokenBucket struct {
	tokens float64
	last   time.Time
}

// NewRateLimiter builds a token-bucket limiter that refills at `limit` tokens
// per `window`, with an initial/peak capacity of `burst` tokens. Sensible
// defaults are applied for non-positive values.
func NewRateLimiter(limit int, burst int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if burst <= 0 {
		burst = 20
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		clients:    make(map[string]*tokenBucket),
		refillRate: float64(limit) / window.Seconds(),
		burst:      float64(burst),
	}
}

// Allow reports whether the client identified by `key` may make a request now,
// consuming one token if permitted.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.clients[key]
	if !ok {
		b = &tokenBucket{tokens: rl.burst - 1, last: now}
		rl.clients[key] = b
		rl.evictLocked(now)
		return true
	}

	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.refillRate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// evictLocked removes idle buckets to bound memory. Caller must hold rl.mu.
func (rl *RateLimiter) evictLocked(now time.Time) {
	if len(rl.clients) <= 4096 { //nolint:mnd // 4096 is a fixed in-memory bucket cap, not a tunable
		return
	}
	for k, b := range rl.clients {
		// Drop buckets that have been full (idle) for over a minute.
		if now.Sub(b.last) > time.Minute && b.tokens >= rl.burst {
			delete(rl.clients, k)
		}
	}
}

// RateLimit returns chi/http middleware that limits each client IP to
// `limit` requests per `window` with an initial burst of `burst`. Throttled
// requests are answered with 429 and recorded by the RED error counter.
func RateLimit(limit int, burst int, window time.Duration) func(http.Handler) http.Handler {
	rl := NewRateLimiter(limit, burst, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(clientIP(r)) {
				route := routeLabel(r)
				httpRequestErrorsTotal.WithLabelValues(r.Method, route).Inc()
				httpRequestsTotal.WithLabelValues(r.Method, route, "429").Inc()
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("rate limit exceeded"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP extracts the client IP from the direct peer address (r.RemoteAddr),
// stripping the port. We intentionally do not trust X-Forwarded-For so clients
// cannot spoof their rate-limit bucket behind a proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// routeLabel returns a coarse route label usable before routing has completed.
func routeLabel(r *http.Request) string {
	if r.URL != nil && r.URL.Path != "" {
		return r.URL.Path
	}
	return "unknown"
}
