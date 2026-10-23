# Pengajuan Registrasi Lahan Hortikultura

## Cakupan yang tersedia

Tahap awal ini menerima pengajuan pemohon terautentikasi, data organisasi dan
kontak, lokasi lahan di DIY, luas, sampai tiga baris komoditas, serta lima
lampiran wajib: proposal, struktur organisasi, daftar anggota, form registrasi
kebun, dan peta kebun. File diunggah melalui endpoint dokumen privat
`POST /v1/dokumen`; database hanya menyimpan metadata dan relasi dokumen.

API pengajuan:

- `POST /v1/horticulture/applications` — hanya role `pelanggan` atau `admin`.
- `GET /v1/horticulture/applications` — pemohon hanya menerima miliknya;
  role petugas dapat melihat daftar lintas pemohon.
- `GET /v1/horticulture/applications/{id}` — pemohon hanya dapat melihat
  permohonan sendiri; petugas terotorisasi dapat melihat permohonan lain.
- `POST /v1/horticulture/applications/{id}/verification` — role
  `verifikator` (atau `admin`) memutuskan `diverifikasi`, `perlu_perbaikan`,
  atau `ditolak`; catatan alasan wajib untuk dua keputusan terakhir dan semua
  lima jenis dokumen wajib harus tertaut sebelum verifikasi lengkap.
- `POST /v1/horticulture/applications/{id}/inspection-schedules` — role
  `inspektor` menjadwalkan inspeksi pada permohonan yang sudah diverifikasi.
  Jadwal disimpan satu per permohonan dan dapat diperbarui selama belum selesai.
- `POST /v1/horticulture/inspection-schedules/{id}/complete` — inspektor yang
  ditugaskan menyimpan 11 indikator checklist GAP, temuan, dan rekomendasi.
- `POST /v1/horticulture/applications/{id}/evaluations` — role `pbt` atau
  `kasi_pmbtp` menyimpan penilaian pohon induk/PIT/RIP/PI/RI dengan skor
  vigoritas, kesehatan, dan kebenaran varietas 1–5 serta rekomendasi.
- Detail permohonan menampilkan jadwal, inspeksi, dan riwayat penilaian.
  Setiap perubahan status ditulis dalam `audit_log` pada transaksi yang sama.

Semua endpoint berada di bawah middleware SSO. Jangan mengirim identitas atau
role melalui header buatan klien. Permohonan dan pengaitan dokumen disimpan
dalam satu transaksi; server memastikan setiap dokumen dimiliki pemohon.

Jalankan migrasi 006-007 setelah koneksi DB tersedia. Migrasi 007 memperluas
ENUM `users.role`; role petugas harus ditetapkan pada akun lokal oleh admin
setelah identitas terverifikasi. SSO tidak memberikan role dari input klien.

```powershell
cd backend
$env:DB_DSN = "user:password@tcp(127.0.0.1:3306)/sitani?parseTime=true&charset=utf8mb4"
go run ./cmd/migrate --dir migrations
```

## Belum termasuk / blocker

Kontrak resmi Sadewa untuk endpoint, autentikasi, dan payload TTE belum tersedia;
pengesahan, penerbitan sertifikat/nomor registrasi final, dan UAT dengan DPKP
belum dilakukan. Nama field, indikator GAP, skala penilaian, dan keputusan
kelayakan masih perlu dikonfirmasi bidang hortikultura sebelum produksi. Nomor
permohonan bukan nomor registrasi resmi.
