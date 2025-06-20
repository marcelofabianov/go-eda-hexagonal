package otel

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/marcelofabianov/redtogreen/internal/platform/config"
)

func InitTracerProvider(cfg config.OtelConfig, logger *slog.Logger) (func(context.Context) error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(cfg.ExporterEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("Failed to create gRPC client for OTLP collector", "error", err, "endpoint", cfg.ExporterEndpoint)
		return nil, fmt.Errorf("failed to create gRPC client for OTLP collector: %w", err)
	}

	otlpClient := otlptracegrpc.NewClient(otlptracegrpc.WithGRPCConn(conn))

	traceExporter, err := otlptrace.New(ctx, otlpClient)
	if err != nil {
		logger.Error("Failed to create OTLP trace exporter", "error", err)
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.ServiceName),
		semconv.ServiceVersionKey.String(cfg.ServiceVersion),
	)

	bsp := tracesdk.NewBatchSpanProcessor(traceExporter)
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(resource),
		tracesdk.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OpenTelemetry TracerProvider initialized", "service_name", cfg.ServiceName, "endpoint", cfg.ExporterEndpoint)

	return func(ctx context.Context) error {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
		defer shutdownCancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down tracer provider", "error", err)
		} else {
			logger.Info("OpenTelemetry TracerProvider shut down successfully")
		}

		if err := conn.Close(); err != nil {
			logger.Error("Error closing OTLP gRPC client connection", "error", err)
			return fmt.Errorf("error closing OTLP gRPC client connection: %w", err)
		}

		return nil
	}, nil
}
