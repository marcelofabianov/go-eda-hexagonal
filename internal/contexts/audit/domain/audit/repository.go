package audit

import "context"

type RegisterAuditLogRepoInput struct {
	AuditLog *AuditLog
}

type RegisterAuditLogRepository interface {
	RegisterAuditLog(ctx context.Context, input RegisterAuditLogRepoInput) error
}
