package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// tokenServer meniru token endpoint Keycloak dan merekam parameter yang
// dikirim, sehingga dapat dipastikan client secret hanya dikirim bila diisi.
func tokenServer(t *testing.T, token string) (*httptest.Server, *url.Values) {
	t.Helper()
	terekam := &url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/protocol/openid-connect/token" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		*terekam = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": token, "expires_in": 300})
	}))
	t.Cleanup(server.Close)
	return server, terekam
}

func TestExchangeCodeMengirimClientSecretHanyaUntukClientConfidential(t *testing.T) {
	server, terekam := tokenServer(t, "token-uji")
	manager := NewManager(Config{BackendURL: "http://localhost:8080"}, nil)

	if _, _, err := manager.exchangeCode(t.Context(), Realm{
		Name: "jsp", Issuer: server.URL, ClientID: "sitani-portal", ClientSecret: "rahasia-client",
	}, "kode", "verifier"); err != nil {
		t.Fatalf("pertukaran token gagal: %v", err)
	}
	if got := terekam.Get("client_secret"); got != "rahasia-client" {
		t.Fatalf("client_secret = %q, ingin terkirim", got)
	}
	if got := terekam.Get("redirect_uri"); got != "http://localhost:8080/auth/callback/jsp" {
		t.Fatalf("redirect_uri = %q", got)
	}

	if _, _, err := manager.exchangeCode(t.Context(), Realm{
		Name: "jsp", Issuer: server.URL, ClientID: "portal",
	}, "kode", "verifier"); err != nil {
		t.Fatalf("pertukaran token tanpa secret gagal: %v", err)
	}
	if terekam.Has("client_secret") {
		t.Fatal("client_secret tidak boleh dikirim untuk client publik")
	}
}

func TestConfigFromEnvMembacaClientSecret(t *testing.T) {
	t.Setenv("SSO_JSP_ISSUER", "http://localhost:8085/realms/sitani")
	t.Setenv("SSO_JSP_CLIENT_ID", "sitani-portal")
	t.Setenv("SSO_JSP_CLIENT_SECRET", "rahasia-lokal")
	t.Setenv("SSO_SATRIYA_ISSUER", "")
	t.Setenv("SSO_SATRIYA_CLIENT_ID", "")

	realm, ok := ConfigFromEnv().Realms["jsp"]
	if !ok {
		t.Fatal("realm jsp seharusnya ada")
	}
	if realm.ClientSecret != "rahasia-lokal" {
		t.Fatalf("ClientSecret = %q", realm.ClientSecret)
	}
}

func TestRealmTanpaClientIDTetapDiabaikan(t *testing.T) {
	t.Setenv("SSO_JSP_ISSUER", "http://localhost:8085/realms/sitani")
	t.Setenv("SSO_JSP_CLIENT_ID", "")
	t.Setenv("SSO_SATRIYA_ISSUER", "")
	t.Setenv("SSO_SATRIYA_CLIENT_ID", "")

	if err := ConfigFromEnv().Validate(); err == nil {
		t.Fatal("realm tanpa client id tidak boleh dianggap terkonfigurasi")
	}
}
