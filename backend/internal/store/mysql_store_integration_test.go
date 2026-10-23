package store

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

// Tes integrasi penomoran permohonan terhadap PostgreSQL sungguhan. Dilewati
// bila SITANI_TEST_DATABASE_URL tidak diatur.
func TestCreatePermohonanMenomoriUnikPerHari(t *testing.T) {
	dsn := os.Getenv("SITANI_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("SITANI_TEST_DATABASE_URL tidak diatur")
	}
	database, err := db.New(dsn, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("buka database: %v", err)
	}
	defer database.Close()
	ctx := context.Background()
	if _, err := database.ExecContext(ctx, `TRUNCATE permohonan_benih RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("bersihkan tabel: %v", err)
	}

	// Pelanggan wajib ada karena foreign key.
	var pelangganID int64
	if err := database.QueryRowContext(ctx, `INSERT INTO users (username,email,password_hash,role) VALUES ('uji','uji@contoh.id','x','pelanggan') RETURNING id`).Scan(&pelangganID); err != nil {
		t.Fatalf("buat pelanggan: %v", err)
	}

	s := NewDBStore(database)
	hari := time.Date(2026, 10, 23, 9, 15, 0, 0, time.UTC)
	hariBerikut := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)

	harapan := []string{"SPB-20261023-0001", "SPB-20261023-0002", "SPB-20261024-0001"}
	for i, ingin := range harapan {
		tanggal := hari
		if ingin[:12] == "SPB-20261024" {
			tanggal = hariBerikut
		}
		id, err := s.CreatePermohonan(ctx, db.PermohonanBenih{
			PelangganID:          pelangganID,
			JenisBenih:           "Tomat",
			Status:               "diajukan",
			PembayaranAwalStatus: "belum",
		}, tanggal)
		if err != nil {
			t.Fatalf("permohonan %d: %v", i, err)
		}
		tersimpan, err := s.GetPermohonan(ctx, id)
		if err != nil {
			t.Fatalf("baca permohonan %d: %v", i, err)
		}
		if tersimpan.NomorPermohonan != ingin {
			t.Fatalf("permohonan %d nomor = %q, ingin %q", i, tersimpan.NomorPermohonan, ingin)
		}
	}

	var sisaSementara int
	if err := database.QueryRowContext(ctx, `SELECT COUNT(*) FROM permohonan_benih WHERE nomor_permohonan LIKE 'TMP-%'`).Scan(&sisaSementara); err != nil {
		t.Fatalf("hitung nomor sementara: %v", err)
	}
	if sisaSementara != 0 {
		t.Fatalf("masih ada %d nomor sementara yang tersimpan", sisaSementara)
	}
}
