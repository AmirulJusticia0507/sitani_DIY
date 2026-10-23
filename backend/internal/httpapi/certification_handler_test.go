package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
)

// palsuStore menyimpan data di memori untuk tes handler tanpa MySQL.
type palsuStore struct {
	permohonan map[int64]*db.PermohonanBenih
	nextID     int64
}

func newPalsuStore() *palsuStore {
	return &palsuStore{permohonan: map[int64]*db.PermohonanBenih{}, nextID: 1}
}

func (p *palsuStore) CreateUser(ctx context.Context, u db.User) (int64, error)                                   { return 0, nil }
func (p *palsuStore) GetUserByID(ctx context.Context, id int64) (*db.User, error)                              { return nil, nil }
func (p *palsuStore) GetUserByUsername(ctx context.Context, username string) (*db.User, error)                 { return nil, nil }
func (p *palsuStore) CreatePermohonan(ctx context.Context, perm db.PermohonanBenih) (int64, error) {
	id := p.nextID
	p.nextID++
	perm.ID = id
	p.permohonan[id] = &perm
	return id, nil
}
func (p *palsuStore) GetPermohonan(ctx context.Context, id int64) (*db.PermohonanBenih, error) {
	if x, ok := p.permohonan[id]; ok {
		return x, nil
	}
	return nil, nil
}
func (p *palsuStore) UpdatePermohonanStatus(ctx context.Context, id int64, status, pembayaranAwalStatus string) error {
	if x, ok := p.permohonan[id]; ok {
		x.Status = status
		x.PembayaranAwalStatus = pembayaranAwalStatus
		return nil
	}
	return nil
}
func (p *palsuStore) CreatePembayaran(ctx context.Context, pb db.Pembayaran) (int64, error)       { return 1, nil }
func (p *palsuStore) CreateInspeksiFase(ctx context.Context, i db.InspeksiFase) (int64, error)    { return 1, nil }
func (p *palsuStore) ListInspeksiFase(ctx context.Context, permohonanID int64) ([]db.InspeksiFase, error) {
	return nil, nil
}
func (p *palsuStore) CreateBukuInduk(ctx context.Context, b db.BukuInduk) (int64, error)   { return 1, nil }
func (p *palsuStore) GetBukuInduk(ctx context.Context, permohonanID int64) (*db.BukuInduk, error) { return nil, nil }
func (p *palsuStore) CreateAuditLog(ctx context.Context, a db.AuditLog) (int64, error)             { return 1, nil }
func (p *palsuStore) Ping(ctx context.Context) error                                              { return nil }

var _ serviceStore = (*palsuStore)(nil)

type serviceStore interface {
	CreatePermohonan(context.Context, db.PermohonanBenih) (int64, error)
	GetPermohonan(context.Context, int64) (*db.PermohonanBenih, error)
	UpdatePermohonanStatus(context.Context, int64, string, string) error
}

func newHandlerWithPalsu() *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: service.NewService(newPalsuStore())}
	newCertificationHandler(h.svc).register(h.mux)
	return h
}

func TestBuatPermohonanSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"jenis_benih":"benih tomat","varietas":"rama","luas_lahan":0.5,"lokasi_kebun":"Sleman"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/permohonan", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "10")
	r.Header.Set("X-User-Role", "pelanggan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["NomorPermohonan"] == nil {
		t.Fatalf("NomorPermohonan tidak ada dalam respons: %s", w.Body.String())
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
}

func TestBuatPermohonanRoleBukanPelangganDitolak(t *testing.T) {
	h := newHandlerWithPalsu()
	r := httptest.NewRequest(http.MethodPost, "/v1/permohonan", nil)
	r.Header.Set("X-User-ID", "10")
	r.Header.Set("X-User-Role", "pengawas")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, ingin %d", w.Code, http.StatusForbidden)
	}
}

func TestHealthSukses(t *testing.T) {
	h := NewHandler(nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d", w.Code, http.StatusOK)
	}
}
