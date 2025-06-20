package gogi

import (
	"context"
	"database/sql"
	"fmt"

	"time"

	"github.com/dejaniskra/go-gi/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

const (
	mysqlDefaultMaxOpenConns    = 10
	mysqlDefaultMaxIdleConns    = 5
	mysqlDefaultConnMaxLifetime = 300 // seconds
)

type MySQLClient struct {
	Writer *sql.DB
	Reader *sql.DB
}

var mysqlClients = make(map[string]*MySQLClient)

func GetMySQLClient(role string) (*MySQLClient, error) {
	if client, exists := mysqlClients[role]; exists {
		return client, nil
	}

	cfg := config.GetConfig().MySQL[role]
	if cfg == nil {
		return nil, fmt.Errorf("no MySQL configuration found for role: %s", role)
	}

	client, err := newMySQLClient(cfg)
	if err != nil {
		return nil, err
	}

	mysqlClients[role] = client
	return client, nil
}

func newMySQLClient(dbRoleConfig *config.DBRoleConfig) (*MySQLClient, error) {
	writer, err := newDbConnection(&dbRoleConfig.Writer)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer connection: %w", err)
	}

	if dbRoleConfig.Reader == nil {
		return &MySQLClient{
			Writer: writer,
			Reader: writer,
		}, nil
	}

	reader, err := newDbConnection(dbRoleConfig.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader connection: %w", err)
	}

	return &MySQLClient{
		Writer: writer,
		Reader: reader,
	}, nil
}

func newDbConnection(cfg *config.DBConnection) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)
	if cfg.Options != nil {
		dsn += "?" + *cfg.Options
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns != nil {
		db.SetMaxOpenConns(*cfg.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(mysqlDefaultMaxOpenConns)
	}

	if cfg.MaxIdleConns != nil {
		db.SetMaxIdleConns(*cfg.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(mysqlDefaultMaxIdleConns)
	}

	if cfg.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(time.Duration(*cfg.ConnMaxLifetime) * time.Second)
	} else {
		db.SetConnMaxLifetime(mysqlDefaultConnMaxLifetime * time.Second)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	GetLogger().Debug(fmt.Sprintf("[MySQL] Connected to %s:%s/%s", cfg.Host, cfg.Port, cfg.DBName))
	return db, nil
}

func (c *MySQLClient) Ping(ctx context.Context) error {
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

func (c *MySQLClient) Close() error {
	if err := c.Writer.Close(); err != nil {
		return err
	}
	if c.Reader != c.Writer {
		return c.Reader.Close()
	}
	return nil
}

func (c *MySQLClient) FindOne(ctx context.Context, query string, args []any, dest ...any) error {
	message := fmt.Sprintf("[MySQL] FindOne: %s | args=%v", query, args)
	GetLogger().Debug(message)
	row := c.Reader.QueryRowContext(ctx, query, args...)
	return row.Scan(dest...)
}

func (c *MySQLClient) FindMany(ctx context.Context, query string, args []any, scanFunc func(*sql.Rows) error) error {
	message := fmt.Sprintf("[MySQL] FindMany: %s | args=%v", query, args)
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

func (c *MySQLClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	message := fmt.Sprintf("[MySQL] Exec: %s | args=%v", query, args)
	GetLogger().Debug(message)

	return c.Writer.ExecContext(ctx, query, args...)
}

func (c *MySQLClient) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
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
