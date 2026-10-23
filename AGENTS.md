# Pedoman Agen — SiTani DIY

## Konteks proyek
- Portal Pertanian dan Ketahanan Pangan DPKP DIY.
- Sumber kebutuhan: `analisis_kebutuhan.md`; daftar pekerjaan: `todo_list.md`.
- Stack yang direncanakan: Go 1.23, Next.js 15, MySQL, MinIO, SSO JSP/Satriya, SPL/Cakra API, Sadewa, Sentry, dan Discord.
- Periksa implementasi aktual sebelum mengklaim fitur tersedia. Dokumen rencana bukan bukti implementasi.

## Alur kerja wajib
1. Baca pedoman ini, kebutuhan, TODO, dan perubahan lokal sebelum bekerja.
2. Kerjakan task berurutan berdasarkan dependensi dan prioritas. Setelah task selesai, langsung lanjutkan task berikutnya yang dapat dikerjakan tanpa meminta konfirmasi rutin. Pengguna telah memberikan persetujuan untuk alur kerja tersebut.
3. Pecah pekerjaan menjadi task yang terbatas, dapat diverifikasi, dan memiliki kriteria selesai yang jelas. Pertahankan perubahan milik pengguna.
4. Implementasikan fungsi secara utuh sesuai cakupan task, termasuk validasi, penanganan kesalahan, dan dokumentasi penggunaan. Jangan menandai placeholder, mock, atau integrasi tanpa kredensial sebagai fitur produksi yang selesai.
5. Jalankan pemeriksaan yang relevan sebelum menyelesaikan task: format, lint/typecheck, build, serta tes perilaku yang terdampak jika tersedia. Untuk perubahan UI, verifikasi melalui browser. Catat pemeriksaan yang tidak dapat dijalankan beserta penyebabnya.
6. Perbarui TODO dan dokumentasi berdasarkan hasil nyata. Kriteria selesai: hasil sesuai kebutuhan task, pemeriksaan relevan lulus, dan tidak ada masalah yang diketahui dalam cakupan tersebut. Jangan menjanjikan bebas bug tanpa bukti.
7. Setiap task wajib diakhiri commit terfokus dengan pesan deskriptif, kemudian push ke remote/branch yang dikonfigurasi. Tinjau diff dan staged files sebelum commit. Jangan sertakan rahasia, dependency terpasang, atau output build.
8. Jika remote, autentikasi, atau jaringan tidak tersedia, simpan commit lokal dan catat push sebagai terblokir. Jangan mengarang URL repository, membuat repository publik, atau mengklaim push berhasil. Lanjutkan pekerjaan independen yang tidak terblokir.
9. Hindari force push, reset destruktif, dan menimpa pekerjaan lain. Persetujuan rutin tidak menggantikan informasi teknis yang belum tersedia.
10. Laporkan ringkas hasil, verifikasi, commit/push, dan blocker konkret. Minta informasi hanya jika benar-benar dibutuhkan untuk pekerjaan yang tidak dapat dilanjutkan.

## Konvensi implementasi
- Struktur target: `backend/`, `frontend/`, `infra/`, dan `docs/`.
- Backend menerapkan otorisasi pada setiap operasi; identitas dari SSO harus diverifikasi, bukan dipercaya dari input client.
- Pisahkan adaptor integrasi eksternal dari aturan bisnis agar dapat diuji secara lokal.
- Alur permohonan harus memiliki transisi status yang eksplisit dan audit trail.
- Gunakan migrasi database terversi, query berparameter, dan transaksi untuk perubahan data yang saling bergantung.
- Berkas dokumen disimpan privat di object storage dengan batas ukuran, validasi tipe, dan pemeriksaan hak akses.
- Konfigurasi melalui environment; dokumentasikan nama variabel dalam `.env.example` tanpa nilai rahasia.
- Jangan mencatat token, kredensial, NIK, atau isi dokumen pribadi ke log.
- Pertahankan Bahasa Indonesia untuk dokumentasi pengguna dan ikuti gaya kode di modul terkait.
- Versi stack dalam kebutuhan merupakan baseline; cek kompatibilitas dan dukungan keamanan sebelum mengunci versi dependensi untuk deployment.

## Integrasi dan deployment
- Tiga container utama: MySQL, backend, frontend. MinIO dan layanan pemerintah adalah dependensi eksternal; bila MinIO lokal diperlukan, dokumentasikan sebagai layanan pengembangan tambahan.
- SSO JSP/Satriya, SPL, dan Sadewa memerlukan kontrak API serta konfigurasi resmi. Jangan menebak endpoint atau format token.
- Tes lokal dan sandbox tidak menggantikan UAT bersama DPKP/Diskominfo.
