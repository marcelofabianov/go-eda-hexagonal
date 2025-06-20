package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/domain/audit"
	pDB "github.com/marcelofabianov/redtogreen/internal/platform/port/database"
)

type PostgresAuditRepository struct {
	db pDB.DB
}

func NewPostgresAuditRepository(db pDB.DB) audit.RegisterAuditLogRepository {
	return &PostgresAuditRepository{db: db}
}

func (r *PostgresAuditRepository) RegisterAuditLog(ctx context.Context, input audit.RegisterAuditLogRepoInput) error {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO audit_logs (id, event_id, event_type, event_version, event_context, trace_id, user_author_id, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	auditLog := input.AuditLog

	_, err := r.db.Conn().ExecContext(
		queryCtx,
		query,
		auditLog.ID,
		auditLog.EventID,
		auditLog.EventType,
		auditLog.EventVersion,
		auditLog.EventContext,
		auditLog.TraceID,
		auditLog.UserAuthorID,
		auditLog.Payload,
		auditLog.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	return nil
}
