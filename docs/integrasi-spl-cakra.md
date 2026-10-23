# Integrasi SPL / Cakra API Gateway

## Status

Transport HTTP JSON generik tersedia di `backend/internal/spl`. Ini **bukan**
integrasi bisnis atau koneksi produksi yang selesai. Belum ada kontrak API
resmi SPL/Cakra di repository, sehingga kode tidak menetapkan endpoint,
payload, skema respons, atau mekanisme autentikasi berdasarkan tebakan.

Client dapat diinisialisasi melalui `spl.ConfigFromEnv()` dan `spl.New()`.
Konfigurasi yang tersedia:

- `SPL_BASE_URL`: URL HTTP(S) resmi per lingkungan; wajib diisi eksplisit.
- `SPL_TIMEOUT_SECONDS`: timeout request dalam detik (default 10).

`Client.Do` menyediakan pengiriman dan decoding JSON generik untuk route
relatif yang ditentukan setelah kontrak diterima. Client tidak otomatis
menambahkan token/header autentikasi. Respons error upstream hanya menyimpan
status HTTP agar body yang mungkin sensitif tidak tersebar ke log/pemanggil.

## Informasi yang harus dikonfirmasi

Sebelum membuat operasi bisnis atau menyatakan integrasi selesai, minta
pengelola SPL/Cakra memberikan:

1. Base URL sandbox dan produksi serta cara akses jaringan/allowlist.
2. Spesifikasi OpenAPI/kontrak versi dan route untuk tiap layanan yang perlu
   dipakai SiTani.
3. Skema request/response, format tanggal/angka, pagination, dan aturan kode
   status/error.
4. Mekanisme autentikasi resmi (mis. OAuth2/mTLS/API key), cara rotasi
   kredensial, audience/scope, dan header yang diwajibkan. Jangan mengirim
   rahasia melalui Git atau log.
5. Batas laju, timeout/retry yang diperbolehkan, idempotency, dan kebijakan
   korelasi/audit.
6. Akun sandbox, data uji, prosedur UAT, PIC teknis, dan kriteria penerimaan.

Setelah kontrak tersedia, implementasikan adaptor per operasi di atas
transport ini, validasi payload secara eksplisit, tambahkan tes kontrak
terhadap sandbox/mocked server, dan lakukan UAT bersama pengelola. Tes lokal
transport tidak menggantikan UAT.
