package devlogin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Body login yang melewati batas 4 KB harus dijawab 413, bukan 400/503, dan
// dicek sebelum database agar permintaan raksasa tidak ditampung utuh.
func TestLoginBodyBesarDijawab413(t *testing.T) {
	m := NewManager(Config{Aktif: true, FrontendURL: "http://127.0.0.1:3002"}, nil)
	besar := []byte(`{"email":"` + strings.Repeat("a", 8<<10) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/dev/login", strings.NewReader(string(besar)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://127.0.0.1:3002")
	rec := httptest.NewRecorder()

	m.login(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, ingin 413 (%s)", rec.Code, rec.Body.String())
	}
}

// Body kecil tanpa database tetap 503 seperti sebelumnya.
func TestLoginBodyKecilTetap503TanpaDatabase(t *testing.T) {
	m := NewManager(Config{Aktif: true, FrontendURL: "http://127.0.0.1:3002"}, nil)
	req := httptest.NewRequest(http.MethodPost, "/auth/dev/login",
		strings.NewReader(`{"email":"a@b.c","password":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://127.0.0.1:3002")
	rec := httptest.NewRecorder()

	m.login(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin 503 (%s)", rec.Code, rec.Body.String())
	}
}
