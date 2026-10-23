-- Sesi login lokal untuk development/UAT (paket backend/internal/auth/devlogin).
-- Hanya berisi hash SHA-256 dari token acak; token mentah hanya ada di cookie
-- HttpOnly. Tabel tidak dipakai oleh alur SSO produksi.
CREATE TABLE IF NOT EXISTS auth_sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    token_hash CHAR(64) NOT NULL UNIQUE COMMENT 'SHA-256 hex dari token cookie',
    realm VARCHAR(32) NOT NULL DEFAULT 'lokal',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_auth_sessions_expires (expires_at),
    CONSTRAINT fk_auth_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
