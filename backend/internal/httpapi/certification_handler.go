package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
)

// certificationHandler menangani endpoint Fase 1 verifikasi lapangan BP3MBTP.
// Identitas pemanggil diambil dari header permintaan (X-User-ID, X-User-Role)
// selama pengembangan; pada produksi identitas diambil dari token SSO yang
// telah diverifikasi melalui adaptor otentikasi.
type certificationHandler struct {
	svc *service.Service
}

func newCertificationHandler(svc *service.Service) *certificationHandler {
	return &certificationHandler{svc: svc}
}

func (h *certificationHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/permohonan", h.buatPermohonan)
	mux.HandleFunc("GET /v1/permohonan/{id}", h.getPermohonan)
	mux.HandleFunc("POST /v1/permohonan/{id}/pembayaran/awal", h.bayarAwal)
	mux.HandleFunc("POST /v1/permohonan/{id}/inspeksi", h.catatInspeksi)
	mux.HandleFunc("POST /v1/permohonan/{id}/buku-induk", h.catatBukuInduk)
}

type permohonanRequest struct {
	JenisBenih   string  `json:"jenis_benih"`
	Varietas     string  `json:"varietas"`
	LuasLahan    float64 `json:"luas_lahan"`
	LokasiKebun  string  `json:"lokasi_kebun"`
}

type pembayaranAwalRequest struct {
	Jumlah float64 `json:"jumlah"`
}

type inspeksiRequest struct {
	Fase            int    `json:"fase"`
	Hasil           string `json:"hasil"`
	Catatan         string `json:"catatan"`
	TanggalInspeksi string `json:"tanggal_inspeksi"`
}

type bukuIndukRequest struct {
	NomorBukuInduk string `json:"nomor_buku_induk"`
	Catatan        string `json:"catatan"`
}

func (h *certificationHandler) buatPermohonan(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if role != "pelanggan" && role != "admin" {
		respondError(w, http.StatusForbidden, "hanya pelanggan atau admin yang dapat membuat permohonan")
		return
	}

	var req permohonanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}

	p, err := h.svc.BuatPermohonan(r.Context(), userID, role, service.BuatPermohonanRequest{
		JenisBenih:   req.JenisBenih,
		Varietas:     req.Varietas,
		LuasLahan:    req.LuasLahan,
		LokasiKebun:  req.LokasiKebun,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, p)
}

func (h *certificationHandler) getPermohonan(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id permohonan tidak valid")
		return
	}
	p, err := h.svc.GetPermohonan(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "permohonan tidak ditemukan")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if role != "admin" && p.PelangganID != userID {
		respondError(w, http.StatusForbidden, "bukan pemilik permohonan")
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *certificationHandler) bayarAwal(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id permohonan tidak valid")
		return
	}

	var req pembayaranAwalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}

	if err := h.svc.BayarAwal(r.Context(), id, userID, role, req.Jumlah); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "pembayaran awal dicatat"})
}

func (h *certificationHandler) catatInspeksi(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id permohonan tidak valid")
		return
	}

	var req inspeksiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}

	if err := h.svc.CatatInspeksi(r.Context(), id, req.Fase, userID, role, req.Hasil, req.Catatan, req.TanggalInspeksi); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": fmt.Sprintf("inspeksi fase %d dicatat", req.Fase)})
}

func (h *certificationHandler) catatBukuInduk(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id permohonan tidak valid")
		return
	}

	var req bukuIndukRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}

	if err := h.svc.CatatBukuInduk(r.Context(), id, req.NomorBukuInduk, req.Catatan, userID, role); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "buku induk dicatat"})
}

// identitasDariHeader membaca identitas pengembang dari header permintaan.
func identitasDariHeader(r *http.Request) (int64, string, error) {
	userIDStr := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")
	if userIDStr == "" || role == "" {
		return 0, "", errors.New("identitas tidak tersedia (atur header X-User-ID dan X-User-Role)")
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, "", errors.New("X-User-ID tidak valid")
	}
	return userID, role, nil
}
