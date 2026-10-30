# SSO Lokal dengan Keycloak

Dokumen ini menjelaskan SSO lokal untuk development/UAT SiTani DIY memakai Keycloak
sendiri. Tujuannya agar alur `/auth/login/jsp` → issuer → callback backend dapat
diuji lengkap **tanpa** memakai identitas atau realm milik Jogjaprov.

Dokumen ini **bukan** panduan produksi. Realm lokal berjalan dengan `start-dev`
(HTTP, tanpa TLS) dan hanya untuk mesin pengembang.

## Mengapa instance terpisah

- Insert realm JSP/Satriya milik `sso.jogjaprov.go.id` **bukan** milik proyek ini.
  Klien `portal` belum terdaftar di sana; membuka authorization endpoint dengan
  `client_id=portal` dijawab `Invalid parameter: redirect_uri`. Pendaftaran client
  hanya bisa dilakukan oleh pengelola SSO Jogjaprov.
- Keycloak milik proyek lain di mesin yang sama (mis. `bprs-keycloak` untuk
  e-notary pada port 8180) tidak boleh dipakai: realm dan datanya milik proyek itu.

Karena itu SiTani memakai container, port, volume, dan realm sendiri:

| Aspek | Nilai |
| --- | --- |
| Compose | `infra/keycloak/compose.yml` |
| Container | `sitani-keycloak` |
| Port host | `8085` (`KEYCLOAK_HOST_PORT`) |
| Volume data | `keycloak_data` |
| Realm | `sitani` (tampilan: "SiTani DIY (lokal)") |
| Client | `sitani-portal` (confidential, PKCE `S256`) |
| Skrip konfigurasi | `infra/keycloak/configure-realm.ps1` |

## Menjalankan

1. Siapkan password administrator Keycloak lokal (hanya berlaku saat volume pertama
   kali dibuat). Simpan di `.env` lokal Anda; jangan commit.

   ```bash
   KEYCLOAK_ADMIN_PASSWORD=<password-admin-lokal>
   KEYCLOAK_HOST_PORT=8085
   ```

2. Nyalakan Keycloak.

   ```bash
   docker compose --env-file .env -f infra/keycloak/compose.yml up -d
   ```

   Propagasi awal (Quarkus augmentation) dapat memakan 1–2 menit; tunggu sampai
   status `healthy` (`docker ps --filter name=sitani-keycloak`).

3. Buat realm, client, dan akun uji. Skrip berjalan berulang tanpa membuat duplikat,
   dan **berhenti dengan galat** bila secret yang tersimpan di Keycloak berbeda dari
   yang dipakai skrip.

   ```powershell
   $env:KEYCLOAK_ADMIN_PASSWORD = '<password-admin-lokal>'
   $env:SEED_DEV_PASSWORD       = '<password-akun-uji-min-12-karakter>'
   ./infra/keycloak/configure-realm.ps1 -ClientSecret '<secret-yang-ingin-dipakai>'
   ```

   Bila `-ClientSecret` dikosongkan, skrip membuat secret acak dan mencetaknya.
   Salin ke `.env` backend (lihat langkah 4). Skrip kompatibel dengan PowerShell 5.1.

4. Arahkan backend ke realm lokal. Dari dalam container, host diissuer **wajib**
   `host.docker.internal`; `localhost` akan menunjuk ke container backend sendiri
   dan token endpoint menjawab `connection refused`.

   ```bash
   SSO_JSP_ISSUER=http://host.docker.internal:8085/realms/sitani
   SSO_JSP_CLIENT_ID=sitani-portal
   SSO_JSP_CLIENT_SECRET=<secret-dari-langkah-3>
   ```

   Bila backend dijalankan langsung di host (bukan container), pakai
   `SSO_JSP_ISSUER=http://localhost:8085/realms/sitani`. Kedua bentuk host sudah
   terdaftar sebagai `redirect_uri` (`localhost` dan `host.docker.internal` × port
   `8080`/`8081` + path `/auth/callback/jsp`).

5. Jalankan ulang backend. Kode backend disalin saat build image, jadi perubahan
   konfigurasi **dan** kode memerlukan rebuild.

   ```bash
   docker compose --env-file .env -f infra/compose.yml up -d --build backend
   ```

6. Buka `http://localhost:3001/masuk`, pilih **Masuk JSP**, lalu login dengan akun
   uji. Callback harus kembali ke portal dan `/auth/me` melaporkan `"realm":"jsp"`.

## Akun uji

`configure-realm.ps1` membuat lima akun dengan password `SEED_DEV_PASSWORD`:

| Email | Peran bisnis (dari tabel `users` lokal) |
| --- | --- |
| `admin@dev.example.invalid` | admin |
| `pengawas@dev.example.invalid` | pengawas |
| `analis@dev.example.invalid` | analis |
| `penandatangan@dev.example.invalid` | penandatangan |
| `pelanggan@dev.example.invalid` | pelanggan |

Peran **tidak** diambil dari claim token. Backend memetakan identitas SSO ke tabel
`users` berdasarkan email, sesuai `docs/login-lokal-development.md`. Karena itu
email akun Keycloak harus sama dengan email di database lokal. Akun yang sudah
dipetakan SSO memiliki `password_hash` berpenanda `SSO-DISABLED-...` sehingga tidak
bisa dipakai untuk login lokal.

## Perilaku backend yang relevan

- **Client secret opsional.** `Realm.ClientSecret` hanya dikirim ke token endpoint
  bila `SSO_*_CLIENT_SECRET` terisi. Client publik (PKCE saja) tidak mengirimnya.
- **Validasi token:** issuer harus sama dengan `SSO_*_ISSUER`, algoritma `RS256`,
  `kid` harus ada di JWKS realm, `exp` wajib, dan `azp` harus sama dengan client id.
  Klaim `aud` **tidak** dipakai karena Keycloak mengisinya dengan peran resource
  (`account`), bukan client id. Token dari client lain pada realm yang sama ditolak
  lewat `azp`.
- **Email terverifikasi wajib.** Claim `sub`, `email`, dan `email_verified=true`
  harus ada.
- **Diagnostik.** Kegagalan pertukaran code, verifikasi token, dan pemetaan identitas
  dicatat di log backend sebagai `WARN` dengan kode kesalahan OAuth saja — token,
  email, dan isi dokumen tidak dicatat.

## Pemecahan masalah

| Gejala | Penyebab umum |
| --- | --- |
| `{"error":"realm SSO tidak tersedia"}` | `SSO_JSP_ISSUER`/`SSO_JSP_CLIENT_ID` kosong; realm tidak valid dibuang saat start |
| Callback `401` + log `dial tcp [::1]:8085: connect: connection refused` | issuer memakai `localhost` dari dalam container; ganti ke `host.docker.internal` |
| Token endpoint `unauthorized_client` | password client di Keycloak berbeda dari `.env`; jalankan ulang skrip dengan `-ClientSecret` yang sama (skrip memverifikasi) |
| Token endpoint `invalid_grant: Code not valid` | authorization code sudah dipakai/kedaluwarsa — perilaku normal saat diuji manual |
| Keycloak menjawab `Invalid parameter: redirect_uri` | `redirect_uri` callback tidak terdaftar. Tambahkan host/port baru di `configure-realm.ps1` lalu jalankan ulang |
| Log `verifikasi token SSO gagal ... invalid audience` | versi backend lama yang masih memvalidasi `aud`; rebuild backend |
| `403` saat login/logout | `FRONTEND_PUBLIC_URL` tidak sama dengan origin frontend yang dipakai |
| `404` untuk `/v1/dokumen?_rsc=...` di konsol | versi frontend lama yang menautkan menu **Dokumen** ke endpoint API. Menu kini mengarah ke halaman `/dashboard/dokumen`; rebuild frontend |

## Menu Dokumen

Menu **Dokumen** di sidebar membuka halaman `/dashboard/dokumen`. Halaman ini
memakai `GET /v1/dokumen` (hanya ditampilkan bila sesi sudah aktif), menampilkan
nama berkas, tujuan, jenis, ukuran, dan waktu unggah, serta tautan unduh ke
`GET /v1/dokumen/{id}/unduh`. Backend tetap memeriksa hak akses; pengguna yang
belum masuk melihat ajakan masuk. MinIO belum dikonfigurasi berarti unggahan baru
ditolak `503`, tetapi daftar dokumen tetap dapat dibuka.

## Batasan

- Realm ini hanya untuk pengembangan/UAT. `start-dev`, HTTP tanpa TLS, dan akun
  uji tidak boleh dipakai untuk identitas produksi.
- UAT dengan SSO Jogjaprov tetap memerlukan client `portal` (atau client resmi lain)
  yang didaftarkan dan diserahkan konfigurasinya oleh pengelola SSO. Endpoint,
  format token, dan `redirect_uri` produksi tidak ditebak di repositori ini.
- Tanpa MinIO, unggahan dokumen ditolak `503 "penyimpanan dokumen belum dikonfigurasi"`
  dan `/health` melaporkan `"storage":"belum-dikonfigurasi"`.
- Kode `configure-realm.ps1` memakai *password grant* `admin-cli` untuk autentikasi
  administrator; ini hanya sesuai Keycloak mode pengembangan.

## Verifikasi yang sudah dijalankan

- `docker compose -f infra/keycloak/compose.yml up -d` → container `healthy`,
  `GET /realms/sitani/.well-known/openid-configuration` `200`.
- `configure-realm.ps1` dijalankan dua kali berturut-turut (idempoten), lulus
  verifikasi "secret client cocok"; uji token endpoint: secret benar → `400
  invalid_grant` (kredensial diterima), secret salah/tanpa secret → `401
  unauthorized_client`.
- `go test ./internal/auth/...` — termasuk regresi token Keycloak (`aud: ["account"]`,
  `azp: sitani-portal` diterima; `azp` client lain atau kosong ditolak), pengiriman
  client secret, dan pembacaan konfigurasi dari environment.
- Uji browser end-to-end: password salah ditolak di halaman Keycloak → login benar →
  `302` ke `http://localhost:3001/dashboard`, `GET /auth/me` `200`
  (`"realm":"jsp"`, `"role":"analis"`), nama tampil di dashboard, dan sesi tetap
  ada setelah reload. Tanpa error konsol maupun permintaan gagal.
- Jalur login portal diuji dua kali (lewat `/dashboard/masuk` → `/masuk` → tombol
  **Masuk dengan JSP**, dan langsung ke `/auth/login/jsp`); keduanya berakhir di
  dashboard sebagai `Analis Development` setelah reload.
- Halaman `/dashboard/dokumen` (menggantikan tautan menu yang dulu menunjuk ke
  `/v1/dokumen`): tamu melihat ajakan masuk, sesi SSO melihat daftar dari
  `GET /v1/dokumen` (`200`, "0 dokumen"). `npm run typecheck` dan `npm run lint`
  lulus.
