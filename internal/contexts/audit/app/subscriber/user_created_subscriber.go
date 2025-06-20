package subscriber

import (
	"context"
	"encoding/json"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/domain/audit"
	identityUser "github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/logger"
	"github.com/marcelofabianov/redtogreen/internal/platform/event"
)

type UserCreatedSubscriber struct {
	repo   audit.RegisterAuditLogRepository
	logger *slog.Logger
	tracer trace.Tracer
}

func NewUserCreatedSubscriber(repo audit.RegisterAuditLogRepository, logger *slog.Logger) *UserCreatedSubscriber {
	return &UserCreatedSubscriber{
		repo:   repo,
		logger: logger,
		tracer: otel.Tracer("audit-subscriber"),
	}
}

func (s *UserCreatedSubscriber) Handle(ctx context.Context, e *event.Event) error {
	ctx, span := s.tracer.Start(ctx, "UserCreatedSubscriber.Handle",
		trace.WithAttributes(
			attribute.String("event.type", string(e.Header.EventType)),
			attribute.String("event.id", e.Header.EventID.String()),
			attribute.String("audit.component", "subscriber"),
		),
	)
	defer span.End()

	loggerWithTrace := logger.WithContext(ctx, s.logger).With(
		logger.Component("audit_subscriber"),
		logger.Action("handle_user_created_event"),
		logger.EventID(e.Header.EventID),
		logger.EventType(string(e.Header.EventType)),
		logger.TraceID(e.Metadata.TraceID.String()),
		logger.UserID(e.Context.UserID),
	)

	var payload identityUser.UserCreatedPayload

	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to unmarshal event payload")
		loggerWithTrace.Error("failed to unmarshal user.created event payload for auditing",
			logger.Err(err),
			slog.String("payload_content", string(e.Payload)),
		)
		return err
	}

	if userID, ok := e.Context.UserID.GetUUID(); ok {
		span.SetAttributes(attribute.String("user.id", userID.String()))
	}

	input := audit.NewAuditLogInput{
		EventID:      e.Header.EventID,
		EventType:    e.Header.EventType,
		EventVersion: e.Header.SchemaVersion,
		EventContext: e.Context,
		TraceID:      e.Metadata.TraceID,
		UserAuthorID: e.Context.UserID,
		Payload:      e.Payload,
	}

	auditLog, err := audit.NewAuditLog(input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create audit log entity")
		loggerWithTrace.Error("failed to create audit log entity from event",
			logger.Err(err),
		)
		return err
	}

	if err := s.repo.RegisterAuditLog(ctx, audit.RegisterAuditLogRepoInput{AuditLog: auditLog}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to save audit log to repository")
		loggerWithTrace.Error("failed to save audit log to repository",
			logger.Err(err),
			logger.EventID(auditLog.EventID),
		)
		return err
	}

	span.SetStatus(codes.Ok, "Audit log saved successfully")
	loggerWithTrace.Info("user.created event audited successfully",
		logger.EventID(auditLog.EventID),
		logger.UserID(auditLog.UserAuthorID),
	)

	return nil
}
