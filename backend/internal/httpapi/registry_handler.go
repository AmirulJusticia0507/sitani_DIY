package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// registryHandler menangani registrasi benih, persediaan, penjualan, dan laporan.
type registryHandler struct {
	svc *service.Service
}

func newRegistryHandler(svc *service.Service) *registryHandler {
	return &registryHandler{svc: svc}
}

func (h *registryHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/registrasi-benih", h.buatRegistrasi)
	mux.HandleFunc("GET /v1/registrasi-benih", h.listRegistrasi)
	mux.HandleFunc("GET /v1/registrasi-benih/{id}", h.getRegistrasi)
	mux.HandleFunc("POST /v1/registrasi-benih/{id}/status", h.ubahStatus)
	mux.HandleFunc("GET /v1/registrasi-benih/{id}/stok", h.stokRegistrasi)
	mux.HandleFunc("POST /v1/penjualan-benih", h.jualBenih)
	mux.HandleFunc("GET /v1/penjualan-benih", h.listPenjualan)
	mux.HandleFunc("POST /v1/penjualan-benih/{id}/batal", h.batalkanPenjualan)
	mux.HandleFunc("GET /v1/laporan/registrasi-benih", h.laporanRegistrasi)
	mux.HandleFunc("GET /v1/laporan/stok-benih", h.laporanStok)
	mux.HandleFunc("GET /v1/laporan/kedaluwarsa-benih", h.laporanKedaluwarsa)
	mux.HandleFunc("GET /v1/laporan/penjualan-benih", h.laporanPenjualan)
}

type registrasiRequest struct {
	PermohonanID       int64   `json:"permohonan_id"`
	Komoditas          string  `json:"komoditas"`
	Varietas           string  `json:"varietas"`
	Produsen           string  `json:"produsen"`
	SumberBenih        string  `json:"sumber_benih"`
	NomorLot           string  `json:"nomor_lot"`
	KelasBenih         string  `json:"kelas_benih"`
	Jumlah             float64 `json:"jumlah"`
	Satuan             string  `json:"satuan"`
	TanggalProduksi    string  `json:"tanggal_produksi"`
	TanggalKedaluwarsa string  `json:"tanggal_kedaluwarsa"`
	StatusEdar         string  `json:"status_edar"`
}

type ubahStatusRequest struct {
	StatusEdar string `json:"status_edar"`
}

type penjualanRequest struct {
	NomorDokumen     string  `json:"nomor_dokumen"`
	RegistrasiID     int64   `json:"registrasi_id"`
	Pembeli          string  `json:"pembeli"`
	Jumlah           float64 `json:"jumlah"`
	HargaSatuan      float64 `json:"harga_satuan"`
	TanggalTransaksi string  `json:"tanggal_transaksi"`
}

type batalPenjualanRequest struct {
	Alasan string `json:"alasan"`
}

type stokRegistrasiResponse struct {
	Registrasi *db.RegistrasiBenih  `json:"registrasi"`
	Stok       *db.StokBenih        `json:"stok"`
	Mutasi     []db.MutasiStokBenih `json:"mutasi"`
}

func (h *registryHandler) buatRegistrasi(w http.ResponseWriter, r *http.Request) {
	actorID, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req registrasiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	registrasi, err := h.svc.BuatRegistrasiBenih(r.Context(), actorID, role, service.BuatRegistrasiRequest{
		PermohonanID:       req.PermohonanID,
		Komoditas:          req.Komoditas,
		Varietas:           req.Varietas,
		Produsen:           req.Produsen,
		SumberBenih:        req.SumberBenih,
		NomorLot:           req.NomorLot,
		KelasBenih:         req.KelasBenih,
		Jumlah:             req.Jumlah,
		Satuan:             req.Satuan,
		TanggalProduksi:    req.TanggalProduksi,
		TanggalKedaluwarsa: req.TanggalKedaluwarsa,
		StatusEdar:         req.StatusEdar,
	})
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, registrasi)
}

func (h *registryHandler) listRegistrasi(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	filter := db.RegistrasiFilter{
		StatusEdar: strings.TrimSpace(r.URL.Query().Get("status_edar")),
		Komoditas:  strings.TrimSpace(r.URL.Query().Get("komoditas")),
		Limit:      parseLimit(r),
	}
	items, err := h.svc.ListRegistrasiBenih(r.Context(), role, filter)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *registryHandler) getRegistrasi(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	id, ok := parsePathID(w, r, "id", "id registrasi tidak valid")
	if !ok {
		return
	}
	registrasi, err := h.svc.GetRegistrasiBenih(r.Context(), role, id)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, registrasi)
}

func (h *registryHandler) ubahStatus(w http.ResponseWriter, r *http.Request) {
	actorID, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	id, ok := parsePathID(w, r, "id", "id registrasi tidak valid")
	if !ok {
		return
	}
	var req ubahStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	if err := h.svc.UbahStatusEdar(r.Context(), actorID, role, id, strings.TrimSpace(req.StatusEdar)); err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "status edar diperbarui"})
}

func (h *registryHandler) stokRegistrasi(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	id, ok := parsePathID(w, r, "id", "id registrasi tidak valid")
	if !ok {
		return
	}
	registrasi, err := h.svc.GetRegistrasiBenih(r.Context(), role, id)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	stok, mutasi, err := h.svc.StokBenih(r.Context(), role, id)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, stokRegistrasiResponse{Registrasi: registrasi, Stok: stok, Mutasi: mutasi})
}

func (h *registryHandler) jualBenih(w http.ResponseWriter, r *http.Request) {
	actorID, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req penjualanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	penjualan, err := h.svc.JualBenih(r.Context(), actorID, role, service.JualBenihRequest{
		NomorDokumen:     req.NomorDokumen,
		RegistrasiID:     req.RegistrasiID,
		Pembeli:          req.Pembeli,
		Jumlah:           req.Jumlah,
		HargaSatuan:      req.HargaSatuan,
		TanggalTransaksi: req.TanggalTransaksi,
	})
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, penjualan)
}

func (h *registryHandler) listPenjualan(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	filter, err := penjualanFilterDariQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.svc.LaporanPenjualan(r.Context(), role, filter)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *registryHandler) batalkanPenjualan(w http.ResponseWriter, r *http.Request) {
	actorID, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	id, ok := parsePathID(w, r, "id", "id penjualan tidak valid")
	if !ok {
		return
	}
	var req batalPenjualanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	penjualan, err := h.svc.BatalkanPenjualan(r.Context(), actorID, role, id, req.Alasan)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, penjualan)
}

func (h *registryHandler) laporanRegistrasi(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	filter := db.RegistrasiFilter{
		StatusEdar: strings.TrimSpace(r.URL.Query().Get("status_edar")),
		Komoditas:  strings.TrimSpace(r.URL.Query().Get("komoditas")),
		Limit:      parseLimit(r),
	}
	items, err := h.svc.LaporanRegistrasi(r.Context(), role, filter)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *registryHandler) laporanStok(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	items, err := h.svc.LaporanStok(r.Context(), role)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *registryHandler) laporanKedaluwarsa(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	hari, _ := strconv.Atoi(r.URL.Query().Get("hari"))
	items, err := h.svc.LaporanKedaluwarsa(r.Context(), role, hari)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *registryHandler) laporanPenjualan(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasAktor(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	filter, err := penjualanFilterDariQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.svc.LaporanPenjualan(r.Context(), role, filter)
	if err != nil {
		respondRegistryError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func penjualanFilterDariQuery(r *http.Request) (db.PenjualanFilter, error) {
	query := r.URL.Query()
	filter := db.PenjualanFilter{
		Status: strings.TrimSpace(query.Get("status")),
		Limit:  parseLimit(r),
	}
	if value := query.Get("registrasi_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			return filter, errors.New("registrasi_id tidak valid")
		}
		filter.RegistrasiID = id
	}
	if value := strings.TrimSpace(query.Get("dari")); value != "" {
		tanggal, err := time.Parse("2006-01-02", value)
		if err != nil {
			return filter, errors.New("parameter dari harus berformat YYYY-MM-DD")
		}
		filter.Dari = &tanggal
	}
	if value := strings.TrimSpace(query.Get("sampai")); value != "" {
		tanggal, err := time.Parse("2006-01-02", value)
		if err != nil {
			return filter, errors.New("parameter sampai harus berformat YYYY-MM-DD")
		}
		filter.Sampai = &tanggal
	}
	return filter, nil
}

func parseLimit(r *http.Request) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	return limit
}

func parsePathID(w http.ResponseWriter, r *http.Request, name, pesan string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, pesan)
		return 0, false
	}
	return id, true
}

func respondRegistryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		respondError(w, http.StatusNotFound, "data tidak ditemukan")
	case errors.Is(err, store.ErrStokTidakCukup):
		respondError(w, http.StatusConflict, "stok benih tidak mencukupi")
	case errors.Is(err, store.ErrSudahDibatalkan):
		respondError(w, http.StatusConflict, "penjualan sudah dibatalkan")
	default:
		respondError(w, http.StatusBadRequest, err.Error())
	}
}

// identitasAktor memakai principal SSO yang sudah diverifikasi. Header hanya
// menjadi fallback untuk pengujian handler yang tidak melewati middleware SSO.
func identitasAktor(r *http.Request) (int64, string, error) {
	if principal, ok := auth.PrincipalFromContext(r.Context()); ok {
		return principal.UserID, principal.Role, nil
	}
	return identitasDariHeader(r)
}
