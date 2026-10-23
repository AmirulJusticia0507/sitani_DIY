package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	DB      string `json:"db,omitempty"`
}

// Handler menampung dependensi adaptor (database, service, integrasi eksternal)
// yang akan digunakan oleh semua handler HTTP.
type Handler struct {
	mux *http.ServeMux
	db  *db.DB
	svc *service.Service
}

// NewHandler membuat handler HTTP dan mendaftarkan semua rute. db boleh nil
// saat database belum dikonfigurasi; service akan mengembalikan kesalahan
// saat store tidak tersedia.
func NewHandler(database *db.DB) *Handler {
	h := &Handler{
		mux: http.NewServeMux(),
		db:  database,
	}
	var st store.Store
	if database != nil {
		st = store.NewDBStore(database)
	}
	h.svc = service.NewService(st)
	h.mux.HandleFunc("GET /health", h.health)
	newCertificationHandler(h.svc).register(h.mux)
	return h
}

// ServeMux mengembalikan http.Handler untuk kompatibilitas dengan http.Server.
func (h *Handler) ServeMux() http.Handler {
	return h.mux
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	resp := healthResponse{Status: "ok", Service: "sitani-api"}
	if h.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := h.db.Health(ctx); err != nil {
			resp.Status = "degraded"
			resp.DB = err.Error()
		} else {
			resp.DB = "ok"
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "gagal menulis respons", http.StatusInternalServerError)
	}
}
