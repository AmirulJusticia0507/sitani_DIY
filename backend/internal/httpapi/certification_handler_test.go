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
	sampel     map[int64]*db.SampelBenih
	nextID     int64
}

func newPalsuStore() *palsuStore {
	permohonan := map[int64]*db.PermohonanBenih{
		1: {ID: 1, NomorPermohonan: "SPB-1-1", PelangganID: 10, Status: "diinspeksi", PembayaranAwalStatus: "lunas"},
	}
	return &palsuStore{
		permohonan: permohonan,
		sampel:     map[int64]*db.SampelBenih{},
		nextID:     1,
	}
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
func (p *palsuStore) CreateSampel(ctx context.Context, s db.SampelBenih) (int64, error) {
	id := p.nextID
	p.nextID++
	s.ID = id
	p.sampel[id] = &s
	return id, nil
}
func (p *palsuStore) GetSampel(ctx context.Context, id int64) (*db.SampelBenih, error) {
	if x, ok := p.sampel[id]; ok {
		return x, nil
	}
	return nil, nil
}
func (p *palsuStore) KodifikasiSampel(ctx context.Context, id int64, kodeSampel string, analisID int64) error {
	if s, ok := p.sampel[id]; ok {
		s.KodeSampel = kodeSampel
		s.AnalisID = &analisID
		s.Status = "dikodekan"
		return nil
	}
	return nil
}
func (p *palsuStore) CreateUjiMutu(ctx context.Context, u db.UjiMutu) (int64, error)   { return 1, nil }
func (p *palsuStore) ListUjiMutu(ctx context.Context, sampelID int64) ([]db.UjiMutu, error) {
	return []db.UjiMutu{{SampelID: sampelID, Parameter: "keutuhan", StatusLulus: true}}, nil
}
func (p *palsuStore) CreateLaporanMutu(ctx context.Context, l db.LaporanMutu) (int64, error) { return 1, nil }
func (p *palsuStore) GetLaporanMutu(ctx context.Context, sampelID int64) (*db.LaporanMutu, error) {
	return nil, nil
}

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

func TestAmbilSampelSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"permohonan_id":1,"jumlah_sampel":2,"tanggal_pengambilan":"2026-10-23"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/sampel", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "20")
	r.Header.Set("X-User-Role", "pengawas")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestKodifikasiSampelSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	// Buat sampel dulu melalui handler agar status tercatat di palsu store.
	body := `{"permohonan_id":1,"jumlah_sampel":2,"tanggal_pengambilan":"2026-10-23"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/sampel", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "20")
	r.Header.Set("X-User-Role", "pengawas")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)

	kode := `{"kode_sampel":"KD-001"}`
	r2 := httptest.NewRequest(http.MethodPost, "/v1/sampel/1/kodifikasi", strings.NewReader(kode))
	r2.Header.Set("Content-Type", "application/json")
	r2.Header.Set("X-User-ID", "30")
	r2.Header.Set("X-User-Role", "analis")
	w2 := httptest.NewRecorder()
	h.mux.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w2.Code, http.StatusOK, w2.Body.String())
	}
}

func TestInputHasilUjiSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"permohonan_id":1,"jumlah_sampel":2,"tanggal_pengambilan":"2026-10-23"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/sampel", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "20")
	r.Header.Set("X-User-Role", "pengawas")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)

	kode := `{"kode_sampel":"KD-001"}`
	r2 := httptest.NewRequest(http.MethodPost, "/v1/sampel/1/kodifikasi", strings.NewReader(kode))
	r2.Header.Set("Content-Type", "application/json")
	r2.Header.Set("X-User-ID", "30")
	r2.Header.Set("X-User-Role", "analis")
	w2 := httptest.NewRecorder()
	h.mux.ServeHTTP(w2, r2)

	uji := `{"sampel_id":1,"parameter":"keutuhan","hasil":"98%","status_lulus":true,"catatan":"baik","tanggal_uji":"2026-10-24"}`
	r3 := httptest.NewRequest(http.MethodPost, "/v1/uji-mutu", strings.NewReader(uji))
	r3.Header.Set("Content-Type", "application/json")
	r3.Header.Set("X-User-ID", "30")
	r3.Header.Set("X-User-Role", "analis")
	w3 := httptest.NewRecorder()
	h.mux.ServeHTTP(w3, r3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w3.Code, http.StatusCreated, w3.Body.String())
	}
}

func TestSusunLaporanMutuSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"permohonan_id":1,"jumlah_sampel":2,"tanggal_pengambilan":"2026-10-23"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/sampel", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "20")
	r.Header.Set("X-User-Role", "pengawas")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)

	kode := `{"kode_sampel":"KD-001"}`
	r2 := httptest.NewRequest(http.MethodPost, "/v1/sampel/1/kodifikasi", strings.NewReader(kode))
	r2.Header.Set("Content-Type", "application/json")
	r2.Header.Set("X-User-ID", "30")
	r2.Header.Set("X-User-Role", "analis")
	w2 := httptest.NewRecorder()
	h.mux.ServeHTTP(w2, r2)

	lapor := `{"sampel_id":1,"kesimpulan":"lulus","isi_laporan":"Sampel memenuhi standar mutu benih.","disusun_oleh":30}`
	r3 := httptest.NewRequest(http.MethodPost, "/v1/laporan-mutu", strings.NewReader(lapor))
	r3.Header.Set("Content-Type", "application/json")
	r3.Header.Set("X-User-ID", "30")
	r3.Header.Set("X-User-Role", "analis")
	w3 := httptest.NewRecorder()
	h.mux.ServeHTTP(w3, r3)
	if w3.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w3.Code, http.StatusCreated, w3.Body.String())
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
