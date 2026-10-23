# Analisis Kebutuhan Sistem: Portal Pertanian dan Ketahanan Pangan DIY

## 1. Latar Belakang & Permasalahan
Dinas Pertanian dan Ketahanan Pangan (DPKP) Pemda DIY memiliki berbagai jenis pelayanan publik (seperti Sertifikasi/Registrasi Lahan Hortikultura, Penilaian Pohon Induk/PIT/RIP/PI/RI, Form Permohonan Producer/Lahan, dsb.). 

**Masalah Utama:**
1. **Penyebaran Data:** Data pertanian dan ketahanan pangan tersebar di berbagai bidang, PIC, serta sebagian masih bersifat manual atau terpisah di aplikasi kementerian/pusat.
2. **Interoperabilitas:** Kesulitan dalam mengompilasi dan menyajikan data terpadu secara lintas bidang dan lintas sektor.
3. **Pengelolaan Layanan Publik:** Proses pengajuan dan penilaian layanan publik memerlukan alur kerja tersistem, mulai dari verifikasi dokumen, inspeksi lapangan, audit catatan budidaya (GAP), hingga penerbitan sertifikat/nomor registrasi dengan TTE (Tanda Tangan Elektronik).

## 2. Tujuan Proyek
Membuat **Portal Pertanian dan Ketahanan Pangan DIY** sebagai *single door platform* / *rumah aplikasi* terpadu yang dilengkapi dengan **Repository Data Lokal** guna menampung, mengolah, dan menampilkan visualisasi data secara terintegrasi.

---

## 3. Identifikasi Pengguna & Akses (Role & ACL)

### A. Pengguna Individu
1. **Individu Umum:** Masyarakat umum luas.
2. **Individu DIY:** Masyarakat ber-KTP DIY.
3. **ASN Pemerintah:** Aparatur Sipil Negara umum.
4. **ASN DIY:** ASN lingkungan Pemda DIY.
5. **Internal OPD (DPKP):** Staf/Pejabat internal DPKP DIY.

### B. Pengguna Organisasi / Pelaku Usaha
1. **Organisasi Umum / Pelaku Usaha / Petani / Kelompok Tani:** Pemohon sertifikasi, registrasi lahan, atau layanan hortikultura.
2. **OPD DIY:** Unit kerja pemda terkait.

---

## 4. Ruang Lingkup Fitur (Tahap 1)

### Modul 1: Repository Data Lokal DPKP
*   **Import Data (Excel/CSV):** Mendukung pengunggahan dan pembacaan berkas hasil ekspor dari aplikasi pusat/kementerian.
*   **API Data Integration:** Jalur sinkronisasi data masuk via API untuk aplikasi pusat yang telah menyediakan interoperabilitas.
*   **Dashboard Visualisasi Data:** Menampilkan rekapitulasi data pertanian, peta sebaran lahan, serta statistik pelayanan publik.

### Modul 2: Layanan Registrasi Lahan Usaha Hortikultura
*   **Pengajuan Permohonan:** Pengunggahan proposal/dokumen (surat permohonan, struktur organisasi, daftar anggota, form registrasi kebun, peta kebun).
*   **Verifikasi Berkas:** Verifikasi kelengkapan oleh Petugas Verifikator.
*   **Inspeksi & Audit Lapangan:** Penjadwalan inspeksi kebun & audit catatan budidaya (Good Agricultural Practices / GAP) oleh Petugas Inspektor.
*   **Evaluasi & Penilaian:** Penilaian kelayakan lahan/pohon induk (seperti parameter vigoritas, kesehatan, kebenaran varietas).
*   **Pengesahan & Penerbitan Nomor Registrasi:** TTE dokumen pengesahan oleh Kepala DPKP DIY via Sadewa.

---

## 5. Arsitektur & Spesifikasi Teknologis

```
[ Pengguna ] 
     │
     ▼
[ SSO Pemda DIY (JSP / Satriya) ]
     │
     ▼
[ SPL / Cakra API Gateway ] (spl.jogjaprov.go.id)
     │
     ├───────────────────────┼───────────────────────┐
     ▼                       ▼                       ▼
┌──────────────┐    ┌─────────────────┐    ┌────────────────┐
│ Container 1  │    │   Container 2   │    │  Container 3   │
│  (Database)  │◄──►│    (Backend)    │◄──►│   (Frontend)   │
│    MySQL     │    │   Golang 1.23   │    │  Next.js 15    │
└──────────────┘    └─────────────────┘    └────────────────┘
                             │
            ┌────────────────┼────────────────┐
            ▼                ▼                ▼
     ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
     │ CentralSign │  │ MinIO S3    │  │ Discord/    │
     │  (Sadewa)   │  │ (Storage)   │  │ Sentry Log  │
     └─────────────┘  └─────────────┘  └─────────────┘
```

*   **Backend:** Golang 1.23
*   **Frontend:** Next.js 15
*   **Database:** MySQL
*   **Containerization:** 3 Container (Database, Backend, Frontend)
*   **Authentication & Access:** SSO Pemda DIY (Realm JSP untuk Umum, Realm Satriya untuk ASN)
*   **API Gateway:** Sistem Penghubung Layanan (SPL / Cakra API)
*   **Digital Signature:** Sadewa Central Sign (TTE)
*   **Object Storage:** MinIO S3
*   **Monitoring & Notification:** Sentry (Error Logging) & Discord Webhook (Notifikasi)