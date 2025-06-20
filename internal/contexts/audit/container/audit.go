package container

import (
	"go.uber.org/dig"

	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/app/subscriber"
	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/domain/audit"
	"github.com/marcelofabianov/redtogreen/internal/contexts/audit/infra/storage"
	pDB "github.com/marcelofabianov/redtogreen/internal/platform/port/database"
)

func Register(container *dig.Container) error {
	type auditRepoParams struct {
		dig.In
		DB pDB.DB `name:"auditDB"`
	}
	if err := container.Provide(func(p auditRepoParams) audit.RegisterAuditLogRepository {
		return storage.NewPostgresAuditRepository(p.DB)
	}); err != nil {
		return err
	}

	if err := container.Provide(subscriber.NewUserCreatedSubscriber); err != nil {
		return err
	}

	return nil
}
