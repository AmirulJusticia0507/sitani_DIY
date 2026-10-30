package store

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

var _ = time.Time{}

var _ = time.Time{}

// bukaDatabaseUji membuka database PostgreSQL uji atau melewati tes bila
// SITANI_TEST_DATABASE_URL tidak diatur. Tabel permohonan_benih dikosongkan
// agar setiap tes berdiri sendiri.
func bukaDatabaseUji(t *testing.T) (*db.DB, *DBStore) {
	t.Helper()
	dsn := os.Getenv("SITANI_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("SITANI_TEST_DATABASE_URL tidak diatur")
	}
	database, err := db.New(dsn, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("buka database: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	ctx := context.Background()
	if _, err := database.ExecContext(ctx, `TRUNCATE permohonan_benih, users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("bersihkan tabel: %v", err)
	}
	return database, NewDBStore(database)
}

// buatPelanggan menyisipkan pengguna pelanggan dan mengembalikan ID-nya.
func buatPelanggan(t *testing.T, database *db.DB, username string) int64 {
	t.Helper()
	var id int64
	if err := database.QueryRowContext(context.Background(),
		`INSERT INTO users (username,email,password_hash,role) VALUES ($1,$2,'x','pelanggan') RETURNING id`,
		username, username+"@contoh.id").Scan(&id); err != nil {
		t.Fatalf("buat pelanggan: %v", err)
	}
	return id
}

// Tes integrasi alur draft terhadap PostgreSQL sungguhan: draft disimpan
// berulang tanpa baris ganda, nomor resmi baru dibuat saat diajukan, dan
// pengajuan ulang atas draft yang sama ditolak.
func TestAlurDraftPermohonan(t *testing.T) {
	database, s := bukaDatabaseUji(t)
	ctx := context.Background()
	pelangganID := buatPelanggan(t, database, "draftuji")
	tanggal := time.Date(2026, 10, 30, 8, 0, 0, 0, time.UTC)

	draftID, err := s.CreatePermohonan(ctx, db.PermohonanBenih{
		PelangganID:          pelangganID,
		JenisBenih:           "Tomat",
		Status:               db.StatusPermohonanDraft,
		PembayaranAwalStatus: "belum",
	}, tanggal)
	if err != nil {
		t.Fatalf("buat draft: %v", err)
	}
	draft, err := s.GetPermohonan(ctx, draftID)
	if err != nil {
		t.Fatalf("baca draft: %v", err)
	}
	if draft.Status != db.StatusPermohonanDraft {
		t.Fatalf("status = %q, ingin draft", draft.Status)
	}
	if draft.DiajukanPada != nil {
		t.Fatalf("draft tidak boleh punya waktu pengajuan: %v", draft.DiajukanPada)
	}

	// Draft dapat diubah berkali-kali.
	if err := s.UpdatePermohonanDraft(ctx, draftID, db.PermohonanBenih{
		PelangganID: pelangganID,
		JenisBenih:  "Tomat",
		Varietas:    "Rama",
		LuasLahan:   0.5,
		LokasiKebun: "Sleman",
	}); err != nil {
		t.Fatalf("ubah draft: %v", err)
	}
	aktif, err := s.GetPermohonanAktifPelanggan(ctx, pelangganID)
	if err != nil || aktif == nil || aktif.ID != draftID {
		t.Fatalf("draft aktif = %+v, %v", aktif, err)
	}
	if aktif.Varietas != "Rama" || aktif.LuasLahan != 0.5 {
		t.Fatalf("isian draft tidak tersimpan: %+v", aktif)
	}

	// Ajukan: nomor resmi dibuat dan isian final ditulis sekaligus.
	if err := s.AjukanPermohonan(ctx, draftID, db.PermohonanBenih{
		PelangganID: pelangganID,
		JenisBenih:  "Tomat",
		Varietas:    "Rama",
		LuasLahan:   0.5,
		LokasiKebun: "Sleman",
	}, tanggal); err != nil {
		t.Fatalf("ajukan permohonan: %v", err)
	}
	resmi, err := s.GetPermohonan(ctx, draftID)
	if err != nil {
		t.Fatalf("baca permohonan resmi: %v", err)
	}
	if resmi.Status != db.StatusPermohonanDiajukan || resmi.NomorPermohonan != "SPB-20261030-0001" {
		t.Fatalf("permohonan resmi = %+v", resmi)
	}
	if resmi.DiajukanPada == nil {
		t.Fatalf("waktu pengajuan tidak tercatat")
	}

	// Draft yang sudah diajukan tidak dapat diubah atau diajukan ulang.
	if err := s.UpdatePermohonanDraft(ctx, draftID, db.PermohonanBenih{PelangganID: pelangganID, JenisBenih: "Padi"}); !errors.Is(err, ErrBukanDraft) {
		t.Fatalf("ubah setelah diajukan: error = %v, ingin ErrBukanDraft", err)
	}
	if err := s.AjukanPermohonan(ctx, draftID, db.PermohonanBenih{PelangganID: pelangganID, JenisBenih: "Padi"}, tanggal); !errors.Is(err, ErrBukanDraft) {
		t.Fatalf("ajukan ulang: error = %v, ingin ErrBukanDraft", err)
	}

	// Pelanggan lain tidak dapat mengubah draft milik orang.
	lainID := buatPelanggan(t, database, "draftuji2")
	draftKedua, err := s.CreatePermohonan(ctx, db.PermohonanBenih{PelangganID: pelangganID, Status: db.StatusPermohonanDraft, PembayaranAwalStatus: "belum"}, tanggal)
	if err != nil {
		t.Fatalf("buat draft kedua: %v", err)
	}
	if err := s.UpdatePermohonanDraft(ctx, draftKedua, db.PermohonanBenih{PelangganID: lainID, JenisBenih: "Padi"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ubah draft orang lain: error = %v, ingin ErrNotFound", err)
	}

	// Draft tidak punya nomor SPB sungguhan (hanya nomor sementara).
	var jumlahDraftSPB int
	if err := database.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM permohonan_benih WHERE status='draft' AND nomor_permohonan LIKE 'SPB-%'`).Scan(&jumlahDraftSPB); err != nil {
		t.Fatalf("hitung draft bernomor: %v", err)
	}
	if jumlahDraftSPB != 0 {
		t.Fatalf("%d draft memakai nomor resmi", jumlahDraftSPB)
	}
}

// Tes integrasi penomoran permohonan terhadap PostgreSQL sungguhan. Dilewati
// bila SITANI_TEST_DATABASE_URL tidak diatur.
func TestCreatePermohonanMenomoriUnikPerHari(t *testing.T) {
	database, s := bukaDatabaseUji(t)
	ctx := context.Background()
	pelangganID := buatPelanggan(t, database, "nomoruji")
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
		t.Fatalf("masih ada %d nomor draft sementara yang tersimpan", sisaSementara)
	}
}
