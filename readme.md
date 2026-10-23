# Portal Pertanian dan Ketahanan Pangan DIY

Portal Pertanian dan Ketahanan Pangan DIY adalah platform terpadu yang dikembangkan untuk Dinas Pertanian dan Ketahanan Pangan (DPKP) Pemda DIY bekerja sama dengan Dinas Komunikasi dan Informatika DIY. Portal ini berfungsi sebagai repository data lokal dan pusat layanan publik di bidang pertanian dan ketahanan pangan.

---

## 🛠 Tech Stack

> Status: tahap perencanaan; stack di bawah merupakan target implementasi.
> Analisis kondisi aktual: [docs/analisis-proyek.md](docs/analisis-proyek.md).
> Pedoman pengerjaan agen: [AGENTS.md](AGENTS.md).

- **Backend:** Golang 1.23
- **Frontend:** Next.js 16 (baseline kebutuhan Next.js 15 dinaikkan karena dukungan keamanan)
- **Database:** MySQL
- **Object Storage:** MinIO
- **TTE / Central Sign:** Sadewa Integrasi
- **API Gateway:** SPL / Cakra API (`spl.jogjaprov.go.id`)
- **Authentication:** SSO Jogjaprov (Realm JSP & Satriya)
- **Monitoring & Log:** Sentry & Discord Webhook

---

## 🏗 Arsitektur Container

Aplikasi dijalankan dalam 3 container utama:
1. **Container 1:** Database Server (MySQL)
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

Alur sertifikasi benih mengikuti 3 fase dan 7 aktor swimlane:

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

---

## 🚀 Panduan Pengembangan Integrasi

- **Panduan Sadewa TTE:** [Integration Guide Sadewa](https://wiki.jogjaprov.go.id/en/diskominfo/panduan/panduan-integrasi-sadewa-central-sign)
- **Standar API Pemda DIY:** [Standard API Documentation](https://wiki.jogjaprov.go.id/en/diskominfo/standard-API)
- **SPL / Cakra API:** [SPL Wiki Documentation](https://wiki.jogjaprov.go.id/diskominfo/sistem-penghubung-layanan)

## Menjalankan fondasi lokal

Backend membutuhkan Go 1.23 atau lebih baru:

```powershell
cd backend
go test ./...
go run ./cmd/api
```

Endpoint pemeriksaan kesehatan tersedia di `GET http://localhost:8080/health`.

Frontend membutuhkan Node.js 20.9 atau lebih baru:

```powershell
cd frontend
npm ci
npm run dev
```

Pemeriksaan frontend dapat dijalankan dengan `npm run lint`, `npm run typecheck`, `npm run build`, dan `npm run test:e2e`. Pengujian browser memakai Chrome yang terpasang melalui channel Playwright `chrome`.

Salin `.env.example` menjadi `.env` sesuai lingkungan. Jangan menyimpan kredensial atau token di repository.

## Menjalankan dengan Docker Compose

Salin `.env.example` menjadi `.env`, ganti seluruh password contoh, lalu jalankan dari root repository:

```powershell
docker compose --env-file .env -f infra/compose.yml up --build -d
docker compose --env-file .env -f infra/compose.yml ps
```

Compose menjalankan tiga container utama: MySQL, API backend, dan frontend. Volume `mysql_data` mempertahankan data database. MinIO tetap merupakan dependensi eksternal dan belum dimasukkan sebelum task integrasinya dikerjakan.
