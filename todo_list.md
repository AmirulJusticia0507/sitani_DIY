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
- [x] Perbaikan alur tombol masuk: navigasi penuh (`window.location.assign`) menggantikan `fetch` ber-redirect sehingga pengalihan ke issuer tidak lagi gagal karena CORS dan tidak memunculkan pesan "Server API tidak dapat dihubungi"; galat JSON backend (4xx/5xx) tetap ditampilkan. Tombol aktif sejak awal (tidak tertahan di "Menyiapkan…"); pemeriksaan `/auth/me` dan `/auth/dev/status` hanya menambah keterangan. `allowedDevOrigins: ["127.0.0.1"]` di `next.config.ts` agar aset dev Next.js 16 tidak diblokir saat portal dibuka lewat `127.0.0.1`. Terverifikasi di browser: klik → `GET /auth/login/jsp` dokumen → `302` → halaman issuer; login lokal dev `200` lalu `/dashboard`.
- [x] Pengerasan backend (opsi A dari `diygov-starter-be`): batas body JSON 1 MB per handler (`bacaJSON` di `internal/httpapi/batas_body.go`) dengan balasan `413` yang menyebut batasnya; rute unggah multipart tetap memakai batas ukurannya sendiri; login lokal dev dibatasi 4 KB dan dicek sebelum akses database. Pendekatan middleware global ditolak karena `r.Pattern` kosong sebelum `ServeMux` memilih rute (rute unggah akan ikut terbatas). Diuji unit (`batas_body_test.go`, `devlogin/batas_test.go`) dan `go test ./...` hijau; uji `413` terhadap backend dev berjalan masih tertunda sampai `task dev` dimulai ulang dengan kode ini.
- [x] Tampilan ulang halaman pengajuan baru hortikultura (`/dashboard/horticulture`): kartu berjudul "Registrasi kebun hortikultura" dengan empat langkah bernomor (dokumen, pemohon, lokasi lahan, komoditas), pemilih jenis lampiran berbentuk pil, kotak unggah dengan nama berkas terpilih, daftar lampiran siap kirim (ikon ekstensi, hitungan, hapus), kartu komoditas dinamis (tambah/hapus, minimal satu), placeholder contoh, pilihan kabupaten/kota DIY, dan bar ringkasan sebelum kirim. Kontrak field/payload ke backend tidak berubah. Terverifikasi di browser: tambah komoditas, ganti jenis lampiran (`farm_map`), tanpa galat konsol.
- [x] Pemisahan portal publik dan area kerja: halaman depan tetap landing (hero "Masuk", kartu layanan mengarah ke `/masuk?redirect=…`), login dua tab di `/masuk` (JSP masyarakat/pelaku usaha, Satriya ASN Pemda DIY), dan modul dipindah ke `/dashboard/{repository-data,horticulture,certification,benih}` dengan sidebar peran. Rute lama `/{repository-data,horticulture,certification,benih}` kini 404 tanpa redirect; sidebar publik tinggal Beranda + Masuk area kerja
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
- [x] Form permohonan memakai autocomplete jenis benih dan varietas: kolom varietas menampilkan saran yang mengikuti jenis benih yang dipilih, dengan isian bebas untuk varietas/jenis di luar daftar (`frontend/components/autocomplete.tsx`, `frontend/lib/jenis-benih.ts`); jenis benih di luar daftar hanya boleh disimpan sebagai draft dan ditolak saat diajukan, dan varietas dikosongkan bila jenis diganti
- [x] Form registrasi benih memakai autocomplete komoditas dan varietas dengan penyaringan bertingkat, nilai tetap terkirim lewat `FormData` (`frontend/app/dashboard/benih/page.tsx`); data saran dipindah ke `frontend/lib/jenis-benih.ts` agar dipakai bersama
- [x] Galat peran pada alur permohonan pelanggan dijawab `403` (bukan `400`) lewat `service.ErrTidakBerhak`/`ErrPeran`, dan form permohonan tidak memuat `/v1/permohonan/aktif` untuk peran non-pelanggan sehingga kartu galat tidak muncul di dashboard peran internal (`TestAlurDraftDitolakUntukRoleInternal`)
- [ ] Validasi daftar resmi jenis/komoditas dan varietas benih: daftar saat ini adalah pilihan praktis (varietas populer DIY), bukan master resmi BP3MBTP; perlu konfirmasi daftar komoditas, varietas terdaftar, dan kelas benih sebelum dipakai sebagai validasi. Penolakan jenis di luar daftar saat ini hanya berlaku di frontend; backend masih menerima isian bebas
- [ ] Validasi format resmi nomor registrasi/dokumen, kelas benih, masa berlaku, dan kontrak SIM Perbenihan bersama BP3MBTP/Diskominfo
- [ ] UAT alur dokumen: unggah berkas asli ke MinIO, lalu buka dari akun peran lain untuk membuktikan pemeriksaan hak akses (belum dapat diuji tanpa instance MinIO siap)

## 📍 Sprint November 2026: Formulir Modul & CRUD Profil
Rencana kerja 1–30 November 2026. Setiap task di bawah masih berupa rencana: diperlukan backend yang sudah berjalan (`task dev`), sesi uji per peran, dan instance MinIO untuk unggah dokumen. Task hanya ditandai selesai setelah implementasi, pemeriksaan (`tsc`, `oxlint`, tes perilaku) dijalankan, dan commit terfokus dipush.

### Pekan 1 (1–7 November): Pondasi formulir bersama
- [ ] **Komponen formulir bersama** — buat `frontend/components/form/` berisi `FormField` (label, keterangan, wajib, galat per kolom), `FormSection`, `FormActions` (simpan draf/kirim, status mengirim), dan `FormAlert`; pakai pola visual yang sama dengan form pengajuan hortikultura (`pengajuan-bagian`, `hortikultura-grid`) agar seluruh modul konsisten. Kriteria selesai: dipakai minimal satu halaman baru, `tsc`/`oxlint` bersih, tanpa galat konsol.
- [ ] **Utilitas validasi & format** — `frontend/lib/validasi.ts` untuk pesan galat Bahasa Indonesia yang konsisten (wajib isi, panjang maksimum, angka/desimal, tanggal), plus `formatTanggal`/`formatAngka` untuk tampilan. Kriteria selesai: unit test perilaku fungsi murni tersedia atau bukti uji manual terdokumentasi.
- [ ] **Pola pengambilan data per modul** — helper `apiFetch` dengan penanganan 401/403/413/503 yang seragam dan pesan dari backend ditampilkan apa adanya; ganti pola `fetch` berulang di modul baru. Kriteria selesai: galat 413 (batas body 1 MB) dan 403 tampil jelas di formulir, bukan pesan umum.
- [ ] **Pemeriksaan akhir pekan 1** — jalankan `npm run typecheck`, `npm run lint`, dan uji browser formulir yang sudah ada (`/dashboard/horticulture`, `/dashboard/certification`) untuk memastikan komponen bersama tidak merusak tampilan lama.

### Pekan 2 (8–14 November): Uji Mutu Lab
- [ ] **Halaman uji mutu lab** (`/dashboard/certification?tab=lab` atau rute `/dashboard/laboratorium`) — formulir input hasil uji sesuai `POST /v1/uji-mutu`: pilih sampel dari daftar (bukan mengetik ID), parameter uji, hasil per parameter, satuan, metode, dan kesimpulan lulus/tidak lulus; tampilkan ringkasan permohonan yang terkait.
- [ ] **Formulir kodifikasi sampel** — UI `POST /v1/sampel/{id}/kodifikasi` bagi analis (kode pengujian, tanggal terima), dengan pratinjau kode yang dihasilkan server.
- [ ] **Formulir laporan mutu** — UI `POST /v1/laporan-mutu`: tarik sampel dan hasil uji yang sudah ada, susun narasi laporan, pratinjau sebelum kirim.
- [ ] **Pemisahan akses analis** — pastikan peran `analis` hanya melihat tab lab (bukan verifikasi lapangan), dan percobaan akses langsung dijawab 403 dengan pesan yang jelas di UI.
- [ ] **Uji uji-mutu end-to-end lokal** — sampel → kodifikasi → hasil uji → laporan mutu dengan akun `analis@sitani.local`; dokumentasikan langkah dan bukti status HTTP. Catatan: perlakuan lulus/gagal memicu notifikasi (Discord) — arahkan ke webhook uji atau matikan kirim nyata.

### Pekan 3 (15–21 November): Penerbitan dokumen & penjualan benih
- [ ] **Halaman penerbitan dokumen** — form sertifikat (`POST /v1/sertifikat`), TTE (`/v1/sertifikat/{permohonan_id}/tte`), draft label (`POST /v1/draft-label`), integrasi SIM (`/sim`), cetak (`/cetak`), dan legalisasi (`/legalisasi`), ditata sebagai langkah bertahap dengan status per langkah (Hormati peran: penandatangan, legalisasi, percetakan).
- [ ] **Prasyarat percabangan lulus** — UI hanya menampilkan langkah penerbitan bila permohonan dinyatakan lulus; untuk jalur tidak lulus tampilkan tautan laporan lengkap (bukan form penerbitan).
- [ ] **Form penjualan benih** (`POST /v1/penjualan-benih`) — pilih lot dari daftar stok tersedia (bukan ID bebas), kuantitas tervalidasi terhadap stok, harga, nomor dokumen, tanggal; cegah kirim bila lot belum punya registrasi/sertifikat/label sah.
- [ ] **Pembatalan penjualan** — dialog konfirmasi `POST /v1/penjualan-benih/{id}/batal` dengan alasan wajib; tampilkan status transaksi pembalik dan jejak audit.
- [ ] **Rekap penjualan** — tabel `GET /v1/penjualan-benih` dengan penyaring tanggal/lot dan ringkasan total.

### Pekan 4 (22–30 November): Laporan, dokumen, dan profil & akses
- [ ] **Halaman laporan** — satu halaman dengan tab Registrasi, Stok, Kedaluwarsa, dan Penjualan (`GET /v1/laporan/*`), penyaring rentang tanggal, ringkasan angka, dan tombol unduh CSV/PDF (bila format ekspor belum ada di backend, tandai sebagai pekerjaan backend terpisah, jangan menampilkan tombol palsu).
- [ ] **Papan status dokumen** (`/dashboard/dokumen`) — gabungkan daftar (`GET /v1/dokumen`), unggah dengan batas ukuran/tipe jelas, status kepemilikan, dan penanganan 413 dari server. Tergantung MinIO siap.
- [ ] **Profil & akses — baca** — rapikan halaman `/dashboard/profil` menjadi kartu identitas, ringkasan modul sesuai peran, dan riwayat sesi (bila tersedia dari backend); jangan menampilkan token atau data pribadi yang tidak diperlukan.
- [ ] **Profil & akses — CRUD pengguna (khusus admin/superadmin)** — halaman kelola pengguna: daftar (cari/filter peran), tambah, ubah peran/status, nonaktifkan, dan reset password dev (khusus lingkungan lokal). **Butuh endpoint backend baru** (`GET/POST/PATCH/DELETE /v1/pengguna`) yang belum ada: kerjakan migrasi, otorisasi admin, audit trail, dan larangan naikkan peran sendiri sebagai satu paket dengan UI.
- [ ] **Profil & akses — perubahan data diri mandiri** — hanya untuk data yang dimiliki portal (bukan identitas SSO): ajukan permintaan perubahan (nama tampilan, kontak) dengan status menunggu persetujuan admin; identitas inti tetap dari JSP/Satriya. **Butuh kontrak backend baru**; jangan menjanjikan perubahan email/peran instan.
- [ ] **Review sprint & dokumentasi** — perbarui `todo_list.md` dan `docs/`, catat task yang tertunda beserta alasan teknis (mis. menunggu MinIO/Keycloak/kontrak SPL), lalu jalankan pemeriksaan penuh sebelum penutupan bulan.

### Risiko & dependensi Sprint November
- Endpoint CRUD pengguna, perubahan data diri, dan ekspor laporan **belum ada di backend**; UI tidak boleh menampilkan tombol yang tidak berfungsi sebelum endpoint tersedia.
- MinIO harus berjalan untuk alur dokumen; SSO Jogjaprov (`client portal`) dan issuer lokal (`127.0.0.1:8085`) masih terblokir dari browser host.
- Notifikasi Discord dan tanda tangan TTE pada alur penerbitan memakai layanan eksternal: gunakan mode uji, jangan kirim pesan atau tanda tangan sungguhan tanpa persetujuan DPKP/Diskominfo.

