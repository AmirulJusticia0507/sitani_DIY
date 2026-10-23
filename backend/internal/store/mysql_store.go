package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

func (s *DBStore) CreateUser(ctx context.Context, u db.User) (int64, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO users (username,email,npp,unit_kerja,role,is_active) VALUES (?,?,?,?,?,?)`,
		u.Username, u.Email, u.NPP, u.UnitKerja, u.Role, boolToInt(u.IsActive))
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return res.LastInsertId()
}

func (s *DBStore) GetUserByID(ctx context.Context, id int64) (*db.User, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,username,email,npp,unit_kerja,role,is_active,created_at,updated_at FROM users WHERE id=?`, id)
	return scanUser(row)
}

func (s *DBStore) GetUserByUsername(ctx context.Context, username string) (*db.User, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,username,email,npp,unit_kerja,role,is_active,created_at,updated_at FROM users WHERE username=?`, username)
	return scanUser(row)
}

func (s *DBStore) CreatePermohonan(ctx context.Context, p db.PermohonanBenih) (int64, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO permohonan_benih (nomor_permohonan,pelanggan_id,jenis_benih,varietas,luas_lahan,lokasi_kebun,status,pembayaran_awal_status) VALUES (?,?,?,?,?,?,?,?)`,
		p.NomorPermohonan, p.PelangganID, p.JenisBenih, p.Varietas, p.LuasLahan, p.LokasiKebun, p.Status, p.PembayaranAwalStatus)
	if err != nil {
		return 0, fmt.Errorf("insert permohonan: %w", err)
	}
	return res.LastInsertId()
}

func (s *DBStore) GetPermohonan(ctx context.Context, id int64) (*db.PermohonanBenih, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,nomor_permohonan,pelanggan_id,jenis_benih,varietas,luas_lahan,lokasi_kebun,status,pembayaran_awal_status,buku_induk_nomor,laporan_verifikasi,created_at,updated_at FROM permohonan_benih WHERE id=?`, id)
	return scanPermohonan(row)
}

func (s *DBStore) UpdatePermohonanStatus(ctx context.Context, id int64, status, pembayaranAwalStatus string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE permohonan_benih SET status=?, pembayaran_awal_status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, status, pembayaranAwalStatus, id)
	if err != nil {
		return fmt.Errorf("update permohonan: %w", err)
	}
	return nil
}

func (s *DBStore) CreatePembayaran(ctx context.Context, pb db.Pembayaran) (int64, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO pembayaran (permohonan_id,jenis,jumlah,status,bukti_transfer) VALUES (?,?,?,?,?)`,
		pb.PermohonanID, pb.Jenis, pb.Jumlah, pb.Status, pb.BuktiTransfer)
	if err != nil {
		return 0, fmt.Errorf("insert pembayaran: %w", err)
	}
	return res.LastInsertId()
}

func (s *DBStore) CreateInspeksiFase(ctx context.Context, i db.InspeksiFase) (int64, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO inspeksi_fase (permohonan_id,fase,pengawas_id,hasil,catatan,tanggal_inspeksi) VALUES (?,?,?,?,?,?)`,
		i.PermohonanID, i.Fase, i.PengawasID, i.Hasil, i.Catatan, i.TanggalInspeksi)
	if err != nil {
		return 0, fmt.Errorf("insert inspeksi: %w", err)
	}
	return res.LastInsertId()
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
	res, err := s.db.ExecContext(ctx, `INSERT INTO buku_induk (permohonan_id,nomor_buku_induk,catatan,dibuat_oleh) VALUES (?,?,?,?)`,
		b.PermohonanID, b.NomorBukuInduk, b.Catatan, b.DibuatOleh)
	if err != nil {
		return 0, fmt.Errorf("insert buku induk: %w", err)
	}
	return res.LastInsertId()
}

func (s *DBStore) GetBukuInduk(ctx context.Context, permohonanID int64) (*db.BukuInduk, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,permohonan_id,nomor_buku_induk,catatan,dibuat_oleh,created_at FROM buku_induk WHERE permohonan_id=?`, permohonanID)
	return scanBukuInduk(row)
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
	res, err := s.db.ExecContext(ctx, `INSERT INTO audit_log (actor_id,actor_role,action,entity,entity_id,detail) VALUES (?,?,?,?,?,?)`,
		actorID, a.ActorRole, a.Action, a.Entity, entityID, a.Detail)
	if err != nil {
		return 0, fmt.Errorf("insert audit log: %w", err)
	}
	return res.LastInsertId()
}

func scanUser(row *sql.Row) (*db.User, error) {
	var u db.User
	var isActive int
	var createdAt, updatedAt string
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.NPP, &u.UnitKerja, &u.Role, &isActive, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.IsActive = isActive == 1
	return &u, nil
}

func scanPermohonan(row *sql.Row) (*db.PermohonanBenih, error) {
	var p db.PermohonanBenih
	var luas sql.NullFloat64
	if err := row.Scan(&p.ID, &p.NomorPermohonan, &p.PelangganID, &p.JenisBenih, &p.Varietas, &luas, &p.LokasiKebun, &p.Status, &p.PembayaranAwalStatus, &p.BukuIndukNomor, &p.LaporanVerifikasi, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan permohonan: %w", err)
	}
	if luas.Valid {
		p.LuasLahan = luas.Float64
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
