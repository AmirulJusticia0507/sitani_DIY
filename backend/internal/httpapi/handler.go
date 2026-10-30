package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/ai"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth/devlogin"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/dataset"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/horticulture"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/notify"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/observability"
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
	mux          *http.ServeMux
	db           *db.DB
	svc          *service.Service
	auth         *auth.Manager
	dokumen      *dokumenHandler
	dataset      *datasetHandler
	horticulture *horticultureHandler
	ai           *aiHandler
	storageSiap  bool
	notify       *notify.Client
	obs          *observability.Client
}

// SetNotify mengatur klien notifikasi Discord.
func (h *Handler) SetNotify(c *notify.Client) { h.notify = c }

// SetObservability mengatur klien pelaporan error.
func (h *Handler) SetObservability(c *observability.Client) { h.obs = c }

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
	devConfig := devlogin.ConfigFromEnv()
	if devConfig.Aktif {
		os.Stderr.WriteString("PERINGATAN: login lokal development AKTIF (/auth/dev/login). " +
			"Jangan aktifkan di lingkungan produksi.\n")
	}
	devManager := devlogin.NewManager(devConfig, database)
	h.auth.SetDevLocal(devManager)
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
	var datasetRepo dataset.Repository
	if database != nil {
		datasetRepo = dataset.NewPostgresRepository(database)
	}
	h.dataset = &datasetHandler{repo: datasetRepo}
	h.dataset.register(h.mux)
	var horticultureRepo horticulture.Repository
	if database != nil {
		horticultureRepo = horticulture.NewPostgresRepository(database)
	}
	h.horticulture = &horticultureHandler{repo: horticultureRepo}
	h.horticulture.register(h.mux)
	h.ai = newAIHandler()
	h.ai.register(h.mux)
	return h
}

// newAIHandler tidak menggagalkan start saat AI_* belum diisi; endpoint tetap
// terdaftar dan menjawab 503 sampai konfigurasi tersedia. Pesan peringatan
// tidak pernah memuat API key.
func newAIHandler() *aiHandler {
	client, err := ai.New(ai.ConfigFromEnv())
	if err != nil {
		os.Stderr.WriteString("asisten AI belum aktif: " + err.Error() + "\n")
		return &aiHandler{limiter: newAILimiterDariEnv()}
	}
	return &aiHandler{client: client, limiter: newAILimiterDariEnv()}
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
			// X-User-Role masih dikirim halaman sertifikasi fase 1; tanpa izin di
			// preflight, browser menolak request sebelum backend memverifikasi
			// sesi. Identitas tetap diambil dari sesi — header X-User-ID dari
			// klien selalu dibuang oleh middleware Require.
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-User-Role")
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
