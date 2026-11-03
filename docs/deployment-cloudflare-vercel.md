# Deployment: Cloudflare, Vercel, dan Backend Go

Dokumen ini menjelaskan cara menaruh SiTani DIY di internet, batasan tiap
platform, dan variabel lingkungan yang harus diisi. Status terakhir diperbarui
3 November 2026.

## Ringkasan status

| Bagian | Status |
| --- | --- |
| PWA frontend (manifest, ikon, service worker, halaman offline) | Tersedia dan diuji lewat build produksi |
| Frontend di Vercel | Belum di-deploy; CLI sudah login, proyek belum ditautkan |
| Variabel lingkungan Vercel | Belum dibuat; menunggu URL publik backend |
| Backend Go di Cloudflare Workers | **Tidak dapat dijalankan langsung** (lihat di bawah) |
| Backend Go di host biasa/VPS | Didukung lewat Docker (`infra/compose.yml`) |
| Cloudflare sebagai DNS + router tepi | Siap dikonfigurasi, belum diterapkan |

Jangan menganggap deployment selesai sebelum ada URL nyata yang bisa dibuka dan
diuji. Dokumen ini adalah rencana yang mengikuti implementasi aktual.

## Topologi yang direkomendasikan

Satu domain publik, dua origin di belakangnya:

```
                    ┌──────────────────────────────┐
Pengguna ── HTTPS ──►│ Cloudflare (DNS + Worker)    │
                    └───────┬──────────────┬───────┘
                            │              │
        /v1/*, /auth/*,     │              │  selain itu
        /health             │              │
                            ▼              ▼
                    Backend Go        Frontend Next.js
                    (Docker/VPS)      (Vercel)
```

Alasannya bukan sekadar rapi: autentikasi SiTani memakai cookie `HttpOnly`
dengan `SameSite=Lax` (`backend/internal/auth/session.go`). Bila frontend dan
backend berada di **domain berbeda**, peramban tidak akan mengirim cookie sesi
pada permintaan `fetch`, sehingga login tampak berhasil tetapi semua endpoint
mengembalikan `401`. Menaruh keduanya di satu domain (backend di balik path
`/v1`, `/auth`, `/health`) menghindari masalah itu tanpa melonggarkan atribut
cookie.

Alternatif domain terpisah (`sso.example.go.id` dan `sitani.example.go.id`)
memerlukan `SameSite=None; Secure` dan tetap berisiko karena pembatasan cookie
pihak ketiga di Safari dan Chrome. Perubahan itu belum diterapkan di kode.

## Mengapa backend Go tidak bisa langsung di Cloudflare Workers

- Backend adalah server `net/http` biasa dengan koneksi TCP ke PostgreSQL dan
  MinIO. Workers hanya menjalankan JavaScript/WASM dengan API web terbatas,
  tanpa socket TCP umum.
- Rewrite ke Workers berarti mengganti lapisan HTTP, akses database, dan storage
  — bukan sekadar deploy ulang. Hasilnya bukan kode yang sama.
- Cloudflare Containers dapat menjalankan image Docker Go, tetapi merupakan
  layanan berbayar dan belum menjadi target deployment yang disetujui.

Yang bisa dilakukan Cloudflare tanpa menulis ulang backend:

1. **DNS + CDN + TLS** untuk domain portal.
2. **Worker router** yang meneruskan `/v1/*`, `/auth/*`, dan `/health` ke host
   backend, sisanya ke Vercel. Berkas contoh ada di `infra/cloudflare/`.
3. **Cloudflare Tunnel** bila backend berjalan di server kantor tanpa IP publik,
   sehingga tidak perlu membuka port.

## Langkah 1 — Deploy backend Go

Backend membutuhkan PostgreSQL dan (bila fitur dokumen dipakai) MinIO. Contoh
lokal ada di `infra/compose.yml`; untuk produksi sesuaikan variabel berikut di
host/container:

| Variabel | Contoh produksi | Catatan |
| --- | --- | --- |
| `PORT` | `8080` | Port internal container |
| `DB_DSN` | `postgres://user:pass@host:5432/sitani?sslmode=require` | Wajib; gunakan TLS |
| `BACKEND_PUBLIC_URL` | `https://sitani.example.go.id` | Dipakai menyusun redirect URI SSO |
| `FRONTEND_PUBLIC_URL` | `https://sitani.example.go.id` | Origin yang diizinkan CORS |
| `SSO_COOKIE_SECURE` | `true` | Wajib di HTTPS |
| `SSO_JSP_ISSUER` / `SSO_JSP_CLIENT_ID` | dari pengelola SSO | Jangan menebak |
| `SSO_SATRIYA_ISSUER` / `SSO_SATRIYA_CLIENT_ID` | dari pengelola SSO | Jangan menebak |
| `SITANI_DEV_LOGIN` | `false` | Harus mati di produksi |
| `MINIO_*` | dari pengelola | Kosong = unggah dokumen menjawab `503` |
| `SENTRY_DSN` | dari pengelola | Opsional |
| `DISCORD_WEBHOOK_URL` | dari pengelola | Opsional |

Karena frontend dan backend berbagi satu domain, `BACKEND_PUBLIC_URL` dan
`FRONTEND_PUBLIC_URL` bernilai sama. Redirect URI yang didaftarkan di client
SSO adalah:

```
https://sitani.example.go.id/auth/callback/jsp
https://sitani.example.go.id/auth/callback/satriya
```

Nilai persisnya harus dikonfirmasi ke pengelola SSO; kode menyusunnya dari
`BACKEND_PUBLIC_URL`.

## Langkah 2 — Deploy frontend ke Vercel

Frontend berada di subdirektori `frontend/`, bukan akar repositori.

1. Tautkan proyek (dari direktori `frontend/`):

   ```powershell
   vercel link
   ```

   Pilih scope akun/organisasi dan buat proyek baru, misalnya
   `sitani-diy-frontend`. Saat diminta *root directory*, gunakan `frontend`.

2. Atur *Root Directory* = `frontend`, *Framework Preset* = Next.js,
   *Build Command* = `npm run build`, *Install Command* = `npm ci`.
   Node.js 22 dipakai saat pengembangan lokal.

3. Tambahkan variabel lingkungan (produksi dan preview):

   ```powershell
   vercel env add NEXT_PUBLIC_BACKEND_URL production
   vercel env add NEXT_PUBLIC_BACKEND_URL preview
   ```

   Nilainya **domain publik portal**, bukan `localhost`, misalnya
   `https://sitani.example.go.id`. Karena frontend memanggil backend lewat
   domain yang sama, nilai ini tidak membocorkan rahasia apa pun.

4. Deploy:

   ```powershell
   vercel --prod
   ```

### Variabel lingkungan frontend

| Nama | Wajib | Contoh | Keterangan |
| --- | --- | --- | --- |
| `NEXT_PUBLIC_BACKEND_URL` | Ya | `https://sitani.example.go.id` | Ikut terkirim ke browser; jangan berisi rahasia |
| `NEXT_PUBLIC_API_BASE` | Tidak | — | Alias lama; dipakai lebih dulu bila diisi |

Tidak ada rahasia di sisi frontend. API key AI, DSN Sentry, kredensial MinIO,
dan secret SSO hanya boleh berada di backend.

Setelah mengubah `NEXT_PUBLIC_*`, deploy ulang: nilainya dibekukan saat build,
bukan dibaca saat runtime.

## Langkah 3 — Cloudflare sebagai router tepi

Berkas contoh:

- `infra/cloudflare/sitani-router.js` — Worker yang memisahkan trafik API dan
  frontend.
- `infra/cloudflare/wrangler.toml` — konfigurasi dasar.

Langkah penerapan (belum dijalankan di lingkungan ini karena `wrangler` belum
terpasang dan kredensial Cloudflare belum tersedia):

```powershell
npm install -g wrangler
wrangler login
cd infra/cloudflare
wrangler deploy

# Isi origin tanpa menyimpannya di repositori:
wrangler secret put BACKEND_ORIGIN    # mis. https://backend.internal.example
wrangler secret put FRONTEND_ORIGIN   # mis. https://sitani-diy-frontend.vercel.app
```

Lalu arahkan domain portal ke Worker (lewat `routes` di `wrangler.toml` atau
menu *Workers Routes* di dasbor Cloudflare).

Catatan penting:

- Worker meneruskan respons apa adanya, termasuk `Set-Cookie` dan `302` ke SSO.
- Bila deployment Vercel memakai *Deployment Protection*, Worker akan menerima
  halaman login Vercel, bukan aplikasi. Nonaktifkan proteksi untuk domain
  produksi atau gunakan *Protection Bypass for Automation*.
- Jangan menaruh Cloudflare di depan backend **dan** mengubah `Host`; backend
  memakai `BACKEND_PUBLIC_URL` dari environment, bukan header `Host`.

## Langkah 4 — Verifikasi setelah deploy

1. `GET https://<domain>/health` mengembalikan status sehat dari backend.
2. Halaman utama memuat dari Vercel (cek header `x-vercel-id`).
3. Login SSO berhasil dan `/auth/me` mengembalikan peran pengguna — ini menguji
   cookie lintas path pada satu domain.
4. `/manifest.webmanifest` dapat dibuka dan `theme_color` bernilai `#0d4938`.
5. DevTools → Application → Service Workers: `/sw.js` aktif, dan halaman
   `/offline` tampil saat jaringan dimatikan.
6. Tidak ada error CORS di konsol; `FRONTEND_PUBLIC_URL` sama dengan origin
   portal.

## Yang belum selesai

- URL publik backend belum ada, sehingga variabel Vercel belum dapat diisi.
- Proyek Vercel belum ditautkan dan belum ada deployment.
- Worker Cloudflare belum di-deploy (butuh akun, zona DNS, dan `wrangler`).
- Domain resmi, sertifikat, dan pendaftaran redirect URI SSO harus melalui
  Diskominfo/DPKP.
- Uji beban, backup PostgreSQL, dan pemantauan produksi belum disiapkan.
