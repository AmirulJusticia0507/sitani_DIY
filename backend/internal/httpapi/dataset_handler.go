package httpapi

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/dataset"
)

type datasetHandler struct{ repo dataset.Repository }

func (h *datasetHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/dataset/import", h.importData)
	mux.HandleFunc("GET /v1/dataset/import", h.listImports)
}

func datasetRole(role string) bool { return role == "admin" || role == "analis" }

func (h *datasetHandler) importData(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID <= 0 {
		respondError(w, http.StatusUnauthorized, "identitas pengguna tidak tersedia")
		return
	}
	if !datasetRole(principal.Role) {
		respondError(w, http.StatusForbidden, "peran tidak diizinkan mengimpor dataset")
		return
	}
	if h.repo == nil {
		respondError(w, http.StatusServiceUnavailable, "repository dataset belum tersedia")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, dataset.MaksBytes+(1<<20))
	if err := r.ParseMultipartForm(dataset.MaksBytes + (1 << 20)); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "permintaan melebihi batas ukuran")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	namaDataset := strings.TrimSpace(r.FormValue("nama_dataset"))
	if namaDataset == "" || utf8.RuneCountInString(namaDataset) > 120 {
		respondError(w, http.StatusBadRequest, "nama_dataset wajib diisi (maks. 120 karakter)")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "field file wajib diisi")
		return
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, dataset.MaksBytes+1))
	if err != nil {
		respondError(w, http.StatusBadRequest, "gagal membaca file import")
		return
	}
	if len(raw) > dataset.MaksBytes {
		respondError(w, http.StatusRequestEntityTooLarge, "file import melebihi batas 10 MiB")
		return
	}
	namaFile := filepath.Base(header.Filename)
	parsed, err := dataset.Parse(namaFile, raw)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, dataset.ErrUnsupported) {
			status = http.StatusUnsupportedMediaType
		}
		respondError(w, status, err.Error())
		return
	}
	item, err := h.repo.Simpan(r.Context(), principal.UserID, namaDataset, namaFile, parsed)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "gagal menyimpan batch import")
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *datasetHandler) listImports(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID <= 0 {
		respondError(w, http.StatusUnauthorized, "identitas pengguna tidak tersedia")
		return
	}
	if !datasetRole(principal.Role) {
		respondError(w, http.StatusForbidden, "peran tidak diizinkan melihat batch import")
		return
	}
	if h.repo == nil {
		respondError(w, http.StatusServiceUnavailable, "repository dataset belum tersedia")
		return
	}
	items, err := h.repo.Daftar(r.Context(), principal.UserID, principal.Role == "admin")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "gagal membaca batch import")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"imports": items})
}
