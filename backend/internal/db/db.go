package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// DB adalah wrapper di atas *sql.DB dengan connection pool yang terkonfigurasi.
type DB struct {
	pool   *sql.DB
	logger *slog.Logger
}

type Tx struct{ tx *sql.Tx }

// New membuka koneksi ke PostgreSQL sesuai DSN dan mengonfigurasi connection pool.
func New(dsn string, logger *slog.Logger) (*DB, error) {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(nil, nil))
	}
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse konfigurasi PostgreSQL: %w", err)
	}
	sqlDB := stdlib.OpenDB(*config)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	logger.Info("koneksi database terbuka")
	return &DB{pool: sqlDB, logger: logger}, nil
}

func Rebind(query string) string {
	var out strings.Builder
	out.Grow(len(query) + 8)
	var quote byte
	lineComment, blockComment := false, false
	placeholder := 0
	for i := 0; i < len(query); i++ {
		c := query[i]
		if lineComment {
			out.WriteByte(c)
			if c == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			out.WriteByte(c)
			if c == '*' && i+1 < len(query) && query[i+1] == '/' {
				out.WriteByte('/')
				i++
				blockComment = false
			}
			continue
		}
		if quote != 0 {
			out.WriteByte(c)
			if c == quote {
				if i+1 < len(query) && query[i+1] == quote {
					out.WriteByte(query[i+1])
					i++
				} else {
					quote = 0
				}
			}
			continue
		}
		if c == '\'' || c == '"' || c == '`' {
			quote = c
			out.WriteByte(c)
			continue
		}
		if c == '-' && i+1 < len(query) && query[i+1] == '-' {
			lineComment = true
			out.WriteByte(c)
			out.WriteByte(query[i+1])
			i++
			continue
		}
		if c == '/' && i+1 < len(query) && query[i+1] == '*' {
			blockComment = true
			out.WriteByte(c)
			out.WriteByte(query[i+1])
			i++
			continue
		}
		if c == '?' {
			placeholder++
			fmt.Fprintf(&out, "$%d", placeholder)
			continue
		}
		out.WriteByte(c)
	}
	return out.String()
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.pool.ExecContext(ctx, Rebind(query), args...)
}
func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.pool.QueryContext(ctx, Rebind(query), args...)
}
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.pool.QueryRowContext(ctx, Rebind(query), args...)
}
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}
func (t *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, Rebind(query), args...)
}
func (t *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, Rebind(query), args...)
}
func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, Rebind(query), args...)
}
func (t *Tx) Commit() error   { return t.tx.Commit() }
func (t *Tx) Rollback() error { return t.tx.Rollback() }

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func InsertID(ctx context.Context, exec queryRower, query string, args ...any) (int64, error) {
	var id int64
	err := exec.QueryRowContext(ctx, Rebind(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(query), ";"))+" RETURNING id"), args...).Scan(&id)
	return id, err
}

// Health memeriksa apakah database masih responsif.
func (d *DB) Health(ctx context.Context) error {
	return d.pool.PingContext(ctx)
}

func (d *DB) Close() error { return d.pool.Close() }
