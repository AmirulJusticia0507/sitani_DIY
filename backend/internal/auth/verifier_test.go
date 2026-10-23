package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifierMenerimaTokenValid(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/protocol/openid-connect/certs" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(keySet{Keys: []jwk{{
			Kid: "test-key", Kty: "RSA", Alg: "RS256", Use: "sig",
			N: base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
		}}})
	}))
	defer server.Close()

	realm := Realm{Name: "jsp", Issuer: server.URL, ClientID: "sitani"}
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: server.URL, Subject: "subject-1",
			Audience:  jwt.ClaimStrings{"sitani"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		AuthorizedParty: "sitani", Email: "pengguna@example.go.id", EmailVerified: true, Name: "Pengguna",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	rawToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	verified, verifiedRealm, err := NewVerifier(map[string]Realm{"jsp": realm}).Verify(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("verifikasi token: %v", err)
	}
	if verified.Subject != "subject-1" || verifiedRealm.Name != "jsp" {
		t.Fatalf("claims/realm tidak sesuai: %+v %+v", verified, verifiedRealm)
	}
}

func TestVerifierMenolakIssuerTidakDikenal(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{RegisteredClaims: jwt.RegisteredClaims{
		Issuer: "https://issuer-tidak-dikenal.example", Subject: "subject-1",
		Audience: jwt.ClaimStrings{"sitani"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}, Email: "pengguna@example.go.id"})
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	token.Header["kid"] = "test-key"
	rawToken, _ := token.SignedString(privateKey)

	if _, _, err := NewVerifier(map[string]Realm{}).Verify(context.Background(), rawToken); err == nil {
		t.Fatal("issuer tidak dikenal seharusnya ditolak")
	}
}

func TestProtectAPIGagalTertutupTanpaKonfigurasi(t *testing.T) {
	manager := NewManager(Config{Realms: map[string]Realm{}}, nil)
	handler := manager.ProtectAPI(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	request := httptest.NewRequest(http.MethodGet, "/v1/permohonan/1", nil)
	request.Header.Set("X-User-ID", "999")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func TestLoginMembuatAuthorizationCodePKCE(t *testing.T) {
	config := Config{
		Realms:      map[string]Realm{"jsp": {Name: "jsp", Issuer: "https://sso.example/realms/jsp", ClientID: "sitani"}},
		FrontendURL: "https://portal.example", BackendURL: "https://api.example", SecureCookie: true,
	}
	manager := NewManager(config, nil)
	mux := http.NewServeMux()
	manager.RegisterRoutes(mux)
	request := httptest.NewRequest(http.MethodGet, "/auth/login/jsp?redirect=/permohonan", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusFound)
	}
	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "response_type=code") || !strings.Contains(location, "code_challenge_method=S256") {
		t.Fatalf("redirect tidak memakai authorization code + PKCE: %s", location)
	}
	if len(recorder.Result().Cookies()) != 3 {
		t.Fatalf("jumlah cookie alur = %d, ingin 3", len(recorder.Result().Cookies()))
	}
}

func TestLogoutMenolakOriginAsing(t *testing.T) {
	manager := NewManager(Config{FrontendURL: "https://portal.example"}, nil)
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.Header.Set("Origin", "https://jahat.example")
	recorder := httptest.NewRecorder()
	manager.logout(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusForbidden)
	}
}
