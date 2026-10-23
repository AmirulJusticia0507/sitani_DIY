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

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
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
	ID             int64                `json:"id"`
	Number         string               `json:"nomor_permohonan"`
	ApplicantID    int64                `json:"-"`
	Organization   string               `json:"nama_organisasi"`
	Contact        string               `json:"nama_kontak"`
	Phone          string               `json:"telepon"`
	Address        string               `json:"alamat"`
	Village        string               `json:"kalurahan"`
	District       string               `json:"kapanewon"`
	Regency        string               `json:"kabupaten_kota"`
	CultivatedArea float64              `json:"luas_lahan_ha"`
	Crops          []Crop               `json:"tanaman"`
	Status         string               `json:"status"`
	Documents      []Document           `json:"dokumen"`
	Schedules      []InspectionSchedule `json:"jadwal_inspeksi,omitempty"`
	Inspections    []Inspection         `json:"inspeksi,omitempty"`
	Evaluations    []Evaluation         `json:"penilaian,omitempty"`
	CreatedAt      time.Time            `json:"dibuat_pada"`
}

type Repository interface {
	Create(context.Context, Application, []Document) (Application, error)
	List(context.Context, int64, bool) ([]Application, error)
	Get(context.Context, int64) (Application, error)
	Verify(context.Context, int64, int64, string, string) (Application, error)
	ScheduleInspection(context.Context, InspectionSchedule, int64) (InspectionSchedule, error)
	CompleteInspection(context.Context, Inspection) (Inspection, error)
	SaveEvaluation(context.Context, Evaluation) (Evaluation, error)
}

type InspectionSchedule struct {
	ID            int64     `json:"id"`
	ApplicationID int64     `json:"permohonan_id"`
	InspectorID   int64     `json:"inspektor_id"`
	ScheduledAt   time.Time `json:"waktu"`
	Location      string    `json:"lokasi"`
	Notes         string    `json:"catatan"`
	Status        string    `json:"status"`
}

type Inspection struct {
	ID              int64           `json:"id"`
	ScheduleID      int64           `json:"jadwal_id"`
	Checklist       map[string]bool `json:"checklist_gap"`
	Findings        string          `json:"temuan"`
	Recommendations string          `json:"rekomendasi"`
	InspectorID     int64           `json:"inspektor_id"`
	CreatedAt       time.Time       `json:"dibuat_pada"`
}

type Evaluation struct {
	ID             int64     `json:"id"`
	ApplicationID  int64     `json:"permohonan_id"`
	AssessmentType string    `json:"jenis_penilaian"`
	PlantName      string    `json:"nama_tanaman"`
	Variety        string    `json:"varietas"`
	VigorScore     int       `json:"skor_vigoritas"`
	HealthScore    int       `json:"skor_kesehatan"`
	VarietyScore   int       `json:"skor_kebenaran_varietas"`
	Recommendation string    `json:"rekomendasi"`
	Notes          string    `json:"catatan"`
	EvaluatorID    int64     `json:"penilai_id"`
	CreatedAt      time.Time `json:"dibuat_pada"`
}

type PostgresRepository struct{ db *db.DB }

func NewPostgresRepository(database *db.DB) *PostgresRepository {
	return &PostgresRepository{db: database}
}

func (r *PostgresRepository) Verify(ctx context.Context, id, actorID int64, decision string, note string) (Application, error) {
	if decision != "diverifikasi" && decision != "perlu_perbaikan" && decision != "ditolak" {
		return Application{}, errors.New("keputusan verifikasi tidak valid")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Application{}, fmt.Errorf("mulai transaksi verifikasi: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM horticulture_applications WHERE id=? FOR UPDATE`, id).Scan(&current); errors.Is(err, sql.ErrNoRows) {
		return Application{}, ErrNotFound
	} else if err != nil {
		return Application{}, err
	}
	if current != "diajukan" && current != "perlu_perbaikan" {
		return Application{}, fmt.Errorf("permohonan berstatus %s tidak dapat diverifikasi", current)
	}
	if (decision == "perlu_perbaikan" || decision == "ditolak") && strings.TrimSpace(note) == "" {
		return Application{}, errors.New("alasan perbaikan atau penolakan wajib diisi")
	}
	if decision == "diverifikasi" {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(DISTINCT document_type) FROM horticulture_application_documents WHERE application_id=? AND document_type IN ('proposal','organization','members','farm_form','farm_map')`, id).Scan(&count); err != nil {
			return Application{}, err
		}
		if count != 5 {
			return Application{}, errors.New("dokumen permohonan belum lengkap")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE horticulture_applications SET status=? WHERE id=?`, decision, id); err != nil {
		return Application{}, fmt.Errorf("simpan keputusan verifikasi: %w", err)
	}
	if err := insertAudit(ctx, tx, actorID, "horticulture:verifikasi:"+decision, id, "status="+decision); err != nil {
		return Application{}, err
	}
	if err := tx.Commit(); err != nil {
		return Application{}, err
	}
	return r.Get(ctx, id)
}

func (r *PostgresRepository) ScheduleInspection(ctx context.Context, schedule InspectionSchedule, actorID int64) (InspectionSchedule, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return InspectionSchedule{}, fmt.Errorf("mulai transaksi jadwal inspeksi: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM horticulture_applications WHERE id=? FOR UPDATE`, schedule.ApplicationID).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return InspectionSchedule{}, ErrNotFound
	} else if err != nil {
		return InspectionSchedule{}, err
	}
	if status != "diverifikasi" && status != "dijadwalkan" {
		return InspectionSchedule{}, fmt.Errorf("permohonan berstatus %s belum siap dijadwalkan", status)
	}
	var existingID int64
	var existingStatus string
	err = tx.QueryRowContext(ctx, `SELECT id,status FROM horticulture_inspection_schedules WHERE application_id=? FOR UPDATE`, schedule.ApplicationID).Scan(&existingID, &existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return InspectionSchedule{}, err
	}
	if err == nil {
		if existingStatus != "terjadwal" {
			return InspectionSchedule{}, errors.New("inspeksi sudah selesai; jadwal tidak dapat diubah")
		}
		_, err = tx.ExecContext(ctx, `UPDATE horticulture_inspection_schedules SET inspector_id=?,scheduled_at=?,location=?,notes=? WHERE id=?`, schedule.InspectorID, schedule.ScheduledAt, schedule.Location, schedule.Notes, existingID)
		schedule.ID = existingID
	} else {
		schedule.ID, err = db.InsertID(ctx, tx, `INSERT INTO horticulture_inspection_schedules (application_id,inspector_id,scheduled_at,location,notes,status) VALUES (?,?,?,?,?,'terjadwal')`, schedule.ApplicationID, schedule.InspectorID, schedule.ScheduledAt, schedule.Location, schedule.Notes)
	}
	if err != nil {
		return InspectionSchedule{}, fmt.Errorf("simpan jadwal inspeksi: %w", err)
	}
	schedule.Status = "terjadwal"
	if _, err := tx.ExecContext(ctx, `UPDATE horticulture_applications SET status='dijadwalkan' WHERE id=?`, schedule.ApplicationID); err != nil {
		return InspectionSchedule{}, err
	}
	if err := insertAudit(ctx, tx, actorID, "horticulture:jadwalkan_inspeksi", schedule.ApplicationID, "status=dijadwalkan"); err != nil {
		return InspectionSchedule{}, err
	}
	if err := tx.Commit(); err != nil {
		return InspectionSchedule{}, err
	}
	return schedule, nil
}

func (r *PostgresRepository) CompleteInspection(ctx context.Context, inspection Inspection) (Inspection, error) {
	checklist, err := json.Marshal(inspection.Checklist)
	if err != nil {
		return Inspection{}, fmt.Errorf("encode checklist GAP: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Inspection{}, fmt.Errorf("mulai transaksi inspeksi: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var appID int64
	var status string
	err = tx.QueryRowContext(ctx, `SELECT s.application_id,s.status FROM horticulture_inspection_schedules s WHERE s.id=? AND s.inspector_id=? FOR UPDATE`, inspection.ScheduleID, inspection.InspectorID).Scan(&appID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return Inspection{}, ErrNotFound
	}
	if err != nil {
		return Inspection{}, err
	}
	if status != "terjadwal" {
		return Inspection{}, fmt.Errorf("jadwal inspeksi berstatus %s tidak dapat diselesaikan", status)
	}
	var lockedAppID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM horticulture_applications WHERE id=? FOR UPDATE`, appID).Scan(&lockedAppID); err != nil {
		return Inspection{}, err
	}
	inspection.ID, err = db.InsertID(ctx, tx, `INSERT INTO horticulture_inspections (schedule_id,checklist_json,findings,recommendations,inspector_id) VALUES (?,?,?,?,?)`, inspection.ScheduleID, string(checklist), inspection.Findings, inspection.Recommendations, inspection.InspectorID)
	if err != nil {
		return Inspection{}, fmt.Errorf("simpan hasil inspeksi GAP: %w", err)
	}
	inspection.CreatedAt = time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE horticulture_inspection_schedules SET status='selesai' WHERE id=?`, inspection.ScheduleID); err != nil {
		return Inspection{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE horticulture_applications SET status='diinspeksi' WHERE id=?`, appID); err != nil {
		return Inspection{}, err
	}
	if err := insertAudit(ctx, tx, inspection.InspectorID, "horticulture:inspeksi_gap", appID, "status=diinspeksi"); err != nil {
		return Inspection{}, err
	}
	if err := tx.Commit(); err != nil {
		return Inspection{}, err
	}
	return inspection, nil
}

func (r *PostgresRepository) SaveEvaluation(ctx context.Context, evaluation Evaluation) (Evaluation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Evaluation{}, fmt.Errorf("mulai transaksi penilaian: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM horticulture_applications WHERE id=? FOR UPDATE`, evaluation.ApplicationID).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return Evaluation{}, ErrNotFound
	} else if err != nil {
		return Evaluation{}, err
	}
	if status != "diinspeksi" && status != "dinilai" {
		return Evaluation{}, fmt.Errorf("permohonan berstatus %s belum dapat dinilai", status)
	}
	evaluation.ID, err = db.InsertID(ctx, tx, `INSERT INTO horticulture_evaluations (application_id,assessment_type,plant_name,variety,vigor_score,health_score,variety_score,recommendation,notes,evaluator_id) VALUES (?,?,?,?,?,?,?,?,?,?)`, evaluation.ApplicationID, evaluation.AssessmentType, evaluation.PlantName, evaluation.Variety, evaluation.VigorScore, evaluation.HealthScore, evaluation.VarietyScore, evaluation.Recommendation, evaluation.Notes, evaluation.EvaluatorID)
	if err != nil {
		return Evaluation{}, fmt.Errorf("simpan penilaian: %w", err)
	}
	evaluation.CreatedAt = time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE horticulture_applications SET status='dinilai' WHERE id=?`, evaluation.ApplicationID); err != nil {
		return Evaluation{}, err
	}
	if err := insertAudit(ctx, tx, evaluation.EvaluatorID, "horticulture:penilaian_pohon_induk", evaluation.ApplicationID, evaluation.Recommendation); err != nil {
		return Evaluation{}, err
	}
	if err := tx.Commit(); err != nil {
		return Evaluation{}, err
	}
	return evaluation, nil
}

func insertAudit(ctx context.Context, tx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, actorID int64, action string, entityID int64, detail string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO audit_log (actor_id,actor_role,action,entity,entity_id,detail) SELECT id,role,?,'horticulture_application',?,? FROM users WHERE id=?`, action, entityID, detail, actorID)
	if err != nil {
		return fmt.Errorf("catat audit %s: %w", action, err)
	}
	return nil
}

func (r *PostgresRepository) Create(ctx context.Context, app Application, documents []Document) (Application, error) {
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
	app.ID, err = db.InsertID(ctx, tx, `INSERT INTO horticulture_applications (application_number,applicant_id,organization_name,contact_name,phone,address,village,district,regency,cultivated_area,crop_details,status) VALUES (?,?,?,?,?,?,?,?,?,?,?,'diajukan')`,
		app.Number, app.ApplicantID, app.Organization, app.Contact, app.Phone, app.Address, app.Village, app.District, app.Regency, app.CultivatedArea, string(crops))
	if err != nil {
		return Application{}, fmt.Errorf("simpan permohonan registrasi lahan: %w", err)
	}
	if err != nil {
		return Application{}, fmt.Errorf("simpan permohonan registrasi lahan: %w", err)
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

func (r *PostgresRepository) List(ctx context.Context, applicantID int64, all bool) ([]Application, error) {
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
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range apps {
		apps[i].ApplicantID = 0
		apps[i].Documents = nil
		apps[i].Crops = nil
		apps[i].Schedules = nil
		apps[i].Inspections = nil
		apps[i].Evaluations = nil
	}
	return apps, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id int64) (Application, error) {
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
	app.Schedules, err = r.listSchedules(ctx, id)
	if err != nil {
		return Application{}, err
	}
	app.Schedules = redactScheduleInspectorIDs(app.Schedules)
	app.Inspections, err = r.listInspections(ctx, id)
	if err != nil {
		return Application{}, err
	}
	app.Evaluations, err = r.listEvaluations(ctx, id)
	if err != nil {
		return Application{}, err
	}
	return app, nil
}

func redactScheduleInspectorIDs(items []InspectionSchedule) []InspectionSchedule {
	for i := range items {
		items[i].InspectorID = 0
	}
	return items
}

func (r *PostgresRepository) listSchedules(ctx context.Context, applicationID int64) ([]InspectionSchedule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,application_id,inspector_id,scheduled_at,location,notes,status FROM horticulture_inspection_schedules WHERE application_id=? ORDER BY scheduled_at`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("daftar jadwal inspeksi: %w", err)
	}
	defer rows.Close()
	items := make([]InspectionSchedule, 0)
	for rows.Next() {
		var item InspectionSchedule
		if err := rows.Scan(&item.ID, &item.ApplicationID, &item.InspectorID, &item.ScheduledAt, &item.Location, &item.Notes, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) listInspections(ctx context.Context, applicationID int64) ([]Inspection, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT i.id,i.schedule_id,i.checklist_json,i.findings,i.recommendations,i.inspector_id,i.created_at FROM horticulture_inspections i JOIN horticulture_inspection_schedules s ON s.id=i.schedule_id WHERE s.application_id=? ORDER BY i.id`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("daftar hasil inspeksi: %w", err)
	}
	defer rows.Close()
	items := make([]Inspection, 0)
	for rows.Next() {
		var item Inspection
		var raw []byte
		if err := rows.Scan(&item.ID, &item.ScheduleID, &raw, &item.Findings, &item.Recommendations, &item.InspectorID, &item.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &item.Checklist); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) listEvaluations(ctx context.Context, applicationID int64) ([]Evaluation, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,application_id,assessment_type,plant_name,variety,vigor_score,health_score,variety_score,recommendation,notes,evaluator_id,created_at FROM horticulture_evaluations WHERE application_id=? ORDER BY id`, applicationID)
	if err != nil {
		return nil, fmt.Errorf("daftar penilaian: %w", err)
	}
	defer rows.Close()
	items := make([]Evaluation, 0)
	for rows.Next() {
		var item Evaluation
		if err := rows.Scan(&item.ID, &item.ApplicationID, &item.AssessmentType, &item.PlantName, &item.Variety, &item.VigorScore, &item.HealthScore, &item.VarietyScore, &item.Recommendation, &item.Notes, &item.EvaluatorID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) listDocuments(ctx context.Context, id int64) ([]Document, error) {
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
