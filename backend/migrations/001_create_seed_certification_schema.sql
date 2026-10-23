-- Skema modul Sertifikasi Benih BP3MBTP (berdasarkan alur XML 7 aktor)
-- Fase 1: Verifikasi Lapangan | Fase 2: Pengujian Lab Mutu | Fase 3: Sertifikasi & Penerbitan Label

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    npp VARCHAR(50),
    unit_kerja VARCHAR(100),
    role ENUM('pelanggan','admin','pengawas','analis','penandatangan','legalisasi','percetakan') NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permohonan_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nomor_permohonan VARCHAR(100) NOT NULL UNIQUE,
    pelanggan_id BIGINT UNSIGNED NOT NULL,
    jenis_benih VARCHAR(100) NOT NULL,
    varietas VARCHAR(100),
    luas_lahan DECIMAL(12,4),
    lokasi_kebun TEXT,
    status ENUM('diajukan','dibayar','diverifikasi','diinspeksi','divalidasi','ditolak','dibatalkan') NOT NULL DEFAULT 'diajukan',
    pembayaran_awal_status ENUM('belum','lunas') NOT NULL DEFAULT 'belum',
    buku_induk_nomor VARCHAR(100),
    laporan_verifikasi TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (pelanggan_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS inspeksi_fase (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    permohonan_id BIGINT UNSIGNED NOT NULL,
    fase TINYINT UNSIGNED NOT NULL COMMENT '1..4 fase inspeksi lapangan',
    pengawas_id BIGINT UNSIGNED NOT NULL,
    hasil ENUM('lulus','perlu_perbaikan','tidak_lulus') NOT NULL,
    catatan TEXT,
    tanggal_inspeksi DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id) ON DELETE CASCADE,
    FOREIGN KEY (pengawas_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS buku_induk (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    permohonan_id BIGINT UNSIGNED NOT NULL UNIQUE,
    nomor_buku_induk VARCHAR(100) NOT NULL UNIQUE,
    catatan TEXT,
    dibuat_oleh BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id) ON DELETE CASCADE,
    FOREIGN KEY (dibuat_oleh) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pembayaran (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    permohonan_id BIGINT UNSIGNED NOT NULL,
    jenis ENUM('awal','akhir') NOT NULL,
    jumlah DECIMAL(18,2) NOT NULL,
    status ENUM('belum','lunas','gagal') NOT NULL DEFAULT 'belum',
    bukti_transfer VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sampel_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    permohonan_id BIGINT UNSIGNED NOT NULL,
    kode_sampel VARCHAR(100) NOT NULL UNIQUE COMMENT 'Kode unik pengujian sampel',
    pengawas_id BIGINT UNSIGNED NOT NULL COMMENT 'Pengawas yang mengambil sampel',
    analis_id BIGINT UNSIGNED COMMENT 'Analis yang menguji sampel',
    jumlah_sampel INT UNSIGNED NOT NULL DEFAULT 1,
    tanggal_pengambilan DATE,
    status ENUM('dikumpulkan','dikodekan','diuji','selesai') NOT NULL DEFAULT 'dikumpulkan',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id) ON DELETE CASCADE,
    FOREIGN KEY (pengawas_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (analis_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS uji_mutu (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sampel_id BIGINT UNSIGNED NOT NULL,
    parameter VARCHAR(100) NOT NULL COMMENT 'misal: keutuhan, vitalitas, kelembaban, variasi genetik',
    hasil TEXT NOT NULL,
    status_lulus TINYINT(1) NOT NULL DEFAULT 0,
    catatan TEXT,
    diuji_oleh BIGINT UNSIGNED NOT NULL,
    tanggal_uji DATE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (sampel_id) REFERENCES sampel_benih(id) ON DELETE CASCADE,
    FOREIGN KEY (diuji_oleh) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS laporan_mutu (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sampel_id BIGINT UNSIGNED NOT NULL UNIQUE,
    kesimpulan ENUM('lulus','tidak_lulus') NOT NULL,
    isi_laporan TEXT NOT NULL,
    disusun_oleh BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (sampel_id) REFERENCES sampel_benih(id) ON DELETE CASCADE,
    FOREIGN KEY (disusun_oleh) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS sertifikat_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    permohonan_id BIGINT UNSIGNED NOT NULL UNIQUE,
    nomor_sertifikat VARCHAR(100) NOT NULL UNIQUE,
    penandatangan_id BIGINT UNSIGNED NOT NULL COMMENT 'Pejabat penandatangan TTE',
    tanggal_ttd DATETIME,
    tte_status ENUM('belum','ditandatangani','gagal') NOT NULL DEFAULT 'belum',
    tte_dokumen_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id) ON DELETE CASCADE,
    FOREIGN KEY (penandatangan_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS draft_label (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sertifikat_id BIGINT UNSIGNED NOT NULL UNIQUE,
    kode_barcode VARCHAR(100) NOT NULL UNIQUE COMMENT 'Kode unik/barcode terintegrasi SIM Perbenihan',
    status ENUM('draft','disetujui','dicetak','legalisasi') NOT NULL DEFAULT 'draft',
    integrasi_sim_status ENUM('belum','terkirim','gagal') NOT NULL DEFAULT 'belum',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (sertifikat_id) REFERENCES sertifikat_benih(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS label_legalisasi (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    draft_label_id BIGINT UNSIGNED NOT NULL UNIQUE,
    legalisasi_id BIGINT UNSIGNED NOT NULL COMMENT 'Petugas legalisasi',
    status ENUM('belum','diverifikasi','ditolak','selesai') NOT NULL DEFAULT 'belum',
    catatan TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (draft_label_id) REFERENCES draft_label(id) ON DELETE CASCADE,
    FOREIGN KEY (legalisasi_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS penyerahan_dokumen (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    permohonan_id BIGINT UNSIGNED NOT NULL UNIQUE,
    sertifikat_dikirim TINYINT(1) NOT NULL DEFAULT 0,
    label_dikirim TINYINT(1) NOT NULL DEFAULT 0,
    pembayaran_akhir_status ENUM('belum','lunas','gagal') NOT NULL DEFAULT 'belum',
    tanggal_penyerahan DATETIME,
    diterima_oleh VARCHAR(255),
    catatan TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    actor_id BIGINT UNSIGNED,
    actor_role VARCHAR(50),
    action VARCHAR(100) NOT NULL,
    entity VARCHAR(100) NOT NULL,
    entity_id BIGINT UNSIGNED,
    detail TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
