package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	otelchi "github.com/riandyrn/otelchi"
	"go.uber.org/dig"

	auditContainer "github.com/marcelofabianov/redtogreen/internal/contexts/audit/container"
	identityContainer "github.com/marcelofabianov/redtogreen/internal/contexts/identity/container"
	identityHttp "github.com/marcelofabianov/redtogreen/internal/contexts/identity/infra/http"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/web"
	"github.com/marcelofabianov/redtogreen/internal/platform/config"
)

type App struct {
	container      *dig.Container
	config         *config.AppConfig
	logger         *slog.Logger
	otelShutdownFn func(context.Context) error
}

func New() (*App, error) {
	container := dig.New()

	if err := providePlatformDependencies(container); err != nil {
		return nil, fmt.Errorf("failed to provide platform dependencies: %w", err)
	}

	if err := identityContainer.Register(container); err != nil {
		return nil, fmt.Errorf("failed to register identity context: %w", err)
	}
	if err := auditContainer.Register(container); err != nil {
		return nil, fmt.Errorf("failed to register audit context: %w", err)
	}

	if err := setupEventSubscriptions(container); err != nil {
		return nil, fmt.Errorf("failed to setup event subscriptions: %w", err)
	}

	app := &App{container: container}
	if err := container.Invoke(func(
		cfg *config.AppConfig,
		log *slog.Logger,
		otelShutdown func(context.Context) error,
	) {
		app.config = cfg
		app.logger = log
		app.otelShutdownFn = otelShutdown
	}); err != nil {
		return nil, fmt.Errorf("failed to invoke app dependencies: %w", err)
	}

	app.logger.Info("application container built successfully")

	return app, nil
}

func (a *App) Run() error {
	mainRouter := web.NewPlatformRouter(a.config, a.logger)
	mainRouter.Use(otelchi.Middleware(a.config.Otel.ServiceName))

	mainRouter.Get("/", DefaultHandler)

	if err := a.container.Invoke(func(identityRouter *identityHttp.Router) {
		mainRouter.Route("/api/v1", func(r chi.Router) {
			r.Mount("/identity", identityRouter)
		})
	}); err != nil {
		return fmt.Errorf("failed to mount identity router: %w", err)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port),
		Handler:      mainRouter,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		a.logger.Info("server is starting", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-stopChan

	a.logger.Info("shutting down server gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.otelShutdownFn(ctx); err != nil {
		a.logger.Error("OpenTelemetry shutdown failed", "error", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	a.logger.Info("server stopped gracefully")
	return nil
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": ".",
		"status":  "ok",
	}
	web.Respond(w, r, http.StatusOK, response)
}
