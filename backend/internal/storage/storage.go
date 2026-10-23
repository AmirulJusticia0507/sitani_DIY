// Package storage menyimpan berkas dokumen ke object storage S3/MinIO.
//
// Berkas disimpan privat: unduhan hanya lewat URL presign bertanda waktu
// yang diterbitkan setelah pemeriksaan hak akses oleh handler.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Jenis berkas yang diizinkan untuk dokumen permohonan/dokumen pendukung.
var jenisDiizinkan = map[string]string{
	"application/pdf":          ".pdf",
	"image/png":                ".png",
	"image/jpeg":               ".jpg",
	"application/zip":          ".zip",
	"text/csv":                 ".csv",
	"application/vnd.ms-excel": ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
	"application/msword": ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
}

// defaultUkuranMaks 20 MB, dapat ditimpa lewat MINIO_MAX_UPLOAD_MB.
const defaultUkuranMaks = 20 << 20

// TTL presign unduhan: pendek agar tautan tidak lama terbuka.
const preSignDurasi = 5 * time.Minute

// Konfirmasi awaran konfigurasi tidak lengkap.
var ErrStorageTidakSiap = errors.New("penyimpanan dokumen belum dikonfigurasi")

// Config menampung parameter koneksi object storage.
type Config struct {
	Endpoint   string
	AksesKunci string
	Rahasia    string
	Bucket     string
	HTTPS      bool
	UkuranMaks int64
}

// Siap melapor benar jika set kunci wajib lengkap.
func (c Config) Siap() bool {
	return c.Endpoint != "" && c.Bucket != "" && c.AksesKunci != "" && c.Rahasia != ""
}

// ConfigFromEnv membaca konfigurasi dari environment variable.
// Ukuran maks boleh 0 → dipakai default 20 MB.
func ConfigFromEnv() Config {
	var ukuran int64 = int64(defaultUkuranMaks)
	if v := os.Getenv("MINIO_MAX_UPLOAD_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ukuran = int64(n) * int64(defaultUkuranMaks) / 20
		}
	}
	return Config{
		Endpoint:   strings.TrimSpace(os.Getenv("MINIO_ENDPOINT")),
		AksesKunci: os.Getenv("MINIO_ACCESS_KEY"),
		Rahasia:    os.Getenv("MINIO_SECRET_KEY"),
		Bucket:     strings.TrimSpace(os.Getenv("MINIO_BUCKET")),
		HTTPS:      strings.EqualFold(os.Getenv("MINIO_SECURE"), "true"),
		UkuranMaks: ukuran,
	}
}

// ukuranTeks menampilkan ukuran dalam MB agar pesan mudah dibaca.
func ukuranTeks(n int64) string {
	return fmt.Sprintf("%d MB", n>>20)
}

// Client antar muka minimum object storage yang dipakai layanan dokumen.
// Diekspor agar mudah dipalsu pada tes tanpa MinIO nyata.
type Client interface {
	Unggah(ctx context.Context, objectKey string, reader io.Reader, ukuran int64, jenis string) error
	URLUnduh(ctx context.Context, objectKey string) (string, error)
	Periksa(ctx context.Context) error
}

// Rencana adalah hasil validasi unggahan siap disimpan.
type Rencana struct {
	ObjectKey string
	NamaAsli  string
	Jenis     string
	Ukuran    int64
	Masukan   io.Reader
}

var namaDihapus = regexp.MustCompile(`[^[:alnum:]\-_ .]+`)

// tandaKonten memetakan jenis yang diizinkan ke tanda baca (magic bytes)
// untuk memastikan isi berkas sesuai deklarasi. Tipe berbasis teks tanpa
// sentinel map ke string kosong sehingga hanya diperiksa deklarasinya.
var tandaKonten = map[string]string{
	"application/pdf": "%PDF-",
	"image/png":       "\x89PNG",
	"image/jpeg":      "\xFF\xD8\xFF",
	"application/zip": "PK\x03\x04",
	// xlsx/docx juga berupa wadah zip, diperiksa sama.
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       "PK\x03\x04",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "PK\x03\x04",
	// .xls dan .doc lama memakai wadah OLE2.
	"application/vnd.ms-excel": "\xD0\xCF\x11\xE0",
	"application/msword":       "\xD0\xCF\x11\xE0",
	"text/csv":                 "",
}

// Siapkan memvalidasi berkas terunggah dan menyiapkan kunci objek.
//
// userID dipakai pada awalan jalur supaya berkas tiap pengguna terpisah;
// namaAsli disanitasi (karakter berbahaya dihapus) hanya untuk penyebutan
// dalam metadata, bukan bagian dari kunci objek. Isi berkas diperiksa
// terhadap tanda baca jenis yang dideklarasikan supaya pembawaan jenis
// berbahaya tidak lolos lewat header palsu.
func Siapkan(ukuranMaks int64, userID int64, berkas multipart.File, header *multipart.FileHeader) (*Rencana, error) {
	if berkas == nil || header == nil {
		return nil, fmt.Errorf("berkas tidak tersedia")
	}
	jenis := header.Header.Get("Content-Type")
	if i := strings.Index(jenis, ";"); i >= 0 {
		jenis = strings.TrimSpace(jenis[:i])
	}
	ekstensi, ok := jenisDiizinkan[jenis]
	if !ok {
		return nil, fmt.Errorf("jenis berkas %q tidak diizinkan", jenis)
	}
	if header.Size > ukuranMaks {
		return nil, fmt.Errorf("ukuran berkas melebihi batas %s", ukuranTeks(ukuranMaks))
	}

	// Membungkus pembacaan dengan satu byte di atas batas supaya berkas
	// melebihi ukuran terdeteksi sebelum diproses lebih jauh.
	tubuh, err := io.ReadAll(io.LimitReader(berkas, ukuranMaks+1))
	if err != nil {
		return nil, fmt.Errorf("baca berkas: %w", err)
	}
	if int64(len(tubuh)) > ukuranMaks {
		return nil, fmt.Errorf("ukuran berkas melebihi batas %s", ukuranTeks(ukuranMaks))
	}
	if len(tubuh) == 0 {
		return nil, fmt.Errorf("berkas kosong")
	}
	if tanda := tandaKonten[jenis]; tanda != "" && !strings.HasPrefix(string(tubuh), tanda) {
		return nil, fmt.Errorf("indikasi isi berkas tidak sesuai dengan jenis %q", jenis)
	}

	// Kunci objek unik per unggahan + terpisah per pemilik; nama asli
	// dipertahankan pada metadata dokumen.
	id := uuid.NewString()
	kunci := fmt.Sprintf("dokumen/%d/%s%s", userID, id, ekstensi)
	namaAsli := filepath.Base(header.Filename)
	namaAsli = namaDihapus.ReplaceAllString(namaAsli, "_")
	if namaAsli == "." || namaAsli == "_" || namaAsli == "" {
		namaAsli = "dokumen" + ekstensi
	}

	return &Rencana{
		ObjectKey: kunci,
		NamaAsli:  namaAsli,
		Jenis:     jenis,
		Ukuran:    int64(len(tubuh)),
		Masukan:   strings.NewReader(string(tubuh)),
	}, nil
}
