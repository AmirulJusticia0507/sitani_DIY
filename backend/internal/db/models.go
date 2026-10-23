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
	ID        int64
	Username  string
	Email     string
	NPP       string
	UnitKerja string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Status permohonan yang menandakan pengajuan sudah selesai atau tidak
// dilanjutkan. Permohonan dengan status ini tidak ditawarkan untuk registrasi
// lot benih.
const (
	StatusPermohonanDitolak    = "ditolak"
	StatusPermohonanDibatalkan = "dibatalkan"
)

// PermohonanBenih merekam pengajuan sertifikasi benih pelanggan (Fase 1).
//
// NomorPermohonan tidak diisi oleh client: nomor dibuat server setelah insert
// dengan format SPB-YYYYMMDD-XXXX lalu disimpan di kolom nomor_permohonan.
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
	ID              int64
	PermohonanID    int64
	Fase            int
	PengawasID      int64
	Hasil           string
	Catatan         string
	TanggalInspeksi string
	CreatedAt       time.Time
}

// BukuInduk mencatat pencatatan buku induk setelah verifikasi lapangan.
type BukuInduk struct {
	ID             int64
	PermohonanID   int64
	NomorBukuInduk string
	Catatan        string
	DibuatOleh     int64
	CreatedAt      time.Time
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
	ID                 int64
	PermohonanID       int64
	KodeSampel         string
	PengawasID         int64
	AnalisID           *int64
	JumlahSampel       int
	TanggalPengambilan string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// UjiMutu mencatat hasil pengujian lab oleh Analis.
type UjiMutu struct {
	ID          int64
	SampelID    int64
	Parameter   string
	Hasil       string
	StatusLulus bool
	Catatan     string
	DiujiOleh   int64
	TanggalUji  string
	CreatedAt   time.Time
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
	ID              int64
	PermohonanID    int64
	NomorSertifikat string
	PenandatanganID int64
	TanggalTTD      *time.Time
	TTEStatus       string
	TTEDokumenURL   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DraftLabel adalah draft label sebelum cetak, dengan kode unik/barcode terintegrasi SIM Perbenihan.
type DraftLabel struct {
	ID                 int64
	SertifikatID       int64
	KodeBarcode        string
	Status             string // draft | disetujui | dicetak | legalisasi
	IntegrasiSIMStatus string // belum | terkirim | gagal
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// LabelLegalisasi mencatat legalisasi label oleh petugas legalisasi.
type LabelLegalisasi struct {
	ID           int64
	DraftLabelID int64
	LegalisasiID int64
	Status       string // belum | diverifikasi | ditolak | selesai
	Catatan      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PenyerahanDokumen mencatat penyerahan sertifikat & label ke pelanggan dan pembayaran akhir.
type PenyerahanDokumen struct {
	ID                    int64
	PermohonanID          int64
	SertifikatDikirim     bool
	LabelDikirim          bool
	PembayaranAkhirStatus string
	TanggalPenyerahan     *time.Time
	DiterimaOleh          string
	Catatan               string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// Kelas benih sesuai penggolongan resmi perbenihan.
const (
	KelasBenihPenjenis = "penjenis"
	KelasBenihDasar    = "dasar"
	KelasBenihPokok    = "pokok"
	KelasBenihSebar    = "sebar"
)

// Status edar lot benih.
const (
	StatusEdarAktif        = "aktif"
	StatusEdarDitangguhkan = "ditangguhkan"
	StatusEdarDicabut      = "dicabut"
	StatusEdarKedaluwarsa  = "kedaluwarsa"
)

// RegistrasiBenih adalah master lot benih yang telah tersertifikasi dan siap diedarkan.
// Nomor registrasi resmi masih menunggu validasi BP3MBTP; format internal memakai "RB-<nomor_lot>".
type RegistrasiBenih struct {
	ID                 int64
	NomorRegistrasi    string
	PermohonanID       int64
	SertifikatID       *int64
	DraftLabelID       *int64
	NomorSertifikat    string
	KodeLabel          string
	Komoditas          string
	Varietas           string
	Produsen           string
	SumberBenih        string
	NomorLot           string
	KelasBenih         string
	Jumlah             float64
	Satuan             string
	TanggalProduksi    string
	TanggalKedaluwarsa string
	StatusEdar         string
	CreatedBy          int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// StokBenih adalah saldo persediaan terkini untuk satu lot benih.
type StokBenih struct {
	ID                 int64
	RegistrasiID       int64
	NomorLot           string
	Komoditas          string
	Varietas           string
	JumlahTersedia     float64
	Satuan             string
	TanggalKedaluwarsa string
	StatusEdar         string
	UpdatedAt          time.Time
}

// MutasiStokBenih mencatat setiap perubahan saldo stok agar dapat diaudit.
type MutasiStokBenih struct {
	ID            int64
	RegistrasiID  int64
	Jenis         string // masuk | keluar | koreksi
	Jumlah        float64
	SaldoSetelah  float64
	ReferensiTipe string
	ReferensiID   *int64
	Catatan       string
	CreatedBy     int64
	CreatedAt     time.Time
}

// PenjualanBenih mencatat transaksi penjualan lot benih. Transaksi tidak pernah
// dihapus; pembatalan mengubah status dan membuat mutasi stok pembalik.
type PenjualanBenih struct {
	ID               int64
	NomorDokumen     string
	RegistrasiID     int64
	NomorLot         string
	Pembeli          string
	Jumlah           float64
	HargaSatuan      float64
	TotalHarga       float64
	TanggalTransaksi string
	Status           string // selesai | dibatalkan
	AlasanPembatalan string
	DibatalkanOleh   *int64
	DibatalkanPada   *time.Time
	CreatedBy        int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// RegistrasiFilter membatasi daftar registrasi benih.
type RegistrasiFilter struct {
	StatusEdar         string
	Komoditas          string
	KedaluwarsaSebelum *time.Time
	Limit              int
}

// SiapRegistrasiPermohonan adalah ringkasan permohonan yang sudah memenuhi
// syarat registrasi lot: sertifikat bertanda tangan TTE, label sudah
// dilegalisasi dan terkirim ke SIM Perbenihan, serta belum punya lot terdaftar.
type SiapRegistrasiPermohonan struct {
	PermohonanID    int64
	NomorPermohonan string
	JenisBenih      string
	Varietas        string
	Status          string
}

// PenjualanFilter membatasi daftar penjualan benih.
type PenjualanFilter struct {
	RegistrasiID int64
	Status       string
	Dari         *time.Time
	Sampai       *time.Time
	Limit        int
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

// Dokumen mencatat metadata lampiran yang fisiknya tersimpan privat di
// object storage. Unduhan dilayani lewat URL presign setelah pemeriksaan
// hak akses; objek tidak dapat diakses langsung.
type Dokumen struct {
	ID        int64
	PemilikID int64
	Tujuan    string
	ObjectKey string
	NamaAsli  string
	Jenis     string
	Ukuran    int64
	CreatedAt time.Time
}
