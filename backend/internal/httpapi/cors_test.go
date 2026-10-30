package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Halaman sertifikasi fase 1 mengirim X-User-ID dan X-User-Role sehingga
// preflight harus mengizinkan keduanya; identitas tetap berasal dari sesi.
func TestCORSAllowHeaderMencakupHeaderHalamanSertifikasi(t *testing.T) {
	t.Setenv("FRONTEND_PUBLIC_URL", "http://localhost:3002")

	req := httptest.NewRequest(http.MethodOptions, "/v1/permohonan/aktif", nil)
	req.Header.Set("Origin", "http://localhost:3002")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	cors(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("preflight tidak boleh diteruskan ke handler")
	})).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, ingin 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3002" {
		t.Fatalf("allow-origin = %q", got)
	}
	izinkan := w.Header().Get("Access-Control-Allow-Headers")
	for _, header := range []string{"Content-Type", "X-User-ID", "X-User-Role"} {
		if !mengandungHeader(izinkan, header) {
			t.Fatalf("header %q tidak diizinkan: %q", header, izinkan)
		}
	}
}

// Origin di luar konfigurasi tidak mendapat header CORS.
func TestCORSOriginTidakDikenalTidakDiizinkan(t *testing.T) {
	t.Setenv("FRONTEND_PUBLIC_URL", "http://localhost:3002")

	req := httptest.NewRequest(http.MethodGet, "/v1/permohonan/aktif", nil)
	req.Header.Set("Origin", "http://contoh-lain.test")
	w := httptest.NewRecorder()
	cors(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow-origin = %q, ingin kosong", got)
	}
}

// mengandungHeader memeriksa daftar header CORS yang dipisah koma tanpa peka
// huruf besar/kecil, sesuai perbandingan browser.
func mengandungHeader(daftar, target string) bool {
	dicari := ""
	for _, r := range target {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		dicari += string(r)
	}
	bagian := ""
	for _, r := range daftar {
		if r == ',' {
			if trimHeader(bagian) == dicari {
				return true
			}
			bagian = ""
			continue
		}
		bagian += string(r)
	}
	return trimHeader(bagian) == dicari
}

func trimHeader(s string) string {
	hasil := ""
	for _, r := range s {
		if r == ' ' || r == '\t' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		hasil += string(r)
	}
	return hasil
}
