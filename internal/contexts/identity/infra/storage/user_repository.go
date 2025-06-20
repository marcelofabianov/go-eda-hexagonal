package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	"github.com/marcelofabianov/redtogreen/internal/platform/port/database"
)

type UserRepository struct {
	db database.DB
}

func NewUserRepository(db database.DB) user.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UserExists(ctx context.Context, input user.UserExistsRepoInput) (bool, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR phone = $2)`

	var exists bool
	err := r.db.Conn().QueryRowContext(queryCtx, query, input.Email, input.Phone).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, input user.CreateUserRepoInput) error {
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (id, name, email, phone, password, preferences, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	u := input.User

	_, err := r.db.Conn().ExecContext(
		queryCtx,
		query,
		u.ID,
		u.Name,
		u.Email,
		u.Phone,
		u.Password,
		u.Preferences,
		u.CreatedAt,
		u.UpdatedAt,
		u.Version,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.ErrCodeUniqueViolation {
			return msg.NewMessageError(err, user.ErrUserAlreadyExists, msg.CodeConflict, nil)
		}
		return err
	}

	return nil
}
