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

   | Email | Peran | Keterangan |
   | --- | --- | --- |
   | `superadmin@dev.example.invalid` | superadmin | Akses penuh, setara admin |
   | `admin@dev.example.invalid` | admin | Semua modul |
   | `pengawas@dev.example.invalid` | pengawas | Inspeksi lapangan, sertifikasi |
   | `analis@dev.example.invalid` | analis | Uji Mutu Lab, Repository Data |
   | `penandatangan@dev.example.invalid` | penandatangan | Penerbitan (TTE) |
   | `legalisasi@dev.example.invalid` | legalisasi | Penerbitan (legalisasi label) |
   | `percetakan@dev.example.invalid` | percetakan | Penerbitan (cetak label) |
   | `pelanggan@dev.example.invalid` | pelanggan | Mengajukan sertifikasi & hortikultura |
   | `verifikator@dev.example.invalid` | verifikator | Verifikasi proposal hortikultura |
   | `inspektor@dev.example.invalid` | inspektor | Penjadwalan & audit GAP hortikultura |
   | `pbt@dev.example.invalid` | pbt | Penilaian pohon induk |
   | `kasi.pmbtp@dev.example.invalid` | kasi_pmbtp | Penilaian lanjutan |

   Selain akun, seed membuat **data contoh** yang saling terhubung:

   - Permohonan `SPB-DEV-0001` (padi Inpari 32) milik akun pelanggan, status
     `diajukan`, pembayaran awal `belum`.
   - Sampel `SMP-DEV-0001` atas permohonan itu, sudah berstatus `dikodekan` dan
     ditugaskan ke akun analis, sehingga akun Analis Lab langsung punya data
     untuk diuji mutunya.

   Perintah ini idempoten: menjalankan ulang akan memperbarui hash password dan
   menyamakan data contoh tanpa membuat akun, permohonan, atau sampel ganda.

   > Catatan: `superadmin` disamakan dengan `admin` pada pemeriksaan otorisasi
   > backend (`peranAdmin`) dan mendapat seluruh modul di dashboard. Peran ini
   > disediakan untuk lokal/UAT; hapus akunnya sebelum produksi.

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
   `GET /auth/dev/status` dengan `{"aktif": true}`. Halaman masuk resmi portal
   ada di `/masuk` (tema SSO Pemda DIY) dan menampilkan tautan ke login lokal
   hanya bila status tersebut aktif.

### Halaman `/masuk` dan navigasi portal

- Sidebar portal (kiri) memuat tautan **Masuk JSP** dan **Masuk Satriya** ke
  `/masuk?realm=jsp|satriya`. Tombol di halaman itu meminta backend melakukan
  pengalihan, lalu browser berpindah ke issuer SSO
  (`/auth/login/{realm}` → `302` → issuer). Frontend tidak menebak endpoint SSO;
  bila backend menjawab JSON `error`, pesannya ditampilkan di halaman.
- Sidebar dapat diciutkan menjadi rail ikon lewat tombol **Ciutkan**; preferensi
  disimpan di `localStorage` (`sitani:sidebar-collapsed`). Di bawah 1024px sidebar
  menjadi drawer yang dibuka lewat tombol **Menu** dan ditutup dengan Escape,
  overlay, tautan, atau tombol ×.

### Memakai stack Docker Compose

Ringkas bila database dijalankan lewat `infra/compose.yml`:

1. Isi `.env`: `POSTGRES_*`, `SITANI_DEV_LOGIN=true`, `SEED_DEV_PASSWORD`, dan
   `FRONTEND_PUBLIC_URL` yang sama dengan port frontend (`FRONTEND_PORT`, mis.
   `http://localhost:3001`). `FRONTEND_PUBLIC_URL` dipakai untuk memeriksa
   `Origin` login/logout, jadi nilai yang salah menyebabkan `403`.
2. `task migrate`, lalu `task seed-dev` (membaca `.env` dan `SEED_DEV_PASSWORD`).
   Keduanya menyusun DSN dari `POSTGRES_HOST_PORT`; ganti port itu bila 5433
   sudah dipakai proyek lain.
3. `docker compose --env-file .env -f infra/compose.yml up -d --build backend frontend`.
   Compose meneruskan `SITANI_DEV_LOGIN` ke backend; log backend menulis
   `PERINGATAN: login lokal development AKTIF`.
4. Buka `http://localhost:3001/dashboard/masuk` (sesuai `FRONTEND_PORT`).

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
- **SSO lokal.** Untuk menguji alur SSO tanpa Jogjaprov, tersedia Keycloak lokal dengan
  realm sendiri; lihat `docs/sso-keycloak-lokal.md`. Realm itu hanya untuk
  development/UAT.

## Hubungan dengan verifikasi API

Saat `SITANI_DEV_LOGIN` aktif, `auth.Manager.Require` memeriksa sesi lokal
(cookie `sitani_dev_session`) lebih dulu; bila tidak ada, alur SSO berjalan
seperti biasa. Tanpa sesi lokal maupun token SSO valid, permintaan ditolak
(`401`/`503`) sehingga sistem tetap gagal tertutup. Bila login lokal nonaktif,
perilakunya sama persis dengan sebelum fitur ini ada: hanya SSO yang diterima.

Di stack Docker Compose, realm SSO selalu terisi (placeholder dari `.env`), jadi
sebelum perbaikan ini sesi lokal ditolak untuk seluruh `/v1/*`. Regresi untuk
kondisi tersebut dicakup `backend/internal/auth/require_devlocal_test.go`.

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
- `go test ./internal/devseed/...` — format dan parameter argon2id, serta salt acak.
- Uji manual terhadap PostgreSQL 17 (migrasi 010): seed 12 akun + permohonan dan
  sampel contoh, dijalankan dua kali tanpa duplikat. Login lokal `superadmin`,
  `analis`, dan `pelanggan` berhasil (cookie `HttpOnly`), lalu analis menginput
  hasil uji pada `SMP-DEV-0001` (`201`) dan superadmin memproses hasil mutu (`200`, `lulus`).
  Dashboard menampilkan menu sesuai peran: superadmin semua modul, analis
  (Uji Mutu Lab, Repository Data), pelanggan (Sertifikasi Benih, Hortikultura).
- `npx playwright test tests/login-lokal.spec.ts` — 6 tes (desktop + mobile).
- Uji di stack Docker Compose (PostgreSQL 17 pada volume `sitani-diy_postgres_data`,
  migrasi 008–009, `task seed-dev`): login admin berhasil, sesi bertahan setelah
  reload, profil terbaca, logout mengakhiri sesi, dan `/v1/*` menerima sesi lokal
  meski realm SSO terisi placeholder.
