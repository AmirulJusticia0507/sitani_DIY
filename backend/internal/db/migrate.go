package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrator menerapkan berkas migrasi SQL terversi ke MySQL.
//
// Berkas diproses berurutan sesuai nama berkas, setiap berkas lewat satu
// transaksi, dan catatannya disimpan di tabel schema_migrations supaya
// migrasi yang sudah diterapkan tidak diproses ulang.
type Migrator struct {
	db  *sql.DB
	dir string
}

// NewMigrator membuat Migrator. dir berisi berkas .sql berurutan.
func NewMigrator(database *sql.DB, dir string) (*Migrator, error) {
	if database == nil {
		return nil, fmt.Errorf("database belum tersedia")
	}
	if dir == "" {
		dir = "migrations"
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("direktori migrasi %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("direktori migrasi %q bukan direktori", dir)
	}
	return &Migrator{db: database, dir: dir}, nil
}

// migration memodelkan satu berkas .sql.
type migration struct {
	name string
	sql  string
}

// listMigrations mengembalikan migrasi berurut nama berkas.
func (m *Migrator) listMigrations() ([]migration, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("baca direktori migrasi: %w", err)
	}
	var list []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		path := filepath.Join(m.dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("baca migrasi %s: %w", entry.Name(), err)
		}
		list = append(list, migration{name: entry.Name(), sql: string(content)})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].name < list[j].name })
	return list, nil
}

// ensureTable membuat tabel pencatat migrasi bila belum ada.
func (m *Migrator) ensureTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`)
	if err != nil {
		return fmt.Errorf("siapkan tabel schema_migrations: %w", err)
	}
	return nil
}

// apply terApplied memeriksa apakah migrasi sudah diterapkan.
func (m *Migrator) applied(ctx context.Context, version string) (bool, error) {
	var exists bool
	err := m.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)`, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("periksa status migrasi %s: %w", version, err)
	}
	return exists, nil
}

// Pending mengembalikan daftar versi migrasi yang belum diterapkan,
// dipakai pemeriksaan tanpa menjalankan apa pun.
func (m *Migrator) Pending(ctx context.Context) ([]string, error) {
	files, err := m.listMigrations()
	if err != nil {
		return nil, err
	}
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	var pending []string
	for _, mig := range files {
		already, err := m.applied(ctx, mig.name)
		if err != nil {
			return pending, err
		}
		if !already {
			pending = append(pending, mig.name)
		}
	}
	return pending, nil
}

// Up menerapkan seluruh migrasi yang belum dijalankan. Mengembalikan daftar
// versi yang baru diterapkan. Berhenti pada migrasi pertama yang gagal; sitani
// tidak meneruskan migrasi berikutnya supaya skema tidak setengah jadi
// tersembunyi di belakang tabel status.
func (m *Migrator) Up(ctx context.Context) ([]string, error) {
	files, err := m.listMigrations()
	if err != nil {
		return nil, err
	}
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	var appliedNow []string
	for _, mig := range files {
		already, err := m.applied(ctx, mig.name)
		if err != nil {
			return appliedNow, err
		}
		if already {
			continue
		}
		if err := m.apply(ctx, mig); err != nil {
			return appliedNow, fmt.Errorf("migrasi %s gagal: %w", mig.name, err)
		}
		appliedNow = append(appliedNow, mig.name)
	}
	return appliedNow, nil
}

// apply menjalankan satu berkas migrasi dalam satu transaksi.
func (m *Migrator) apply(ctx context.Context, mig migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("buka transaksi: %w", err)
	}
	rollback := true
	defer func() {
		if rollback {
			// Setelah Commit', Rollback' aman (ErrTxDone diabaikan eksplisit).
			_ = tx.Rollback()
		}
	}()

	statements, err := splitStatements(mig.sql)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("parsing pernyataan: %w", err)
	}
	for i, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("pernyataan %d: %w", i+1, err)
		}
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version) VALUES (?)`, mig.name); err != nil {
		return fmt.Errorf("catat versi migrasi: %w", err)
	}
	rollback = false
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("komit transaksi: %w", err)
	}
	return nil
}

// splitStatements memecah teks SQL menjadi pernyataan pada titik koma yang
// berada di luar string literal. Komentar baris (--) dan blok (/* */) dihapus.
// Untuk skema proyek ini tidak ada prosedur tersimpan, sehingga pemisahan
// sederhana cukup aman.
func splitStatements(sqlText string) ([]string, error) {
	var statements []string
	var current strings.Builder

	flush := func() {
		if trimmed := strings.TrimSpace(current.String()); trimmed != "" {
			statements = append(statements, trimmed)
		}
		current.Reset()
	}

	var (
		inQuote     rune
		inLineNote  bool
		inBlockNote bool
	)
	runes := []rune(sqlText)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case inLineNote:
			if r == '\n' {
				inLineNote = false
				current.WriteRune(r)
			}
		case inBlockNote:
			if r == '*' && i+1 < len(runes) && runes[i+1] == '/' {
				inBlockNote = false
				i++
			}
		case inQuote != 0:
			current.WriteRune(r)
			if r == inQuote {
				// Dukungan escape '' di dalam string literal.
				if i+1 < len(runes) && runes[i+1] == inQuote {
					current.WriteRune(inQuote)
					i++
					continue
				}
				inQuote = 0
			}
		case r == '\'' || r == '"':
			inQuote = r
			current.WriteRune(r)
		case r == '-' && i+1 < len(runes) && runes[i+1] == '-':
			inLineNote = true
			i++
		case r == '/' && i+1 < len(runes) && runes[i+1] == '*':
			inBlockNote = true
			i++
		case r == ';':
			flush()
		default:
			current.WriteRune(r)
		}
	}
	if inQuote != 0 {
		return nil, fmt.Errorf("string literal tidak ditutup dalam berkas SQL")
	}
	flush()
	return statements, nil
}
