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
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// realmKeycloak menyiapkan realm uji dengan JWKS lokal agar token dapat
// ditandatangani dan diverifikasi tanpa Keycloak sungguhan.
func realmKeycloak(t *testing.T) (Realm, *rsa.PrivateKey) {
	t.Helper()
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
			Kid: "keycloak-uji", Kty: "RSA", Alg: "RS256", Use: "sig",
			N: base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
		}}})
	}))
	t.Cleanup(server.Close)
	return Realm{Name: "jsp", Issuer: server.URL, ClientID: "sitani-portal"}, privateKey
}

func signedKeycloakToken(t *testing.T, realm Realm, privateKey *rsa.PrivateKey, aud jwt.ClaimStrings, azp string) string {
	t.Helper()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: realm.Issuer, Subject: "subject-uji", Audience: aud,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		AuthorizedParty: azp, Email: "analis@dev.example.invalid", EmailVerified: true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "keycloak-uji"
	raw, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// Keycloak tidak mengisi aud dengan client id, melainkan peran resource
// ("account"), sedangkan client id berada di azp. Token seperti ini harus
// diterima agar alur SSO selesai.
func TestVerifierMenerimaTokenKeycloakDenganAudPeranResource(t *testing.T) {
	realm, privateKey := realmKeycloak(t)
	raw := signedKeycloakToken(t, realm, privateKey, jwt.ClaimStrings{"account"}, "sitani-portal")

	claims, verifiedRealm, err := NewVerifier(map[string]Realm{"jsp": realm}).Verify(context.Background(), raw)
	if err != nil {
		t.Fatalf("token Keycloak ditolak: %v", err)
	}
	if claims.Email != "analis@dev.example.invalid" || verifiedRealm.Name != "jsp" {
		t.Fatalf("claims/realm tidak sesuai: %+v %+v", claims, verifiedRealm)
	}
}

func TestVerifierMenolakTokenDariClientLain(t *testing.T) {
	realm, privateKey := realmKeycloak(t)
	raw := signedKeycloakToken(t, realm, privateKey, jwt.ClaimStrings{"account"}, "client-lain")

	if _, _, err := NewVerifier(map[string]Realm{"jsp": realm}).Verify(context.Background(), raw); err == nil {
		t.Fatal("token dengan azp client lain seharusnya ditolak")
	}
}

func TestVerifierMenolakTokenTanpaAuthorizedParty(t *testing.T) {
	realm, privateKey := realmKeycloak(t)
	raw := signedKeycloakToken(t, realm, privateKey, jwt.ClaimStrings{"sitani-portal"}, "")

	if _, _, err := NewVerifier(map[string]Realm{"jsp": realm}).Verify(context.Background(), raw); err == nil {
		t.Fatal("token tanpa azp seharusnya ditolak")
	}
}
