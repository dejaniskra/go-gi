package gogi

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dejaniskra/go-gi/internal/config"
	_ "github.com/lib/pq"
)

const (
	pgDefaultMaxOpenConns    = 10
	pgDefaultMaxIdleConns    = 5
	pgDefaultConnMaxLifetime = 300 // seconds
)

type PostgresClient struct {
	Writer *sql.DB
	Reader *sql.DB
}

var postgresClients = make(map[string]*PostgresClient)

func GetPostgresClient(role string) (*PostgresClient, error) {
	if client, exists := postgresClients[role]; exists {
		return client, nil
	}

	cfg := config.GetConfig().Postgres[role]
	if cfg == nil {
		return nil, fmt.Errorf("no Postgres configuration found for role: %s", role)
	}

	client, err := newPostgresClient(cfg)
	if err != nil {
		return nil, err
	}

	postgresClients[role] = client
	return client, nil
}

func newPostgresClient(dbRoleConfig *config.DBRoleConfig) (*PostgresClient, error) {
	writer, err := newPgConnection(&dbRoleConfig.Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer connection: %w", err)
	}

	if dbRoleConfig.Reader == nil {
		return &PostgresClient{
			Writer: writer,
			Reader: writer,
		}, nil
	}

	reader, err := newPgConnection(dbRoleConfig.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader connection: %w", err)
	}

	return &PostgresClient{
		Writer: writer,
		Reader: reader,
	}, nil
}

func newPgConnection(cfg *config.DBConnection) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)
	if cfg.Options != nil {
		dsn += "?" + *cfg.Options
	} else {
		dsn += "?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns != nil {
		db.SetMaxOpenConns(*cfg.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(pgDefaultMaxOpenConns)
	}

	if cfg.MaxIdleConns != nil {
		db.SetMaxIdleConns(*cfg.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(pgDefaultMaxIdleConns)
	}

	if cfg.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(time.Duration(*cfg.ConnMaxLifetime) * time.Second)
	} else {
		db.SetConnMaxLifetime(pgDefaultConnMaxLifetime * time.Second)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	GetLogger().Debug(fmt.Sprintf("[Postgres] Connected to %s:%s/%s", cfg.Host, cfg.Port, cfg.DBName))
	return db, nil
}

func (c *PostgresClient) Ping(ctx context.Context) error {
	if err := c.Writer.PingContext(ctx); err != nil {
		return fmt.Errorf("writer DB ping failed: %w", err)
	}

	if c.Reader != c.Writer {
		if err := c.Reader.PingContext(ctx); err != nil {
			return fmt.Errorf("reader DB ping failed: %w", err)
		}
	}
	return nil
}

func (c *PostgresClient) Close() error {
	if err := c.Writer.Close(); err != nil {
		return err
	}
	if c.Reader != c.Writer {
		return c.Reader.Close()
	}
	return nil
}

func (c *PostgresClient) FindOne(ctx context.Context, query string, args []any, dest ...any) error {
	message := fmt.Sprintf("[Postgres] FindOne: %s | args=%v", query, args)
	GetLogger().Debug(message)
	row := c.Reader.QueryRowContext(ctx, query, args...)
	return row.Scan(dest...)
}

func (c *PostgresClient) FindMany(ctx context.Context, query string, args []any, scanFunc func(*sql.Rows) error) error {
	message := fmt.Sprintf("[Postgres] FindMany: %s | args=%v", query, args)
	GetLogger().Debug(message)
	rows, err := c.Reader.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanFunc(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (c *PostgresClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	message := fmt.Sprintf("[Postgres] Exec: %s | args=%v", query, args)
	GetLogger().Debug(message)

	return c.Writer.ExecContext(ctx, query, args...)
}

func (c *PostgresClient) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := c.Writer.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
