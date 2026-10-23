-- Nomor permohonan dibuat server dengan format SPB-YYYYMMDD-XXXX berdasarkan
-- tanggal pengajuan dan urutan harian. Migrasi ini (PostgreSQL) menyeragamkan
-- nomor lama, misalnya SPB-1-1761000000, menjadi format baru. Nomor yang sudah
-- sesuai format dibiarkan.
WITH normalisasi AS (
    SELECT id,
           'SPB-' || to_char(created_at, 'YYYYMMDD') AS prefix,
           ROW_NUMBER() OVER (
               PARTITION BY to_char(created_at, 'YYYYMMDD')
               ORDER BY created_at, id
           ) AS urut
    FROM permohonan_benih
    WHERE nomor_permohonan !~ '^SPB-[0-9]{8}-[0-9]{4,}$'
)
UPDATE permohonan_benih p
SET nomor_permohonan = n.prefix || '-' || lpad(n.urut::text, 4, '0')
FROM normalisasi n
WHERE p.id = n.id;
