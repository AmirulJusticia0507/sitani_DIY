-- Peran internal untuk alur hortikultura.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('pelanggan','admin','pengawas','analis','penandatangan','legalisasi','percetakan','verifikator','inspektor','pbt','kasi_pmbtp'));
