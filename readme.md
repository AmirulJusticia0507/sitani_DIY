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
