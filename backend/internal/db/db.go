package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB adalah wrapper di atas *sql.DB dengan connection pool yang terkonfigurasi.
type DB struct {
	*sql.DB
	logger *slog.Logger
}

// New membuka koneksi ke MySQL sesuai DSN dan mengonfigurasi connection pool.
func New(dsn string, logger *slog.Logger) (*DB, error) {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(nil, nil))
	}
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("buka koneksi database: %w", err)
	}
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
	return &DB{sqlDB, logger}, nil
}

// Health memeriksa apakah database masih responsif.
func (d *DB) Health(ctx context.Context) error {
	return d.PingContext(ctx)
}
