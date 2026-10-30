-- Peran superadmin: administrator tertinggi untuk lingkungan lokal/UAT.
-- Tetap berada di tabel users yang sama; pemetaan modulnya ada di frontend
-- (role-access.ts) dan tidak otomatis mendapat izin baru di backend kecuali
-- ditambahkan eksplisit.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('pelanggan','admin','pengawas','analis','penandatangan','legalisasi','percetakan','verifikator','inspektor','pbt','kasi_pmbtp','superadmin'));
