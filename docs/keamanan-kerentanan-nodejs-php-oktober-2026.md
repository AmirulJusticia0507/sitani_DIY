# Antisipasi Kerentanan Keamanan Node.js dan PHP — Oktober 2026

Tanggal dokumentasi: 23 Oktober 2026  
Sumber verifikasi: rilis keamanan resmi Node.js (21 Oktober 2026), pengumuman PHP (21 Oktober 2026), dan peringatan CERT-EU 2026-018 (21 Oktober 2026).

Dokumen ini mencatat dua berita kerentanan yang diterbitkan sekitar 21–22 Oktober 2026 dan langkah anticipasi yang relevan bagi proyek Portal Pertanian dan Ketahanan Pangan DIY.

---

## 1. Ringkasan Kerentanan

| # | Produk | Tanggal Rilis Keamanan | ID Kerentanan | Tingkat Tertinggi | Versi Terkena Dampak | Versi Tanggal |
|---|--------|------------------------|---------------|-------------------|----------------------|---------------|
| 1 | Node.js | 21 Oktober 2026 | CVE-2026-111497 dan 9 CVE lain | HIGH | 26.x, 24.x, 22.x | v22.23.4, v24.22.1, v26.12.1 |
| 2 | PHP | 21–22 Oktober 2026 | CVE-2026-109618, CVE-2026-109619 | Critical (CVSS 9.2) | 8.5 < 8.5.12, 8.4 < 8.4.27, 8.3 < 8.3.36, 8.2 < 8.2.35, cabang EOL | 8.5.12, 8.4.27, 8.3.36, 8.2.35 |

---

## 2. Detail Kerentanan

### 2.1 Node.js — Security Release Oktober 2026

Rilis keamanan Node.js diterbitkan 21 Oktober 2026 untuk cabang rilis 26.x, 24.x, dan 22.x. Terdapat 10 kerentanan dengan tingkat tertinggi **HIGH**.

Kerentanan utama yang relevan:

- **CVE-2026-111497 (HIGH):** Node.js Brotli decoder out-of-bounds write yang memungkinkan remote code execution. Mempengaruhi Node.js 26.x, 24.x, dan 22.x.
- **CVE-2026-105482 (HIGH):** `dns.resolve` memotong nama yang mengandung karakter NUL. Mempengaruhi Node.js 26.x, 24.x, dan 22.x.
- **CVE-2026-105483 (HIGH):** Pembatalan verifikasi hostname sertifikat TLS saat IP ditulis dengan digit lebar (fullwidth digits). Mempengaruhi Node.js 26.x dan 24.x.
- **CVE-2026-106069 (HIGH):** Pembatalan verifikasi hostname sertifikat TLS saat host longgar meresolusi ke alamat IP. Mempengaruhi Node.js 26.x, 24.x, dan 22.x.
- **CVE-2026-106074 (MEDIUM):** Cache entropi `crypto.randomUUID()` dan `randomInt()` disimpan dalam snapshot startup user-land, sehingga setiap proses menghasilkan nilai yang sama.
- **CVE-2026-105481, CVE-2026-106072 (MEDIUM):** Masalah soket keep-alive dan validasi sertifikat untuk hostname IPv4 non-kanonik.
- **CVE-2026-106070, CVE-2026-106071, CVE-2026-106073 (LOW):** Masalah ZipFile, Permission Model wildcard radix, dan Permission Model bypass via `fs.realpathSync`.

**Versi yang diperbaiki:**
- Node.js 22.x → **v22.23.4**
- Node.js 24.x → **v24.22.1**
- Node.js 26.x → **v26.12.1**

Sumber resmi: https://nodejs.org/en/blog/vulnerability/october-2026-security-releases

### 2.2 PHP — PHP2Shell (CVE-2026-109618 dan CVE-2026-109619)

Pada 21 Oktober 2026, PHP merilis pembaruan darurat untuk dua kerentanan kritis yang bersama dikenal sebagai **PHP2Shell**. Kedua kerentanan memungkinkan **remote code execution tanpa autentikasi** dan telah diamati dieksploitasi dalam liar sejak 21 Oktober 2026 (menurut watchTowr). Sebagai tanggal 22 Oktober 2026, keduanya belum masuk katalog Known Exploited Vulnerabilities CISA.

- **CVE-2026-109618 (CVSS v4.0: 9.2):** Remote code execution melalui korupsi memori saat penanganan out-of-memory (OOM). Terjadi saat PHP melebihi `memory_limit` di tengah pemrosesan permintaan, sehingga data internal hanya sebagian dibangun dan kode pembersihan mempercayainya sebagai valid.
- **CVE-2026-109619 (CVSS v4.0: 9.2):** Remote code execution melalui status global yang tidak segar (stale global state). Status internal dari permintaan dapat bertahan setelah memori pendukung dibebaskan, dan permintaan berikutnya yang ditangani worker yang sama dapat memakai status lama sebelum inisialisasi normal di-reset.

**Versi yang terpengaruh dan diperbaiki:**

| Cabang PHP | Versi yang Terpengaruh | Versi yang Diperbaiki |
|------------|------------------------|-----------------------|
| PHP 8.5.x | < 8.5.12 | 8.5.12 |
| PHP 8.4.x | < 8.4.27 | 8.4.27 |
| PHP 8.3.x | < 8.3.36 | 8.3.36 |
| PHP 8.2.x | < 8.2.35 | 8.2.35 |
| PHP 8.1 dan lebih lama (EOL) | Semua versi | Tidak ada rilis resmi; gunakan paket distribusi/LTS vendor, backport minimal sumber PHP, atau upgrade |

Setiap instalasi PHP yang melayani permintaan web berada dalam risiko, termasuk PHP-FPM, Apache `mod_php`, `php-cgi`, FrankenPHP, dan LiteSpeed. PHP menilai `mod_php` sebagai konfigurasi tertinggi risiko. PHP menyatakan bahwa **tidak ada perubahan konfigurasi yang sepenuhnya mitigasi** kedua kerentanan ini; pembaruan adalah solusi tunggal yang cukup.

Sumber resmi: https://cert.europa.eu/publications/security-advisories/2026-018

---

## 3. Dampak terhadap Proyek SiTani DIY

Berdasarkan `analisis_kebutuhan.md` dan `docs/analisis-proyek.md`, stack proyek saat ini adalah:

| Komponen | Teknologi | Terpengaruh CVE Node.js? | Terpengaruh CVE PHP? |
|----------|-----------|--------------------------|----------------------|
| Backend API | Go 1.23 | Tidak | Tidak |
| Frontend Web | Next.js 16 (berjalan di atas Node.js) | **Ya** (runtime Node.js) | Tidak |
| Database | MySQL | Tidak | Tidak |
| Object Storage | MinIO | Tidak | Tidak |
| TTE / Central Sign | Sadewa | Tidak | Tidak |
| API Gateway | SPL / Cakra API | Tidak | Tidak |
| Autentikasi | SSO JSP/Satriya | Tidak | Tidak |

**Kesimpulan dampak:**

1. **Node.js relevan** karena frontend Next.js di-build dan dijalankan menggunakan runtime Node.js. Versi Node yang dipakai di workstation pengembang, CI/CD, dan container frontend harus diperbaruhi ke versi yang terpatch.
2. **PHP bukan bagian dari stack inti proyek.** Tidak ada komponen PHP dalam arsitektur target. Namun, dokumentasi ini tetap dibuat sebagai referensi administrasi server dan pengawasan aset yang mungkin menyisipkan PHP (misalnya alat pihak ketiga, panel kontrol, atau legacy tooling di lingkungan operasi DPKP).
3. Backend Go 1.23 tidak langsung terpengaruh oleh keduanya, namun dependensi transitif dan alat build yang berbasis Node.js perlu di-audit.

---

## 4. Langkah Antisipasi dan Remediasi

### 4.1 Untuk Node.js (frontend dan tooling build)

1. **Periksa versi Node.js yang digunakan:**
   ```powershell
   node --version
   npm --version
   ```
2. **Perbarui ke versi yang aman:**
   - Jika menggunakan Node.js 22.x → minimal **v22.23.4**
   - Jika menggunakan Node.js 24.x → minimal **v24.22.1**
   - Jika menggunakan Node.js 26.x → minimal **v26.12.1**
3. **Perbarui dependensi frontend:**
   ```powershell
   cd frontend
   npm ci
   npm audit
   npm audit fix
   ```
4. **Perbarui image container frontend** di `infra/compose.yml` agar menggunakan base image Node.js yang telah diperbaiki, lalu rebuild dan redeploy.
5. **Verifikasi build frontend:**
   ```powershell
   cd frontend
   npm run build
   npm run lint
   npm run typecheck
   ```

### 4.2 Untuk PHP (referensi administrasi server)

PHP tidak digunakan dalam stack inti, tetapi jika terdapat instalasi PHP di lingkungan operasi atau server pendukung:

1. **Inventarisasi setiap server dan aplikasi yang menjalankan PHP**, mulai dari sistem yang menghadap internet.
2. **Perbarui ke versi yang aman** sesuai cabang yang digunakan (8.5.12, 8.4.27, 8.3.36, atau 8.2.35).
3. **Restart setiap worker PHP** setelah pembaruan agar pemeriksaan berlaku.
4. **Untuk versi EOL (8.1 dan lebih lama):** gunakan paket distribusi atau LTS vendor yang menyediakan backport, atau upgrade ke cabang yang didukung.
5. **Batasi akses internet** ke sistem PHP yang tidak dapat diperbarui segera, sampai patch terverifikasi.
6. Perhatikan bahwa tidak ada perubahan konfigurasi PHP yang sepenuhnya mitigasi CVE-2026-109618 dan CVE-2026-109619.

### 4.3 Monitoring dan Observabilitas

1. Pasang **Sentry** untuk pelacakan error frontend/backend sesuai rencana di `todo_list.md` Tahap 4.
2. Aktifkan notifikasi **Discord Webhook** untuk peringatan kegagalan build atau anomalis runtime.
3. Pantasi rilis keamanan Node.js melalui mailing list `nodejs-sec` dan rilis PHP melalui `php.net/archive`.
4. Catat versi runtime Node.js dan PHP dalam dokumen deployment setiap environment.

---

## 5. Catatan Verifikasi

- Informasi di atas diverifikasi dari sumber resmi Node.js, PHP, dan CERT-EU pada 23 Oktober 2026.
- Proyek saat ini masih pada tahap perencanaan/fondasi; integrasi SSO, SPL, Sadewa, MinIO, dan containerisasi belum tersedia sepenuhnya. Dokumen ini akan disesuaikan saat implementasi container frontend dan backend tersedia.
- Tidak ada kredensial, token, atau data pribadi yang dicatat dalam dokumen ini, sesuai konvensi keamanan proyek.

---

## Referensi

1. Node.js Security Releases — Oktober 2026: https://nodejs.org/en/blog/vulnerability/october-2026-security-releases
2. CERT-EU Security Advisory 2026-018 — Critical Vulnerabilities in PHP: https://cert.europa.eu/publications/security-advisories/2026-018
3. watchTowr PHP2Shell FAQ — CVE-2026-109618 & CVE-2026-109619: https://watchtowr.com/intelligence/php-cve-2026-109618-faq/
4. NVD — CVE-2026-109618: https://nvd.nist.gov/vuln/detail/CVE-2026-109618
5. NVD — CVE-2026-109619: https://nvd.nist.gov/vuln/detail/CVE-2026-109619
6. PHP Security Advisory — GHSA-w8hw-w9fx-j79w: https://github.com/php/php-src/security/advisories/GHSA-w8hw-w9fx-j79w
7. PHP Security Advisory — GHSA-6wmf-4rv5-p3rp: https://github.com/php/php-src/security/advisories/GHSA-6wmf-4rv5-p3rp
