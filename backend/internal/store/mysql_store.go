package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

func (s *DBStore) CreateUser(ctx context.Context, u db.User) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO users (username,email,npp,unit_kerja,role,is_active) VALUES (?,?,?,?,?,?)`,
		u.Username, u.Email, u.NPP, u.UnitKerja, u.Role, u.IsActive)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetUserByID(ctx context.Context, id int64) (*db.User, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,username,email,npp,unit_kerja,role,is_active,created_at,updated_at FROM users WHERE id=?`, id)
	return scanUser(row)
}

func (s *DBStore) GetUserByUsername(ctx context.Context, username string) (*db.User, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,username,email,npp,unit_kerja,role,is_active,created_at,updated_at FROM users WHERE username=?`, username)
	return scanUser(row)
}

// FormatNomorPermohonan membentuk nomor permohonan resmi dari tanggal pengajuan
// dan nomor urut harian, contoh: SPB-20261023-0001.
func FormatNomorPermohonan(tanggal time.Time, urut int64) string {
	return fmt.Sprintf("SPB-%s-%04d", tanggal.Format("20060102"), urut)
}

// CreatePermohonan menyimpan pengajuan baru beserta nomor permohonannya dalam
// satu transaksi. Urutan harian dihitung dari baris terakhir pada tanggal yang
// sama sehingga tidak bergantung pada jam server maupun ID dari client.
func (s *DBStore) CreatePermohonan(ctx context.Context, p db.PermohonanBenih, tanggal time.Time) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("mulai transaksi permohonan: %w", err)
	}
	defer tx.Rollback()

	// Nomor sementara wajib unik (kolom NOT NULL UNIQUE) dan tidak pernah
	// ditampilkan. Isinya langsung diganti nomor resmi dalam transaksi ini.
	sementara := fmt.Sprintf("TMP-%d", tanggal.UnixNano())
	if _, err := tx.ExecContext(ctx, `INSERT INTO permohonan_benih (nomor_permohonan,pelanggan_id,jenis_benih,varietas,luas_lahan,lokasi_kebun,status,pembayaran_awal_status) VALUES (?,?,?,?,?,?,?,?)`,
		sementara, p.PelangganID, p.JenisBenih, p.Varietas, p.LuasLahan, p.LokasiKebun, p.Status, p.PembayaranAwalStatus); err != nil {
		return 0, fmt.Errorf("insert permohonan: %w", err)
	}

	// Kunci advisory per tanggal menyerialkan penomoran sehingga dua pengajuan
	// bersamaan tidak mendapat nomor urut yang sama.
	kunci := "permohonan:" + tanggal.Format("20060102")
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext(?))`, kunci); err != nil {
		return 0, fmt.Errorf("kunci penomoran permohonan: %w", err)
	}

	prefix := FormatNomorPermohonan(tanggal, 0)[:len("SPB-")+8+1]
	var nomorTerakhir sql.NullString
	if err := tx.QueryRowContext(ctx, `SELECT MAX(nomor_permohonan) FROM permohonan_benih WHERE nomor_permohonan LIKE ?`, prefix+"%").Scan(&nomorTerakhir); err != nil {
		return 0, fmt.Errorf("ambil urutan nomor permohonan: %w", err)
	}
	urut := int64(1)
	if nomorTerakhir.Valid && len(nomorTerakhir.String) > len(prefix) {
		if n, err := strconv.ParseInt(nomorTerakhir.String[len(prefix):], 10, 64); err == nil {
			urut = n + 1
		}
	}
	nomor := FormatNomorPermohonan(tanggal, urut)

	var id int64
	if err := tx.QueryRowContext(ctx, `UPDATE permohonan_benih SET nomor_permohonan=? WHERE nomor_permohonan=? RETURNING id`, nomor, sementara).Scan(&id); err != nil {
		return 0, fmt.Errorf("tetapkan nomor permohonan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit permohonan: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetPermohonan(ctx context.Context, id int64) (*db.PermohonanBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+permohonanColumns+` FROM permohonan_benih WHERE id=?`, id)
	return scanPermohonan(row)
}

const permohonanColumns = `id,nomor_permohonan,pelanggan_id,jenis_benih,varietas,luas_lahan,lokasi_kebun,status,pembayaran_awal_status,buku_induk_nomor,laporan_verifikasi,created_at,updated_at`

// SetNomorPermohonan mengganti nomor permohonan, misalnya format lama ke
// format SPB-YYYYMMDD-XXXX saat migrasi data. Nomor tidak berasal dari client.
func (s *DBStore) SetNomorPermohonan(ctx context.Context, id int64, nomor string) error {
	result, err := s.db.ExecContext(ctx, `UPDATE permohonan_benih SET nomor_permohonan=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, nomor, id)
	if err != nil {
		return fmt.Errorf("simpan nomor permohonan: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListPermohonanSiapRegistrasi mengembalikan permohonan yang dokumen
// sertifikasinya sudah lengkap sehingga boleh diregistrasi menjadi lot benih.
// semua=false membatasi hasil pada pemilik tertentu.
func (s *DBStore) ListPermohonanSiapRegistrasi(ctx context.Context, pemilikID int64, semua bool, limit int) ([]db.SiapRegistrasiPermohonan, error) {
	query := `SELECT p.id, p.nomor_permohonan, p.jenis_benih, COALESCE(p.varietas,''), p.status
		FROM permohonan_benih p
		JOIN sertifikat_benih s ON s.permohonan_id = p.id
		JOIN draft_label d ON d.sertifikat_id = s.id
		LEFT JOIN registrasi_benih r ON r.permohonan_id = p.id
		WHERE s.tte_status = 'ditandatangani'
		  AND d.status = 'legalisasi'
		  AND d.integrasi_sim_status = 'terkirim'
		  AND p.status NOT IN (?, ?)
		  AND r.id IS NULL`
	args := []any{db.StatusPermohonanDitolak, db.StatusPermohonanDibatalkan}
	if !semua {
		query += ` AND p.pelanggan_id = ?`
		args = append(args, pemilikID)
	}
	query += ` ORDER BY p.created_at DESC, p.id DESC LIMIT ?`
	args = append(args, batasBaris(limit))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query permohonan siap registrasi: %w", err)
	}
	defer rows.Close()
	hasil := make([]db.SiapRegistrasiPermohonan, 0)
	for rows.Next() {
		var item db.SiapRegistrasiPermohonan
		if err := rows.Scan(&item.PermohonanID, &item.NomorPermohonan, &item.JenisBenih, &item.Varietas, &item.Status); err != nil {
			return nil, fmt.Errorf("pindai permohonan siap registrasi: %w", err)
		}
		hasil = append(hasil, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi permohonan siap registrasi: %w", err)
	}
	return hasil, nil
}

func (s *DBStore) UpdatePermohonanStatus(ctx context.Context, id int64, status, pembayaranAwalStatus string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE permohonan_benih SET status=?, pembayaran_awal_status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, status, pembayaranAwalStatus, id)
	if err != nil {
		return fmt.Errorf("update permohonan: %w", err)
	}
	return nil
}

func (s *DBStore) CreatePembayaran(ctx context.Context, pb db.Pembayaran) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO pembayaran (permohonan_id,jenis,jumlah,status,bukti_transfer) VALUES (?,?,?,?,?)`,
		pb.PermohonanID, pb.Jenis, pb.Jumlah, pb.Status, pb.BuktiTransfer)
	if err != nil {
		return 0, fmt.Errorf("insert pembayaran: %w", err)
	}
	return id, nil
}

func (s *DBStore) CreateInspeksiFase(ctx context.Context, i db.InspeksiFase) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO inspeksi_fase (permohonan_id,fase,pengawas_id,hasil,catatan,tanggal_inspeksi) VALUES (?,?,?,?,?,?)`,
		i.PermohonanID, i.Fase, i.PengawasID, i.Hasil, i.Catatan, i.TanggalInspeksi)
	if err != nil {
		return 0, fmt.Errorf("insert inspeksi: %w", err)
	}
	return id, nil
}

func (s *DBStore) ListInspeksiFase(ctx context.Context, permohonanID int64) ([]db.InspeksiFase, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,permohonan_id,fase,pengawas_id,hasil,catatan,tanggal_inspeksi,created_at FROM inspeksi_fase WHERE permohonan_id=? ORDER BY fase`, permohonanID)
	if err != nil {
		return nil, fmt.Errorf("query inspeksi: %w", err)
	}
	defer rows.Close()
	var out []db.InspeksiFase
	for rows.Next() {
		var f db.InspeksiFase
		if err := rows.Scan(&f.ID, &f.PermohonanID, &f.Fase, &f.PengawasID, &f.Hasil, &f.Catatan, &f.TanggalInspeksi, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan inspeksi: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *DBStore) CreateBukuInduk(ctx context.Context, b db.BukuInduk) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO buku_induk (permohonan_id,nomor_buku_induk,catatan,dibuat_oleh) VALUES (?,?,?,?)`,
		b.PermohonanID, b.NomorBukuInduk, b.Catatan, b.DibuatOleh)
	if err != nil {
		return 0, fmt.Errorf("insert buku induk: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetBukuInduk(ctx context.Context, permohonanID int64) (*db.BukuInduk, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,permohonan_id,nomor_buku_induk,catatan,dibuat_oleh,created_at FROM buku_induk WHERE permohonan_id=?`, permohonanID)
	return scanBukuInduk(row)
}

func (s *DBStore) CreateSampel(ctx context.Context, s2 db.SampelBenih) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO sampel_benih (permohonan_id,kode_sampel,pengawas_id,analis_id,jumlah_sampel,tanggal_pengambilan,status) VALUES (?,?,?,?,?,?,?)`,
		s2.PermohonanID, s2.KodeSampel, s2.PengawasID, s2.AnalisID, s2.JumlahSampel, s2.TanggalPengambilan, s2.Status)
	if err != nil {
		return 0, fmt.Errorf("insert sampel: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetSampel(ctx context.Context, id int64) (*db.SampelBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,permohonan_id,kode_sampel,pengawas_id,analis_id,jumlah_sampel,tanggal_pengambilan,status,created_at,updated_at FROM sampel_benih WHERE id=?`, id)
	var samp db.SampelBenih
	var analisID sql.NullInt64
	if err := row.Scan(&samp.ID, &samp.PermohonanID, &samp.KodeSampel, &samp.PengawasID, &analisID, &samp.JumlahSampel, &samp.TanggalPengambilan, &samp.Status, &samp.CreatedAt, &samp.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan sampel: %w", err)
	}
	if analisID.Valid {
		samp.AnalisID = &analisID.Int64
	}
	return &samp, nil
}

func (s *DBStore) KodifikasiSampel(ctx context.Context, id int64, kodeSampel string, analisID int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sampel_benih SET kode_sampel=?, analis_id=?, status='dikodekan', updated_at=CURRENT_TIMESTAMP WHERE id=?`, kodeSampel, analisID, id)
	if err != nil {
		return fmt.Errorf("kodifikasi sampel: %w", err)
	}
	return nil
}

func (s *DBStore) CreateUjiMutu(ctx context.Context, u db.UjiMutu) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO uji_mutu (sampel_id,parameter,hasil,status_lulus,catatan,diuji_oleh,tanggal_uji) VALUES (?,?,?,?,?,?,?)`,
		u.SampelID, u.Parameter, u.Hasil, u.StatusLulus, u.Catatan, u.DiujiOleh, u.TanggalUji)
	if err != nil {
		return 0, fmt.Errorf("insert uji mutu: %w", err)
	}
	return id, nil
}

func (s *DBStore) ListUjiMutu(ctx context.Context, sampelID int64) ([]db.UjiMutu, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,sampel_id,parameter,hasil,status_lulus,catatan,diuji_oleh,tanggal_uji,created_at FROM uji_mutu WHERE sampel_id=?`, sampelID)
	if err != nil {
		return nil, fmt.Errorf("query uji mutu: %w", err)
	}
	defer rows.Close()
	var out []db.UjiMutu
	for rows.Next() {
		var u db.UjiMutu
		if err := rows.Scan(&u.ID, &u.SampelID, &u.Parameter, &u.Hasil, &u.StatusLulus, &u.Catatan, &u.DiujiOleh, &u.TanggalUji, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan uji mutu: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *DBStore) CreateLaporanMutu(ctx context.Context, l db.LaporanMutu) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO laporan_mutu (sampel_id,kesimpulan,isi_laporan,disusun_oleh) VALUES (?,?,?,?)`,
		l.SampelID, l.Kesimpulan, l.IsiLaporan, l.DisusunOleh)
	if err != nil {
		return 0, fmt.Errorf("insert laporan mutu: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetLaporanMutu(ctx context.Context, sampelID int64) (*db.LaporanMutu, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,sampel_id,kesimpulan,isi_laporan,disusun_oleh,created_at FROM laporan_mutu WHERE sampel_id=?`, sampelID)
	var l db.LaporanMutu
	if err := row.Scan(&l.ID, &l.SampelID, &l.Kesimpulan, &l.IsiLaporan, &l.DisusunOleh, &l.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan laporan mutu: %w", err)
	}
	return &l, nil
}

func (s *DBStore) CreateSertifikat(ctx context.Context, s2 db.SertifikatBenih) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO sertifikat_benih (permohonan_id,nomor_sertifikat,penandatangan_id,tanggal_ttd,tte_status,tte_dokumen_url) VALUES (?,?,?,?,?,?)`,
		s2.PermohonanID, s2.NomorSertifikat, s2.PenandatanganID, s2.TanggalTTD, s2.TTEStatus, s2.TTEDokumenURL)
	if err != nil {
		return 0, fmt.Errorf("insert sertifikat: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetSertifikat(ctx context.Context, permohonanID int64) (*db.SertifikatBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,permohonan_id,nomor_sertifikat,penandatangan_id,tanggal_ttd,tte_status,tte_dokumen_url,created_at,updated_at FROM sertifikat_benih WHERE permohonan_id=?`, permohonanID)
	return scanSertifikat(row)
}

func (s *DBStore) GetSertifikatByID(ctx context.Context, id int64) (*db.SertifikatBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,permohonan_id,nomor_sertifikat,penandatangan_id,tanggal_ttd,tte_status,tte_dokumen_url,created_at,updated_at FROM sertifikat_benih WHERE id=?`, id)
	return scanSertifikat(row)
}

func (s *DBStore) UpdateSertifikatTTE(ctx context.Context, permohonanID int64, tteStatus, tteDokumenURL string, ttd *time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sertifikat_benih SET tte_status=?, tte_dokumen_url=?, tanggal_ttd=?, updated_at=CURRENT_TIMESTAMP WHERE permohonan_id=?`, tteStatus, tteDokumenURL, ttd, permohonanID)
	if err != nil {
		return fmt.Errorf("update sertifikat TTE: %w", err)
	}
	return nil
}

func (s *DBStore) CreateDraftLabel(ctx context.Context, l db.DraftLabel) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO draft_label (sertifikat_id,kode_barcode,status,integrasi_sim_status) VALUES (?,?,?,?)`,
		l.SertifikatID, l.KodeBarcode, l.Status, l.IntegrasiSIMStatus)
	if err != nil {
		return 0, fmt.Errorf("insert draft label: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetDraftLabel(ctx context.Context, sertifikatID int64) (*db.DraftLabel, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,sertifikat_id,kode_barcode,status,integrasi_sim_status,created_at,updated_at FROM draft_label WHERE sertifikat_id=?`, sertifikatID)
	return scanDraftLabel(row)
}

func (s *DBStore) GetDraftLabelByID(ctx context.Context, id int64) (*db.DraftLabel, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,sertifikat_id,kode_barcode,status,integrasi_sim_status,created_at,updated_at FROM draft_label WHERE id=?`, id)
	return scanDraftLabel(row)
}

func (s *DBStore) UpdateDraftLabel(ctx context.Context, id int64, status, integrasiSIM, kodeBarcode string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE draft_label SET status=?, integrasi_sim_status=?, kode_barcode=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, status, integrasiSIM, kodeBarcode, id)
	if err != nil {
		return fmt.Errorf("update draft label: %w", err)
	}
	return nil
}

func (s *DBStore) CreateLegalisasi(ctx context.Context, l db.LabelLegalisasi) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO label_legalisasi (draft_label_id,legalisasi_id,status,catatan) VALUES (?,?,?,?)`,
		l.DraftLabelID, l.LegalisasiID, l.Status, l.Catatan)
	if err != nil {
		return 0, fmt.Errorf("insert legalisasi: %w", err)
	}
	return id, nil
}

func (s *DBStore) UpdateLegalisasi(ctx context.Context, id int64, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE label_legalisasi SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, status, id)
	if err != nil {
		return fmt.Errorf("update legalisasi: %w", err)
	}
	return nil
}

func (s *DBStore) CreatePenyerahan(ctx context.Context, p db.PenyerahanDokumen) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO penyerahan_dokumen (permohonan_id,sertifikat_dikirim,label_dikirim,pembayaran_akhir_status,tanggal_penyerahan,diterima_oleh,catatan) VALUES (?,?,?,?,?,?,?)`,
		p.PermohonanID, p.SertifikatDikirim, p.LabelDikirim, p.PembayaranAkhirStatus, p.TanggalPenyerahan, p.DiterimaOleh, p.Catatan)
	if err != nil {
		return 0, fmt.Errorf("insert penyerahan: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetPenyerahan(ctx context.Context, permohonanID int64) (*db.PenyerahanDokumen, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,permohonan_id,sertifikat_dikirim,label_dikirim,pembayaran_akhir_status,tanggal_penyerahan,diterima_oleh,catatan,created_at,updated_at FROM penyerahan_dokumen WHERE permohonan_id=?`, permohonanID)
	var p db.PenyerahanDokumen
	var tgl sql.NullTime
	if err := row.Scan(&p.ID, &p.PermohonanID, &p.SertifikatDikirim, &p.LabelDikirim, &p.PembayaranAkhirStatus, &tgl, &p.DiterimaOleh, &p.Catatan, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan penyerahan: %w", err)
	}
	if tgl.Valid {
		p.TanggalPenyerahan = &tgl.Time
	}
	return &p, nil
}

func (s *DBStore) UpdatePenyerahan(ctx context.Context, permohonanID int64, sertifikatDikirim, labelDikirim bool, pembayaranAkhirStatus string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE penyerahan_dokumen SET sertifikat_dikirim=?, label_dikirim=?, pembayaran_akhir_status=?, updated_at=CURRENT_TIMESTAMP WHERE permohonan_id=?`, sertifikatDikirim, labelDikirim, pembayaranAkhirStatus, permohonanID)
	if err != nil {
		return fmt.Errorf("update penyerahan: %w", err)
	}
	return nil
}

func (s *DBStore) CreateAuditLog(ctx context.Context, a db.AuditLog) (int64, error) {
	var actorID *int64
	if a.ActorID != nil {
		actorID = a.ActorID
	}
	var entityID *int64
	if a.EntityID != nil {
		entityID = a.EntityID
	}
	id, err := db.InsertID(ctx, s.db, `INSERT INTO audit_log (actor_id,actor_role,action,entity,entity_id,detail) VALUES (?,?,?,?,?,?)`,
		actorID, a.ActorRole, a.Action, a.Entity, entityID, a.Detail)
	if err != nil {
		return 0, fmt.Errorf("insert audit log: %w", err)
	}
	return id, nil
}

func (s *DBStore) CreateDokumen(ctx context.Context, d db.Dokumen) (int64, error) {
	id, err := db.InsertID(ctx, s.db, `INSERT INTO dokumen (pemilik_id,tujuan,object_key,nama_asli,jenis,ukuran) VALUES (?,?,?,?,?,?)`,
		d.PemilikID, d.Tujuan, d.ObjectKey, d.NamaAsli, d.Jenis, d.Ukuran)
	if err != nil {
		return 0, fmt.Errorf("insert dokumen: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetDokumen(ctx context.Context, id int64) (*db.Dokumen, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,pemilik_id,COALESCE(tujuan,''),object_key,nama_asli,jenis,ukuran,created_at FROM dokumen WHERE id=?`, id)
	var dok db.Dokumen
	var createdAt time.Time
	if err := row.Scan(&dok.ID, &dok.PemilikID, &dok.Tujuan, &dok.ObjectKey, &dok.NamaAsli, &dok.Jenis, &dok.Ukuran, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ambil dokumen: %w", err)
	}
	dok.CreatedAt = createdAt
	return &dok, nil
}

func (s *DBStore) ListDokumen(ctx context.Context, userID int64, semua bool) ([]db.Dokumen, error) {
	query := `SELECT id,pemilik_id,COALESCE(tujuan,''),object_key,nama_asli,jenis,ukuran,created_at FROM dokumen`
	args := []any{}
	if !semua {
		query += ` WHERE pemilik_id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY id DESC LIMIT 200`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query daftar dokumen: %w", err)
	}
	defer rows.Close()
	var daftar []db.Dokumen
	for rows.Next() {
		var dok db.Dokumen
		var createdAt time.Time
		if err := rows.Scan(&dok.ID, &dok.PemilikID, &dok.Tujuan, &dok.ObjectKey, &dok.NamaAsli, &dok.Jenis, &dok.Ukuran, &createdAt); err != nil {
			return nil, fmt.Errorf("scan dokumen: %w", err)
		}
		dok.CreatedAt = createdAt
		daftar = append(daftar, dok)
	}
	return daftar, rows.Err()
}

func scanUser(row *sql.Row) (*db.User, error) {
	var u db.User
	var isActive bool
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.NPP, &u.UnitKerja, &u.Role, &isActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.IsActive = isActive
	return &u, nil
}

func scanPermohonan(row *sql.Row) (*db.PermohonanBenih, error) {
	var p db.PermohonanBenih
	var luas sql.NullFloat64
	var varietas, lokasi, bukuInduk, laporan sql.NullString
	if err := row.Scan(&p.ID, &p.NomorPermohonan, &p.PelangganID, &p.JenisBenih, &varietas, &luas, &lokasi, &p.Status, &p.PembayaranAwalStatus, &bukuInduk, &laporan, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan permohonan: %w", err)
	}
	if luas.Valid {
		p.LuasLahan = luas.Float64
	}
	if varietas.Valid {
		p.Varietas = varietas.String
	}
	if lokasi.Valid {
		p.LokasiKebun = lokasi.String
	}
	if bukuInduk.Valid {
		p.BukuIndukNomor = bukuInduk.String
	}
	if laporan.Valid {
		p.LaporanVerifikasi = laporan.String
	}
	return &p, nil
}

func scanBukuInduk(row *sql.Row) (*db.BukuInduk, error) {
	var b db.BukuInduk
	if err := row.Scan(&b.ID, &b.PermohonanID, &b.NomorBukuInduk, &b.Catatan, &b.DibuatOleh, &b.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan buku induk: %w", err)
	}
	return &b, nil
}

func scanSertifikat(row *sql.Row) (*db.SertifikatBenih, error) {
	var srt db.SertifikatBenih
	var ttd sql.NullTime
	if err := row.Scan(&srt.ID, &srt.PermohonanID, &srt.NomorSertifikat, &srt.PenandatanganID, &ttd, &srt.TTEStatus, &srt.TTEDokumenURL, &srt.CreatedAt, &srt.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan sertifikat: %w", err)
	}
	if ttd.Valid {
		srt.TanggalTTD = &ttd.Time
	}
	return &srt, nil
}

func scanDraftLabel(row *sql.Row) (*db.DraftLabel, error) {
	var l db.DraftLabel
	if err := row.Scan(&l.ID, &l.SertifikatID, &l.KodeBarcode, &l.Status, &l.IntegrasiSIMStatus, &l.CreatedAt, &l.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan draft label: %w", err)
	}
	return &l, nil
}
