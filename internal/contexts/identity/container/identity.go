package container

import (
	"go.uber.org/dig"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/app/command"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/app/publisher"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/app/usecase"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/infra/http"
	storage "github.com/marcelofabianov/redtogreen/internal/contexts/identity/infra/storage"
	pDB "github.com/marcelofabianov/redtogreen/internal/platform/port/database"
)

func Register(container *dig.Container) error {
	if err := registerApp(container); err != nil {
		return err
	}
	if err := registerInfra(container); err != nil {
		return err
	}
	return nil
}

func registerApp(container *dig.Container) error {
	if err := container.Provide(publisher.NewUserPublisher); err != nil {
		return err
	}
	if err := container.Provide(func(p *publisher.UserPublisher) user.UserPublisher { return p }); err != nil {
		return err
	}
	if err := container.Provide(func(p user.UserPublisher) user.UserCreatedEventPublisher { return p }); err != nil {
		return err
	}
	if err := container.Provide(usecase.NewCreateUserUseCase); err != nil {
		return err
	}
	if err := container.Provide(command.NewCreateUserCommand); err != nil {
		return err
	}
	return nil
}

func registerInfra(container *dig.Container) error {
	type userRepoParams struct {
		dig.In
		DB pDB.DB `name:"mainDB"`
	}
	if err := container.Provide(func(p userRepoParams) user.UserRepository {
		return storage.NewUserRepository(p.DB)
	}); err != nil {
		return err
	}

	if err := container.Provide(func(repo user.UserRepository) user.CreateUserRepository { return repo }); err != nil {
		return err
	}

	if err := container.Provide(http.NewCreateUserHandler); err != nil {
		return err
	}
	if err := container.Provide(http.NewIdentityRouter); err != nil {
		return err
	}
	return nil
}
