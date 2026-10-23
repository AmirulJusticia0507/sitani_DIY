# Portal Pertanian dan Ketahanan Pangan DIY

Portal Pertanian dan Ketahanan Pangan DIY adalah platform terpadu yang dikembangkan untuk Dinas Pertanian dan Ketahanan Pangan (DPKP) Pemda DIY bekerja sama dengan Dinas Komunikasi dan Informatika DIY. Portal ini berfungsi sebagai repository data lokal dan pusat layanan publik di bidang pertanian dan ketahanan pangan.

---

## 🛠 Tech Stack

> Status: tahap perencanaan; stack di bawah merupakan target implementasi.
> Analisis kondisi aktual: [docs/analisis-proyek.md](docs/analisis-proyek.md).
> Pedoman pengerjaan agen: [AGENTS.md](AGENTS.md).

- **Backend:** Golang 1.23
- **Frontend:** Next.js 16 (baseline kebutuhan Next.js 15 dinaikkan karena dukungan keamanan)
- **Database development baru:** PostgreSQL 17; volume MySQL lama dipertahankan dan tidak dimigrasikan.
- **Object Storage:** MinIO
- **TTE / Central Sign:** Sadewa Integrasi
- **API Gateway:** SPL / Cakra API (`spl.jogjaprov.go.id`)
- **Authentication:** SSO Jogjaprov (Realm JSP & Satriya)
- **Monitoring & Log:** Sentry & Discord Webhook

---

## 🏗 Arsitektur Container

Aplikasi dijalankan dalam 3 container utama:
1. **Container 1:** Database Server (PostgreSQL)
2. **Container 2:** Backend API Service (Golang 1.23)
3. **Container 3:** Frontend Web Application (Next.js 15)

---

## 📋 Alur Utama Sistem

1. **Authentication:**
   - Pengguna Umum/Pelaku Usaha menggunakan realm **JSP**.
   - ASN / Internal OPD menggunakan realm **Satriya**.
2. **API Routing:**
   - Seluruh request dari client diarahkan melalui Gateway **SPL (Cakra API)**.
3. **Layanan Registrasi & Sertifikasi:**
   - Pemohon mengajukan proposal/berkas secara elektronik.
   - Verifikasi kelengkapan berkas oleh Petugas Verifikator.
   - Inspeksi kebun dan audit GAP oleh Petugas Inspektor.
   - Evaluasi kelulusan & penetapan register.
   - Penerbitan sertifikat/dokumen resmi dengan **TTE Sadewa**.

---

## 🌾 Arsitektur Alur Sertifikasi Benih BP3MBTP (berdasarkan Diagram XML)

Fitur **Pengelolaan Benih Tanaman** mencakup registrasi benih, sertifikasi BP3MBTP, pengelolaan lot/stok, dan pencatatan penjualan. Alur sertifikasi mengikuti 3 fase dan 7 aktor swimlane. Bagian ini adalah arsitektur target; status implementasi aktual tetap mengikuti `todo_list.md`.

**Fase 1 — Verifikasi Lapangan (Pengawas/PBT)**
Permohonan & Pembayaran Awal → Verifikasi Dokumen → Inspeksi 4 Fase → Pencatatan Buku Induk & Laporan Verifikasi

**Fase 2 — Pengujian Lab Mutu (Pengawas + Analis Lab)**
Pengambilan Sampel Benih → Kodifikasi Sampel → Pengujian Lab → Penyusunan Laporan Mutu

**Fase 3 — Sertifikasi & Penerbitan Label**
- Jalur Lulus: Pembuatan Sertifikat Benih → Penandatanganan Pejabat (TTE) → Draft Label & Integrasi SIM Perbenihan → Cetak di Percetakan → Legalisasi Label → Penyerahan Sertifikat/Label ke Pelanggan & Pembayaran Akhir
- Jalur Tidak Lulus: Pengiriman Laporan Lengkap Pengujian ke Pelanggan → Proses selesai

### Komponen Integrasi
- **SIM Perbenihan:** integrasi kode unik/barcode label dan data sertifikasi benih.
- **Modul Legalisasi:** verifikasi dan legalisasi label sebelum penyerahan.
- **Percetakan:** cetak label resmi setelah draft disetujui.
- **Gateway Pembayaran:** pembayaran awal pada permohonan dan pembayaran akhir pada penyerahan dokumen.
- **TTE Sadewa:** penandatanganan sertifikat oleh pejabat berwenang.

7 Aktor: Pelanggan, Admin, Pengawas/PBT, Analis Lab, Penandatangan TTE, Legalisasi, Percetakan.

### Registrasi dan Penjualan Benih

- Registrasi menyimpan komoditas/varietas, produsen, sumber benih, nomor lot, kelas, jumlah, satuan, tanggal produksi, masa berlaku, dan status edar.
- Lot yang dijual harus memiliki registrasi, sertifikat, dan label yang masih sah sesuai aturan bisnis yang disetujui BP3MBTP.
- Penjualan menyimpan pembeli, lot, kuantitas, harga, nomor dokumen, tanggal transaksi, dan perubahan stok secara transaksional.
- Koreksi atau pembatalan penjualan wajib meninggalkan audit trail; transaksi historis tidak dihapus secara fisik.
- Format resmi nomor registrasi, aturan kelas benih, masa berlaku, dan kontrak SIM Perbenihan masih memerlukan validasi BP3MBTP/Diskominfo.

Implementasi backend, migrasi, dan halaman `/benih` didokumentasikan di [docs/registrasi-penjualan-benih.md](docs/registrasi-penjualan-benih.md).

---

## 🚀 Panduan Pengembangan Integrasi

- **Panduan Sadewa TTE:** [Integration Guide Sadewa](https://wiki.jogjaprov.go.id/en/diskominfo/panduan/panduan-integrasi-sadewa-central-sign)
- **Standar API Pemda DIY:** [Standard API Documentation](https://wiki.jogjaprov.go.id/en/diskominfo/standard-API)
- **SPL / Cakra API:** [SPL Wiki Documentation](https://wiki.jogjaprov.go.id/diskominfo/sistem-penghubung-layanan)

## Menjalankan pengembangan lokal

Persyaratan: Go sesuai `backend/go.mod`, Node.js 20.9 atau lebih baru, npm,
dan Task (`go-task`). Siapkan `.env` di root repository (salin dari
`.env.example`), lalu buka **dua terminal** dari root repository:

```powershell
# Terminal 1 — backend API (task `dev` menjalankan Go API)
task dev
```

```powershell
# Terminal 2 — frontend Next.js
task frontend
```

Frontend lokal tersedia di `http://localhost:3002`, backend di
`http://localhost:8081`, dan pemeriksaan backend di
`http://localhost:8081/health`. Port tersebut dipisah dari Docker Compose,
yang secara default memakai 3000/8080. Task `dev` memuat variabel dari `.env`
root sebelum menjalankan backend. Task `frontend` mengarahkan browser ke backend
lokal melalui `NEXT_PUBLIC_BACKEND_URL`.

Perintah pengembangan tambahan: `task test` untuk unit test backend dan
`task check` untuk `go vet` serta build. Perintah Next.js langsung tetap tersedia
di folder `frontend` (`npm run dev`, `npm run lint`, `npm run typecheck`,
`npm run build`, dan `npm run test:e2e`).

Task `dev` dan `migrate` menyusun DSN PostgreSQL dari `POSTGRES_USER`,
`POSTGRES_PASSWORD`, `POSTGRES_DB`, dan `POSTGRES_HOST_PORT` di `.env`; nilai
`DB_DSN` eksplisit tetap dapat dipakai. Jalankan database baru dengan
`docker compose --env-file .env -f infra/compose.yml up -d database`, terapkan
migrasi memakai `task migrate`, kemudian jalankan `task dev`. PostgreSQL
dipublikasikan pada port host 5433 secara default agar tidak bentrok dengan
instance lokal. Untuk unggah dokumen, MinIO juga harus tersedia dan `MINIO_*`
di `.env` dikonfigurasi.

> Perhatian: konfigurasi PostgreSQL ini membuat volume `postgres_data` baru.
> Volume MySQL lama tidak dihapus, diubah, atau disalin. Migrasi
> data dari MySQL tidak termasuk dalam perpindahan backend ini dan harus
> direncanakan/diuji terpisah. Backend lokal juga tidak akan bisa bind bila
> `DEV_BACKEND_PORT` sedang dipakai; ubah nilainya atau hentikan proses pemilik port.

### Skema Database & Migrasi

Skema PostgreSQL dikelola sebagai berkas SQL terversi di
`backend/migrations/postgres` (modul sertifikasi, identitas SSO,
registrasi/persediaan/penjualan benih, dokumen, import dataset, dan hortikultura).
Pemasangan mandiri:

```powershell
cd backend
DB_DSN="postgres://sitani:password@127.0.0.1:5433/sitani?sslmode=disable" go run ./cmd/migrate -dry-run
DB_DSN="postgres://sitani:password@127.0.0.1:5433/sitani?sslmode=disable" go run ./cmd/migrate
```

Runner menerapkan berkas berurut nama dalam satu transaksi per berkas dan mencatat versi di tabel `schema_migrations`; penerapan idempoten. Jalankan `task migrate` saat database kosong atau setelah menambah migrasi.

Endpoint pemeriksaan kesehatan tersedia di `GET http://localhost:8080/health`.

### SSO Jogjaprov

Backend menerapkan OpenID Connect Authorization Code dengan PKCE untuk realm JSP dan Satriya. Token diverifikasi terhadap issuer, audience, masa berlaku, algoritma RS256, JWKS, dan email terverifikasi. Token sesi disimpan dalam cookie `HttpOnly`; API `/v1/*` menolak identitas dari header client dan hanya memakai identitas hasil verifikasi SSO. Role aplikasi tetap dikelola pada database SiTani agar role dari token eksternal tidak otomatis memperoleh kewenangan lokal.

Konfigurasikan `SSO_JSP_ISSUER`, `SSO_JSP_CLIENT_ID`, `SSO_SATRIYA_ISSUER`, `SSO_SATRIYA_CLIENT_ID`, `BACKEND_PUBLIC_URL`, dan `FRONTEND_PUBLIC_URL` melalui environment. Untuk produksi HTTPS, wajib gunakan `SSO_COOKIE_SECURE=true`. Client publik di masing-masing realm harus mengaktifkan Authorization Code dan PKCE serta mendaftarkan redirect URI berikut:

```text
${BACKEND_PUBLIC_URL}/auth/callback/jsp
${BACKEND_PUBLIC_URL}/auth/callback/satriya
```

Endpoint aplikasi: `GET /auth/login/{jsp|satriya}`, `GET /auth/callback/{realm}`, `GET /auth/me`, dan `POST /auth/logout`. Nilai issuer/client ID di `.env.example` berasal dari pola integrasi proyek referensi dan tetap harus dikonfirmasi serta diuji bersama pengelola SSO Jogjaprov sebelum produksi.

Frontend membutuhkan Node.js 20.9 atau lebih baru:

```powershell
cd frontend
npm ci
npm run dev
```

Pemeriksaan frontend dapat dijalankan dengan `npm run lint`, `npm run typecheck`, `npm run build`, dan `npm run test:e2e`. Pengujian browser memakai Chrome yang terpasang melalui channel Playwright `chrome`.

Salin `.env.example` menjadi `.env` sesuai lingkungan. Jangan menyimpan kredensial atau token di repository.

### Object Storage MinIO (Dokumen/Lampiran)

Dokumen dan lampiran disimpan privat di bucket MinIO/S3; objek tidak dapat diakses langsung dari luar. Unduhan hanya melalui `GET /v1/dokumen/{id}/unduh` yang memeriksa hak akses (pemilik atau petugas internal: admin, pengawas, analis, penandatangan, legalisasi, percetakan) dan menerbitkan URL presign berumur 5 menit.

Endpoint:
- `POST /v1/dokumen` — unggah multipart (`file`; `tujuan` opsional maks. 255 karakter). Batas ukuran default 20 MB (`MINIO_MAX_UPLOAD_MB`), jenis yang diizinkan: PDF, PNG/JPEG, ZIP, CSV, XLS/XLSX, DOC/DOCX; isi berkas diverifikasi terhadap tanda baca jenis untuk menolak header palsu.
- `GET /v1/dokumen` — daftar metadata (pelanggan: miliknya; petugas: seluruhnya).
- `GET /v1/dokumen/{id}/unduh` — URL presign setelah pemeriksaan hak akses.

Konfigurasi: `MINIO_ENDPOINT` (host:port), `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `MINIO_SECURE`, `MINIO_MAX_UPLOAD_MB`. Tanpa konfigurasi, endpoint dokumen menjawab 503 dan `/health` melaporkan `storage: "belum-dikonfigurasi"`.

Untuk pengembangan lokal, layanan MinIO tersedia lewat profil compose:

```powershell
# .env: tambahkan MINIO_ENDPOINT=storage:9000, MINIO_ACCESS_KEY, MINIO_SECRET_KEY
docker compose --env-file .env --profile minio -f infra/compose.yml up --build -d
```

Bucket dibuat otomatis saat backend pertama kali terhubung. Instance MinIO resmi untuk produksi diatur bersama pengelola; kredensial produksi tidak disimpan di repository.

## Menjalankan dengan Docker Compose

Modul Repository Data Lokal tersedia untuk import batch CSV/XLSX. Batasan format, role yang diizinkan, validasi, endpoint, dan spesifikasi yang masih menunggu DPKP dijelaskan di [docs/repository-data.md](docs/repository-data.md).

Salin `.env.example` menjadi `.env`, ganti seluruh password contoh, lalu jalankan dari root repository:

```powershell
docker compose --env-file .env -f infra/compose.yml up --build -d
docker compose --env-file .env -f infra/compose.yml ps
```

Compose menjalankan tiga container utama: PostgreSQL, API backend, dan frontend. Volume baru `postgres_data` menyimpan database development. Volume `mysql_data` lama bukan bagian layanan Compose baru dan tidak disentuh oleh konfigurasi ini. Layanan MinIO untuk pengembangan dapat diaktifkan dengan profil `minio` (lihat bagian Object Storage MinIO di atas).
