-- Permohonan registrasi lahan hortikultura dan tautan lampiran privat.
CREATE TABLE IF NOT EXISTS horticulture_applications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    application_number VARCHAR(40) NOT NULL UNIQUE,
    applicant_id BIGINT UNSIGNED NOT NULL,
    organization_name VARCHAR(200) NOT NULL,
    contact_name VARCHAR(150) NOT NULL,
    phone VARCHAR(40) NOT NULL,
    address TEXT NOT NULL,
    village VARCHAR(120) NOT NULL,
    district VARCHAR(120) NOT NULL,
    regency VARCHAR(120) NOT NULL,
    cultivated_area DECIMAL(12,4) NOT NULL,
    crop_details JSON NOT NULL,
    status ENUM('diajukan','diverifikasi','perlu_perbaikan','dijadwalkan','diinspeksi','dinilai','disetujui','ditolak') NOT NULL DEFAULT 'diajukan',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_horticulture_applicant (applicant_id, id),
    KEY idx_horticulture_status (status, id),
    CONSTRAINT fk_horticulture_applicant FOREIGN KEY (applicant_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS horticulture_application_documents (
    application_id BIGINT UNSIGNED NOT NULL,
    document_id BIGINT UNSIGNED NOT NULL,
    document_type ENUM('proposal','organization','members','farm_form','farm_map','other') NOT NULL,
    PRIMARY KEY (application_id, document_id),
    KEY idx_horticulture_document (document_id),
    CONSTRAINT fk_horticulture_document_application FOREIGN KEY (application_id) REFERENCES horticulture_applications(id) ON DELETE CASCADE,
    CONSTRAINT fk_horticulture_document_file FOREIGN KEY (document_id) REFERENCES dokumen(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
