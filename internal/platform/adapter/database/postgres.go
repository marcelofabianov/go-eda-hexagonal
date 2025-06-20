package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	otelpgx "github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/marcelofabianov/redtogreen/internal/platform/config"
	platformDB "github.com/marcelofabianov/redtogreen/internal/platform/port/database"
)

const (
	ErrFailedToOpenConnection      = "failed to open database connection: %w"
	ErrDatabasePingFailed          = "database ping failed: %w"
	ErrCouldNotBeginTransaction    = "could not begin transaction: %w"
	ErrFailedToRollbackTransaction = "failed to rollback transaction (original error: %v): %w"
	logClosingConnection           = "closing database connection pool"
	logPanicRecovered              = "panic recovered during transaction, rolling back"
	logRollbackAfterPanicFailed    = "failed to rollback transaction after panic"
	logRollbackFailed              = "failed to rollback transaction"
)

type Database struct {
	conn *sql.DB
	log  *slog.Logger
}

func New(db *sql.DB, logger *slog.Logger) *Database {
	return &Database{
		conn: db,
		log:  logger,
	}
}

var _ platformDB.DB = (*Database)(nil)

func Connect(cfg config.DatabaseConfig, logger *slog.Logger) (platformDB.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	connConfig, err := pgx.ParseConfig(connStr)
	if err != nil {
		logger.Error("Failed to parse pgx config for database connection", "error", err)
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	connConfig.Tracer = otelpgx.NewTracer()

	conn := stdlib.OpenDB(*connConfig)

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)
	conn.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		logger.Error("Database ping failed", "error", err, "db_host", cfg.Host, "db_name", cfg.Name)
		return nil, fmt.Errorf(ErrDatabasePingFailed, err)
	}

	logger.Info("Database connection established successfully", "db_host", cfg.Host, "db_name", cfg.Name)

	return New(conn, logger), nil
}

func (d *Database) Conn() *sql.DB {
	return d.conn
}

func (d *Database) Close() error {
	d.log.Info(logClosingConnection)
	return d.conn.Close()
}

func (d *Database) WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(tx *sql.Tx) error) error {
	tx, err := d.conn.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf(ErrCouldNotBeginTransaction, err)
	}

	defer func() {
		if p := recover(); p != nil {
			d.log.Error(logPanicRecovered, "panic", p)
			if rbErr := tx.Rollback(); rbErr != nil {
				d.log.Error(logRollbackAfterPanicFailed, "error", rbErr)
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			d.log.Error(logRollbackFailed, "original_error", err, "rollback_error", rbErr)
			return fmt.Errorf(ErrFailedToRollbackTransaction, err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
