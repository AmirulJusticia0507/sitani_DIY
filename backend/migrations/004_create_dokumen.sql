-- Metadata dokumen/lampiran yang berkas fisiknya tersimpan privat di
-- object storage (MinIO/S3). Unduhan selalu lewat pemeriksaan hak akses
-- dan URL presign berumur pendek; tidak ada akses langsung ke objek.

CREATE TABLE IF NOT EXISTS dokumen (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    pemilik_id BIGINT UNSIGNED NOT NULL,
    tujuan VARCHAR(255) NULL COMMENT 'lokasi penggunaan dokumen, mis. permohonan_elemen',
    object_key VARCHAR(500) NOT NULL UNIQUE COMMENT 'kunci objek di bucket (terpisah tiap pemilik)',
    nama_asli VARCHAR(255) NOT NULL,
    jenis VARCHAR(100) NOT NULL COMMENT 'tipe konten yang diizinkan',
    ukuran BIGINT UNSIGNED NOT NULL COMMENT 'ukuran dalam byte',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_dokumen_pemilik (pemilik_id),
    CONSTRAINT fk_dokumen_pemilik FOREIGN KEY (pemilik_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
