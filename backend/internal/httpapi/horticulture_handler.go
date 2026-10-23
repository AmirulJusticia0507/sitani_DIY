package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/horticulture"
)

type horticultureHandler struct{ repo horticulture.Repository }

func (h *horticultureHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/horticulture/applications", h.create)
	mux.HandleFunc("GET /v1/horticulture/applications", h.list)
	mux.HandleFunc("GET /v1/horticulture/applications/{id}", h.get)
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
	if principal.Role != "pelanggan" && principal.Role != "admin" && principal.Role != "pengawas" {
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
	if principal.Role != "pelanggan" && principal.Role != "admin" && principal.Role != "pengawas" {
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
