# Repository Data Lokal — Import CSV/XLSX

## Status dan batasan

Fondasi import lokal tersedia melalui `POST /v1/dataset/import` dan
`GET /v1/dataset/import`. Import menyimpan batch dan barisnya dalam JSON di
MySQL agar header dataset fleksibel selama template resmi belum ditetapkan.
Ini belum memetakan data ke model pertanian kanonis, menyinkronkan API
kementerian, atau menyediakan dashboard agregat.

Sebelum mengunci mapper dan dashboard, DPKP perlu menetapkan template per
sumber, satuan, periode, referensi wilayah, definisi metrik, aturan duplikasi/
idempotensi, dan kebijakan retensi. Saat ini parser mempertahankan header
sumber dan nilainya sebagai string; tidak melakukan normalisasi domain.

## API

Semua route dilindungi middleware SSO. Hanya role `admin` dan `analis` dapat
mengimpor/melihat batch. `analis` hanya melihat batch yang diunggahnya,
sedangkan `admin` dapat melihat seluruh batch.

`POST /v1/dataset/import` menerima `multipart/form-data`:

- `nama_dataset`: wajib, maksimal 120 karakter.
- `file`: CSV UTF-8 atau XLSX; XLSX membaca sheet pertama dan baris pertama
  sebagai header.

Batas file 10 MiB, maksimum 10.000 baris data dan 100 kolom. Header kosong,
duplikat (case-insensitive), terlalu panjang, dan baris dengan kolom melebihi
header ditolak. Baris kosong dilewati. Respons `201` memuat metadata batch,
bukan isi dataset.

`GET /v1/dataset/import` mengembalikan metadata batch terbaru (maksimal 100).
Saat ini belum tersedia API untuk membaca baris impor atau melakukan koreksi/
penghapusan batch.

## Verifikasi lokal

Setelah database tersedia, jalankan migrasi baru melalui runner:

```powershell
cd backend
go run ./cmd/migrate -dry-run
go run ./cmd/migrate
```

Tabel `dataset_imports` dibuat oleh migrasi `005_create_dataset_imports.sql`.
Endpoint memerlukan pengguna terautentikasi yang sudah dipetakan ke role
internal melalui SSO.
