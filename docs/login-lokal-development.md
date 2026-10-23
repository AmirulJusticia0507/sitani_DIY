# Login Lokal Development

Dokumen ini menjelaskan login lokal yang tersedia untuk development dan UAT selama
SSO Jogjaprov (JSP/Satriya) belum diaktifkan oleh pengelola. Login lokal **bukan**
jalur masuk produksi.

## Ringkasan

| Aspek | Kondisi |
| --- | --- |
| Paket | `backend/internal/auth/devlogin` |
| Perintah seed | `backend/cmd/seed-dev` |
| Halaman | `/dashboard/masuk` (frontend) |
| Tabel | `auth_sessions` (migrasi 009) |
| Flag aktif | `SITANI_DEV_LOGIN=true` |
| Pematian paksa | `APP_ENV`/`GO_ENV`/`ENVIRONMENT`/`SITANI_ENV` = `production`/`prod`, atau `SENTRY_ENVIRONMENT=production` |
| Algoritma password | argon2id (dibuat oleh `seed-dev`); bcrypt tetap diterima bila sudah ada |

## Cara mengaktifkan

1. Jalankan migrasi terbaru.

   ```bash
   DB_DSN='postgres://sitani:<password>@127.0.0.1:5433/sitani?sslmode=disable' go run ./cmd/migrate
   ```

2. Buat akun uji. Password dibaca dari environment, tidak disimpan di repositori.

   ```bash
   SEED_DEV_PASSWORD='password-lokal-minimal-12-karakter' \
   DB_DSN='postgres://sitani:<password>@127.0.0.1:5433/sitani?sslmode=disable' \
   go run ./cmd/seed-dev
   ```

   Akun yang dibuat (semua memakai password yang sama):

   | Email | Peran |
   | --- | --- |
   | `admin@dev.example.invalid` | admin |
   | `pengawas@dev.example.invalid` | pengawas |
   | `analis@dev.example.invalid` | analis |
   | `penandatangan@dev.example.invalid` | penandatangan |
   | `legalisasi@dev.example.invalid` | legalisasi |
   | `percetakan@dev.example.invalid` | percetakan |
   | `pelanggan@dev.example.invalid` | pelanggan |

   Perintah ini idempoten: menjalankan ulang akan memperbarui hash password sesuai
   `SEED_DEV_PASSWORD` terbaru.

3. Jalankan backend dengan flag aktif.

   ```bash
   SITANI_DEV_LOGIN=true \
   APP_ENV=development \
   FRONTEND_PUBLIC_URL=http://localhost:3002 \
   DB_DSN='postgres://sitani:<password>@127.0.0.1:5433/sitani?sslmode=disable' \
   go run ./cmd/api
   ```

   Saat aktif, stderr menulis peringatan `login lokal development AKTIF`. Ini
   sengaja agar terlihat jelas di log.

4. Buka `/dashboard/masuk`, masukkan email dan password, lalu masuk. Tautan
   "Masuk lokal" hanya muncul di antarmuka bila backend menjawab
   `GET /auth/dev/status` dengan `{"aktif": true}`.

## Perilaku keamanan

- **Gagal tertutup.** Saat `SITANI_DEV_LOGIN` tidak diset atau lingkungan terdeteksi
  produksi, ketiga endpoint menjawab `404`. Halaman masuk menampilkan pesan bahwa
  login lokal tidak tersedia dan mengarahkan ke SSO.
- **CSRF.** Login dan logout mewajibkan header `Origin` sama dengan
  `FRONTEND_PUBLIC_URL`; selain itu ditolak `403`.
- **Pesan seragam.** Email tidak dikenal dan password salah sama-sama dijawab
  `401 "email atau password salah"`. Saat email tidak ditemukan, backend tetap
  menjalankan perbandingan bcrypt agar waktu respons tidak membocorkan keberadaan akun.
- **Akun SSO tidak bisa dipakai.** `password_hash` bernilai `SSO-DISABLED-...` untuk
  akun hasil pemetaan SSO; login lokal selalu menolak nilai itu.
- **Token sesi.** Cookie `sitani_dev_session` bersifat `HttpOnly` dan
  `SameSite=Lax`. Database hanya menyimpan hash SHA-256 dari token. Masa berlaku 8 jam.
  Logout menghapus baris sesi, sehingga token lama tidak dapat dipakai ulang.
- **Terpisah dari SSO.** Cookie dan alur berbeda; tidak menggantikan verifikasi token SSO.

## Hubungan dengan verifikasi API

`auth.Manager.Require` tetap memakai SSO sebagai jalur utama. Sesi lokal hanya
diperiksa bila konfigurasi realm SSO belum lengkap. Bila SSO sudah dikonfigurasi,
permintaan tanpa token SSO valid tetap ditolak.

## Yang harus dilakukan sebelum produksi

1. Pastikan `SITANI_DEV_LOGIN` tidak diset (atau `APP_ENV=production`).
2. Hapus akun `*@dev.example.invalid` dari database produksi.
3. Tetapkan password akun nyata lewat SSO; jangan membuat akun dengan password di produksi.

## Batasan

- Belum ada alur reset password, kunci akun setelah gagal berulang, dan pembatasan laju
  (rate limit). Karena itu login ini tidak layak untuk produksi.
- Hanya satu realm lokal (`lokal`). Pemetaan peran tetap dari `users.role` di database.

## Verifikasi yang sudah dijalankan

- `go test ./internal/auth/...` — unit test argon2id/bcrypt, penolakan penanda SSO,
  origin asing, endpoint nonaktif, dan pematian otomatis di produksi.
- Uji manual terhadap PostgreSQL 17: migrasi 009, seed 7 akun, login berhasil/gagal,
  `GET /auth/me`, dan logout. Dengan `APP_ENV=production` login ditolak `404`.
- `npx playwright test tests/login-lokal.spec.ts` — 6 tes (desktop + mobile).
