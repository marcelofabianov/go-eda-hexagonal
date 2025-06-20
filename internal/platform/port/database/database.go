package database

import (
	"context"
	"database/sql"
)

const (
	ErrCodeUniqueViolation = "23505"
)

type DB interface {
	Conn() *sql.DB
	WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(tx *sql.Tx) error) error
	Close() error
}
