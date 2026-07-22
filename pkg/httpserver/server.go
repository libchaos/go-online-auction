package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
)

const (
	DefaultReadTimeout       = 15
	DefaultWriteTimeout      = 15
	DefaultIdleTimeout       = 60
	DefaultShutdownTimeout   = 10
	DefaultWSReadBufferSize  = 1024
	DefaultWSWriteBufferSize = 1024
)

// Server wraps an HTTP server with Chi router and websocket support.
type Server struct {
	server   *http.Server
	router   *chi.Mux
	upgrader *websocket.Upgrader
	config   Config
}

// New creates a new HTTP server with Chi router and websocket upgrader.
func New(cfg Config) (*Server, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	// Default middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Security: per-IP rate limiting (~60 req/min, burst 20) and RED metrics.
	// Rate limiting keys on r.RemoteAddr directly (no header-based RealIP) to
	// avoid client-spoofable X-Forwarded-For bypassing the limiter.
	router.Use(RateLimit(60, 20, time.Minute)) //nolint:mnd // rate-limit defaults are fixed operational constants
	// Observability: per-request trace span (telemetry module) and RED metrics.
	router.Use(Trace)
	router.Use(Metrics)

	// CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:     cfg.CORS.AllowedOrigins,
		AllowedMethods:     cfg.CORS.AllowedMethods,
		AllowedHeaders:     cfg.CORS.AllowedHeaders,
		ExposedHeaders:     cfg.CORS.ExposedHeaders,
		AllowCredentials:   cfg.CORS.AllowCredentials,
		MaxAge:             cfg.CORS.MaxAge,
		OptionsPassthrough: cfg.CORS.OptionsPassthrough,
		Debug:              cfg.CORS.Debug,
	}))

	// add a healthy check endpoint
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  DefaultWSReadBufferSize,
		WriteBufferSize: DefaultWSWriteBufferSize,
		CheckOrigin: func(_ *http.Request) bool {
			// Allow all connections by default; override via SetCheckOrigin
			return true
		},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		server:   srv,
		router:   router,
		upgrader: upgrader,
		config:   cfg,
	}, nil
}

// Router returns the Chi router for registering routes.
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Upgrader returns the websocket upgrader.
func (s *Server) Upgrader() *websocket.Upgrader {
	return s.upgrader
}

// SetCheckOrigin sets a custom origin checker for websocket connections.
func (s *Server) SetCheckOrigin(fn func(r *http.Request) bool) {
	s.upgrader.CheckOrigin = fn
}

// Start begins listening and serving HTTP requests.
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.server.Addr
}

func validateConfig(cfg Config) error {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.Port)
	}
	// validate host
	if strings.TrimSpace(cfg.Host) == "" {
		return errors.New("host cannot be empty")
	}
	return nil
}
