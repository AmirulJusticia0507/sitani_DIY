package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/horticulture"
)

type horticultureHandler struct{ repo horticulture.Repository }

func (h *horticultureHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/horticulture/applications", h.create)
	mux.HandleFunc("GET /v1/horticulture/applications", h.list)
	mux.HandleFunc("GET /v1/horticulture/applications/{id}", h.get)
	mux.HandleFunc("POST /v1/horticulture/applications/{id}/verification", h.verify)
	mux.HandleFunc("POST /v1/horticulture/applications/{id}/inspection-schedules", h.schedule)
	mux.HandleFunc("POST /v1/horticulture/inspection-schedules/{id}/complete", h.completeInspection)
	mux.HandleFunc("POST /v1/horticulture/applications/{id}/evaluations", h.evaluate)
	mux.HandleFunc("GET /v1/horticulture/applications/{id}/evaluations", h.listEvaluations)
}

type horticultureRequest struct {
	Organization   string                  `json:"nama_organisasi"`
	Contact        string                  `json:"nama_kontak"`
	Phone          string                  `json:"telepon"`
	Address        string                  `json:"alamat"`
	Village        string                  `json:"kalurahan"`
	District       string                  `json:"kapanewon"`
	Regency        string                  `json:"kabupaten_kota"`
	CultivatedArea float64                 `json:"luas_lahan_ha"`
	Crops          []horticulture.Crop     `json:"tanaman"`
	Documents      []horticulture.Document `json:"dokumen"`
}

type verificationRequest struct {
	Decision string `json:"keputusan"`
	Note     string `json:"catatan"`
}
type scheduleRequest struct {
	InspectorID int64  `json:"inspektor_id"`
	ScheduledAt string `json:"waktu"`
	Location    string `json:"lokasi"`
	Notes       string `json:"catatan"`
}
type inspectionRequest struct {
	Checklist       map[string]bool `json:"checklist_gap"`
	Findings        string          `json:"temuan"`
	Recommendations string          `json:"rekomendasi"`
}
type evaluationRequest struct {
	AssessmentType string `json:"jenis_penilaian"`
	PlantName      string `json:"nama_tanaman"`
	Variety        string `json:"varietas"`
	VigorScore     int    `json:"skor_vigoritas"`
	HealthScore    int    `json:"skor_kesehatan"`
	VarietyScore   int    `json:"skor_kebenaran_varietas"`
	Recommendation string `json:"rekomendasi"`
	Notes          string `json:"catatan"`
}

func (h *horticultureHandler) create(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID < 1 {
		respondError(w, http.StatusUnauthorized, "identitas SSO tidak tersedia")
		return
	}
	if principal.Role != "pelanggan" && principal.Role != "admin" {
		respondError(w, http.StatusForbidden, "peran tidak diizinkan mengajukan registrasi lahan")
		return
	}
	if h.repo == nil {
		respondError(w, http.StatusServiceUnavailable, "database belum tersedia")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var req horticultureRequest
	if err := decoder.Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permohonan tidak valid")
		return
	}
	if err := validateHorticultureRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	app, err := h.repo.Create(r.Context(), horticulture.Application{
		ApplicantID: principal.UserID, Organization: strings.TrimSpace(req.Organization),
		Contact: strings.TrimSpace(req.Contact), Phone: strings.TrimSpace(req.Phone),
		Address: strings.TrimSpace(req.Address), Village: strings.TrimSpace(req.Village),
		District: strings.TrimSpace(req.District), Regency: strings.TrimSpace(req.Regency),
		CultivatedArea: req.CultivatedArea, Crops: req.Crops,
	}, req.Documents)
	if err != nil {
		respondError(w, http.StatusBadRequest, "gagal menyimpan permohonan atau dokumen tidak dimiliki pemohon")
		return
	}
	respondJSON(w, http.StatusCreated, app)
}

func validateHorticultureRequest(req horticultureRequest) error {
	for label, value := range map[string]string{
		"nama_organisasi": req.Organization, "nama_kontak": req.Contact, "telepon": req.Phone,
		"alamat": req.Address, "kalurahan": req.Village, "kapanewon": req.District, "kabupaten_kota": req.Regency,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s wajib diisi", label)
		}
		if utf8.RuneCountInString(strings.TrimSpace(value)) > 500 {
			return fmt.Errorf("%s terlalu panjang", label)
		}
	}
	if req.CultivatedArea <= 0 || req.CultivatedArea > 100000 {
		return errors.New("luas_lahan_ha harus lebih dari 0 dan maksimal 100000")
	}
	if len(req.Crops) == 0 || len(req.Crops) > 100 {
		return errors.New("tanaman wajib memuat 1 sampai 100 komoditas")
	}
	for i, crop := range req.Crops {
		if strings.TrimSpace(crop.Name) == "" || strings.TrimSpace(crop.Variety) == "" || crop.PlantCount < 0 || crop.PlantCount > 100000000 {
			return fmt.Errorf("detail tanaman ke-%d tidak valid", i+1)
		}
	}
	if len(req.Documents) < 1 || len(req.Documents) > 20 {
		return errors.New("lampiran wajib memuat 1 sampai 20 dokumen")
	}
	seen := map[int64]bool{}
	seenTypes := map[string]bool{}
	allowed := map[string]bool{"proposal": true, "organization": true, "members": true, "farm_form": true, "farm_map": true, "other": true}
	for _, doc := range req.Documents {
		if doc.ID < 1 || !allowed[doc.Type] || seen[doc.ID] {
			return errors.New("daftar lampiran tidak valid")
		}
		seen[doc.ID] = true
		seenTypes[doc.Type] = true
	}
	for _, required := range []string{"proposal", "organization", "members", "farm_form", "farm_map"} {
		if !seenTypes[required] {
			return fmt.Errorf("lampiran %s wajib diunggah", required)
		}
	}
	return nil
}

func (h *horticultureHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID < 1 {
		respondError(w, http.StatusUnauthorized, "identitas SSO tidak tersedia")
		return
	}
	if principal.Role != "pelanggan" && principal.Role != "admin" && principal.Role != "pengawas" && principal.Role != "verifikator" && principal.Role != "inspektor" && principal.Role != "pbt" && principal.Role != "kasi_pmbtp" {
		respondError(w, http.StatusForbidden, "peran tidak diizinkan melihat permohonan")
		return
	}
	if h.repo == nil {
		respondError(w, http.StatusServiceUnavailable, "database belum tersedia")
		return
	}
	items, err := h.repo.List(r.Context(), principal.UserID, principal.Role == "admin" || principal.Role == "pengawas")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "gagal membaca permohonan")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"permohonan": items})
}

func (h *horticultureHandler) get(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID < 1 {
		respondError(w, http.StatusUnauthorized, "identitas SSO tidak tersedia")
		return
	}
	if principal.Role != "pelanggan" && principal.Role != "admin" && principal.Role != "pengawas" && principal.Role != "verifikator" && principal.Role != "inspektor" && principal.Role != "pbt" && principal.Role != "kasi_pmbtp" {
		respondError(w, http.StatusForbidden, "peran tidak diizinkan melihat permohonan")
		return
	}
	if h.repo == nil {
		respondError(w, http.StatusServiceUnavailable, "database belum tersedia")
		return
	}
	id, valid := parsePathID(w, r, "id", "ID permohonan tidak valid")
	if !valid {
		return
	}
	app, err := h.repo.Get(r.Context(), id)
	if errors.Is(err, horticulture.ErrNotFound) {
		respondError(w, http.StatusNotFound, "permohonan tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "gagal membaca permohonan")
		return
	}
	if principal.Role == "pelanggan" && app.ApplicantID != principal.UserID {
		respondError(w, http.StatusNotFound, "permohonan tidak ditemukan")
		return
	}
	respondJSON(w, http.StatusOK, app)
}

func decodeHorticultureJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return false
	}
	return true
}

func (h *horticultureHandler) verify(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID < 1 {
		respondError(w, 401, "identitas SSO tidak tersedia")
		return
	}
	if p.Role != "admin" && p.Role != "verifikator" {
		respondError(w, 403, "hanya verifikator yang dapat memverifikasi proposal")
		return
	}
	if h.repo == nil { respondError(w, http.StatusServiceUnavailable, "database belum tersedia"); return }
	id, valid := parsePathID(w, r, "id", "ID permohonan tidak valid")
	if !valid {
		return
	}
	var req verificationRequest
	if !decodeHorticultureJSON(w, r, &req) {
		return
	}
	if req.Decision != "diverifikasi" && req.Decision != "perlu_perbaikan" && req.Decision != "ditolak" {
		respondError(w, http.StatusBadRequest, "keputusan verifikasi tidak valid")
		return
	}
	if (req.Decision == "perlu_perbaikan" || req.Decision == "ditolak") && strings.TrimSpace(req.Note) == "" {
		respondError(w, http.StatusBadRequest, "catatan alasan wajib diisi")
		return
	}
	app, err := h.repo.Verify(r.Context(), id, p.UserID, req.Decision, req.Note)
	if errors.Is(err, horticulture.ErrNotFound) {
		respondError(w, 404, "permohonan tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, 400, err.Error())
		return
	}
	respondJSON(w, 200, app)
}

func (h *horticultureHandler) schedule(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID < 1 {
		respondError(w, 401, "identitas SSO tidak tersedia")
		return
	}
	if p.Role != "admin" && p.Role != "inspektor" {
		respondError(w, 403, "hanya inspektor yang dapat menjadwalkan inspeksi")
		return
	}
	if h.repo == nil { respondError(w, http.StatusServiceUnavailable, "database belum tersedia"); return }
	id, valid := parsePathID(w, r, "id", "ID permohonan tidak valid")
	if !valid {
		return
	}
	var req scheduleRequest
	if !decodeHorticultureJSON(w, r, &req) {
		return
	}
	when, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil || !when.After(time.Now()) {
		respondError(w, 400, "waktu inspeksi harus berupa RFC3339 dan berada di masa depan")
		return
	}
	if strings.TrimSpace(req.Location) == "" || len(req.Location) > 255 {
		respondError(w, 400, "lokasi inspeksi wajib diisi (maks. 255 karakter)")
		return
	}
	if req.InspectorID != 0 && req.InspectorID != p.UserID {
		respondError(w, 403, "inspektor hanya dapat menjadwalkan dirinya sendiri")
		return
	}
	schedule, err := h.repo.ScheduleInspection(r.Context(), horticulture.InspectionSchedule{ApplicationID: id, InspectorID: p.UserID, ScheduledAt: when, Location: strings.TrimSpace(req.Location), Notes: strings.TrimSpace(req.Notes)}, p.UserID)
	if errors.Is(err, horticulture.ErrNotFound) {
		respondError(w, 404, "permohonan tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, 400, err.Error())
		return
	}
	respondJSON(w, 201, schedule)
}

func (h *horticultureHandler) completeInspection(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID < 1 {
		respondError(w, 401, "identitas SSO tidak tersedia")
		return
	}
	if p.Role != "admin" && p.Role != "inspektor" {
		respondError(w, 403, "hanya inspektor yang dapat mengisi audit GAP")
		return
	}
	if h.repo == nil { respondError(w, http.StatusServiceUnavailable, "database belum tersedia"); return }
	if h.repo == nil { respondError(w, http.StatusServiceUnavailable, "database belum tersedia"); return }
	id, valid := parsePathID(w, r, "id", "ID jadwal tidak valid")
	if !valid {
		return
	}
	var req inspectionRequest
	if !decodeHorticultureJSON(w, r, &req) {
		return
	}
	if err := validateGAPChecklist(req.Checklist); err != nil {
		respondError(w, 400, err.Error())
		return
	}
	if strings.TrimSpace(req.Findings) == "" || strings.TrimSpace(req.Recommendations) == "" {
		respondError(w, 400, "temuan dan rekomendasi wajib diisi")
		return
	}
	inspection, err := h.repo.CompleteInspection(r.Context(), horticulture.Inspection{ScheduleID: id, Checklist: req.Checklist, Findings: strings.TrimSpace(req.Findings), Recommendations: strings.TrimSpace(req.Recommendations), InspectorID: p.UserID})
	if errors.Is(err, horticulture.ErrNotFound) {
		respondError(w, 404, "jadwal inspeksi tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, 400, err.Error())
		return
	}
	respondJSON(w, 201, inspection)
}

var requiredGAPChecks = []string{"legalitas_lahan", "riwayat_lahan", "kondisi_lahan", "sumber_air", "drainase", "konservasi_tanah", "benih_bibit", "pemupukan", "pengendalian_hama", "panen_pascapanen", "pencatatan_budidaya"}

func validateGAPChecklist(checklist map[string]bool) error {
	if len(checklist) != len(requiredGAPChecks) {
		return errors.New("checklist audit GAP harus mengisi seluruh indikator")
	}
	for _, key := range requiredGAPChecks {
		if _, ok := checklist[key]; !ok {
			return fmt.Errorf("indikator GAP %s wajib diisi", key)
		}
	}
	return nil
}

func (h *horticultureHandler) evaluate(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID < 1 {
		respondError(w, 401, "identitas SSO tidak tersedia")
		return
	}
	if p.Role != "admin" && p.Role != "pbt" && p.Role != "kasi_pmbtp" {
		respondError(w, 403, "hanya PBT atau Kasi PMBTP yang dapat menilai")
		return
	}
	if h.repo == nil { respondError(w, http.StatusServiceUnavailable, "database belum tersedia"); return }
	id, valid := parsePathID(w, r, "id", "ID permohonan tidak valid")
	if !valid {
		return
	}
	var req evaluationRequest
	if !decodeHorticultureJSON(w, r, &req) {
		return
	}
	allowedType := map[string]bool{"pohon_induk": true, "pit": true, "rip": true, "pi": true, "ri": true}
	allowedRecommendation := map[string]bool{"layak": true, "perlu_perbaikan": true, "tidak_layak": true}
	if !allowedType[req.AssessmentType] || !allowedRecommendation[req.Recommendation] || strings.TrimSpace(req.PlantName) == "" || strings.TrimSpace(req.Variety) == "" || strings.TrimSpace(req.Notes) == "" || req.VigorScore < 1 || req.VigorScore > 5 || req.HealthScore < 1 || req.HealthScore > 5 || req.VarietyScore < 1 || req.VarietyScore > 5 {
		respondError(w, 400, "data penilaian tidak valid; skor setiap parameter harus 1 sampai 5")
		return
	}
	evaluation, err := h.repo.SaveEvaluation(r.Context(), horticulture.Evaluation{ApplicationID: id, AssessmentType: req.AssessmentType, PlantName: strings.TrimSpace(req.PlantName), Variety: strings.TrimSpace(req.Variety), VigorScore: req.VigorScore, HealthScore: req.HealthScore, VarietyScore: req.VarietyScore, Recommendation: req.Recommendation, Notes: strings.TrimSpace(req.Notes), EvaluatorID: p.UserID})
	if errors.Is(err, horticulture.ErrNotFound) {
		respondError(w, 404, "permohonan tidak ditemukan")
		return
	}
	if err != nil {
		respondError(w, 400, err.Error())
		return
	}
	respondJSON(w, 201, evaluation)
}

func (h *horticultureHandler) listEvaluations(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok || p.UserID < 1 { respondError(w,401,"identitas SSO tidak tersedia"); return }
	if p.Role!="pelanggan" && p.Role!="admin" && p.Role!="pengawas" && p.Role!="pbt" && p.Role!="kasi_pmbtp" { respondError(w,403,"peran tidak diizinkan melihat hasil penilaian"); return }
	if h.repo == nil { respondError(w,http.StatusServiceUnavailable,"database belum tersedia"); return }
	id, valid := parsePathID(w,r,"id","ID permohonan tidak valid"); if !valid { return }
	app, err := h.repo.Get(r.Context(),id)
	if errors.Is(err,horticulture.ErrNotFound) { respondError(w,404,"permohonan tidak ditemukan"); return }; if err!=nil { respondError(w,500,"gagal membaca hasil penilaian"); return }
	if p.Role=="pelanggan" && app.ApplicantID!=p.UserID { respondError(w,404,"permohonan tidak ditemukan"); return }
	respondJSON(w,200,map[string]any{"penilaian":app.Evaluations})
}
