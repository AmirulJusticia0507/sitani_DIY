-- Skema registrasi benih, persediaan per lot, dan penjualan benih.
-- Catatan: format nomor registrasi/dokumen resmi masih menunggu validasi BP3MBTP.

CREATE TABLE IF NOT EXISTS registrasi_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nomor_registrasi VARCHAR(120) NOT NULL UNIQUE COMMENT 'Format internal RB-<nomor_lot>',
    permohonan_id BIGINT UNSIGNED NOT NULL,
    sertifikat_id BIGINT UNSIGNED,
    draft_label_id BIGINT UNSIGNED,
    nomor_sertifikat VARCHAR(100) NOT NULL,
    kode_label VARCHAR(100) NOT NULL,
    komoditas VARCHAR(100) NOT NULL,
    varietas VARCHAR(100) NOT NULL,
    produsen VARCHAR(200) NOT NULL,
    sumber_benih VARCHAR(200) NOT NULL,
    nomor_lot VARCHAR(100) NOT NULL UNIQUE,
    kelas_benih ENUM('penjenis','dasar','pokok','sebar') NOT NULL,
    jumlah DECIMAL(18,3) NOT NULL,
    satuan VARCHAR(20) NOT NULL,
    tanggal_produksi DATE NOT NULL,
    tanggal_kedaluwarsa DATE NOT NULL,
    status_edar ENUM('aktif','ditangguhkan','dicabut','kedaluwarsa') NOT NULL DEFAULT 'aktif',
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (permohonan_id) REFERENCES permohonan_benih(id),
    FOREIGN KEY (sertifikat_id) REFERENCES sertifikat_benih(id),
    FOREIGN KEY (draft_label_id) REFERENCES draft_label(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_registrasi_status_edar (status_edar),
    INDEX idx_registrasi_kedaluwarsa (tanggal_kedaluwarsa)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stok_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    registrasi_id BIGINT UNSIGNED NOT NULL UNIQUE,
    jumlah_tersedia DECIMAL(18,3) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (registrasi_id) REFERENCES registrasi_benih(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS mutasi_stok_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    registrasi_id BIGINT UNSIGNED NOT NULL,
    jenis ENUM('masuk','keluar','koreksi') NOT NULL,
    jumlah DECIMAL(18,3) NOT NULL,
    saldo_setelah DECIMAL(18,3) NOT NULL,
    referensi_tipe VARCHAR(50) NOT NULL,
    referensi_id BIGINT UNSIGNED,
    catatan TEXT,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (registrasi_id) REFERENCES registrasi_benih(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_mutasi_registrasi (registrasi_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS penjualan_benih (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nomor_dokumen VARCHAR(120) NOT NULL UNIQUE,
    registrasi_id BIGINT UNSIGNED NOT NULL,
    nomor_lot VARCHAR(100) NOT NULL COMMENT 'Snapshot nomor lot saat transaksi',
    pembeli VARCHAR(200) NOT NULL,
    jumlah DECIMAL(18,3) NOT NULL,
    harga_satuan DECIMAL(18,2) NOT NULL,
    total_harga DECIMAL(20,2) NOT NULL,
    tanggal_transaksi DATE NOT NULL,
    status ENUM('selesai','dibatalkan') NOT NULL DEFAULT 'selesai',
    alasan_pembatalan TEXT,
    dibatalkan_oleh BIGINT UNSIGNED,
    dibatalkan_pada DATETIME,
    created_by BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (registrasi_id) REFERENCES registrasi_benih(id),
    FOREIGN KEY (dibatalkan_oleh) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    INDEX idx_penjualan_registrasi (registrasi_id),
    INDEX idx_penjualan_tanggal (tanggal_transaksi),
    INDEX idx_penjualan_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
