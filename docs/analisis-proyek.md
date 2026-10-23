# Analisis kondisi proyek

Tanggal pemeriksaan: 23 Oktober 2026.

## Ringkasan

Proyek bermula pada tahap perencanaan. Saat pemeriksaan awal, direktori hanya berisi `readme.md`, `analisis_kebutuhan.md`, dan `todo_list.md`. Fondasi backend dan frontend kini telah tersedia: API Go dengan health endpoint teruji serta halaman awal Next.js dengan lint, typecheck, build, dan smoke test browser. Skema database, container, serta integrasi eksternal belum tersedia.

## Kebutuhan yang sudah tersedia

- Dua domain utama: repository data lokal serta layanan registrasi hortikultura.
- Tahapan layanan: pengajuan, verifikasi, inspeksi/audit GAP, evaluasi, dan penerbitan dokumen dengan TTE.
- Kelompok pengguna masyarakat, pelaku usaha/organisasi, ASN, dan internal OPD.
- Stack dan sistem integrasi pemerintah telah disebutkan.

## Kekurangan spesifikasi dan tindak lanjut

| Area | Kondisi saat ini | Tindak lanjut |
| --- | --- | --- |
| Hak akses | Daftar kategori pengguna, belum matriks izin | Definisikan pemohon, verifikator, inspektor, evaluator, penandatangan, admin; petakan izin dan lingkup organisasi |
| Workflow | Alur sukses tersedia | Tentukan revisi, penolakan, pembatalan, pengajuan ulang, kewenangan transisi, dan SLA |
| Data lokal | Import Excel/CSV disebutkan | Tetapkan dataset, template, unit, periode, referensi wilayah, aturan duplikasi, dan pelaporan baris gagal |
| Registrasi | Daftar dokumen masih umum | Definisikan field wajib, jenis layanan, validasi lahan, checklist audit, dan format nomor registrasi |
| SSO | Nama realm tersedia | Dapatkan issuer, client, redirect URI, discovery, claims dan pemetaan akun dari pengelola |
| SPL/Sadewa | Referensi dokumentasi tersedia | Verifikasi kontrak, sandbox, autentikasi, callback, retry dan idempotensi |
| Penyimpanan | MinIO dipilih | Tentukan bucket privat, batas file, retensi, backup, dan akses unduhan |
| Operasional | Logging disebutkan | Definisikan health/readiness, redaksi log, backup/restore, dan parameter deployment |
| QA | Tahap testing/UAT tersedia | Tetapkan acceptance criteria per task dan skenario lintas peran |

Kategori pengguna tidak otomatis menjadi role operasional. Contohnya, ASN DIY tidak otomatis berhak memverifikasi atau menandatangani permohonan.

## Arsitektur yang disarankan

Gunakan monorepo dengan backend modular dan frontend terpisah. Modul awal: identitas/akses, katalog layanan, permohonan, dokumen, workflow/audit, dataset/import, dan integrasi. Backend menjadi sumber validasi dan otorisasi. Frontend mengakses API melalui konfigurasi gateway sesuai lingkungan.

Tiga container pada dokumen awal merujuk pada komponen utama aplikasi. MinIO tetap membutuhkan instance eksternal atau container tambahan untuk pengembangan. SSO dan SPL memiliki peran berbeda: SSO menangani identitas; SPL menangani routing/integrasi API. Urutan diagram kebutuhan tidak cukup untuk menetapkan protokol autentikasi.

Entitas kandidat untuk desain skema: pengguna, organisasi, keanggotaan, role/izin, layanan, permohonan, lahan, tanaman, lampiran, riwayat status, inspeksi, penilaian, sertifikat, dataset, batch import, dan audit event. Skema final mengikuti aturan bisnis yang telah divalidasi.

## Urutan pengerjaan

1. Dokumentasikan aturan agen dan kondisi awal; inisialisasi Git.
2. Siapkan struktur backend/frontend dan endpoint health dengan build yang dapat direproduksi.
3. Siapkan environment example, container, MySQL, dan mekanisme migrasi.
4. Definisikan kontrak akses/workflow; implementasikan adaptor SSO setelah konfigurasi resmi tersedia.
5. Implementasikan satu alur vertikal permohonan sampai verifikasi beserta penyimpanan dokumen dan audit.
6. Lanjutkan inspeksi, evaluasi, dan integrasi TTE melalui sandbox.
7. Implementasikan dataset/import dan dashboard berdasarkan template data yang tersedia.
8. Lengkapi observabilitas, pengujian integrasi, backup/restore, dan UAT.

Task yang membutuhkan layanan eksternal harus ditandai terblokir sampai konfigurasi tersedia, sementara pekerjaan lokal yang independen dapat dilanjutkan.

## Kriteria verifikasi

- Fondasi: backend/frontend dapat dibangun; health endpoint teruji; konfigurasi terdokumentasi.
- Workflow: transisi valid berhasil, transisi ilegal dan akses lintas pemohon ditolak, audit tercatat.
- Import: data valid diterima, kesalahan menunjuk baris/field, dan strategi duplikasi diuji.
- Dokumen: unggah/unduh mengikuti hak akses dan kegagalan storage ditangani.
- Integrasi: bukti pengujian sandbox mencakup sukses, timeout, retry, dan kegagalan autentikasi.
- UI: halaman benar-benar dirender dan alur utama diverifikasi lewat browser.

## Dependensi eksternal yang belum tersedia

Pada pemeriksaan lanjutan, repository Git telah tersedia dengan commit awal dan remote `https://github.com/AmirulJusticia0507/sitani_DIY.git`. Kredensial dan konfigurasi SSO, SPL, Sadewa, MinIO, serta dataset contoh belum terdapat dalam proyek. Ini membatasi verifikasi integrasi nyata, bukan pekerjaan dokumentasi atau fondasi lokal. Baseline Next.js 15 dinaikkan ke Next.js 16 karena patch terbaru mayor 15 masih membawa advisory dependency; audit fondasi Next.js 16 bersih pada 23 Oktober 2026.

## Rencana Pemisahan Repository (Backend & Frontend)

Secara saat ini proyek masih dalam bentuk monorepo dengan struktur `backend/` dan `frontend/` di dalam satu repository. Rencana pemisahan akan dilakukan setelah seluruh fitur inti selesai dan stabil, dengan tujuan masing-masing direktori menjadi repository terpisah:

- **Repository Backend:** berisi `backend/` (Go 1.23, database, migrasi, integrasi SSO/SPL/Sadewa/MinIO)
- **Repository Frontend:** berisi `frontend/` (Next.js 16, UI, komponen, tes browser)

Langkah pemisahan:
1. Pastikan semua fitur inti selesai dan di-commit pada branch `main`.
2. Ekstrak history Git masing-masing subdirektori agar tidak menyisipkan commit yang tidak terkait (mis. menggunakan `git filter-repo` atau pendekatan setara).
3. Buat dua repository baru di remote, satu untuk backend dan satu untuk frontend.
4. Update `.env.example`, `readme.md`, dan `infra/compose.yml` agar referensi path dan build menyesuaikan struktur terpisah.
5. Dokumentasikan cara menjalankan masing-masing repository secara independen maupun bersama melalui Docker Compose.

Pemisahan belum dilakukan saat ini untuk meminimalkan risiko gangguan durante pengembangan aktif.
