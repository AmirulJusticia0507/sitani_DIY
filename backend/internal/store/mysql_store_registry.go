package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

const registrasiColumns = `id,nomor_registrasi,permohonan_id,sertifikat_id,draft_label_id,nomor_sertifikat,kode_label,komoditas,varietas,produsen,sumber_benih,nomor_lot,kelas_benih,jumlah,satuan,tanggal_produksi,tanggal_kedaluwarsa,status_edar,created_by,created_at,updated_at`

const penjualanColumns = `id,nomor_dokumen,registrasi_id,nomor_lot,pembeli,jumlah,harga_satuan,total_harga,tanggal_transaksi,status,alasan_pembatalan,dibatalkan_oleh,dibatalkan_pada,created_by,created_at,updated_at`

// CreateRegistrasiBenih menyimpan lot baru, saldo stok awal, dan mutasi masuk
// dalam satu transaksi agar persediaan tidak pernah terpisah dari registrasinya.
func (s *DBStore) CreateRegistrasiBenih(ctx context.Context, r db.RegistrasiBenih) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("mulai transaksi registrasi: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO registrasi_benih
		(nomor_registrasi,permohonan_id,sertifikat_id,draft_label_id,nomor_sertifikat,kode_label,komoditas,varietas,produsen,sumber_benih,nomor_lot,kelas_benih,jumlah,satuan,tanggal_produksi,tanggal_kedaluwarsa,status_edar,created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.NomorRegistrasi, r.PermohonanID, r.SertifikatID, r.DraftLabelID, r.NomorSertifikat, r.KodeLabel,
		r.Komoditas, r.Varietas, r.Produsen, r.SumberBenih, r.NomorLot, r.KelasBenih, r.Jumlah, r.Satuan,
		r.TanggalProduksi, r.TanggalKedaluwarsa, r.StatusEdar, r.CreatedBy)
	if err != nil {
		return 0, fmt.Errorf("insert registrasi benih: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("id registrasi benih: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO stok_benih (registrasi_id,jumlah_tersedia) VALUES (?,?)`, id, r.Jumlah); err != nil {
		return 0, fmt.Errorf("insert stok awal: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO mutasi_stok_benih
		(registrasi_id,jenis,jumlah,saldo_setelah,referensi_tipe,referensi_id,catatan,created_by)
		VALUES (?,?,?,?,?,?,?,?)`,
		id, "masuk", r.Jumlah, r.Jumlah, "registrasi", id, "stok awal registrasi", r.CreatedBy); err != nil {
		return 0, fmt.Errorf("insert mutasi stok awal: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit registrasi benih: %w", err)
	}
	return id, nil
}

func (s *DBStore) GetRegistrasiBenih(ctx context.Context, id int64) (*db.RegistrasiBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+registrasiColumns+` FROM registrasi_benih WHERE id=?`, id)
	return scanRegistrasi(row)
}

func (s *DBStore) GetRegistrasiBenihByNomorLot(ctx context.Context, nomorLot string) (*db.RegistrasiBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+registrasiColumns+` FROM registrasi_benih WHERE nomor_lot=?`, nomorLot)
	return scanRegistrasi(row)
}

func (s *DBStore) ListRegistrasiBenih(ctx context.Context, filter db.RegistrasiFilter) ([]db.RegistrasiBenih, error) {
	query := `SELECT ` + registrasiColumns + ` FROM registrasi_benih WHERE 1=1`
	var args []any
	if filter.StatusEdar != "" {
		query += ` AND status_edar=?`
		args = append(args, filter.StatusEdar)
	}
	if filter.Komoditas != "" {
		query += ` AND komoditas LIKE ?`
		args = append(args, "%"+filter.Komoditas+"%")
	}
	if filter.KedaluwarsaSebelum != nil {
		query += ` AND tanggal_kedaluwarsa <= ?`
		args = append(args, filter.KedaluwarsaSebelum.Format("2006-01-02"))
	}
	query += ` ORDER BY created_at DESC, id DESC LIMIT ?`
	args = append(args, batasBaris(filter.Limit))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query registrasi benih: %w", err)
	}
	defer rows.Close()
	var out []db.RegistrasiBenih
	for rows.Next() {
		r, err := scanRegistrasi(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

func (s *DBStore) UpdateRegistrasiStatus(ctx context.Context, id int64, status string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE registrasi_benih SET status_edar=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, status, id)
	if err != nil {
		return fmt.Errorf("update status registrasi: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *DBStore) GetStokBenih(ctx context.Context, registrasiID int64) (*db.StokBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT s.id,s.registrasi_id,r.nomor_lot,r.komoditas,r.varietas,s.jumlah_tersedia,r.satuan,r.tanggal_kedaluwarsa,r.status_edar,s.updated_at
		FROM stok_benih s JOIN registrasi_benih r ON r.id=s.registrasi_id WHERE s.registrasi_id=?`, registrasiID)
	return scanStok(row)
}

func (s *DBStore) ListStokBenih(ctx context.Context) ([]db.StokBenih, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT s.id,s.registrasi_id,r.nomor_lot,r.komoditas,r.varietas,s.jumlah_tersedia,r.satuan,r.tanggal_kedaluwarsa,r.status_edar,s.updated_at
		FROM stok_benih s JOIN registrasi_benih r ON r.id=s.registrasi_id ORDER BY r.tanggal_kedaluwarsa, r.nomor_lot`)
	if err != nil {
		return nil, fmt.Errorf("query stok benih: %w", err)
	}
	defer rows.Close()
	var out []db.StokBenih
	for rows.Next() {
		stok, err := scanStok(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *stok)
	}
	return out, rows.Err()
}

func (s *DBStore) ListMutasiStok(ctx context.Context, registrasiID int64) ([]db.MutasiStokBenih, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,registrasi_id,jenis,jumlah,saldo_setelah,referensi_tipe,referensi_id,catatan,created_by,created_at
		FROM mutasi_stok_benih WHERE registrasi_id=? ORDER BY id`, registrasiID)
	if err != nil {
		return nil, fmt.Errorf("query mutasi stok: %w", err)
	}
	defer rows.Close()
	var out []db.MutasiStokBenih
	for rows.Next() {
		var m db.MutasiStokBenih
		var refID sql.NullInt64
		if err := rows.Scan(&m.ID, &m.RegistrasiID, &m.Jenis, &m.Jumlah, &m.SaldoSetelah, &m.ReferensiTipe, &refID, &m.Catatan, &m.CreatedBy, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan mutasi stok: %w", err)
		}
		if refID.Valid {
			m.ReferensiID = &refID.Int64
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CreatePenjualanBenih mengunci saldo stok, memvalidasi kecukupan, lalu
// menyimpan penjualan dan mutasi keluar dalam satu transaksi.
func (s *DBStore) CreatePenjualanBenih(ctx context.Context, p db.PenjualanBenih) (*db.PenjualanBenih, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("mulai transaksi penjualan: %w", err)
	}
	defer tx.Rollback()

	var saldo float64
	if err := tx.QueryRowContext(ctx, `SELECT jumlah_tersedia FROM stok_benih WHERE registrasi_id=? FOR UPDATE`, p.RegistrasiID).Scan(&saldo); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ambil stok: %w", err)
	}
	if saldo < p.Jumlah {
		return nil, ErrStokTidakCukup
	}
	saldoBaru := saldo - p.Jumlah
	if _, err := tx.ExecContext(ctx, `UPDATE stok_benih SET jumlah_tersedia=?, updated_at=CURRENT_TIMESTAMP WHERE registrasi_id=?`, saldoBaru, p.RegistrasiID); err != nil {
		return nil, fmt.Errorf("kurangi stok: %w", err)
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO penjualan_benih
		(nomor_dokumen,registrasi_id,nomor_lot,pembeli,jumlah,harga_satuan,total_harga,tanggal_transaksi,status,created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		p.NomorDokumen, p.RegistrasiID, p.NomorLot, p.Pembeli, p.Jumlah, p.HargaSatuan, p.TotalHarga, p.TanggalTransaksi, "selesai", p.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("insert penjualan: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("id penjualan: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO mutasi_stok_benih
		(registrasi_id,jenis,jumlah,saldo_setelah,referensi_tipe,referensi_id,catatan,created_by)
		VALUES (?,?,?,?,?,?,?,?)`,
		p.RegistrasiID, "keluar", p.Jumlah, saldoBaru, "penjualan", id, "penjualan "+p.NomorDokumen, p.CreatedBy); err != nil {
		return nil, fmt.Errorf("insert mutasi penjualan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit penjualan: %w", err)
	}
	p.ID = id
	p.Status = "selesai"
	return &p, nil
}

func (s *DBStore) GetPenjualanBenih(ctx context.Context, id int64) (*db.PenjualanBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+penjualanColumns+` FROM penjualan_benih WHERE id=?`, id)
	return scanPenjualan(row)
}

func (s *DBStore) ListPenjualanBenih(ctx context.Context, filter db.PenjualanFilter) ([]db.PenjualanBenih, error) {
	query := `SELECT ` + penjualanColumns + ` FROM penjualan_benih WHERE 1=1`
	var args []any
	if filter.RegistrasiID > 0 {
		query += ` AND registrasi_id=?`
		args = append(args, filter.RegistrasiID)
	}
	if filter.Status != "" {
		query += ` AND status=?`
		args = append(args, filter.Status)
	}
	if filter.Dari != nil {
		query += ` AND tanggal_transaksi >= ?`
		args = append(args, filter.Dari.Format("2006-01-02"))
	}
	if filter.Sampai != nil {
		query += ` AND tanggal_transaksi <= ?`
		args = append(args, filter.Sampai.Format("2006-01-02"))
	}
	query += ` ORDER BY tanggal_transaksi DESC, id DESC LIMIT ?`
	args = append(args, batasBaris(filter.Limit))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query penjualan: %w", err)
	}
	defer rows.Close()
	var out []db.PenjualanBenih
	for rows.Next() {
		p, err := scanPenjualan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

// BatalkanPenjualanBenih mempertahankan transaksi asli, menandainya batal,
// dan menambahkan stok kembali melalui mutasi koreksi.
func (s *DBStore) BatalkanPenjualanBenih(ctx context.Context, id int64, actorID int64, alasan string, waktu time.Time) (*db.PenjualanBenih, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("mulai transaksi pembatalan: %w", err)
	}
	defer tx.Rollback()

	var p db.PenjualanBenih
	var alasanLama sql.NullString
	var dibatalkanOleh sql.NullInt64
	var dibatalkanPada sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT `+penjualanColumns+` FROM penjualan_benih WHERE id=? FOR UPDATE`, id).
		Scan(&p.ID, &p.NomorDokumen, &p.RegistrasiID, &p.NomorLot, &p.Pembeli, &p.Jumlah, &p.HargaSatuan, &p.TotalHarga,
			&p.TanggalTransaksi, &p.Status, &alasanLama, &dibatalkanOleh, &dibatalkanPada, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("ambil penjualan: %w", err)
	}
	if alasanLama.Valid {
		p.AlasanPembatalan = alasanLama.String
	}
	if dibatalkanOleh.Valid {
		p.DibatalkanOleh = &dibatalkanOleh.Int64
	}
	if dibatalkanPada.Valid {
		p.DibatalkanPada = &dibatalkanPada.Time
	}
	if p.Status == "dibatalkan" {
		return nil, ErrSudahDibatalkan
	}

	var saldo float64
	if err := tx.QueryRowContext(ctx, `SELECT jumlah_tersedia FROM stok_benih WHERE registrasi_id=? FOR UPDATE`, p.RegistrasiID).Scan(&saldo); err != nil {
		return nil, fmt.Errorf("ambil stok pembatalan: %w", err)
	}
	saldoBaru := saldo + p.Jumlah
	if _, err := tx.ExecContext(ctx, `UPDATE stok_benih SET jumlah_tersedia=?, updated_at=CURRENT_TIMESTAMP WHERE registrasi_id=?`, saldoBaru, p.RegistrasiID); err != nil {
		return nil, fmt.Errorf("kembalikan stok: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE penjualan_benih SET status='dibatalkan', alasan_pembatalan=?, dibatalkan_oleh=?, dibatalkan_pada=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		alasan, actorID, waktu, id); err != nil {
		return nil, fmt.Errorf("update status penjualan: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO mutasi_stok_benih
		(registrasi_id,jenis,jumlah,saldo_setelah,referensi_tipe,referensi_id,catatan,created_by)
		VALUES (?,?,?,?,?,?,?,?)`,
		p.RegistrasiID, "koreksi", p.Jumlah, saldoBaru, "pembatalan_penjualan", id, alasan, actorID); err != nil {
		return nil, fmt.Errorf("insert mutasi pembatalan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pembatalan: %w", err)
	}
	p.Status = "dibatalkan"
	p.AlasanPembatalan = alasan
	p.DibatalkanOleh = &actorID
	p.DibatalkanPada = &waktu
	return &p, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRegistrasi(row rowScanner) (*db.RegistrasiBenih, error) {
	var r db.RegistrasiBenih
	var sertifikatID, labelID sql.NullInt64
	var produksi, kedaluwarsa sql.NullTime
	if err := row.Scan(&r.ID, &r.NomorRegistrasi, &r.PermohonanID, &sertifikatID, &labelID, &r.NomorSertifikat, &r.KodeLabel,
		&r.Komoditas, &r.Varietas, &r.Produsen, &r.SumberBenih, &r.NomorLot, &r.KelasBenih, &r.Jumlah, &r.Satuan,
		&produksi, &kedaluwarsa, &r.StatusEdar, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan registrasi benih: %w", err)
	}
	if sertifikatID.Valid {
		r.SertifikatID = &sertifikatID.Int64
	}
	if labelID.Valid {
		r.DraftLabelID = &labelID.Int64
	}
	if produksi.Valid {
		r.TanggalProduksi = produksi.Time.Format("2006-01-02")
	}
	if kedaluwarsa.Valid {
		r.TanggalKedaluwarsa = kedaluwarsa.Time.Format("2006-01-02")
	}
	return &r, nil
}

func scanStok(row rowScanner) (*db.StokBenih, error) {
	var stok db.StokBenih
	var kedaluwarsa sql.NullTime
	if err := row.Scan(&stok.ID, &stok.RegistrasiID, &stok.NomorLot, &stok.Komoditas, &stok.Varietas, &stok.JumlahTersedia,
		&stok.Satuan, &kedaluwarsa, &stok.StatusEdar, &stok.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan stok benih: %w", err)
	}
	if kedaluwarsa.Valid {
		stok.TanggalKedaluwarsa = kedaluwarsa.Time.Format("2006-01-02")
	}
	return &stok, nil
}

func scanPenjualan(row rowScanner) (*db.PenjualanBenih, error) {
	var p db.PenjualanBenih
	var tanggal sql.NullTime
	var dibatalkanPada sql.NullTime
	var dibatalkanOleh sql.NullInt64
	var alasan sql.NullString
	if err := row.Scan(&p.ID, &p.NomorDokumen, &p.RegistrasiID, &p.NomorLot, &p.Pembeli, &p.Jumlah, &p.HargaSatuan,
		&p.TotalHarga, &tanggal, &p.Status, &alasan, &dibatalkanOleh, &dibatalkanPada, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan penjualan: %w", err)
	}
	if tanggal.Valid {
		p.TanggalTransaksi = tanggal.Time.Format("2006-01-02")
	}
	if alasan.Valid {
		p.AlasanPembatalan = alasan.String
	}
	if dibatalkanOleh.Valid {
		p.DibatalkanOleh = &dibatalkanOleh.Int64
	}
	if dibatalkanPada.Valid {
		p.DibatalkanPada = &dibatalkanPada.Time
	}
	return &p, nil
}

func batasBaris(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}
