# TODO List Pengembangan Portal Pertanian & Ketahanan Pangan DIY

## 📍 Tahap 1: Setup Lingkungan & Arsitektur dasar
- [x] Analisis kondisi awal dan pedoman kerja agen (`docs/analisis-proyek.md`, `AGENTS.md`)
- [x] Setup Repository & Struktur Proyek (Backend Go 1.23 & Frontend Next.js; aktual Next.js 16 karena dukungan keamanan)
- [ ] Konfigurasi Docker / Containerization (3 Container: DB, BE, FE)
- [ ] Integrasi SSO Jogjaprov (Realm JSP untuk Umum & Realm Satriya untuk ASN)
- [ ] Setup Integrasi SPL / Cakra API Gateway (`spl.jogjaprov.go.id`)
- [ ] Setup Database MySQL & Skema Dasar
- [ ] Integrasi MinIO Object Storage untuk dokumen/lampiran

## 📍 Tahap 2: Modul Repository Data Lokal
- [ ] Implementasi Service Import Data (Excel/CSV Parser & Mapper)
- [ ] Development Rest API data terintegrasi
- [ ] Development Dashboard & Visualisasi Data Pertanian

## 📍 Tahap 3: Modul Layanan Registrasi Lahan & Pohon Induk (Hortikultura)
- [ ] Form Pengajuan Permohonan Registrasi Lahan (Upload Proposal, Peta Kebun, Detail Tanaman)
- [ ] Modul Verifikasi Proposal (Petugas Verifikator)
- [ ] Modul Jadwal & Form Inspeksi Lapangan / Audit GAP (Petugas Inspektor)
- [ ] Modul Form Penilaian Pohon Induk / PIT / RIP / PI / RI (PBT & Kasi PMBTP)
- [ ] Integrasi TTE Sadewa untuk pengesahan & penerbitan Surat/Sertifikat Nomor Registrasi

## 📍 Tahap 4: Notifikasi, Logging & Quality Assurance
- [ ] Integrasi Discord Webhook Notifikasi
- [ ] Integrasi Sentry Error Logging
- [x] Dokumentasi anticipasi kerentanan Node.js & PHP Oktober 2026 (`docs/keamanan-kerentanan-nodejs-php-oktober-2026.md`)
- [ ] Testing Unit & Integration Test
- [ ] User Acceptance Testing (UAT) bersama DPKP DIY & Diskominfo DIY

## 📍 Tahap 5: Modul Sertifikasi Benih BP3MBTP (Alur XML 7 Aktor)
- [ ] Implementasi Swimlane Roles & RBAC (Pelanggan, Admin, Pengawas/PBT, Analis Lab, Penandatangan TTE, Legalisasi, Percetakan)
- [ ] Fase 1 Verifikasi Lapangan: Permohonan & Pembayaran Awal, Verifikasi Dokumen & Inspeksi 4 Fase, Pencatatan Buku Induk & Laporan Verifikasi
- [ ] Fase 2 Pengujian Lab Mutu: Pengambilan Sampel oleh Pengawas, Kodifikasi Sampel oleh Analis, Pengujian Lab, Penyusunan Laporan Mutu
- [ ] Modul Manajemen Uji Mutu Lab: Pemberian Kode Pengujian Sampel, Input Hasil Uji, Laporan Mutu
- [ ] Logic Percabangan Lulus / Tidak Lulus: Auto-trigger Sertifikat & Label vs Notifikasi Gagal ke Pelanggan
- [ ] Fase 3 Sertifikasi & Penerbitan Label (Jalur Lulus): Pembuatan Sertifikat Benih, Penandatanganan Pejabat (TTE), Draft Label & Integrasi SIM Perbenihan, Cetak di Percetakan, Legalisasi Label, Penyerahan Sertifikat/Label & Pembayaran Akhir
- [ ] Jalur Tidak Lulus: Pengiriman Laporan Lengkap Pengujian ke Pelanggan (Proses selesai)
- [ ] Modul Draft Label, Kode Unik/Barcode SIM Perbenihan, Legalisasi Label, dan Penyerahan Dokumen
