-- Nomor permohonan dibuat server dengan format SPB-YYYYMMDD-XXXX berdasarkan
-- tanggal pengajuan dan urutan harian. Migrasi ini menyeragamkan nomor lama
-- (format SPB-<id>-<epoch Unix>) menjadi format baru. Nomor yang sudah sesuai
-- format dibiarkan.
WITH normalisasi AS (
    SELECT id,
           CONCAT('SPB-', DATE_FORMAT(created_at, '%Y%m%d')) AS prefix,
           ROW_NUMBER() OVER (
               PARTITION BY DATE_FORMAT(created_at, '%Y%m%d')
               ORDER BY created_at, id
           ) AS urut
    FROM permohonan_benih
)
UPDATE permohonan_benih p
JOIN normalisasi n ON n.id = p.id
SET p.nomor_permohonan = CONCAT(n.prefix, '-', LPAD(n.urut, 4, '0'))
WHERE p.nomor_permohonan NOT REGEXP '^SPB-[0-9]{8}-[0-9]{4,}$';
