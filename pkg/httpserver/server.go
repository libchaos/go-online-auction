package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

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
		return fmt.Errorf("host cannot be empty")
	}
	return nil
}
