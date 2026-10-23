package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/store"
)

// palsuStore menyimpan data di memori untuk tes handler tanpa MySQL.
type palsuStore struct {
	permohonan   map[int64]*db.PermohonanBenih
	sampel       map[int64]*db.SampelBenih
	registrasi   map[int64]*db.RegistrasiBenih
	stok         map[int64]*db.StokBenih
	mutasi       map[int64][]db.MutasiStokBenih
	penjualan    map[int64]*db.PenjualanBenih
	dokumen      map[int64]*db.Dokumen
	nextID       int64
	nextRegID    int64
	nextSaleID   int64
	dokumenValid bool
}

func newPalsuStore() *palsuStore {
	permohonan := map[int64]*db.PermohonanBenih{
		1: {ID: 1, NomorPermohonan: "SPB-1-1", PelangganID: 10, Status: "diinspeksi", PembayaranAwalStatus: "lunas"},
	}
	return &palsuStore{
		permohonan:   permohonan,
		sampel:       map[int64]*db.SampelBenih{},
		registrasi:   map[int64]*db.RegistrasiBenih{},
		stok:         map[int64]*db.StokBenih{},
		mutasi:       map[int64][]db.MutasiStokBenih{},
		penjualan:    map[int64]*db.PenjualanBenih{},
		dokumen:      map[int64]*db.Dokumen{},
		nextID:       1,
		nextRegID:    1,
		nextSaleID:   1,
		dokumenValid: true,
	}
}

// CreateDokumen menyimpan metadata dokumen di memori.
func (p *palsuStore) CreateDokumen(ctx context.Context, d db.Dokumen) (int64, error) {
	id := p.nextID
	p.nextID++
	d.ID = id
	p.dokumen[id] = &d
	return id, nil
}

// GetDokumen mengembalikan metadata dokumen atau ErrNotFound.
func (p *palsuStore) GetDokumen(ctx context.Context, id int64) (*db.Dokumen, error) {
	if d, ok := p.dokumen[id]; ok {
		return d, nil
	}
	return nil, store.ErrNotFound
}

// ListDokumen mengembalikan dokumen milik satu pengguna atau seluruhnya.
func (p *palsuStore) ListDokumen(ctx context.Context, userID int64, semua bool) ([]db.Dokumen, error) {
	var daftar []db.Dokumen
	for _, d := range p.dokumen {
		if semua || d.PemilikID == userID {
			daftar = append(daftar, *d)
		}
	}
	for i := 0; i < len(daftar)-1; i++ {
		for j := i + 1; j < len(daftar); j++ {
			if daftar[i].ID < daftar[j].ID {
				daftar[i], daftar[j] = daftar[j], daftar[i]
			}
		}
	}
	return daftar, nil
}

func (p *palsuStore) CreateUser(ctx context.Context, u db.User) (int64, error)    { return 0, nil }
func (p *palsuStore) GetUserByID(ctx context.Context, id int64) (*db.User, error) { return nil, nil }
func (p *palsuStore) GetUserByUsername(ctx context.Context, username string) (*db.User, error) {
	return nil, nil
}
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
func (p *palsuStore) CreatePembayaran(ctx context.Context, pb db.Pembayaran) (int64, error) {
	return 1, nil
}
func (p *palsuStore) CreateInspeksiFase(ctx context.Context, i db.InspeksiFase) (int64, error) {
	return 1, nil
}
func (p *palsuStore) ListInspeksiFase(ctx context.Context, permohonanID int64) ([]db.InspeksiFase, error) {
	return nil, nil
}
func (p *palsuStore) CreateBukuInduk(ctx context.Context, b db.BukuInduk) (int64, error) {
	return 1, nil
}
func (p *palsuStore) GetBukuInduk(ctx context.Context, permohonanID int64) (*db.BukuInduk, error) {
	return nil, nil
}
func (p *palsuStore) CreateAuditLog(ctx context.Context, a db.AuditLog) (int64, error) { return 1, nil }
func (p *palsuStore) Ping(ctx context.Context) error                                   { return nil }
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
func (p *palsuStore) CreateUjiMutu(ctx context.Context, u db.UjiMutu) (int64, error) { return 1, nil }
func (p *palsuStore) ListUjiMutu(ctx context.Context, sampelID int64) ([]db.UjiMutu, error) {
	return []db.UjiMutu{{SampelID: sampelID, Parameter: "keutuhan", StatusLulus: true}}, nil
}
func (p *palsuStore) CreateLaporanMutu(ctx context.Context, l db.LaporanMutu) (int64, error) {
	return 1, nil
}
func (p *palsuStore) GetLaporanMutu(ctx context.Context, sampelID int64) (*db.LaporanMutu, error) {
	kesimpulan := "lulus"
	if sampelID == 2 {
		kesimpulan = "tidak_lulus"
	}
	return &db.LaporanMutu{ID: 1, SampelID: sampelID, Kesimpulan: kesimpulan}, nil
}
func (p *palsuStore) CreateSertifikat(ctx context.Context, s db.SertifikatBenih) (int64, error) {
	id := p.nextID
	p.nextID++
	s.ID = id
	return id, nil
}
func (p *palsuStore) GetSertifikat(ctx context.Context, permohonanID int64) (*db.SertifikatBenih, error) {
	status := "belum"
	if p.dokumenValid {
		status = "ditandatangani"
	}
	return &db.SertifikatBenih{ID: 1, PermohonanID: permohonanID, NomorSertifikat: "SRT-001", TTEStatus: status}, nil
}
func (p *palsuStore) UpdateSertifikatTTE(ctx context.Context, permohonanID int64, tteStatus, tteDokumenURL string, ttd *time.Time) error {
	return nil
}
func (p *palsuStore) CreateDraftLabel(ctx context.Context, l db.DraftLabel) (int64, error) {
	id := p.nextID
	p.nextID++
	l.ID = id
	return id, nil
}
func (p *palsuStore) GetDraftLabel(ctx context.Context, sertifikatID int64) (*db.DraftLabel, error) {
	status := "draft"
	integrasi := "belum"
	if p.dokumenValid {
		status = "legalisasi"
		integrasi = "terkirim"
	}
	return &db.DraftLabel{ID: 1, SertifikatID: sertifikatID, KodeBarcode: "SB-2026-0001", Status: status, IntegrasiSIMStatus: integrasi}, nil
}
func (p *palsuStore) UpdateDraftLabel(ctx context.Context, id int64, status, integrasiSIM, kodeBarcode string) error {
	return nil
}
func (p *palsuStore) CreateLegalisasi(ctx context.Context, l db.LabelLegalisasi) (int64, error) {
	return 1, nil
}
func (p *palsuStore) UpdateLegalisasi(ctx context.Context, id int64, status string) error { return nil }
func (p *palsuStore) CreatePenyerahan(ctx context.Context, p2 db.PenyerahanDokumen) (int64, error) {
	return 1, nil
}
func (p *palsuStore) GetPenyerahan(ctx context.Context, permohonanID int64) (*db.PenyerahanDokumen, error) {
	return nil, nil
}
func (p *palsuStore) UpdatePenyerahan(ctx context.Context, permohonanID int64, sertifikatDikirim, labelDikirim bool, pembayaranAkhirStatus string) error {
	return nil
}

func (p *palsuStore) CreateRegistrasiBenih(ctx context.Context, r db.RegistrasiBenih) (int64, error) {
	id := p.nextRegID
	p.nextRegID++
	r.ID = id
	salinan := r
	p.registrasi[id] = &salinan
	p.stok[id] = &db.StokBenih{
		ID:                 id,
		RegistrasiID:       id,
		NomorLot:           r.NomorLot,
		Komoditas:          r.Komoditas,
		Varietas:           r.Varietas,
		JumlahTersedia:     r.Jumlah,
		Satuan:             r.Satuan,
		TanggalKedaluwarsa: r.TanggalKedaluwarsa,
		StatusEdar:         r.StatusEdar,
		UpdatedAt:          time.Now(),
	}
	refID := id
	p.mutasi[id] = append(p.mutasi[id], db.MutasiStokBenih{
		ID:            int64(len(p.mutasi[id]) + 1),
		RegistrasiID:  id,
		Jenis:         "masuk",
		Jumlah:        r.Jumlah,
		SaldoSetelah:  r.Jumlah,
		ReferensiTipe: "registrasi",
		ReferensiID:   &refID,
		Catatan:       "stok awal registrasi",
		CreatedBy:     r.CreatedBy,
		CreatedAt:     time.Now(),
	})
	return id, nil
}

func (p *palsuStore) GetRegistrasiBenih(ctx context.Context, id int64) (*db.RegistrasiBenih, error) {
	return p.registrasi[id], nil
}

func (p *palsuStore) GetRegistrasiBenihByNomorLot(ctx context.Context, nomorLot string) (*db.RegistrasiBenih, error) {
	for _, r := range p.registrasi {
		if r.NomorLot == nomorLot {
			return r, nil
		}
	}
	return nil, nil
}

func (p *palsuStore) ListRegistrasiBenih(ctx context.Context, filter db.RegistrasiFilter) ([]db.RegistrasiBenih, error) {
	var out []db.RegistrasiBenih
	for id := int64(1); id < p.nextRegID; id++ {
		r, ok := p.registrasi[id]
		if !ok {
			continue
		}
		if filter.StatusEdar != "" && r.StatusEdar != filter.StatusEdar {
			continue
		}
		if filter.Komoditas != "" && !strings.Contains(strings.ToLower(r.Komoditas), strings.ToLower(filter.Komoditas)) {
			continue
		}
		if filter.KedaluwarsaSebelum != nil && r.TanggalKedaluwarsa > filter.KedaluwarsaSebelum.Format("2006-01-02") {
			continue
		}
		out = append(out, *r)
	}
	return out, nil
}

func (p *palsuStore) UpdateRegistrasiStatus(ctx context.Context, id int64, status string) error {
	r, ok := p.registrasi[id]
	if !ok {
		return nil
	}
	r.StatusEdar = status
	if stok, ok := p.stok[id]; ok {
		stok.StatusEdar = status
	}
	return nil
}

func (p *palsuStore) GetStokBenih(ctx context.Context, registrasiID int64) (*db.StokBenih, error) {
	return p.stok[registrasiID], nil
}

func (p *palsuStore) ListStokBenih(ctx context.Context) ([]db.StokBenih, error) {
	var out []db.StokBenih
	for id := int64(1); id < p.nextRegID; id++ {
		if stok, ok := p.stok[id]; ok {
			out = append(out, *stok)
		}
	}
	return out, nil
}

func (p *palsuStore) ListMutasiStok(ctx context.Context, registrasiID int64) ([]db.MutasiStokBenih, error) {
	return p.mutasi[registrasiID], nil
}

func (p *palsuStore) CreatePenjualanBenih(ctx context.Context, penjualan db.PenjualanBenih) (*db.PenjualanBenih, error) {
	stok, ok := p.stok[penjualan.RegistrasiID]
	if !ok {
		return nil, store.ErrStokTidakCukup
	}
	if stok.JumlahTersedia < penjualan.Jumlah {
		return nil, store.ErrStokTidakCukup
	}
	id := p.nextSaleID
	p.nextSaleID++
	stok.JumlahTersedia -= penjualan.Jumlah
	penjualan.ID = id
	penjualan.Status = "selesai"
	salinan := penjualan
	p.penjualan[id] = &salinan
	refID := id
	p.mutasi[penjualan.RegistrasiID] = append(p.mutasi[penjualan.RegistrasiID], db.MutasiStokBenih{
		ID:            int64(len(p.mutasi[penjualan.RegistrasiID]) + 1),
		RegistrasiID:  penjualan.RegistrasiID,
		Jenis:         "keluar",
		Jumlah:        penjualan.Jumlah,
		SaldoSetelah:  stok.JumlahTersedia,
		ReferensiTipe: "penjualan",
		ReferensiID:   &refID,
		Catatan:       "penjualan " + penjualan.NomorDokumen,
		CreatedBy:     penjualan.CreatedBy,
		CreatedAt:     time.Now(),
	})
	return &salinan, nil
}

func (p *palsuStore) GetPenjualanBenih(ctx context.Context, id int64) (*db.PenjualanBenih, error) {
	return p.penjualan[id], nil
}

func (p *palsuStore) BatalkanPenjualanBenih(ctx context.Context, id int64, actorID int64, alasan string, waktu time.Time) (*db.PenjualanBenih, error) {
	penjualan, ok := p.penjualan[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	if penjualan.Status == "dibatalkan" {
		return nil, store.ErrSudahDibatalkan
	}
	stok := p.stok[penjualan.RegistrasiID]
	stok.JumlahTersedia += penjualan.Jumlah
	penjualan.Status = "dibatalkan"
	penjualan.AlasanPembatalan = alasan
	penjualan.DibatalkanOleh = &actorID
	penjualan.DibatalkanPada = &waktu
	refID := id
	p.mutasi[penjualan.RegistrasiID] = append(p.mutasi[penjualan.RegistrasiID], db.MutasiStokBenih{
		ID:            int64(len(p.mutasi[penjualan.RegistrasiID]) + 1),
		RegistrasiID:  penjualan.RegistrasiID,
		Jenis:         "koreksi",
		Jumlah:        penjualan.Jumlah,
		SaldoSetelah:  stok.JumlahTersedia,
		ReferensiTipe: "pembatalan_penjualan",
		ReferensiID:   &refID,
		Catatan:       alasan,
		CreatedBy:     actorID,
		CreatedAt:     waktu,
	})
	return penjualan, nil
}

func (p *palsuStore) ListPenjualanBenih(ctx context.Context, filter db.PenjualanFilter) ([]db.PenjualanBenih, error) {
	var out []db.PenjualanBenih
	for id := int64(1); id < p.nextSaleID; id++ {
		penjualan, ok := p.penjualan[id]
		if !ok {
			continue
		}
		if filter.RegistrasiID > 0 && penjualan.RegistrasiID != filter.RegistrasiID {
			continue
		}
		if filter.Status != "" && penjualan.Status != filter.Status {
			continue
		}
		out = append(out, *penjualan)
	}
	return out, nil
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
	newRegistryHandler(h.svc).register(h.mux)
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

func TestBuatSertifikatSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"permohonan_id":1,"nomor_sertifikat":"SRT-001"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/sertifikat", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "40")
	r.Header.Set("X-User-Role", "penandatangan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestTandaTanganiTTESukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"ttd":"2026-10-23T10:00:00Z","dokumen_url":"https://minio/sertifikat.pdf"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/sertifikat/1/tte", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "40")
	r.Header.Set("X-User-Role", "penandatangan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestBuatDraftLabelSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"sertifikat_id":1,"kode_barcode":"SB-2026-0001"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/draft-label", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "50")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
}

func TestIntegrasiSIMSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	r := httptest.NewRequest(http.MethodPost, "/v1/draft-label/1/sim", nil)
	r.Header.Set("X-User-ID", "50")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestCetakLabelSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	r := httptest.NewRequest(http.MethodPost, "/v1/draft-label/1/cetak", nil)
	r.Header.Set("X-User-ID", "60")
	r.Header.Set("X-User-Role", "percetakan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestLegalisasiLabelSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"status":"diverifikasi","catatan":"label sudah sesuai standar"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/draft-label/1/legalisasi", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "70")
	r.Header.Set("X-User-Role", "legalisasi")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestProsesMutuLulus(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"sampel_id":1}`
	r := httptest.NewRequest(http.MethodPost, "/v1/proses-mutu", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "80")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["hasil"] != "lulus" {
		t.Fatalf("hasil = %q, ingin lulus", got["hasil"])
	}
}

func TestProsesMutuTidakLulus(t *testing.T) {
	h := newHandlerWithPalsu()
	body := `{"sampel_id":2}`
	r := httptest.NewRequest(http.MethodPost, "/v1/proses-mutu", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "80")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["hasil"] != "tidak_lulus" {
		t.Fatalf("hasil = %q, ingin tidak_lulus", got["hasil"])
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
