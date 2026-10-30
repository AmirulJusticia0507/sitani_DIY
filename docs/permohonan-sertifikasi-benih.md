# Permohonan Sertifikasi Benih (Fase 1)

Dokumen ini menjelaskan alur permohonan sertifikasi benih di halaman `/dashboard/certification`
beserta API yang dipakainya. Fase 2 (pengujian lab mutu) dan Fase 3 (sertifikat, TTE,
label) **bukan** form pelanggan: keduanya dijalankan petugas internal atas permohonan
yang sudah diajukan.

## Alur status permohonan

```
draft ──ajukan──> diajukan ──bayar──> dibayar ──verifikasi──> diverifikasi
                    │
                    ├──inspeksi──> diinspeksi ──validasi──> divalidasi (selesai)
                    └──tolak/batal──> ditolak | dibatalkan (selesai)
```

- `draft`: hanya tersimpan di aplikasi, belum punya nomor resmi, dan **belum**
  membuka langkah Analis Lab maupun Penandatangan.
- Isian permohonan yang sudah diajukan dikunci bagi pelanggan. Perbaikan atau
  pembatalan dilakukan petugas dan tercatat pada audit trail.
- Satu pelanggan hanya boleh punya satu permohonan aktif. Pengajuan baru saat
  permohonan sebelumnya masih berjalan ditolak dengan `409`.

## API

Semua endpoint membutuhkan sesi login (cookie SSO/login lokal) dengan peran
`pelanggan`; identitas diambil dari sesi, bukan dari header yang dikirim browser.

| Metode | Path | Keterangan |
|---|---|---|
| `GET` | `/v1/permohonan/aktif` | Permohonan aktif milik pelanggan; `204` bila belum ada draft maupun pengajuan. |
| `POST` | `/v1/permohonan/draft` | Simpan draft. Isian boleh sebagian. Body: `draft_id` (0 untuk draft baru), `jenis_benih`, `varietas`, `luas_lahan`, `lokasi_kebun`. |
| `POST` | `/v1/permohonan/ajukan` | Ajukan permohonan. Semua isian wajib lengkap. `draft_id > 0` mengajukan draft tersimpan; `0` memakai draft aktif bila ada, atau membuat lalu langsung mengajukan. Nomor `SPB-YYYYMMDD-XXXX` dibuat server di dalam satu transaksi. |
| `GET` | `/v1/permohonan/{id}` | Detail permohonan. |

Kode galat penting:

- `400`: isian tidak lengkap/tidak valid.
- `409`: `draft_id` bukan draft atau pelanggan sudah punya pengajuan berjalan.
- `401`/`403`: sesi belum login atau peran tidak berhak.

## Antarmuka pelanggan

1. Halaman memuat `/v1/permohonan/aktif`; bila ada draft, isian langsung terisi dan
   tombol **Edit Draft** serta **Ajukan Permohonan** tersedia.
2. **Simpan Draft** menyimpan isian walau belum lengkap. Draft ditandai
   `DRAFT (BELUM DIAJUKAN)` dan belum memiliki nomor `SPB`.
3. **Ajukan Permohonan** memvalidasi isian, memberi nomor resmi, mengunci form, dan
   membuka langkah Analis Lab serta Penandatangan pada stepper.

## Batasan saat ini

Unggahan dokumen (proposal, struktur organisasi, daftar anggota, form registrasi
kebun, peta kebun) belum tersedia pada form ini; lihat `todo_list.md`.
