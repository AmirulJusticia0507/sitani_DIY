-- Batch impor repository data lokal. Kolom dataset tetap fleksibel sampai
-- template resmi, satuan, wilayah, dan aturan duplikasi ditetapkan.
CREATE TABLE IF NOT EXISTS dataset_imports (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nama_dataset VARCHAR(120) NOT NULL,
    nama_file VARCHAR(255) NOT NULL,
    headers_json JSON NOT NULL,
    jumlah_baris INT UNSIGNED NOT NULL,
    baris_json JSON NOT NULL,
    diunggah_oleh BIGINT UNSIGNED NOT NULL,
    dibuat_pada TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_dataset_imports_uploader (diunggah_oleh),
    KEY idx_dataset_imports_name (nama_dataset),
    CONSTRAINT fk_dataset_imports_user FOREIGN KEY (diunggah_oleh) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
