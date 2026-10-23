package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	DB      string `json:"db,omitempty"`
}

// Handler menampung dependensi adaptor (database, integrasi eksternal) yang
// akan digunakan oleh semua handler HTTP.
type Handler struct {
	mux *http.ServeMux
	db  *db.DB
}

// NewHandler membuat handler HTTP. db boleh nil saat database belum dikonfigurasi.
func NewHandler(database *db.DB) *Handler {
	h := &Handler{
		mux: http.NewServeMux(),
		db:  database,
	}
	h.mux.HandleFunc("GET /health", h.health)
	return h
}

// ServeMux mengembalikan http.Handler untuk kompatibilitas dengan http.Server.
func (h *Handler) ServeMux() http.Handler {
	return h.mux
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
