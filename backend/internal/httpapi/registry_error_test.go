package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Tes pemetaan kesalahan handler registrasi benih memastikan klien menerima
// status HTTP yang tepat untuk penolakan otorisasi, validasi, dan data hilang.

func registryRequest(metode, target, body, role string) *http.Request {
	r := httptest.NewRequest(metode, target, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if role != "" {
		r.Header.Set("X-User-ID", "1")
		r.Header.Set("X-User-Role", role)
	}
	return r
}

func serveRegistry(t *testing.T, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	newHandlerWithPalsu().mux.ServeHTTP(w, r)
	return w
}

func TestRegistryMembutuhkanIdentitas(t *testing.T) {
	kasus := []struct{ metode, target, body string }{
		{http.MethodPost, "/v1/registrasi-benih", bodyRegistrasiUji()},
		{http.MethodGet, "/v1/registrasi-benih", ""},
		{http.MethodGet, "/v1/registrasi-benih/1", ""},
		{http.MethodGet, "/v1/registrasi-benih/1/stok", ""},
		{http.MethodPost, "/v1/penjualan-benih", `{}`},
		{http.MethodGet, "/v1/penjualan-benih", ""},
		{http.MethodGet, "/v1/laporan/stok-benih", ""},
	}
	for _, k := range kasus {
		t.Run(k.metode+k.target, func(t *testing.T) {
			w := serveRegistry(t, registryRequest(k.metode, k.target, k.body, ""))
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, ingin %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestRegistryMenolakRoleTanpaKewenangan(t *testing.T) {
	w := serveRegistry(t, registryRequest(http.MethodGet, "/v1/laporan/stok-benih", "", "pelanggan"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "tidak diizinkan") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestRegistryMenolakBodyRusak(t *testing.T) {
	w := serveRegistry(t, registryRequest(http.MethodPost, "/v1/registrasi-benih", `{bukan json`, "admin"))
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "body permintaan tidak valid") {
		t.Fatalf("status = %d; body=%s", w.Code, w.Body.String())
	}
}

func TestRegistryIDTidakValid(t *testing.T) {
	for _, target := range []string{"/v1/registrasi-benih/abc", "/v1/registrasi-benih/-1/stok"} {
		w := serveRegistry(t, registryRequest(http.MethodGet, target, "", "admin"))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, ingin %d", target, w.Code, http.StatusBadRequest)
		}
	}
}

func TestRegistryDataTidakDitemukan(t *testing.T) {
	for _, target := range []string{"/v1/registrasi-benih/999", "/v1/registrasi-benih/999/stok"} {
		w := serveRegistry(t, registryRequest(http.MethodGet, target, "", "admin"))
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, ingin %d; body=%s", target, w.Code, http.StatusNotFound, w.Body.String())
		}
	}
}

func TestRegistryStokTidakCukupMenghasilkanKonflik(t *testing.T) {
	h := newHandlerWithPalsu()
	if w := h.mux; w == nil {
		t.Fatal("mux tidak tersedia")
	}
	buatRegistrasiUji(t, h)

	body := `{"nomor_dokumen":"PJ-002","registrasi_id":1,"pembeli":"Petani","jumlah":500,"harga_satuan":1000,"tanggal_transaksi":"` + tanggalRelatif(0) + `"}`
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, registryRequest(http.MethodPost, "/v1/penjualan-benih", body, "admin"))
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "stok benih tidak mencukupi") {
		t.Fatalf("status = %d; body=%s", w.Code, w.Body.String())
	}
}

func TestRegistryParameterFilterTidakValid(t *testing.T) {
	for _, target := range []string{
		"/v1/penjualan-benih?registrasi_id=abc",
		"/v1/penjualan-benih?dari=23-10-2026",
		"/v1/penjualan-benih?sampai=bukan-tanggal",
		"/v1/penjualan-benih?status=proses",
		"/v1/laporan/kedaluwarsa-benih?hari=99999",
		"/v1/registrasi-benih?status_edar=karantina",
	} {
		w := serveRegistry(t, registryRequest(http.MethodGet, target, "", "admin"))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, ingin %d; body=%s", target, w.Code, http.StatusBadRequest, w.Body.String())
		}
	}
}
