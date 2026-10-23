package dataset

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Import struct {
	ID           int64     `json:"id"`
	NamaDataset  string    `json:"nama_dataset"`
	NamaFile     string    `json:"nama_file"`
	JumlahBaris  int       `json:"jumlah_baris"`
	DiunggahOleh int64     `json:"diunggah_oleh"`
	DibuatPada   time.Time `json:"dibuat_pada"`
}

type Repository interface {
	Simpan(context.Context, int64, string, string, Result) (Import, error)
	Daftar(context.Context, int64, bool) ([]Import, error)
}

type MySQLRepository struct{ db *sql.DB }

func NewMySQLRepository(db *sql.DB) *MySQLRepository { return &MySQLRepository{db: db} }

func (r *MySQLRepository) Simpan(ctx context.Context, uploaderID int64, namaDataset, namaFile string, parsed Result) (Import, error) {
	headers, err := json.Marshal(parsed.Headers)
	if err != nil {
		return Import{}, fmt.Errorf("encode header dataset: %w", err)
	}
	rows, err := json.Marshal(parsed.Rows)
	if err != nil {
		return Import{}, fmt.Errorf("encode baris dataset: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Import{}, fmt.Errorf("mulai transaksi import dataset: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `INSERT INTO dataset_imports (nama_dataset,nama_file,headers_json,jumlah_baris,baris_json,diunggah_oleh) VALUES (?,?,?,?,?,?)`,
		namaDataset, namaFile, string(headers), len(parsed.Rows), string(rows), uploaderID)
	if err != nil {
		return Import{}, fmt.Errorf("simpan batch import dataset: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Import{}, fmt.Errorf("baca ID batch import: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Import{}, fmt.Errorf("commit batch import: %w", err)
	}
	return Import{ID: id, NamaDataset: namaDataset, NamaFile: namaFile, JumlahBaris: len(parsed.Rows), DiunggahOleh: uploaderID, DibuatPada: time.Now()}, nil
}

func (r *MySQLRepository) Daftar(ctx context.Context, userID int64, semua bool) ([]Import, error) {
	query := `SELECT id,nama_dataset,nama_file,jumlah_baris,diunggah_oleh,dibuat_pada FROM dataset_imports`
	args := []any{}
	if !semua {
		query += ` WHERE diunggah_oleh=?`
		args = append(args, userID)
	}
	query += ` ORDER BY id DESC LIMIT 100`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("daftar import dataset: %w", err)
	}
	defer rows.Close()
	result := make([]Import, 0)
	for rows.Next() {
		var item Import
		if err := rows.Scan(&item.ID, &item.NamaDataset, &item.NamaFile, &item.JumlahBaris, &item.DiunggahOleh, &item.DibuatPada); err != nil {
			return nil, fmt.Errorf("baca batch dataset: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi batch dataset: %w", err)
	}
	return result, nil
}
