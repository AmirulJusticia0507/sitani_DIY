// Perintah seed-dev membuat akun login lokal untuk development/UAT saat SSO
// Jogjaprov belum aktif. Password dibaca dari environment (tidak pernah
// di-hardcode di repositori) dan disimpan sebagai hash argon2id.
//
// Contoh pemakaian:
//
//	SITANI_DEV_LOGIN=true SEED_DEV_PASSWORD=... DB_DSN=postgres://sitani:...@127.0.0.1:5433/sitani?sslmode=disable go run ./cmd/seed-dev
//
// Perintah ini menolak berjalan bila lingkungan terdeteksi produksi.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

// akunDev adalah akun uji, satu per peran utama. Email memakai domain contoh
// (example.invalid) agar tidak menyerupai alamat dinas sungguhan.
var akunDev = []struct {
	Username string
	Email    string
	Role     string
}{
	{"admin.dev", "admin@dev.example.invalid", "admin"},
	{"pengawas.dev", "pengawas@dev.example.invalid", "pengawas"},
	{"analis.dev", "analis@dev.example.invalid", "analis"},
	{"penandatangan.dev", "penandatangan@dev.example.invalid", "penandatangan"},
	{"legalisasi.dev", "legalisasi@dev.example.invalid", "legalisasi"},
	{"percetakan.dev", "percetakan@dev.example.invalid", "percetakan"},
	{"pelanggan.dev", "pelanggan@dev.example.invalid", "pelanggan"},
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

	hash, err := hashPassword(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gagal membuat hash password:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, akun := range akunDev {
		// Password diperbarui agar seed idempoten dan akun lama langsung memakai
		// password terbaru dari environment.
		if _, err := database.ExecContext(ctx, `
			INSERT INTO users (username, email, password_hash, role, is_active)
			VALUES (?, ?, ?, ?, TRUE)
			ON CONFLICT (email) DO UPDATE
			SET password_hash = EXCLUDED.password_hash,
			    role = EXCLUDED.role,
			    is_active = TRUE`,
			akun.Username, akun.Email, hash, akun.Role,
		); err != nil {
			fmt.Fprintf(os.Stderr, "gagal menyimpan akun %s: %v\n", akun.Email, err)
			os.Exit(1)
		}
		fmt.Printf("akun siap: %s (%s)\n", akun.Email, akun.Role)
	}
	fmt.Println("selesai. Password seluruh akun = SEED_DEV_PASSWORD.")
}

// hashPassword membuat hash argon2id dengan parameter yang sepadan dengan
// OWASP Password Storage Cheat Sheet (m=64 MiB, t=3, p=2).
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	const mem uint32 = 64 * 1024
	const iter uint32 = 3
	const par uint8 = 2
	kunci := argon2.IDKey([]byte(password), salt, iter, mem, par, 32)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		mem, iter, par,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(kunci)), nil
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
