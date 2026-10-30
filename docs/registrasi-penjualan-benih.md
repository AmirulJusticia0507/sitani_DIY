# Registrasi, Persediaan, dan Penjualan Benih

Dokumen ini menjelaskan implementasi modul registrasi lot benih, persediaan per lot, penjualan, pembatalan, dan laporan. Modul ini melengkapi alur sertifikasi BP3MBTP pada `docs/` dan `todo_list.md`.

## Status

- Implementasi backend, migrasi, halaman frontend, dan tes handler tersedia.
- Format resmi nomor registrasi, nomor dokumen penjualan, kelas benih, masa berlaku, serta kontrak SIM Perbenihan **belum** tervalidasi BP3MBTP/Diskominfo. Format internal saat ini adalah `RB-<nomor_lot>`.
- Aktivasi SSO JSP/Satriya dan UAT login tetap merupakan prasyarat produksi. Endpoint memakai principal SSO yang sudah diverifikasi; header `X-User-ID`/`X-User-Role` hanya fallback untuk pengujian handler.

## Model data

Migrasi: `backend/migrations/003_create_seed_registry_schema.sql`.

| Tabel | Isi |
| --- | --- |
| `registrasi_benih` | Master lot: nomor registrasi, permohonan, sertifikat, label, komoditas, varietas, produsen, sumber benih, nomor lot, kelas, jumlah, satuan, tanggal produksi, tanggal kedaluwarsa, status edar |
| `stok_benih` | Saldo tersedia per lot (satu baris per registrasi) |
| `mutasi_stok_benih` | Ledger mutasi `masuk`, `keluar`, `koreksi` beserta saldo setelah mutasi dan referensi transaksi |
| `penjualan_benih` | Transaksi penjualan: dokumen, lot, pembeli, jumlah, harga, total, tanggal, status, alasan pembatalan |

Status edar: `aktif`, `ditangguhkan`, `dicabut`, `kedaluwarsa`. Penjualan hanya diterima untuk status `aktif`.

## Aturan bisnis

1. Registrasi lot hanya dapat dibuat jika:
   - permohonan sertifikasi ada dan belum `ditolak`/`dibatalkan`;
   - sertifikat sudah berstatus TTE `ditandatangani`;
   - label sudah `legalisasi` dan integrasi SIM berstatus `terkirim`;
   - nomor lot belum pernah terdaftar.
2. Stok awal dibuat dari jumlah registrasi melalui transaksi database. Saldo tidak dapat berubah tanpa baris mutasi.
3. Penjualan divalidasi terhadap status edar, masa berlaku, kecocokan sertifikat/label, dan kecukupan stok. Pengurangan stok dan pencatatan penjualan berada dalam satu transaksi dengan `SELECT ... FOR UPDATE`.
4. Pembatalan tidak menghapus transaksi. Status berubah menjadi `dibatalkan`, alasan dan aktor dicatat, dan stok dikembalikan melalui mutasi `koreksi`.
5. Setiap aksi penting menulis `audit_log`.

## Endpoint

Semua endpoint berada di bawah `/v1` dan dilindungi middleware SSO. Peran yang diizinkan:

| Metode | Path | Peran | Keterangan |
| --- | --- | --- | --- |
| GET | `/v1/permohonan-siap-registrasi` | admin, pengawas, pelanggan | Pilihan permohonan untuk form registrasi |
| POST | `/v1/registrasi-benih` | admin, pengawas | Daftarkan lot |
| GET | `/v1/registrasi-benih` | admin, pengawas | Filter `status_edar`, `komoditas`, `limit` |
| GET | `/v1/registrasi-benih/{id}` | admin, pengawas | Detail lot |
| POST | `/v1/registrasi-benih/{id}/status` | admin | Ubah status edar |
| GET | `/v1/registrasi-benih/{id}/stok` | admin, pengawas | Saldo dan riwayat mutasi |
| POST | `/v1/penjualan-benih` | admin | Catat penjualan |
| GET | `/v1/penjualan-benih` | admin, pengawas | Filter `registrasi_id`, `status`, `dari`, `sampai`, `limit` |
| POST | `/v1/penjualan-benih/{id}/batal` | admin | Batalkan dengan alasan |
| GET | `/v1/laporan/registrasi-benih` | admin, pengawas | Laporan registrasi |
| GET | `/v1/laporan/stok-benih` | admin, pengawas | Laporan stok |
| GET | `/v1/laporan/kedaluwarsa-benih?hari=90` | admin, pengawas | Lot yang kedaluwarsa ≤ N hari |
| GET | `/v1/laporan/penjualan-benih` | admin, pengawas | Laporan penjualan |

Contoh registrasi. `permohonan_id` dipilih dari respons `/v1/permohonan-siap-registrasi`;
halaman frontend menampilkan nomor permohonan dan mengirim ID di balik layar:

```json
{
  "permohonan_id": 1,
  "komoditas": "Tomat",
  "varietas": "Rama",
  "produsen": "Kelompok Tani Sleman",
  "sumber_benih": "BP3MBTP",
  "nomor_lot": "LOT-001",
  "kelas_benih": "sebar",
  "jumlah": 100,
  "satuan": "kg",
  "tanggal_produksi": "2026-09-23",
  "tanggal_kedaluwarsa": "2027-04-21",
  "status_edar": "aktif"
}
```

Contoh penjualan:

```json
{
  "nomor_dokumen": "PJ-001",
  "registrasi_id": 1,
  "pembeli": "Petani Bantul",
  "jumlah": 60,
  "harga_satuan": 25000,
  "tanggal_transaksi": "2026-10-23"
}
```

## Nomor permohonan

Nomor permohonan **dibuat server** setelah permohonan tersimpan, dengan format
`SPB-YYYYMMDD-XXXX`:

- `YYYYMMDD` adalah tanggal pengajuan;
- `XXXX` adalah urutan harian (dimulai `0001`, minimal 4 digit, bertambah bila
  lebih dari 9999 pengajuan dalam sehari).

Contoh: permohonan pertama pada 23 Oktober 2026 bernomor `SPB-20261023-0001`.
Nomor tidak diterima dari client dan tidak memakai kolom `id` database. Nomor
disimpan di kolom `nomor_permohonan` dalam transaksi yang sama dengan insert.
Penomoran diserialkan per tanggal memakai advisory lock PostgreSQL
(`pg_advisory_xact_lock`), sehingga dua pengajuan bersamaan tidak memperoleh
nomor yang sama.

Klien tidak memakai nomor ini sebagai kunci form. Untuk mendaftarkan lot,
frontend memanggil `/v1/permohonan-siap-registrasi`, yang hanya mengembalikan
permohonan dengan sertifikat TTE `ditandatangani`, label `legalisasi` +
integrasi SIM `terkirim`, belum `ditolak`/`dibatalkan`, dan belum punya lot.

Migrasi `008_formalize_nomor_permohonan.sql` menyeragamkan nomor lama
(`SPB-<id>-<epoch Unix>`, format yang salah) ke format baru.

> Format ini adalah konvensi internal. Bila BP3MBTP/Diskominfo menetapkan format
> resmi, generator di `store.FormatNomorPermohonan` dan migrasi 008 perlu
> disesuaikan.

## Antarmuka

Halaman `/dashboard/benih` menyediakan formulir registrasi, daftar lot, detail stok dan mutasi, formulir penjualan, pembatalan, serta laporan stok/kedaluwarsa/penjualan. Halaman memakai sesi SSO (`credentials: include`) dan tidak menyimpan token di browser.

## Menjalankan migrasi

```powershell
cd backend
$env:DB_DSN="user:pass@tcp(127.0.0.1:3306)/sitani?parseTime=true"
go run ./cmd/migrate
```

Periksa tanpa menerapkan:

```powershell
go run ./cmd/migrate -dry-run
```

## Verifikasi yang dijalankan

- `go build ./...`, `go vet ./...`, `go test ./... -count=1` (backend)
- `npm run lint`, `npm run typecheck`, `npm run build` (frontend)
- `npm run test:e2e` (Playwright desktop dan mobile)
- Migrasi 001–003 diterapkan ke MySQL 8.4 lokal melalui `go run ./cmd/migrate`.
- Uji alur nyata terhadap MySQL: registrasi lot stok 100, penjualan 60 menyisakan stok 40 dengan 2 mutasi, pembatalan mengembalikan stok 100 dengan 3 mutasi; laporan penjualan dan stok mengembalikan data.
- Penomoran permohonan diuji terhadap PostgreSQL 18 sungguhan (`SITANI_TEST_DATABASE_URL`): `CreatePermohonan` menghasilkan `SPB-20261023-0001`, `SPB-20261023-0002`, lalu `SPB-20261024-0001` pada hari berikutnya, dan tidak menyisakan nomor sementara.
- Migrasi `008_formalize_nomor_permohonan.sql` diuji pada data lama `SPB-<id>-<epoch>`: nama lama dinomori ulang per tanggal (mis. `SPB-20261020-0001`), sedangkan nomor yang sudah berformat baru tidak diubah.

Uji tersebut memakai data sintetis dan bukan UAT. Format resmi serta integrasi SIM Perbenihan tetap menunggu validasi pihak berwenang. Uji browser halaman `/dashboard/benih` menunjukkan 0 error konsol dan 0 permintaan gagal.
