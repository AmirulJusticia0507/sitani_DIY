package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
)

func newRegistryHandlerWithStore(st *palsuStore) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: service.NewService(st)}
	newRegistryHandler(h.svc).register(h.mux)
	return h
}

func tanggalRelatif(hari int) string {
	return time.Now().AddDate(0, 0, hari).Format("2006-01-02")
}

func bodyRegistrasiUji() string {
	return `{"permohonan_id":1,"komoditas":"Tomat","varietas":"Rama","produsen":"Kelompok Tani Sleman","sumber_benih":"BP3MBTP","nomor_lot":"LOT-001","kelas_benih":"sebar","jumlah":100,"satuan":"kg","tanggal_produksi":"` + tanggalRelatif(-30) + `","tanggal_kedaluwarsa":"` + tanggalRelatif(180) + `","status_edar":"aktif"}`
}

func buatRegistrasiUji(t *testing.T, h *Handler) map[string]interface{} {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/v1/registrasi-benih", strings.NewReader(bodyRegistrasiUji()))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("buat registrasi status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode registrasi: %v", err)
	}
	return got
}

func jualBenihUji(t *testing.T, h *Handler, jumlah float64) *httptest.ResponseRecorder {
	t.Helper()
	body := `{"nomor_dokumen":"PJ-001","registrasi_id":1,"pembeli":"Petani Bantul","jumlah":` +
		formatAngka(jumlah) + `,"harga_satuan":25000,"tanggal_transaksi":"` + tanggalRelatif(0) + `"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/penjualan-benih", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	return w
}

func formatAngka(v float64) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func TestBuatRegistrasiBenihSukses(t *testing.T) {
	h := newHandlerWithPalsu()
	got := buatRegistrasiUji(t, h)
	if got["NomorRegistrasi"] != "RB-LOT-001" {
		t.Fatalf("NomorRegistrasi = %v, ingin RB-LOT-001", got["NomorRegistrasi"])
	}
	if got["NomorSertifikat"] != "SRT-001" || got["KodeLabel"] != "SB-2026-0001" {
		t.Fatalf("dokumen tidak terhubung: %v", got)
	}
	if got["Jumlah"] != float64(100) {
		t.Fatalf("Jumlah = %v, ingin 100", got["Jumlah"])
	}
}

func TestBuatRegistrasiDokumenBelumSahDitolak(t *testing.T) {
	st := newPalsuStore()
	st.dokumenValid = false
	h := newRegistryHandlerWithStore(st)
	r := httptest.NewRequest(http.MethodPost, "/v1/registrasi-benih", strings.NewReader(bodyRegistrasiUji()))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "TTE") {
		t.Fatalf("pesan kesalahan tidak menjelaskan TTE: %s", w.Body.String())
	}
}

func TestBuatRegistrasiNomorLotGandaDitolak(t *testing.T) {
	h := newHandlerWithPalsu()
	buatRegistrasiUji(t, h)
	r := httptest.NewRequest(http.MethodPost, "/v1/registrasi-benih", strings.NewReader(bodyRegistrasiUji()))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "sudah terdaftar") {
		t.Fatalf("pesan tidak sesuai: %s", w.Body.String())
	}
}

func TestJualBenihMengurangiStok(t *testing.T) {
	h := newHandlerWithPalsu()
	buatRegistrasiUji(t, h)
	w := jualBenihUji(t, h, 60)
	if w.Code != http.StatusCreated {
		t.Fatalf("jual status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var penjualan map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &penjualan); err != nil {
		t.Fatalf("decode penjualan: %v", err)
	}
	if penjualan["TotalHarga"] != float64(1500000) {
		t.Fatalf("TotalHarga = %v, ingin 1500000", penjualan["TotalHarga"])
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/registrasi-benih/1/stok", nil)
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w2 := httptest.NewRecorder()
	h.mux.ServeHTTP(w2, r)
	if w2.Code != http.StatusOK {
		t.Fatalf("stok status = %d; body=%s", w2.Code, w2.Body.String())
	}
	var stok map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &stok); err != nil {
		t.Fatalf("decode stok: %v", err)
	}
	stokData, ok := stok["stok"].(map[string]interface{})
	if !ok {
		t.Fatalf("respons stok tidak lengkap: %v", stok)
	}
	if stokData["JumlahTersedia"] != float64(40) {
		t.Fatalf("JumlahTersedia = %v, ingin 40", stokData["JumlahTersedia"])
	}
	mutasi, ok := stok["mutasi"].([]interface{})
	if !ok || len(mutasi) != 2 {
		t.Fatalf("mutasi = %v, ingin 2 entri", stok["mutasi"])
	}
}

func TestJualBenihMelebihiStokDitolak(t *testing.T) {
	h := newHandlerWithPalsu()
	buatRegistrasiUji(t, h)
	w := jualBenihUji(t, h, 120)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
}

func TestJualBenihDokumenTidakSahDitolak(t *testing.T) {
	st := newPalsuStore()
	h := newRegistryHandlerWithStore(st)
	buatRegistrasiUji(t, h)
	st.dokumenValid = false
	w := jualBenihUji(t, h, 10)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestBatalkanPenjualanMengembalikanStok(t *testing.T) {
	h := newHandlerWithPalsu()
	buatRegistrasiUji(t, h)
	jual := jualBenihUji(t, h, 60)
	if jual.Code != http.StatusCreated {
		t.Fatalf("jual status = %d; body=%s", jual.Code, jual.Body.String())
	}

	body := `{"alasan":"pembeli membatalkan pesanan"}`
	r := httptest.NewRequest(http.MethodPost, "/v1/penjualan-benih/1/batal", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("batal status = %d, ingin %d; body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	r2 := httptest.NewRequest(http.MethodGet, "/v1/registrasi-benih/1/stok", nil)
	r2.Header.Set("X-User-ID", "1")
	r2.Header.Set("X-User-Role", "admin")
	w2 := httptest.NewRecorder()
	h.mux.ServeHTTP(w2, r2)
	var stok map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &stok); err != nil {
		t.Fatalf("decode stok: %v", err)
	}
	stokData := stok["stok"].(map[string]interface{})
	if stokData["JumlahTersedia"] != float64(100) {
		t.Fatalf("JumlahTersedia = %v, ingin 100", stokData["JumlahTersedia"])
	}
}

func TestLaporanKedaluwarsaBenih(t *testing.T) {
	h := newHandlerWithPalsu()
	buatRegistrasiUji(t, h)
	r := httptest.NewRequest(http.MethodGet, "/v1/laporan/kedaluwarsa-benih?hari=365", nil)
	r.Header.Set("X-User-ID", "1")
	r.Header.Set("X-User-Role", "admin")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", w.Code, w.Body.String())
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode laporan: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("jumlah lot = %d, ingin 1", len(items))
	}
}

func TestRegistrasiDitolakUntukRoleTidakBerwenang(t *testing.T) {
	h := newHandlerWithPalsu()
	r := httptest.NewRequest(http.MethodPost, "/v1/registrasi-benih", strings.NewReader(bodyRegistrasiUji()))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "10")
	r.Header.Set("X-User-Role", "pelanggan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}
