package httpserver

import (
	"time"

	"auction/internal/shared/modules/config"
	"auction/pkg/httpserver"
)

func New(cfg config.Config) (*httpserver.Server, error) {
	return httpserver.New(httpserver.Config{
		Host:            cfg.HTTPServer.Host,
		Port:            cfg.HTTPServer.Port,
		ReadTimeout:     time.Duration(cfg.HTTPServer.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.HTTPServer.WriteTimeout) * time.Second,
		IdleTimeout:     time.Duration(cfg.HTTPServer.IdleTimeout) * time.Second,
		ShutdownTimeout: time.Duration(cfg.HTTPServer.ShutdownTimeout) * time.Second,
		CORS: httpserver.CORSConfig{
			AllowedOrigins:     cfg.CORS.GetAllowedOrigins(),
			AllowedMethods:     cfg.CORS.GetAllowedMethods(),
			AllowedHeaders:     cfg.CORS.GetAllowedHeaders(),
			ExposedHeaders:     cfg.CORS.GetExposedHeaders(),
			AllowCredentials:   cfg.CORS.AllowCredentials,
			MaxAge:             cfg.CORS.MaxAge,
			OptionsPassthrough: cfg.CORS.OptionsPassthrough,
			Debug:              cfg.CORS.Debug,
		},
	})
}
