-- Status draft pada permohonan benih.
--
-- Pelanggan dapat menyimpan permohonan sebagai draft sebelum mengajukannya.
-- Draft belum dianggap pengajuan resmi: nomor resmi SPB-YYYYMMDD-XXXX baru
-- dibuat saat draft diajukan (kolom nomor_permohonan masih NOT NULL UNIQUE,
-- jadi draft memakai nomor sementara berawalan 'DRAFT-' yang tidak pernah
-- ditampilkan ke pengguna). Draft juga belum dapat dibayar, diinspeksi, atau
-- diregistrasi lot.
--
-- Status tetap VARCHAR dengan pemeriksaan nilai; transisi yang sah dijaga di
-- lapisan service. Status bawaan baru adalah 'draft'.
--
-- Kolom diajukan_pada mencatat saat draft resmi diajukan, terpisah dari
-- created_at (saat draft pertama kali dibuat).

ALTER TABLE permohonan_benih ALTER COLUMN status SET DEFAULT 'draft';

ALTER TABLE permohonan_benih DROP CONSTRAINT IF EXISTS permohonan_benih_status_check;
ALTER TABLE permohonan_benih ADD CONSTRAINT permohonan_benih_status_check
    CHECK (status IN ('draft','diajukan','dibayar','diverifikasi','diinspeksi','divalidasi','ditolak','dibatalkan'));

ALTER TABLE permohonan_benih
    ADD COLUMN IF NOT EXISTS diajukan_pada TIMESTAMP NULL DEFAULT NULL;
