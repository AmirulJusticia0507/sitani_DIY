package devlogin

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// hashArgon2id membuat hash PHC argon2id untuk pengujian.
func hashArgon2id(t *testing.T, password string) string {
	t.Helper()
	salt := []byte("sitani-dev-salt")
	kunci := argon2.IDKey([]byte(password), salt, 1, 64*1024, 2, 32)
	return "$argon2id$v=19$m=65536,t=1,p=2$" +
		base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(kunci)
}

func TestVerifikasiPasswordMenerimaBcryptDanArgon2id(t *testing.T) {
	hashBcrypt, err := bcrypt.GenerateFromPassword([]byte("rahasia-benar"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	kasus := map[string]struct {
		hash     string
		password string
		ingin    bool
	}{
		"bcrypt benar":             {string(hashBcrypt), "rahasia-benar", true},
		"bcrypt salah":             {string(hashBcrypt), "rahasia-salah", false},
		"argon2id benar":           {hashArgon2id(t, "rahasia-benar"), "rahasia-benar", true},
		"argon2id salah":           {hashArgon2id(t, "rahasia-benar"), "rahasia-salah", false},
		"penanda akun SSO ditolak": {"SSO-DISABLED-abc123", "apa-pun", false},
		"hash kosong":              {"", "apa-pun", false},
		"format argon2id rusak":    {"$argon2id$v=19$rusak$abc$def", "rahasia-benar", false},
	}
	for nama, kasus := range kasus {
		hasil := verifikasiPassword(kasus.hash, kasus.password)
		if hasil != kasus.ingin {
			t.Fatalf("%s: hasil = %v, ingin %v", nama, hasil, kasus.ingin)
		}
	}
}

func TestHandlerMenolakTanpaOriginCocok(t *testing.T) {
	manager := NewManager(Config{Aktif: true, FrontendURL: "http://localhost:3002"}, nil)
	request := httptest.NewRequest(http.MethodPost, "/auth/dev/login",
		strings.NewReader(`{"email":"a@b.id","password":"x"}`))
	request.Header.Set("Origin", "https://jahat.example")
	recorder := httptest.NewRecorder()
	manager.login(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusForbidden)
	}

	logout := httptest.NewRequest(http.MethodPost, "/auth/dev/logout", nil)
	logout.Header.Set("Origin", "https://jahat.example")
	recorder = httptest.NewRecorder()
	manager.logout(recorder, logout)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("logout status = %d, ingin %d", recorder.Code, http.StatusForbidden)
	}
}

func TestEndpointTidakTersediaSaatNonaktif(t *testing.T) {
	manager := NewManager(Config{Aktif: false}, nil)
	mux := http.NewServeMux()
	manager.RegisterRoutes(mux)

	for _, kasus := range []struct{ method, path string }{
		{http.MethodGet, "/auth/dev/status"},
		{http.MethodPost, "/auth/dev/login"},
		{http.MethodPost, "/auth/dev/logout"},
	} {
		request := httptest.NewRequest(kasus.method, kasus.path, strings.NewReader("{}"))
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("%s %s status = %d, ingin 404", kasus.method, kasus.path, recorder.Code)
		}
	}
}

func TestLoginTanpaDatabaseMenjawab503(t *testing.T) {
	manager := NewManager(Config{Aktif: true, FrontendURL: "http://localhost:3002"}, nil)
	request := httptest.NewRequest(http.MethodPost, "/auth/dev/login",
		strings.NewReader(`{"email":"a@b.id","password":"x"}`))
	request.Header.Set("Origin", "http://localhost:3002")
	recorder := httptest.NewRecorder()
	manager.login(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestConfigFromEnvMenolakProduksi(t *testing.T) {
	t.Setenv("SITANI_DEV_LOGIN", "true")
	t.Setenv("APP_ENV", "production")
	if config := ConfigFromEnv(); config.Aktif {
		t.Fatal("login lokal tidak boleh aktif saat APP_ENV=production")
	}

	t.Setenv("APP_ENV", "development")
	if config := ConfigFromEnv(); !config.Aktif {
		t.Fatalf("login lokal seharusnya aktif di development, alasan: %s", config.Alasan)
	}

	t.Setenv("SITANI_DEV_LOGIN", "")
	if config := ConfigFromEnv(); config.Aktif {
		t.Fatal("login lokal tidak boleh aktif tanpa SITANI_DEV_LOGIN=true")
	}
}

func TestHashTokenTidakMenyimpanTokenMentah(t *testing.T) {
	token := "token-rahasia"
	if hashToken(token) == token || len(hashToken(token)) != 64 {
		t.Fatalf("token harus disimpan sebagai hash SHA-256 hex: %s", hashToken(token))
	}
}
