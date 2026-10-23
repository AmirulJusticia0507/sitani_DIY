package store

import (
	"context"
	"errors"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

// Kesalahan domain yang perlu dibedakan oleh lapisan service dan HTTP.
var (
	ErrNotFound        = errors.New("data tidak ditemukan")
	ErrStokTidakCukup  = errors.New("stok benih tidak mencukupi")
	ErrSudahDibatalkan = errors.New("penjualan sudah dibatalkan")
)

// Store adalah antarmuka akses data untuk modul sertifikasi benih BP3MBTP.
// Diekspor agar mudah diimplementasikan ulang (misal in-memory untuk tes)
// tanpa mengubah lapisan service atau handler.
type Store interface {
	// User / RBAC
	CreateUser(ctx context.Context, u db.User) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*db.User, error)
	GetUserByUsername(ctx context.Context, username string) (*db.User, error)

	// Fase 1: Permohonan & pembayaran
	// CreatePermohonan mengisi nomor_permohonan secara atomik dengan format
	// SPB-YYYYMMDD-XXXX berdasarkan tanggal pengajuan.
	CreatePermohonan(ctx context.Context, p db.PermohonanBenih, tanggal time.Time) (int64, error)
	GetPermohonan(ctx context.Context, id int64) (*db.PermohonanBenih, error)
	SetNomorPermohonan(ctx context.Context, id int64, nomor string) error
	UpdatePermohonanStatus(ctx context.Context, id int64, status, pembayaranAwalStatus string) error
	ListPermohonanSiapRegistrasi(ctx context.Context, pemilikID int64, semua bool, limit int) ([]db.SiapRegistrasiPermohonan, error)
	CreatePembayaran(ctx context.Context, pb db.Pembayaran) (int64, error)

	// Fase 1: Inspeksi 4 fase & buku induk
	CreateInspeksiFase(ctx context.Context, i db.InspeksiFase) (int64, error)
	ListInspeksiFase(ctx context.Context, permohonanID int64) ([]db.InspeksiFase, error)
	CreateBukuInduk(ctx context.Context, b db.BukuInduk) (int64, error)
	GetBukuInduk(ctx context.Context, permohonanID int64) (*db.BukuInduk, error)

	// Fase 2: Pengujian Lab Mutu
	CreateSampel(ctx context.Context, s db.SampelBenih) (int64, error)
	GetSampel(ctx context.Context, id int64) (*db.SampelBenih, error)
	KodifikasiSampel(ctx context.Context, id int64, kodeSampel string, analisID int64) error
	CreateUjiMutu(ctx context.Context, u db.UjiMutu) (int64, error)
	ListUjiMutu(ctx context.Context, sampelID int64) ([]db.UjiMutu, error)
	CreateLaporanMutu(ctx context.Context, l db.LaporanMutu) (int64, error)
	GetLaporanMutu(ctx context.Context, sampelID int64) (*db.LaporanMutu, error)

	// Fase 3: Sertifikasi & Penerbitan Label
	CreateSertifikat(ctx context.Context, s db.SertifikatBenih) (int64, error)
	GetSertifikat(ctx context.Context, permohonanID int64) (*db.SertifikatBenih, error)
	GetSertifikatByID(ctx context.Context, id int64) (*db.SertifikatBenih, error)
	UpdateSertifikatTTE(ctx context.Context, permohonanID int64, tteStatus, tteDokumenURL string, ttd *time.Time) error
	CreateDraftLabel(ctx context.Context, l db.DraftLabel) (int64, error)
	GetDraftLabel(ctx context.Context, sertifikatID int64) (*db.DraftLabel, error)
	GetDraftLabelByID(ctx context.Context, id int64) (*db.DraftLabel, error)
	UpdateDraftLabel(ctx context.Context, id int64, status, integrasiSIM, kodeBarcode string) error
	CreateLegalisasi(ctx context.Context, l db.LabelLegalisasi) (int64, error)
	UpdateLegalisasi(ctx context.Context, id int64, status string) error
	CreatePenyerahan(ctx context.Context, p db.PenyerahanDokumen) (int64, error)
	GetPenyerahan(ctx context.Context, permohonanID int64) (*db.PenyerahanDokumen, error)
	UpdatePenyerahan(ctx context.Context, permohonanID int64, sertifikatDikirim, labelDikirim bool, pembayaranAkhirStatus string) error

	// Registrasi, persediaan, dan penjualan benih
	CreateRegistrasiBenih(ctx context.Context, r db.RegistrasiBenih) (int64, error)
	GetRegistrasiBenih(ctx context.Context, id int64) (*db.RegistrasiBenih, error)
	GetRegistrasiBenihByNomorLot(ctx context.Context, nomorLot string) (*db.RegistrasiBenih, error)
	ListRegistrasiBenih(ctx context.Context, filter db.RegistrasiFilter) ([]db.RegistrasiBenih, error)
	UpdateRegistrasiStatus(ctx context.Context, id int64, status string) error
	GetStokBenih(ctx context.Context, registrasiID int64) (*db.StokBenih, error)
	ListStokBenih(ctx context.Context) ([]db.StokBenih, error)
	ListMutasiStok(ctx context.Context, registrasiID int64) ([]db.MutasiStokBenih, error)
	CreatePenjualanBenih(ctx context.Context, p db.PenjualanBenih) (*db.PenjualanBenih, error)
	GetPenjualanBenih(ctx context.Context, id int64) (*db.PenjualanBenih, error)
	BatalkanPenjualanBenih(ctx context.Context, id int64, actorID int64, alasan string, waktu time.Time) (*db.PenjualanBenih, error)
	ListPenjualanBenih(ctx context.Context, filter db.PenjualanFilter) ([]db.PenjualanBenih, error)

	// Audit
	CreateAuditLog(ctx context.Context, a db.AuditLog) (int64, error)

	// Dokumen (lampiran privat di object storage)
	CreateDokumen(ctx context.Context, d db.Dokumen) (int64, error)
	GetDokumen(ctx context.Context, id int64) (*db.Dokumen, error)
	ListDokumen(ctx context.Context, userID int64, semua bool) ([]db.Dokumen, error)

	// Health
	Ping(ctx context.Context) error
}

// DBStore adalah implementasi Store menggunakan PostgreSQL melalui db.DB.
type DBStore struct {
	db *db.DB
}

func NewDBStore(database *db.DB) *DBStore {
	return &DBStore{db: database}
}

// Ping memeriksa koneksi database (diisi implementasi penuh progresif per tabel).
func (s *DBStore) Ping(ctx context.Context) error {
	return s.db.Health(ctx)
}
