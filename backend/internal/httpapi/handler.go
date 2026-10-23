package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/storage"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	DB      string `json:"db,omitempty"`
	Storage string `json:"storage,omitempty"`
}

// Handler menampung dependensi adaptor (database, service, integrasi eksternal)
// yang akan digunakan oleh semua handler HTTP.
type Handler struct {
	mux         *http.ServeMux
	db          *db.DB
	svc         *service.Service
	auth        *auth.Manager
	dokumen     *dokumenHandler
	storageSiap bool
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
	h.auth = auth.NewManager(auth.ConfigFromEnv(), database)
	h.mux.HandleFunc("GET /health", h.health)
	h.auth.RegisterRoutes(h.mux)
	newCertificationHandler(h.svc).register(h.mux)
	newRegistryHandler(h.svc).register(h.mux)

	// Object storage dokumen aktif saat MINIO_* terkonfigurasi; jika koneksi
	// gagal, endpoint dokumen tetap terdaftar tetapi menjawab 503.
	conf := storage.ConfigFromEnv()
	if conf.Siap() {
		klien, err := storage.NewMinioClient(context.Background(), conf)
		if err != nil {
			os.Stderr.WriteString("object storage tidak tersedia: " + err.Error() + "\n")
			h.dokumen = newDokumenHandler(st, nil, conf)
		} else {
			h.dokumen = newDokumenHandler(st, klien, conf)
			h.storageSiap = true
		}
	} else {
		h.dokumen = newDokumenHandler(st, nil, conf)
	}
	h.dokumen.register(h.mux)
	return h
}

// ServeMux mengembalikan http.Handler untuk kompatibilitas dengan http.Server.
func (h *Handler) ServeMux() http.Handler {
	return cors(h.auth.ProtectAPI(h.mux))
}

func cors(next http.Handler) http.Handler {
	allowedOrigin := strings.TrimRight(os.Getenv("FRONTEND_PUBLIC_URL"), "/")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		if allowedOrigin != "" && origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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
	if h.dokumen != nil {
		switch penyErr := h.dokumen.periksaPenyimpanan(context.Background()); {
		case penyErr == nil:
			resp.Storage = "ok"
		case errors.Is(penyErr, storage.ErrStorageTidakSiap):
			resp.Storage = "belum-dikonfigurasi"
		default:
			resp.Status = "degraded"
			resp.Storage = "error"
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "gagal menulis respons", http.StatusInternalServerError)
	}
}
