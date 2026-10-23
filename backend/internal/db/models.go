package db

import (
	"time"
)

// Role sesuai 7 aktor swimlane alur sertifikasi benih BP3MBTP.
const (
	RolePelanggan     = "pelanggan"
	RoleAdmin         = "admin"
	RolePengawas      = "pengawas"
	RoleAnalis        = "analis"
	RolePenandatangan = "penandatangan"
	RoleLegalisasi    = "legalisasi"
	RolePercetakan    = "percetakan"
)

// User merekam akter dalam sistem sertifikasi benih.
type User struct {
	ID         int64
	Username   string
	Email      string
	NPP        string
	UnitKerja  string
	Role       string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PermohonanBenih merekam pengajuan sertifikasi benih pelanggan (Fase 1).
type PermohonanBenih struct {
	ID                   int64
	NomorPermohonan      string
	PelangganID          int64
	JenisBenih           string
	Varietas             string
	LuasLahan            float64
	LokasiKebun          string
	Status               string
	PembayaranAwalStatus string
	BukuIndukNomor       string
	LaporanVerifikasi    string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// InspeksiFase merekam hasil salah satu dari 4 fase inspeksi lapangan oleh Pengawas.
type InspeksiFase struct {
	ID            int64
	PermohonanID  int64
	Fase          int
	PengawasID    int64
	Hasil         string
	Catatan       string
	TanggalInspeksi string
	CreatedAt     time.Time
}

// BukuInduk mencatat pencatatan buku induk setelah verifikasi lapangan.
type BukuInduk struct {
	ID            int64
	PermohonanID  int64
	NomorBukuInduk string
	Catatan       string
	DibuatOleh    int64
	CreatedAt     time.Time
}

// Pembayaran mencatat pembayaran awal dan akhir melalui Gateway Pembayaran.
type Pembayaran struct {
	ID            int64
	PermohonanID  int64
	Jenis         string // awal | akhir
	Jumlah        float64
	Status        string // belum | lunas | gagal
	BuktiTransfer string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SampelBenih merekam sampel benih yang diambil Pengawas dan dikodekan (Fase 2).
type SampelBenih struct {
	ID               int64
	PermohonanID     int64
	KodeSampel       string
	PengawasID       int64
	AnalisID         *int64
	JumlahSampel     int
	TanggalPengambilan string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UjiMutu mencatat hasil pengujian lab oleh Analis.
type UjiMutu struct {
	ID           int64
	SampelID     int64
	Parameter    string
	Hasil        string
	StatusLulus  bool
	Catatan      string
	DiujiOleh    int64
	TanggalUji   string
	CreatedAt    time.Time
}

// LaporanMutu adalah kesimpulan pengujian mutu yang disusun Analis.
type LaporanMutu struct {
	ID          int64
	SampelID    int64
	Kesimpulan  string // lulus | tidak_lulus
	IsiLaporan  string
	DisusunOleh int64
	CreatedAt   time.Time
}

// SertifikatBenih adalah dokumen resmi yang dibuat jika lulus, ditandatangani Pejabat via TTE.
type SertifikatBenih struct {
	ID            int64
	PermohonanID  int64
	NomorSertifikat string
	PenandatanganID int64
	TanggalTTD    *time.Time
	TTEStatus     string
	TTEDokumenURL string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DraftLabel adalah draft label sebelum cetak, dengan kode unik/barcode terintegrasi SIM Perbenihan.
type DraftLabel struct {
	ID               int64
	SertifikatID     int64
	KodeBarcode      string
	Status           string // draft | disetujui | dicetak | legalisasi
	IntegrasiSIMStatus string // belum | terkirim | gagal
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// LabelLegalisasi mencatat legalisasi label oleh petugas legalisasi.
type LabelLegalisasi struct {
	ID            int64
	DraftLabelID  int64
	LegalisasiID  int64
	Status        string // belum | diverifikasi | ditolak | selesai
	Catatan       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PenyerahanDokumen mencatat penyerahan sertifikat & label ke pelanggan dan pembayaran akhir.
type PenyerahanDokumen struct {
	ID                int64
	PermohonanID      int64
	SertifikatDikirim bool
	LabelDikirim      bool
	PembayaranAkhirStatus string
	TanggalPenyerahan *time.Time
	DiterimaOleh      string
	Catatan           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// AuditLog mencatat aksi lintas aktor untuk traceabilitas.
type AuditLog struct {
	ID        int64
	ActorID   *int64
	ActorRole string
	Action    string
	Entity    string
	EntityID  *int64
	Detail    string
	CreatedAt time.Time
}
