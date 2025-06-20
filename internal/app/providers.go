package app

import (
	"context"
	"log/slog"

	"go.uber.org/dig"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/bus"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/database"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/hasher"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/logger"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/otel"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/validator"
	"github.com/marcelofabianov/redtogreen/internal/platform/config"
	platformBus "github.com/marcelofabianov/redtogreen/internal/platform/port/bus"
	platformDB "github.com/marcelofabianov/redtogreen/internal/platform/port/database"
	platformHasher "github.com/marcelofabianov/redtogreen/internal/platform/port/hasher"
)

func providePlatformDependencies(container *dig.Container) error {
	if err := provideConfig(container); err != nil {
		return err
	}
	if err := provideAdapters(container); err != nil {
		return err
	}
	if err := provideEventBus(container); err != nil {
		return err
	}
	if err := provideOtel(container); err != nil {
		return err
	}
	return nil
}

func provideConfig(container *dig.Container) error {
	if err := container.Provide(config.LoadConfig); err != nil {
		return err
	}

	if err := container.Provide(func(cfg *config.AppConfig) config.DatabaseConfig {
		return cfg.Database
	}, dig.Name("mainDBConfig")); err != nil {
		return err
	}

	if err := container.Provide(func(cfg *config.AppConfig) config.DatabaseConfig {
		return cfg.AuditDatabase
	}, dig.Name("auditDBConfig")); err != nil {
		return err
	}

	if err := container.Provide(func(cfg *config.AppConfig) config.LoggerConfig { return cfg.Logger }); err != nil {
		return err
	}
	if err := container.Provide(func(cfg *config.AppConfig) config.NATSConfig { return cfg.NATS }); err != nil {
		return err
	}
	if err := container.Provide(func(cfg *config.AppConfig) config.OtelConfig { return cfg.Otel }); err != nil {
		return err
	}
	return nil
}

func provideAdapters(container *dig.Container) error {
	if err := container.Provide(logger.NewSlogLogger); err != nil {
		return err
	}
	if err := container.Provide(hasher.NewHasher); err != nil {
		return err
	}
	if err := container.Provide(func(h *hasher.Hasher) platformHasher.Hasher { return h }); err != nil {
		return err
	}
	if err := container.Provide(validator.New); err != nil {
		return err
	}

	type mainDBParams struct {
		dig.In
		Config config.DatabaseConfig `name:"mainDBConfig"`
		Logger *slog.Logger
	}
	if err := container.Provide(func(p mainDBParams) (platformDB.DB, error) {
		db, err := database.Connect(p.Config, p.Logger)
		return db, err
	}, dig.Name("mainDB")); err != nil {
		return err
	}

	type auditDBParams struct {
		dig.In
		Config config.DatabaseConfig `name:"auditDBConfig"`
		Logger *slog.Logger
	}
	if err := container.Provide(func(p auditDBParams) (platformDB.DB, error) {
		db, err := database.Connect(p.Config, p.Logger)
		return db, err
	}, dig.Name("auditDB")); err != nil {
		return err
	}

	return nil
}

func provideEventBus(container *dig.Container) error {
	if err := container.Provide(func(cfg config.NATSConfig, logger *slog.Logger) (*bus.NatsEventBus, error) {
		return bus.NewNatsEventBus(&cfg, logger)
	}); err != nil {
		return err
	}
	if err := container.Provide(func(b *bus.NatsEventBus) platformBus.EventBus { return b }); err != nil {
		return err
	}
	if err := container.Provide(func(b platformBus.EventBus) platformBus.EventBusPublisher { return b }); err != nil {
		return err
	}
	if err := container.Provide(func(b platformBus.EventBus) platformBus.EventBusSubscriber { return b }); err != nil {
		return err
	}
	return nil
}

func provideOtel(container *dig.Container) error {
	if err := container.Provide(func(cfg config.OtelConfig, logger *slog.Logger) (func(context.Context) error, error) {
		return otel.InitTracerProvider(cfg, logger)
	}); err != nil {
		return err
	}
	return nil
}
