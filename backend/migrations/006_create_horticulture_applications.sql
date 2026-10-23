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
    status ENUM('diajukan','perlu_perbaikan','diverifikasi','dijadwalkan','diinspeksi','dinilai','disetujui','ditolak') NOT NULL DEFAULT 'diajukan',
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

CREATE TABLE IF NOT EXISTS horticulture_inspection_schedules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    application_id BIGINT UNSIGNED NOT NULL,
    inspector_id BIGINT UNSIGNED NOT NULL,
    scheduled_at DATETIME NOT NULL,
    location VARCHAR(255) NOT NULL,
    notes TEXT NULL,
    status ENUM('terjadwal','selesai','dibatalkan') NOT NULL DEFAULT 'terjadwal',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_horticulture_schedule_application_once (application_id),
    KEY idx_horticulture_schedule_application (application_id, scheduled_at),
    KEY idx_horticulture_schedule_inspector (inspector_id, scheduled_at),
    CONSTRAINT fk_horticulture_schedule_application FOREIGN KEY (application_id) REFERENCES horticulture_applications(id) ON DELETE CASCADE,
    CONSTRAINT fk_horticulture_schedule_inspector FOREIGN KEY (inspector_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS horticulture_inspections (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    schedule_id BIGINT UNSIGNED NOT NULL UNIQUE,
    checklist_json JSON NOT NULL,
    findings TEXT NOT NULL,
    recommendations TEXT NOT NULL,
    inspector_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT fk_horticulture_inspection_schedule FOREIGN KEY (schedule_id) REFERENCES horticulture_inspection_schedules(id) ON DELETE RESTRICT,
    CONSTRAINT fk_horticulture_inspection_inspector FOREIGN KEY (inspector_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS horticulture_evaluations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    application_id BIGINT UNSIGNED NOT NULL,
    assessment_type ENUM('pohon_induk','pit','rip','pi','ri') NOT NULL,
    plant_name VARCHAR(160) NOT NULL,
    variety VARCHAR(160) NOT NULL,
    vigor_score TINYINT UNSIGNED NOT NULL,
    health_score TINYINT UNSIGNED NOT NULL,
    variety_score TINYINT UNSIGNED NOT NULL,
    recommendation ENUM('layak','perlu_perbaikan','tidak_layak') NOT NULL,
    notes TEXT NOT NULL,
    evaluator_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_horticulture_evaluation_application (application_id, id),
    CONSTRAINT fk_horticulture_evaluation_application FOREIGN KEY (application_id) REFERENCES horticulture_applications(id) ON DELETE CASCADE,
    CONSTRAINT fk_horticulture_evaluation_user FOREIGN KEY (evaluator_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
