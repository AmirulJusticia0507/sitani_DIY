package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

// CLI migrasi skema: menerapkan berkas .sql di direktori migrations secara
// berurutan ke MySQL yang dituju DB_DSN. Contoh:
//
//	DB_DSN=user:pass@tcp(127.0.0.1:3306)/sitani?parseTime=true go run ./cmd/migrate
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DB_DSN")
	dir := flag.String("dir", "migrations", "direktori berkas migrasi .sql")
	dryRun := flag.Bool("dry-run", false, "hanya daftar migrasi yang belum diterapkan, tanpa menjalankan")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "perintah tidak dikenal, gunakan flag --help")
		os.Exit(2)
	}
	if dsn == "" {
		logger.Error("DB_DSN tidak diatur")
		os.Exit(1)
	}

	database, err := db.New(dsn, logger)
	if err != nil {
		logger.Error("gagal membuka database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	migrator, err := db.NewMigrator(database.DB, *dir)
	if err != nil {
		logger.Error("gagal membuat migrator", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if *dryRun {
		pending, listErr := migrator.Pending(ctx)
		if listErr != nil {
			logger.Error("gagal membaca daftar migrasi", "error", listErr)
			os.Exit(1)
		}
		logger.Info("migrasi tertunda", "jumlah", len(pending), "berkas", pending)
		return
	}

	applied, err := migrator.Up(ctx)
	if err != nil {
		logger.Error("migrasi gagal", "error", err, "diterapkan", applied)
		os.Exit(1)
	}
	logger.Info("migrasi selesai", "baru diterapkan", applied, "jumlah", len(applied))
}
