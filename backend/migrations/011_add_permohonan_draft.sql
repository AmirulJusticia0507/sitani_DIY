-- Status draft pada permohonan benih.
--
-- Pelanggan dapat menyimpan permohonan sebagai draft sebelum mengajukannya.
-- Draft belum dianggap pengajuan resmi: nomor resmi SPB-YYYYMMDD-XXXX baru
-- dibuat saat draft diajukan (kolom nomor_permohonan masih NOT NULL UNIQUE,
-- jadi draft memakai nomor sementara berawalan 'DRAFT-' yang tidak pernah
-- ditampilkan ke pengguna). Draft juga belum dapat dibayar, diinspeksi, atau
-- diregistrasi lot.
--
-- Status diubah dari ENUM menjadi VARCHAR agar transisi berikutnya dapat
-- ditambahkan tanpa mengunci daftar nilai pada tipe kolom; transisi yang sah
-- dijaga di lapisan service. Status bawaan baru adalah 'draft'.
--
-- Kolom diajukan_pada mencatat saat draft resmi diajukan, terpisah dari
-- created_at (saat draft pertama kali dibuat).

ALTER TABLE permohonan_benih MODIFY COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';

ALTER TABLE permohonan_benih
    ADD COLUMN diajukan_pada TIMESTAMP NULL DEFAULT NULL AFTER status;
