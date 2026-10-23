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
  `admin` dan `pengawas` dapat melihat daftar lintas pemohon.
- `GET /v1/horticulture/applications/{id}` — pemohon hanya dapat melihat
  permohonan sendiri; `admin` dan `pengawas` dapat melihat permohonan lain.

Semua endpoint berada di bawah middleware SSO. Jangan mengirim identitas atau
role melalui header buatan klien. Permohonan dan pengaitan dokumen disimpan
dalam satu transaksi; server memastikan setiap dokumen dimiliki pemohon.

Jalankan migrasi setelah koneksi DB tersedia:

```powershell
cd backend
$env:DB_DSN = "user:password@tcp(127.0.0.1:3306)/sitani?parseTime=true&charset=utf8mb4"
go run ./cmd/migrate --dir migrations
```

## Belum termasuk

Status baru `diajukan`. Verifikasi proposal, permintaan perbaikan, jadwal
inspeksi, checklist/audit GAP, penilaian PIT/RIP/PI/RI, transisi status
petugas, nomor registrasi final, dan TTE Sadewa merupakan pekerjaan lanjutan.
Nama field dan kelengkapan lampiran perlu dikonfirmasi oleh bidang hortikultura
DPKP sebelum dijadikan formulir final/UAT. Jangan menganggap nomor pengajuan
sebagai nomor registrasi resmi.
