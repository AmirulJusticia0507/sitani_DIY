package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth/devlogin"
)

// resolverUji meniru devlogin.Manager tanpa database.
type resolverUji struct {
	aktif     bool
	cocok     bool
	principal devlogin.Principal
	rute      *http.ServeMux
}

func (r *resolverUji) Aktif() bool { return r.aktif }

func (r *resolverUji) Resolve(_ context.Context, _ *http.Request) (devlogin.Principal, bool) {
	return r.principal, r.cocok
}

func (r *resolverUji) RegisterRoutes(mux *http.ServeMux) {
	if r.rute != nil {
		mux.Handle("GET /auth/dev/status", r.rute)
	}
}

// managerUjiSSOTerkonfigurasi membuat Manager dengan realm SSO terisi seperti
// pada stack Docker, sehingga Validate() berhasil.
func managerUjiSSOTerkonfigurasi() *Manager {
	return NewManager(Config{
		Realms: map[string]Realm{
			"jsp": {Name: "jsp", Issuer: "https://sso.example/realms/jsp", ClientID: "portal"},
		},
		FrontendURL: "http://localhost:3001",
	}, nil)
}

func TestRequireMenerimaSesiLokalSaatSSOTerkonfigurasi(t *testing.T) {
	manager := managerUjiSSOTerkonfigurasi()
	manager.SetDevLocal(&resolverUji{
		aktif:     true,
		cocok:     true,
		principal: devlogin.Principal{UserID: 7, Email: "admin@dev.example.invalid", Role: "admin", Realm: "lokal"},
	})

	var principal Principal
	handler := manager.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, _ = PrincipalFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/permohonan", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200: %s", recorder.Code, recorder.Body.String())
	}
	if principal.UserID != 7 || principal.Role != "admin" || principal.Realm != "lokal" {
		t.Fatalf("principal = %+v, ingin user 7 peran admin realm lokal", principal)
	}
}

func TestRequireMenolakTanpaSesiLokalMeskiSSOTerkonfigurasi(t *testing.T) {
	manager := managerUjiSSOTerkonfigurasi()
	manager.SetDevLocal(&resolverUji{aktif: true})
	// Aktif tetapi tidak ada cookie/token lokal yang cocok.
	// Tanpa database, alur SSO menjawab 503 sebelum memetakan identitas; yang
	// diuji di sini adalah sesi lokal tidak dipakai untuk melewati verifikasi.
	manager.database = nil

	handler := manager.Require(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler tidak boleh dipanggil tanpa token SSO")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/permohonan", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin 503", recorder.Code)
	}
}

func TestRequireMengabaikanLoginLokalNonaktif(t *testing.T) {
	manager := managerUjiSSOTerkonfigurasi()
	// Resolve akan mengembalikan true bila dipanggil; Aktif() false memastikan
	// cabang lokal tidak dipakai sehingga tidak ada sesi yang diterima.
	manager.SetDevLocal(&resolverUji{
		aktif:     false,
		cocok:     true,
		principal: devlogin.Principal{UserID: 7, Role: "admin"},
	})

	handler := manager.Require(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler tidak boleh dipanggil")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/permohonan", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin 503", recorder.Code)
	}
}

func TestRequireGagalTertutupTanpaSSODanTanpaSesiLokal(t *testing.T) {
	manager := NewManager(Config{FrontendURL: "http://localhost:3001"}, nil)

	recorder := httptest.NewRecorder()
	manager.Require(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler tidak boleh dipanggil")
	})).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/permohonan", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin 503", recorder.Code)
	}
}
