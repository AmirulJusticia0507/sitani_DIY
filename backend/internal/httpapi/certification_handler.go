package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	mux.HandleFunc("POST /v1/sampel", h.ambilSampel)
	mux.HandleFunc("POST /v1/sampel/{id}/kodifikasi", h.kodifikasiSampel)
	mux.HandleFunc("POST /v1/uji-mutu", h.inputHasilUji)
	mux.HandleFunc("POST /v1/laporan-mutu", h.susunLaporanMutu)
	mux.HandleFunc("POST /v1/sertifikat", h.buatSertifikat)
	mux.HandleFunc("POST /v1/sertifikat/{permohonan_id}/tte", h.tandaTanganiTTE)
	mux.HandleFunc("POST /v1/draft-label", h.buatDraftLabel)
	mux.HandleFunc("POST /v1/draft-label/{id}/sim", h.integrasiSIM)
	mux.HandleFunc("POST /v1/draft-label/{id}/cetak", h.cetakLabel)
	mux.HandleFunc("POST /v1/draft-label/{id}/legalisasi", h.legalisasiLabel)
	mux.HandleFunc("POST /v1/penyerahan", h.terimaDokumenPembayaranAkhir)
	mux.HandleFunc("POST /v1/proses-mutu", h.prosesHasilMutu)
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

type ambilSampelRequest struct {
	PermohonanID       int64  `json:"permohonan_id"`
	JumlahSampel       int    `json:"jumlah_sampel"`
	TanggalPengambilan string `json:"tanggal_pengambilan"`
}

type kodifikasiSampelRequest struct {
	KodeSampel string `json:"kode_sampel"`
}

type inputHasilUjiRequest struct {
	SampelID    int64  `json:"sampel_id"`
	Parameter   string `json:"parameter"`
	Hasil       string `json:"hasil"`
	StatusLulus bool   `json:"status_lulus"`
	Catatan     string `json:"catatan"`
	TanggalUji  string `json:"tanggal_uji"`
}

type susunLaporanMutuRequest struct {
	SampelID   int64  `json:"sampel_id"`
	Kesimpulan string `json:"kesimpulan"`
	IsiLaporan string `json:"isi_laporan"`
}

type buatSertifikatRequest struct {
	PermohonanID    int64  `json:"permohonan_id"`
	NomorSertifikat string `json:"nomor_sertifikat"`
}

type tandaTanganiTTERequest struct {
	TTD       string `json:"ttd"`
	DokumenURL string `json:"dokumen_url"`
}

type buatDraftLabelRequest struct {
	SertifikatID int64  `json:"sertifikat_id"`
	KodeBarcode  string `json:"kode_barcode"`
}

type legalisasiLabelRequest struct {
	Status  string `json:"status"`
	Catatan string `json:"catatan"`
}

type terimaDokumenRequest struct {
	SertifikatDikirim     bool   `json:"sertifikat_dikirim"`
	LabelDikirim          bool   `json:"label_dikirim"`
	PembayaranAkhirStatus string `json:"pembayaran_akhir_status"`
	DiterimaOleh          string `json:"diterima_oleh"`
}

type prosesMutuRequest struct {
	SampelID int64 `json:"sampel_id"`
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

func (h *certificationHandler) ambilSampel(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req ambilSampelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	samp, err := h.svc.AmbilSampel(r.Context(), req.PermohonanID, userID, role, req.JumlahSampel, req.TanggalPengambilan)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, samp)
}

func (h *certificationHandler) kodifikasiSampel(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id sampel tidak valid")
		return
	}
	var req kodifikasiSampelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	if err := h.svc.KodifikasiSampel(r.Context(), id, req.KodeSampel, userID, role); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "sampel dikodekan"})
}

func (h *certificationHandler) inputHasilUji(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req inputHasilUjiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	uji, err := h.svc.InputHasilUji(r.Context(), req.SampelID, req.Parameter, req.Hasil, req.StatusLulus, req.Catatan, userID, role, req.TanggalUji)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, uji)
}

func (h *certificationHandler) susunLaporanMutu(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req susunLaporanMutuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	laporan, err := h.svc.SusunLaporanMutu(r.Context(), req.SampelID, req.Kesimpulan, req.IsiLaporan, userID, role)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, laporan)
}

func (h *certificationHandler) buatSertifikat(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req buatSertifikatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	srt, err := h.svc.BuatSertifikat(r.Context(), req.PermohonanID, userID, role, req.NomorSertifikat)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, srt)
}

func (h *certificationHandler) tandaTanganiTTE(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	permohonanIDStr := r.PathValue("permohonan_id")
	permohonanID, err := strconv.ParseInt(permohonanIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id permohonan tidak valid")
		return
	}
	var req tandaTanganiTTERequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	ttd, err := time.Parse(time.RFC3339, req.TTD)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ttd harus format RFC3339")
		return
	}
	if err := h.svc.TandaTanganiTTE(r.Context(), permohonanID, userID, role, ttd, req.DokumenURL); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "sertifikat ditandatangani TTE"})
}

func (h *certificationHandler) buatDraftLabel(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req buatDraftLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	lbl, err := h.svc.BuatDraftLabel(r.Context(), req.SertifikatID, req.KodeBarcode, userID, role)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, lbl)
}

func (h *certificationHandler) integrasiSIM(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id draft label tidak valid")
		return
	}
	if err := h.svc.IntegrasiSIM(r.Context(), id, userID, role); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "label terintegrasi SIM Perbenihan"})
}

func (h *certificationHandler) cetakLabel(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id draft label tidak valid")
		return
	}
	if err := h.svc.CetakLabel(r.Context(), id, userID, role); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "label dicetak"})
}

func (h *certificationHandler) legalisasiLabel(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "id draft label tidak valid")
		return
	}
	var req legalisasiLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	if err := h.svc.LegalisasiLabel(r.Context(), id, userID, role, req.Status, req.Catatan); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "label " + req.Status})
}

func (h *certificationHandler) terimaDokumenPembayaranAkhir(w http.ResponseWriter, r *http.Request) {
	_, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req terimaDokumenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	if err := h.svc.TerimaDokumenPembayaranAkhir(r.Context(), 0, req.DiterimaOleh, role, req.SertifikatDikirim, req.LabelDikirim, req.PembayaranAkhirStatus); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "dokumen diserahkan dan pembayaran akhir dicatat"})
}

func (h *certificationHandler) prosesHasilMutu(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasDariHeader(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req prosesMutuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
		return
	}
	hasil, err := h.svc.ProsesHasilMutu(r.Context(), req.SampelID, userID, role)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"hasil": hasil})
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
