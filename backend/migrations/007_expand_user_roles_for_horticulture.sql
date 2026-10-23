-- Role operasional registrasi hortikultura. Akun SSO tetap diberi role
-- oleh administrator; role claims dari client/SSO tidak dipakai sebagai otorisasi.
ALTER TABLE users
    MODIFY COLUMN role ENUM(
        'pelanggan','admin','pengawas','analis','penandatangan','legalisasi','percetakan',
        'verifikator','inspektor','pbt','kasi_pmbtp'
    ) NOT NULL;
