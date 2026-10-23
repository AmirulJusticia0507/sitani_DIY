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
- [ ] Testing Unit & Integration Test
- [ ] User Acceptance Testing (UAT) bersama DPKP DIY & Diskominfo DIY
