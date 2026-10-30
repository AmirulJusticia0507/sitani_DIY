# TODO List Pengembangan Portal Pertanian & Ketahanan Pangan DIY

## 📍 Tahap 1: Setup Lingkungan & Arsitektur dasar
- [x] Analisis kondisi awal dan pedoman kerja agen (`docs/analisis-proyek.md`, `AGENTS.md`)
- [x] Setup Repository & Struktur Proyek (Backend Go 1.23 & Frontend Next.js 16; aktual Next.js 16 karena dukungan keamanan)
- [x] Verifikasi fondasi: backend `go build`/`go vet`/`go test` lulus; endpoint `GET /health` tersedia; frontend `typecheck`/`lint`/`build` lulus
- [x] Konfigurasi Docker / Containerization (3 Container: DB, BE, FE)
- [x] Implementasi teknis SSO Jogjaprov: OIDC Authorization Code + PKCE, verifikasi JWT/JWKS, cookie HttpOnly, pemetaan user lokal, dan proteksi API (Realm JSP & Satriya)
- [ ] Aktivasi client, konfirmasi kontrak resmi, dan UAT login langsung bersama pengelola SSO Jogjaprov (JSP & Satriya) — terblokir menunggu aktivasi/kredensial dan jadwal UAT dari pengelola; implementasi teknis lokal bukan bukti UAT.
- [x] Login lokal development/UAT (di luar produksi, flag `SITANI_DEV_LOGIN`): argon2id/bcrypt, cookie sesi ber-hash, seed akun per peran, halaman `/dashboard/masuk`, dan pematian otomatis di produksi (lihat `docs/login-lokal-development.md`). Belum ada reset password dan rate limit sehingga bukan jalur masuk produksi.
- [x] Halaman masuk bertema SSO Pemda DIY (`/masuk`) menggantikan tautan SSO langsung ke backend, serta sidebar portal yang dapat diciutkan (rail ikon, preferensi tersimpan) dan menjadi drawer di layar sempit. Klik tombol login terbukti berpindah ke issuer (`302` backend → `/realms/{jsp,satriya}`); UAT login sungguhan tetap menunggu aktivasi SSO Jogjaprov.
- [x] SSO lokal dengan Keycloak terpisah untuk development/UAT (`infra/keycloak/compose.yml`, realm `sitani`, client confidential `sitani-portal` + PKCE, 5 akun uji, skrip idempoten `configure-realm.ps1`); backend mendukung client secret dan memvalidasi `azp` (bukan `aud`). Alur login browser terbukti end-to-end (`/auth/me` → `realm: jsp`, sesi bertahan setelah reload). Realm ini bukan identitas produksi dan tidak menggantikan UAT Jogjaprov (lihat `docs/sso-keycloak-lokal.md`).
- [x] Halaman `/dashboard/dokumen` menggantikan tautan menu **Dokumen** yang sebelumnya menunjuk ke endpoint API `/v1/dokumen` (menyebabkan `404` saat navigasi). Daftar dan tautan unduh memakai `GET /v1/dokumen`; tamu diarahkan masuk. Unggah baru masih menunggu MinIO dikonfigurasi.
- [x] Siklus sesi SSO: perpanjangan access token lewat refresh token (cookie HttpOnly terbatas path `/auth/`, ambang `SSO_REFRESH_SKEW_SECONDS`, toleransi jam `SSO_TOKEN_LEEWAY_SECONDS`), endpoint `POST /auth/refresh`, dan logout yang mencabut token lalu mengarahkan ke `end_session_endpoint` (`id_token_hint`, `sid`). Terbukti nyata terhadap Keycloak 26.0 (refresh `200`, logout `200`, refresh setelah logout `401`, `Origin` asing `403`). Catatan: refresh lintas host issuer (`localhost` vs `host.docker.internal`) ditolak Keycloak — samakan issuer atau setel `KC_HOSTNAME`. UAT logout bersama Jogjaprov menunggu kebijakan session/logout pengelola.
- [ ] Setup Integrasi SPL / Cakra API Gateway (`spl.jogjaprov.go.id`) — fondasi transport JSON generik tersedia; endpoint, autentikasi, dan operasi bisnis menunggu kontrak resmi (lihat `docs/integrasi-spl-cakra.md`).
- [x] Setup Database MySQL & Skema Dasar (migrasi terversi 001-003: modul sertifikasi, sso_identities, registrasi/persediaan/penjualan benih; runner `backend/cmd/migrate` sudah diuji ke MySQL 8.4 dan idempoten)
- [x] Backend PostgreSQL baru untuk development: driver pgx, repository/query, migrasi PostgreSQL 001-007, konfigurasi Compose/task dan dokumentasi; diuji pada PostgreSQL 17 kosong dan migrasi kedua idempoten. Volume/data MySQL lama tidak diubah atau dipindahkan.
- [x] Integrasi MinIO Object Storage untuk dokumen/lampiran (adaptor `internal/storage`, metadata `dokumen` migrasi 004, endpoint `/v1/dokumen` dengan pemeriksaan hak akses + URL presign 5 menit, profil compose `minio` untuk dev; diuji unggah/presign ke MinIO container lokal — instance produksi menunggu pengelola)

## 📍 Tahap 2: Modul Repository Data Lokal
- [x] Fondasi import repository lokal CSV/XLSX: parser/header validation, batas ukuran/baris, batch tersimpan di MySQL, role admin/analis dan tes; template/mapping kanonis menunggu ketetapan DPKP
- [ ] Development Rest API data terintegrasi — sinkronisasi/API sumber eksternal menunggu kontrak resmi; batch import lokal tersedia, belum ada model/endpoint baca baris kanonis
- [x] Kerangka dashboard area kerja: layout segmen `app/dashboard` dengan header, sidebar sesuai peran (peta `components/dashboard/role-access.ts`), footer ringkas terpisah dari portal publik, halaman Ringkasan dan Profil & Akses; sesi dibaca dari `GET /auth/me`, tanpa angka statistik karena endpoint agregat belum ada
- [ ] Development Dashboard & Visualisasi Data Pertanian — belum ada endpoint agregat/visualisasi; menunggu model data kanonis dan kontrak sumber eksternal

## 📍 Tahap 3: Modul Layanan Registrasi Lahan & Pohon Induk (Hortikultura)
- [x] Form Pengajuan Permohonan Registrasi Lahan (data pemohon/lahan/tanaman dan lampiran privat proposal, struktur organisasi, daftar anggota, form kebun, peta; API simpan, akses pemohon, audit dan migrasi 006)
- [x] Modul Verifikasi Proposal (Petugas Verifikator; kelengkapan lampiran wajib, keputusan/perbaikan/penolakan, audit; role diaktifkan via migrasi 007)
- [x] Modul Jadwal & Form Inspeksi Lapangan / Audit GAP (Petugas Inspektor; jadwal, 11 indikator, temuan/rekomendasi dan audit)
- [x] Modul Form Penilaian Pohon Induk / PIT / RIP / PI / RI (PBT & Kasi PMBTP; parameter skor 1–5, rekomendasi, riwayat dan audit)
- [ ] Integrasi TTE Sadewa untuk pengesahan & penerbitan Surat/Sertifikat Nomor Registrasi

## 📍 Tahap 4: Notifikasi, Logging & Quality Assurance
- [x] Integrasi Discord Webhook Notifikasi
- [x] Fondasi proxy asisten AI: adaptor `internal/ai` (OpenAI-compatible), endpoint `POST /v1/ai/chat` terautentikasi SSO, batas pesan/laju, dan tes dengan server palsu — panggilan langsung & pemilihan model menunggu `AI_API_KEY` serta evaluasi kualitas
- [x] Integrasi Sentry Error Logging
- [x] Dokumentasi anticipasi kerentanan Node.js & PHP Oktober 2026 (`docs/keamanan-kerentanan-nodejs-php-oktober-2026.md`)
- [x] Fondasi PWA (manifest, ikon, service worker terbatas ke aset publik) serta halaman Tentang, Kebijakan Privasi (draf), dan Bantuan
- [ ] Testing Unit & Integration Test — cakupan unit backend berjalan (paket `auth`, `dataset`, `db`, `httpapi`, `service`, `spl`; `go test ./...` lulus). Tes E2E frontend (Playwright, desktop + mobile) mencakup portal publik dan kerangka dashboard (`tests/dashboard.spec.ts`). Modul `horticulture`, `notify`, `observability`, `storage`, `store`, dan `cmd/*` belum memiliki tes; tes integrasi ke PostgreSQL/MinIO serta CI otomatis belum tersedia.
- [ ] User Acceptance Testing (UAT) bersama DPKP DIY & Diskominfo DIY

## 📍 Tahap 5: Modul Sertifikasi Benih BP3MBTP (Alur XML 7 Aktor)
- [x] Implementasi Swimlane Roles & RBAC (Pelanggan, Admin, Pengawas/PBT, Analis Lab, Penandatangan TTE, Legalisasi, Percetakan)
- [x] Fase 1 Verifikasi Lapangan: Permohonan & Pembayaran Awal, Verifikasi Dokumen & Inspeksi 4 Fase, Pencatatan Buku Induk & Laporan Verifikasi
- [x] Fase 2 Pengujian Lab Mutu: Pengambilan Sampel oleh Pengawas, Kodifikasi Sampel oleh Analis, Pengujian Lab, Penyusunan Laporan Mutu
- [x] Modul Manajemen Uji Mutu Lab: Pemberian Kode Pengujian Sampel, Input Hasil Uji, Laporan Mutu
- [x] Logic Percabangan Lulus / Tidak Lulus: Auto-trigger Sertifikat & Label vs Notifikasi Gagal ke Pelanggan
- [x] Fase 3 Sertifikasi & Penerbitan Label (Jalur Lulus): Pembuatan Sertifikat Benih, Penandatanganan Pejabat (TTE), Draft Label & Integrasi SIM Perbenihan, Cetak di Percetakan, Legalisasi Label, Penyerahan Sertifikat/Label & Pembayaran Akhir
- [x] Jalur Tidak Lulus: Pengiriman Laporan Lengkap Pengujian ke Pelanggan (Proses selesai)
- [x] Modul Draft Label, Kode Unik/Barcode SIM Perbenihan, Legalisasi Label, dan Penyerahan Dokumen
- [x] Registrasi benih: master komoditas/varietas, produsen, sumber benih, lot, kelas, jumlah/satuan, masa berlaku, dan status edar
- [x] Persediaan benih per lot dengan mutasi stok yang transaksional dan dapat diaudit
- [x] Penjualan benih: pembeli, lot, kuantitas, harga, nomor dokumen, tanggal transaksi, dan pengurangan stok
- [x] Validasi penjualan hanya untuk lot dengan registrasi, sertifikat, dan label yang sah
- [x] Koreksi/pembatalan penjualan melalui transaksi pembalik dan audit trail tanpa menghapus histori
- [x] Laporan registrasi, stok, masa berlaku, dan penjualan benih
- [x] Nomor permohonan dibuat server dengan format `SPB-YYYYMMDD-XXXX` dan tersimpan di database
- [x] Draft permohonan: pelanggan dapat menyimpan draf, mengeditnya, lalu mengajukannya; langkah Analis Lab dan Penandatangan baru terbuka setelah permohonan diajukan (migrasi 011, `POST /v1/permohonan/draft`, `POST /v1/permohonan/ajukan`, `GET /v1/permohonan/aktif`)
- [ ] Unggahan dokumen permohonan Fase 1 (proposal, struktur organisasi, daftar anggota, form registrasi kebun, peta kebun) pada form sertifikasi
- [x] Form registrasi benih memilih permohonan (bukan mengetik ID permohonan)
- [x] Form permohonan memakai autocomplete jenis benih dan varietas: kolom varietas menampilkan saran yang mengikuti jenis benih yang dipilih, dengan isian bebas untuk varietas/jenis di luar daftar (`frontend/components/autocomplete.tsx`, `frontend/app/certification/jenis-benih.ts`)
- [ ] Validasi daftar resmi jenis/komoditas dan varietas benih: daftar saat ini adalah pilihan praktis (varietas populer DIY), bukan master resmi BP3MBTP; perlu konfirmasi daftar komoditas, varietas terdaftar, dan kelas benih sebelum dipakai sebagai validasi
- [ ] Validasi format resmi nomor registrasi/dokumen, kelas benih, masa berlaku, dan kontrak SIM Perbenihan bersama BP3MBTP/Diskominfo
