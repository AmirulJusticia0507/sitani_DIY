# Antisipasi Kerentanan Keamanan Go, Node.js, dan PHP — Oktober 2026

Tanggal dokumentasi: 23 Oktober 2026  
Sumber verifikasi: advisory resmi Go (20–21 Oktober 2026), rilis keamanan resmi Node.js (21 Oktober 2026), pengumuman PHP (21 Oktober 2026), dan peringatan CERT-EU 2026-018 (21 Oktober 2026).

Dokumen ini mencatat tiga berita kerentanan yang diterbitkan sekitar 21–22 Oktober 2026 dan langkah anticipasi yang relevan bagi proyek Portal Pertanian dan Ketahanan Pangan DIY.

---

## 1. Ringkasan Kerentanan

| # | Produk | Tanggal Rilis Keamanan | ID Kerentanan | Tingkat Tertinggi | Versi Terkena Dampak | Versi Tanggal |
|---|--------|------------------------|---------------|-------------------|----------------------|---------------|
| 1 | Go (Golang) | 20–21 Oktober 2026 | CVE-2026-111943 | HIGH / Critical | Go 1.26, 1.27, dan 1.25 atau lebih lama | Go 1.26.10, Go 1.27.3 |
| 2 | Node.js | 21 Oktober 2026 | CVE-2026-111497 dan 9 CVE lain | HIGH | 26.x, 24.x, 22.x | v22.23.4, v24.22.1, v26.12.1 |
| 3 | PHP | 21–22 Oktober 2026 | CVE-2026-109618, CVE-2026-109619 | Critical (CVSS 9.2) | 8.5 < 8.5.12, 8.4 < 8.4.27, 8.3 < 8.3.36, 8.2 < 8.2.35, cabang EOL | 8.5.12, 8.4.27, 8.3.36, 8.2.35 |

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

### 2.3 Go (Golang) — CVE-2026-111943 (eMudhra Intermediate CA Compromise)

Pada 20–21 Oktober 2026, Go merilis pembaruan darurat untuk kerentanan kepercayaan sertifikat yang dikaitkan dengan kompromi sertifikat perantara (intermediate CA) **eMudhra**. Go menggunakan sistem root certificates untuk verifikasi TLS, sehingga aplikasi yang menggunakan verifier sertifikat Go (`crypto/x509`) dapat mempercayai rantai sertifikat yang seharusnya tidak diinginkan.

- **CVE-2026-111943 (HIGH/Critical):** Kerentanan bypass integritas karena validasi checksum modul yang tidak tepat melalui modul berbahaya, serta kompromi sertifikat perantara eMudhra yang menyebabkan rantai sertifikat TLS dipercayai meskipun tidak valid. Patch Go memfilter rantai sertifikat menggunakan delapan hash SHA-256 SPKI (SubjectPublicKeyInfo) yang terlarang di `crypto/x509`, mencakup verifier Go dan jalur verifikasi macOS serta Windows.

**Versi yang terpengaruh dan diperbaiki:**

| Cabang Go | Versi yang Terpengaruh | Versi yang Diperbaiki |
|-----------|------------------------|-----------------------|
| Go 1.27.x | Sebelum 1.27.3 | 1.27.3 atau lebih baru |
| Go 1.26.x | Sebelum 1.26.10 | 1.26.10 atau lebih baru |
| Go 1.25 dan lebih lama | Semua versi yang rentan | Pindah ke cabang yang didukung, atau bukti backport dari vendor |

**Relevansi untuk SiTani DIY:** Backend SiTani DIY menggunakan **Go 1.23** (lihat `analisis_kebutuhan.md` dan `readme.md`). Go 1.23 termasuk dalam cakupan "Go 1.25 dan lebih lama" yang perlu diperbarui ke cabang yang didukung atau memperoleh bukti backport dari vendor. Toolchain Go yang digunakan di CI/CD, workstation pengembang, dan container backend harus dicek dan diperbarui.

Sumber resmi:
- https://github.com/golang/go/issues/82345
- https://pkg.go.dev/vuln/list
- https://www.mozilla.org/en-US/security/advisories/mfsa2026-111/ (konfirmasi kompromi eMudhra)

---

## 3. Dampak terhadap Proyek SiTani DIY

Berdasarkan `analisis_kebutuhan.md` dan `docs/analisis-proyek.md`, stack proyek saat ini adalah:

| Komponen | Teknologi | Terpengaruh CVE Go? | Terpengaruh CVE Node.js? | Terpengaruh CVE PHP? |
|----------|-----------|---------------------|--------------------------|----------------------|
| Backend API | Go 1.23 | **Ya** (CVE-2026-111943) | Tidak | Tidak |
| Frontend Web | Next.js 16 (berjalan di atas Node.js) | Tidak | **Ya** (runtime Node.js) | Tidak |
| Database | MySQL | Tidak | Tidak | Tidak |
| Object Storage | MinIO | Tidak | Tidak | Tidak |
| TTE / Central Sign | Sadewa | Tidak | Tidak | Tidak |
| API Gateway | SPL / Cakra API | Tidak | Tidak | Tidak |
| Autentikasi | SSO JSP/Satriya | Tidak | Tidak | Tidak |

**Kesimpulan dampak:**

1. **Go relevan dan PRIORITAS TERTINGGI** karena backend API SiTani DIY menggunakan **Go 1.23**, yang termasuk dalam cakupan CVE-2026-111943 (Go 1.25 dan lebih lama). Toolchain Go di workstation pengembang, CI/CD, dan container backend harus diperbaruhi ke cabang yang didukung (minimal Go 1.26.10 atau 1.27.3, atau rilis terbaru yang didukung).
2. **Node.js relevan** karena frontend Next.js di-build dan dijalankan menggunakan runtime Node.js. Versi Node yang dipakai di workstation pengembang, CI/CD, dan container frontend harus diperbaruhi ke versi yang terpatch.
3. **PHP bukan bagian dari stack inti proyek.** Tidak ada komponen PHP dalam arsitektur target. Namun, dokumentasi ini tetap dibuat sebagai referensi administrasi server dan pengawasan aset yang mungkin menyisipkan PHP (misalnya alat pihak ketiga, panel kontrol, atau legacy tooling di lingkungan operasi DPKP).

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

### 4.3 Untuk Go (backend API — prioritas tertinggi)

Backend SiTani DIY menggunakan **Go 1.23**, yang terkena dampak CVE-2026-111943. Langkah remediasi:

1. **Periksa versi Go yang digunakan di workstation, CI/CD, dan container backend:**
   ```powershell
   go version
   ```
   Untuk memeriksa versi Go yang tertanam dalam binary yang sudah dibangun:
   ```powershell
   go version -m /path/to/your-application
   ```

2. **Perbarui toolchain Go ke versi yang aman:**
   - Go 1.26.x → minimal **1.26.10**
   - Go 1.27.x → minimal **1.27.3**
   - Go 1.25 dan lebih lama (termasuk Go 1.23 yang dipakai SiTani) → pindah ke cabang yang didukung, atau dapatkan bukti backport dari vendor

3. **Perbarui `GOTOOLCHAIN` dan pengaturan build di CI, Docker, serta file version-manager:**
   ```powershell
   GOTOOLCHAIN=go1.27.3 go version
   GOTOOLCHAIN=go1.27.3 go test ./...
   GOTOOLCHAIN=go1.27.3 go build ./...
   ```

4. **Perbarui image container backend** di `infra/compose.yml` agar menggunakan base image Go yang telah diperbaiki, lalu rebuild dan redeploy.

5. **Perbarui dependensi Go modul backend:**
   ```powershell
   cd backend
   go mod tidy
   go mod verify
   ```

6. **Jalankan pemeriksaan kerentanan Go:**
   ```powershell
   go install golang.org/x/vuln/cmd/govulncheck@latest
   govulncheck ./...
   ```

7. **Verifikasi build backend:**
   ```powershell
   cd backend
   go build ./...
   go test ./...
   ```

### 4.4 Monitoring dan Observabilitas

1. Pasang **Sentry** untuk pelacakan error frontend/backend sesuai rencana di `todo_list.md` Tahap 4.
2. Aktifkan notifikasi **Discord Webhook** untuk peringatan kegagalan build atau anomalis runtime.
3. Pantasi rilis keamanan Go melalui `go.dev/doc/devel/release`, rilis Node.js melalui mailing list `nodejs-sec`, dan rilis PHP melalui `php.net/archive`.
4. Catat versi runtime Go, Node.js, dan PHP dalam dokumen deployment setiap environment.

---

## 5. Catatan Verifikasi

- Informasi di atas diverifikasi dari sumber resmi Go, Node.js, PHP, dan CERT-EU pada 23 Oktober 2026.
- Proyek saat ini masih pada tahap perencanaan/fondasi; integrasi SSO, SPL, Sadewa, MinIO, dan containerisasi belum tersedia sepenuhnya. Dokumen ini akan disesuaikan saat implementasi container frontend dan backend tersedia.
- Tidak ada kredensial, token, atau data pribadi yang dicatat dalam dokumen ini, sesuai konvensi keamanan proyek.

---

## Referensi

1. Go Security Advisory — CVE-2026-111943: https://github.com/golang/go/issues/82345
2. Go Vulnerability Database: https://pkg.go.dev/vuln/list
3. Mozilla Advisory — eMudhra Intermediate CA Compromise: https://www.mozilla.org/en-US/security/advisories/mfsa2026-111/
4. Node.js Security Releases — Oktober 2026: https://nodejs.org/en/blog/vulnerability/october-2026-security-releases
5. CERT-EU Security Advisory 2026-018 — Critical Vulnerabilities in PHP: https://cert.europa.eu/publications/security-advisories/2026-018
6. watchTowr PHP2Shell FAQ — CVE-2026-109618 & CVE-2026-109619: https://watchtowr.com/intelligence/php-cve-2026-109618-faq/
7. NVD — CVE-2026-109618: https://nvd.nist.gov/vuln/detail/CVE-2026-109618
8. NVD — CVE-2026-109619: https://nvd.nist.gov/vuln/detail/CVE-2026-109619
9. PHP Security Advisory — GHSA-w8hw-w9fx-j79w: https://github.com/php/php-src/security/advisories/GHSA-w8hw-w9fx-j79w
10. PHP Security Advisory — GHSA-6wmf-4rv5-p3rp: https://github.com/php/php-src/security/advisories/GHSA-6wmf-4rv5-p3rp
