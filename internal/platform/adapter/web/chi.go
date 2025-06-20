package web

import (
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/marcelofabianov/redtogreen/internal/platform/config"
)

const MaxBodySize = 1048576 // 1MB

func NewPlatformRouter(cfg *config.AppConfig, logger *slog.Logger) *chi.Mux {
	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(SlogLoggerMiddleware(logger))
	r.Use(middleware.Heartbeat("/ping"))

	r.Use(middleware.RequestSize(MaxBodySize))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(cors.Handler(setCorsOptions(cfg.Auth.Cors)))

	r.Use(httprate.Limit(
		cfg.Server.RateLimit,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithResponseHeaders(headersRateLimit()),
	))

	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	r.Use(middleware.SetHeader("X-Frame-Options", "deny"))
	r.Use(middleware.SetHeader("Content-Security-Policy", "default-src 'self'; script-src 'self'"))
	r.Use(middleware.SetHeader("Cache-Control", "no-store, no-cache"))

	return r
}

func headersRateLimit() httprate.ResponseHeaders {
	return httprate.ResponseHeaders{
		Limit:     "X-RateLimit-Limit",
		Remaining: "X-RateLimit-Remaining",
		Reset:     "X-RateLimit-Reset",
	}
}

func setCorsOptions(cfg config.CorsConfig) cors.Options {
	return cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
	}
}
