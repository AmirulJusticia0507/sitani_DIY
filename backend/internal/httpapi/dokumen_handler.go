package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/storage"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// dokumenHandler melayani unggah, daftar, dan unduh lampiran yang fisiknya
// disimpan privat di object storage. Unduhan hanya diberikan setelah
// pemeriksaan hak akses: pemilik dokumen atau petugas internal.
type dokumenHandler struct {
	st          store.Store
	penyimpanan storage.Client
	ukuranMaks  int64
}

func newDokumenHandler(st store.Store, penyimpanan storage.Client, conf storage.Config) *dokumenHandler {
	ukuran := conf.UkuranMaks
	if ukuran <= 0 {
		ukuran = 20 << 20
	}
	return &dokumenHandler{st: st, penyimpanan: penyimpanan, ukuranMaks: ukuran}
}

func (h *dokumenHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/dokumen", h.unggah)
	mux.HandleFunc("GET /v1/dokumen", h.daftar)
	mux.HandleFunc("GET /v1/dokumen/{id}/unduh", h.unduh)
}

// internalRole adalah peran petugas yang boleh melihat semua dokumen dan
// mengunduh dokumen pelanggan untuk kebutuhan verifikasi/penilaian.
func internalRole(role string) bool {
	switch role {
	case db.RoleAdmin, db.RolePengawas, db.RoleAnalis, db.RolePenandatangan, db.RoleLegalisasi, db.RolePercetakan:
		return true
	default:
		return false
	}
}

// identitasPeminta mengambil principal dari context SSO, dengan fallback
// identitasDariHeader untuk kompatibilitas handler lama dalam tes.
func identitasPeminta(r *http.Request) (int64, string, error) {
	if p, ok := auth.PrincipalFromContext(r.Context()); ok {
		return p.UserID, p.Role, nil
	}
	return identitasDariHeader(r)
}

func (h *dokumenHandler) unggah(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasPeminta(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if h.penyimpanan == nil {
		respondError(w, http.StatusServiceUnavailable, storage.ErrStorageTidakSiap.Error())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.ukuranMaks+(1<<20))
	if err := r.ParseMultipartForm(h.ukuranMaks + (1 << 20)); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge,
			"berkas atau formulir melebihi batas ukuran yang diizinkan")
		return
	}
	defer func() {
		_ = r.MultipartForm.RemoveAll()
	}()

	berkas, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "field \"file\" wajib diisi sebagai multipart")
		return
	}
	defer berkas.Close()

	tujuan := strings.TrimSpace(r.FormValue("tujuan"))
	if len([]rune(tujuan)) > 255 {
		tujuan = string([]rune(tujuan)[:255])
	}

	rencana, err := storage.Siapkan(h.ukuranMaks, userID, berkas, header)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.penyimpanan.Unggah(r.Context(), rencana.ObjectKey, rencana.Masukan, rencana.Ukuran, rencana.Jenis); err != nil {
		respondError(w, http.StatusBadGateway, "gagal menyimpan berkas ke object storage")
		return
	}

	dok := db.Dokumen{
		PemilikID: userID,
		Tujuan:    tujuan,
		ObjectKey: rencana.ObjectKey,
		NamaAsli:  rencana.NamaAsli,
		Jenis:     rencana.Jenis,
		Ukuran:    rencana.Ukuran,
	}
	id, err := h.st.CreateDokumen(r.Context(), dok)
	if err != nil {
		// Berkas sudah tersimpan namun metadata gagal: beri tahu kegagalan
		// supaya tidak ada berkas "tanpa catatan" yang tak terlacak.
		respondError(w, http.StatusInternalServerError, "gagal mencatat metadata dokumen")
		return
	}
	_, _ = h.st.CreateAuditLog(r.Context(), db.AuditLog{
		ActorID:   &userID,
		ActorRole: role,
		Action:    "dokumen:unggah",
		Entity:    "dokumen",
		EntityID:  &id,
		Detail:    "unggah " + dok.NamaAsli,
	})

	respondJSON(w, http.StatusCreated, map[string]any{
		"id":         id,
		"nama":       dok.NamaAsli,
		"jenis":      dok.Jenis,
		"ukuran":     dok.Ukuran,
		"tujuan":     dok.Tujuan,
		"object_key": dok.ObjectKey,
	})
}

func (h *dokumenHandler) daftar(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasPeminta(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if h.st == nil {
		respondError(w, http.StatusServiceUnavailable, "database belum tersedia")
		return
	}
	daftar, err := h.st.ListDokumen(r.Context(), userID, internalRole(role))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "gagal membaca daftar dokumen")
		return
	}
	if daftar == nil {
		daftar = []db.Dokumen{}
	}
	respondJSON(w, http.StatusOK, map[string]any{"dokumen": daftar})
}

func (h *dokumenHandler) unduh(w http.ResponseWriter, r *http.Request) {
	userID, role, err := identitasPeminta(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if h.penyimpanan == nil {
		respondError(w, http.StatusServiceUnavailable, storage.ErrStorageTidakSiap.Error())
		return
	}
	id, ok := parsePathID(w, r, "id", "ID dokumen tidak valid")
	if !ok {
		return
	}
	dok, err := h.st.GetDokumen(r.Context(), id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		respondError(w, http.StatusNotFound, "dokumen tidak ditemukan")
		return
	case err != nil:
		respondError(w, http.StatusInternalServerError, "gagal membaca metadata dokumen")
		return
	}
	if !internalRole(role) && dok.PemilikID != userID {
		respondError(w, http.StatusForbidden, "dokumen bukan milik Anda")
		return
	}
	urlUnduh, err := h.penyimpanan.URLUnduh(r.Context(), dok.ObjectKey)
	if err != nil {
		respondError(w, http.StatusBadGateway, "gagal menerbitkan tautan unduhan")
		return
	}
	_, _ = h.st.CreateAuditLog(r.Context(), db.AuditLog{
		ActorID:   &userID,
		ActorRole: role,
		Action:    "dokumen:unduh",
		Entity:    "dokumen",
		EntityID:  &dok.ID,
		Detail:    "terbitkan URL unduh " + dok.NamaAsli,
	})
	respondJSON(w, http.StatusOK, map[string]any{
		"id":    dok.ID,
		"nama":  dok.NamaAsli,
		"jenis": dok.Jenis,
		"url":   urlUnduh,
	})
}

// periksaPenyimpanan dipakai health untuk melaporkan status bucket.
func (h *dokumenHandler) periksaPenyimpanan(ctx context.Context) error {
	if h.penyimpanan == nil {
		return storage.ErrStorageTidakSiap
	}
	return h.penyimpanan.Periksa(ctx)
}
