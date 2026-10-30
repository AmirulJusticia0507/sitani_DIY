// Perintah seed-dev membuat akun login lokal dan data contoh untuk
// development/UAT saat SSO Jogjaprov belum aktif. Password dibaca dari
// environment (tidak pernah di-hardcode di repositori) dan disimpan sebagai
// hash argon2id.
//
// Contoh pemakaian:
//
//	SITANI_DEV_LOGIN=true SEED_DEV_PASSWORD=... DB_DSN=postgres://sitani:...@127.0.0.1:5435/sitani?sslmode=disable go run ./cmd/seed-dev
//
// Perintah ini menolak berjalan bila lingkungan terdeteksi produksi.
// Seluruhnya idempoten: menjalankan ulang memperbarui password dan
// menyamakan data contoh tanpa membuat duplikat.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/devseed"
)

// akunDev adalah akun uji, satu per peran. Email memakai domain lokal
// sitani.local (TLD .local tidak pernah diarahkan ke internet) agar tidak
// menyerupai alamat dinas sungguhan, dan username
// berbasis peran (*.dev) agar tidak bentrok dengan username akun SSO nyata.
var akunDev = []struct {
	Username  string
	Email     string
	Role      string
	NPP       string
	UnitKerja string
}{
	{Username: "superadmin.dev", Email: "superadmin@sitani.local", Role: "superadmin", UnitKerja: "Diskominfo DIY (lokal)"},
	{Username: "admin.dev", Email: "admin@sitani.local", Role: "admin", UnitKerja: "BP3MBTP DIY"},
	{Username: "pengawas.dev", Email: "pengawas@sitani.local", Role: "pengawas", UnitKerja: "BP3MBTP DIY"},
	{Username: "analis.dev", Email: "analis@sitani.local", Role: "analis", UnitKerja: "Laboratorium BP3MBTP DIY"},
	{Username: "penandatangan.dev", Email: "penandatangan@sitani.local", Role: "penandatangan", UnitKerja: "BP3MBTP DIY"},
	{Username: "legalisasi.dev", Email: "legalisasi@sitani.local", Role: "legalisasi", UnitKerja: "BP3MBTP DIY"},
	{Username: "percetakan.dev", Email: "percetakan@sitani.local", Role: "percetakan", UnitKerja: "Percetakan rekanan"},
	{Username: "pelanggan.dev", Email: "pelanggan@sitani.local", Role: "pelanggan", UnitKerja: "Kelompok Tani Sido Makmur"},
	{Username: "verifikator.dev", Email: "verifikator@sitani.local", Role: "verifikator", UnitKerja: "Dinas Pertanian DIY"},
	{Username: "inspektor.dev", Email: "inspektor@sitani.local", Role: "inspektor", UnitKerja: "Dinas Pertanian DIY"},
	{Username: "pbt.dev", Email: "pbt@sitani.local", Role: "pbt", UnitKerja: "Dinas Pertanian DIY"},
	{Username: "kasi.dev", Email: "kasi.pmbtp@sitani.local", Role: "kasi_pmbtp", UnitKerja: "Dinas Pertanian DIY"},
}

// permohonanContoh menghubungkan akun pelanggan dengan satu permohonan
// sertifikasi agar halaman /certification tidak kosong saat login sebagai
// Pelanggan, dan penugasan analis agar akun Analis Lab punya sampel untuk diuji.
var permohonanContoh = struct {
	NomorPermohonan string
	JenisBenih      string
	Varietas        string
	LuasLahan       float64
	LokasiKebun     string
	Status          string
	PembayaranAwal  string
}{
	NomorPermohonan: "SPB-DEV-0001",
	JenisBenih:      "Padi",
	Varietas:        "Inpari 32",
	LuasLahan:       1.5,
	LokasiKebun:     "Sleman (data contoh lokal)",
	Status:          "diajukan",
	PembayaranAwal:  "belum",
}

func main() {
	if lingkunganProduksi() {
		fmt.Fprintln(os.Stderr, "seed-dev menolak berjalan di lingkungan produksi")
		os.Exit(1)
	}
	password := os.Getenv("SEED_DEV_PASSWORD")
	if len(password) < 12 {
		fmt.Fprintln(os.Stderr, "SEED_DEV_PASSWORD wajib diisi minimal 12 karakter")
		os.Exit(1)
	}
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DB_DSN wajib diisi")
		os.Exit(1)
	}

	database, err := db.New(dsn, slog.New(slog.NewTextHandler(os.Stderr, nil)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "koneksi database gagal:", err)
		os.Exit(1)
	}
	defer database.Close()

	hash, err := devseed.HashPassword(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gagal membuat hash password:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := seed(ctx, database, hash); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// seed menjalankan seluruh langkah; dipisah dari main agar dapat diuji.
func seed(ctx context.Context, database *db.DB, hash string) error {
	idPerPeran := make(map[string]int64, len(akunDev))
	for _, akun := range akunDev {
		// Password diperbarui agar seed idempoten dan akun lama langsung memakai
		// password terbaru dari environment.
		id, err := db.InsertID(ctx, database, `
			INSERT INTO users (username, email, password_hash, role, npp, unit_kerja, is_active)
			VALUES (?, ?, ?, ?, ?, ?, TRUE)
			ON CONFLICT (email) DO UPDATE
			SET username = EXCLUDED.username,
			    password_hash = EXCLUDED.password_hash,
			    role = EXCLUDED.role,
			    npp = EXCLUDED.npp,
			    unit_kerja = EXCLUDED.unit_kerja,
			    is_active = TRUE`,
			akun.Username, akun.Email, hash, akun.Role, nullable(akun.NPP), nullable(akun.UnitKerja),
		)
		if err != nil {
			return fmt.Errorf("gagal menyimpan akun %s: %w", akun.Email, err)
		}
		idPerPeran[akun.Role] = id
		fmt.Printf("akun siap: %-34s %-13s id=%d\n", akun.Email, akun.Role, id)
	}

	pelangganID, okPelanggan := idPerPeran["pelanggan"]
	analisID, okAnalis := idPerPeran["analis"]
	pengawasID, okPengawas := idPerPeran["pengawas"]
	if !okPelanggan || !okAnalis || !okPengawas {
		return errors.New("akun pelanggan, pengawas, dan analis wajib ada untuk data contoh")
	}

	// ON CONFLICT (nomor_permohonan) menjaga agar seed berulang tidak membuat
	// permohonan ganda.
	var permohonanID int64
	if err := database.QueryRowContext(ctx, `
		INSERT INTO permohonan_benih
			(nomor_permohonan, pelanggan_id, jenis_benih, varietas, luas_lahan, lokasi_kebun, status, pembayaran_awal_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (nomor_permohonan) DO UPDATE
		SET pelanggan_id = EXCLUDED.pelanggan_id,
		    jenis_benih = EXCLUDED.jenis_benih,
		    varietas = EXCLUDED.varietas,
		    luas_lahan = EXCLUDED.luas_lahan,
		    lokasi_kebun = EXCLUDED.lokasi_kebun
		RETURNING id`,
		permohonanContoh.NomorPermohonan, pelangganID, permohonanContoh.JenisBenih, permohonanContoh.Varietas,
		permohonanContoh.LuasLahan, permohonanContoh.LokasiKebun, permohonanContoh.Status, permohonanContoh.PembayaranAwal,
	).Scan(&permohonanID); err != nil {
		return fmt.Errorf("gagal menyimpan permohonan contoh: %w", err)
	}
	fmt.Printf("permohonan contoh: %s (id=%d) milik pelanggan@sitani.local\n", permohonanContoh.NomorPermohonan, permohonanID)

	// Sampel contoh sudah dikodekan dan ditugaskan ke analis, sehingga akun
	// Analis Lab punya data untuk diuji mutunya. Kode sampel unik menjadi kunci
	// idempotensi agar seed berulang tidak menggandakan sampel.
	if _, err := database.ExecContext(ctx, `
		INSERT INTO sampel_benih (permohonan_id, kode_sampel, pengawas_id, analis_id, jumlah_sampel, tanggal_pengambilan, status)
		VALUES (?, ?, ?, ?, ?, CURRENT_DATE, 'dikodekan')
		ON CONFLICT (kode_sampel) DO UPDATE
		SET permohonan_id = EXCLUDED.permohonan_id,
		    pengawas_id = EXCLUDED.pengawas_id,
		    analis_id = EXCLUDED.analis_id,
		    status = 'dikodekan'`,
		permohonanID, "SMP-DEV-0001", pengawasID, analisID, 1,
	); err != nil {
		return fmt.Errorf("gagal menyimpan sampel contoh: %w", err)
	}
	fmt.Printf("sampel contoh: %s ditugaskan ke analis@sitani.local (id=%d)\n", "SMP-DEV-0001", analisID)

	fmt.Println("selesai. Password seluruh akun = SEED_DEV_PASSWORD.")
	return nil
}

// nullable mengubah string kosong menjadi NULL agar kolom opsional tidak berisi
// string kosong yang membingungkan.
func nullable(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func lingkunganProduksi() bool {
	for _, nama := range []string{"APP_ENV", "GO_ENV", "ENVIRONMENT", "SITANI_ENV"} {
		switch strings.ToLower(strings.TrimSpace(os.Getenv(nama))) {
		case "production", "prod":
			return true
		}
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("SENTRY_ENVIRONMENT")), "production")
}
