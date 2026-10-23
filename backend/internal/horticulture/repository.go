package horticulture

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("permohonan tidak ditemukan")

type Crop struct {
	Name       string `json:"nama"`
	Variety    string `json:"varietas"`
	PlantCount int    `json:"jumlah_tanaman"`
	SeedSource string `json:"sumber_benih"`
}

type Document struct {
	ID   int64  `json:"id"`
	Type string `json:"jenis"`
}

type Application struct {
	ID             int64      `json:"id"`
	Number         string     `json:"nomor_permohonan"`
	ApplicantID    int64      `json:"-"`
	Organization   string     `json:"nama_organisasi"`
	Contact        string     `json:"nama_kontak"`
	Phone          string     `json:"telepon"`
	Address        string     `json:"alamat"`
	Village        string     `json:"kalurahan"`
	District       string     `json:"kapanewon"`
	Regency        string     `json:"kabupaten_kota"`
	CultivatedArea float64    `json:"luas_lahan_ha"`
	Crops          []Crop     `json:"tanaman"`
	Status         string     `json:"status"`
	Documents      []Document `json:"dokumen"`
	CreatedAt      time.Time  `json:"dibuat_pada"`
}

type Repository interface {
	Create(context.Context, Application, []Document) (Application, error)
	List(context.Context, int64, bool) ([]Application, error)
	Get(context.Context, int64) (Application, error)
}

type MySQLRepository struct{ db *sql.DB }

func NewMySQLRepository(db *sql.DB) *MySQLRepository { return &MySQLRepository{db: db} }

func (r *MySQLRepository) Create(ctx context.Context, app Application, documents []Document) (Application, error) {
	if app.Crops == nil {
		app.Crops = []Crop{}
	}
	crops, err := json.Marshal(app.Crops)
	if err != nil {
		return Application{}, fmt.Errorf("encode detail tanaman: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Application{}, fmt.Errorf("mulai transaksi permohonan: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var suffix [4]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return Application{}, fmt.Errorf("buat nomor permohonan: %w", err)
	}
	app.Number = fmt.Sprintf("RH-%d-%08X", time.Now().UTC().Year(), suffix)
	result, err := tx.ExecContext(ctx, `INSERT INTO horticulture_applications (application_number,applicant_id,organization_name,contact_name,phone,address,village,district,regency,cultivated_area,crop_details,status) VALUES (?,?,?,?,?,?,?,?,?,?,?,'diajukan')`,
		app.Number, app.ApplicantID, app.Organization, app.Contact, app.Phone, app.Address, app.Village, app.District, app.Regency, app.CultivatedArea, string(crops))
	if err != nil {
		return Application{}, fmt.Errorf("simpan permohonan registrasi lahan: %w", err)
	}
	app.ID, err = result.LastInsertId()
	if err != nil {
		return Application{}, fmt.Errorf("baca ID permohonan: %w", err)
	}
	for _, doc := range documents {
		if _, err := tx.ExecContext(ctx, `INSERT INTO horticulture_application_documents (application_id,document_id,document_type) SELECT ?,id,? FROM dokumen WHERE id=? AND pemilik_id=?`, app.ID, doc.Type, doc.ID, app.ApplicantID); err != nil {
			return Application{}, fmt.Errorf("tautkan dokumen permohonan: %w", err)
		}
		var linked int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM horticulture_application_documents WHERE application_id=? AND document_id=?`, app.ID, doc.ID).Scan(&linked); err != nil || linked != 1 {
			if err != nil {
				return Application{}, fmt.Errorf("verifikasi tautan dokumen: %w", err)
			}
			return Application{}, fmt.Errorf("dokumen %d tidak ditemukan atau bukan milik pemohon", doc.ID)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_log (actor_id,actor_role,action,entity,entity_id,detail) SELECT id,role,'horticulture_application:create','horticulture_application',?,? FROM users WHERE id=?`, app.ID, "pengajuan registrasi lahan", app.ApplicantID); err != nil {
		return Application{}, fmt.Errorf("catat audit permohonan: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Application{}, fmt.Errorf("commit permohonan: %w", err)
	}
	app.Documents = append(make([]Document, 0, len(documents)), documents...)
	app.Status = "diajukan"
	app.CreatedAt = time.Now().UTC()
	return app, nil
}

func (r *MySQLRepository) List(ctx context.Context, applicantID int64, all bool) ([]Application, error) {
	query := `SELECT id,application_number,applicant_id,organization_name,contact_name,phone,address,village,district,regency,cultivated_area,crop_details,status,created_at FROM horticulture_applications`
	args := []any{}
	if !all {
		query += ` WHERE applicant_id=?`
		args = append(args, applicantID)
	}
	query += ` ORDER BY id DESC LIMIT 100`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("daftar permohonan: %w", err)
	}
	defer rows.Close()
	apps := make([]Application, 0)
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, err
		}
		app.Documents, err = r.listDocuments(ctx, app.ID)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (r *MySQLRepository) Get(ctx context.Context, id int64) (Application, error) {
	app, err := scanApplication(r.db.QueryRowContext(ctx, `SELECT id,application_number,applicant_id,organization_name,contact_name,phone,address,village,district,regency,cultivated_area,crop_details,status,created_at FROM horticulture_applications WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	app.Documents, err = r.listDocuments(ctx, id)
	if err != nil {
		return Application{}, err
	}
	return app, nil
}

func (r *MySQLRepository) listDocuments(ctx context.Context, id int64) ([]Document, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT document_id,document_type FROM horticulture_application_documents WHERE application_id=? ORDER BY document_type,document_id`, id)
	if err != nil {
		return nil, fmt.Errorf("daftar dokumen permohonan: %w", err)
	}
	defer rows.Close()
	documents := make([]Document, 0)
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.ID, &doc.Type); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, rows.Err()
}

type scanner interface{ Scan(...any) error }

func scanApplication(row scanner) (Application, error) {
	var app Application
	var crops []byte
	err := row.Scan(&app.ID, &app.Number, &app.ApplicantID, &app.Organization, &app.Contact, &app.Phone, &app.Address, &app.Village, &app.District, &app.Regency, &app.CultivatedArea, &crops, &app.Status, &app.CreatedAt)
	if err != nil {
		return Application{}, err
	}
	if err := json.Unmarshal(crops, &app.Crops); err != nil {
		return Application{}, fmt.Errorf("decode detail tanaman: %w", err)
	}
	if app.Crops == nil {
		app.Crops = []Crop{}
	}
	app.Organization = strings.TrimSpace(app.Organization)
	return app, nil
}
