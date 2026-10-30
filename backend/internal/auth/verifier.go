package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	AuthorizedParty string `json:"azp"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	Username        string `json:"preferred_username"`
	Name            string `json:"name"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type keySet struct {
	Keys []jwk `json:"keys"`
}

type cachedKeys struct {
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}

type Verifier struct {
	realms map[string]Realm
	client *http.Client
	mu     sync.RWMutex
	cache  map[string]cachedKeys
}

func NewVerifier(realms map[string]Realm) *Verifier {
	return &Verifier{
		realms: realms,
		client: &http.Client{Timeout: 10 * time.Second},
		cache:  make(map[string]cachedKeys),
	}
}

func (v *Verifier) Verify(ctx context.Context, rawToken string) (*Claims, Realm, error) {
	if rawToken == "" {
		return nil, Realm{}, errors.New("token tidak tersedia")
	}
	unverified := &Claims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(rawToken, unverified); err != nil {
		return nil, Realm{}, errors.New("format token tidak valid")
	}
	realm, ok := v.realmByIssuer(unverified.Issuer)
	if !ok {
		return nil, Realm{}, errors.New("issuer token tidak diizinkan")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("algoritma token harus RS256")
		}
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("header kid tidak tersedia")
		}
		return v.publicKey(ctx, realm, kid)
	}, jwt.WithIssuer(realm.Issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, Realm{}, fmt.Errorf("verifikasi token gagal: %w", err)
	}
	// Keycloak mengisi aud dengan peran resource (mis. "account") dan menaruh
	// client id pada azp, sehingga klaim azp yang diwajibkan cocok. Dengan
	// begitu token dari client lain pada realm yang sama tetap ditolak.
	if claims.AuthorizedParty != realm.ClientID {
		return nil, Realm{}, errors.New("authorized party token tidak sesuai")
	}
	if claims.Subject == "" || claims.Email == "" || !claims.EmailVerified {
		return nil, Realm{}, errors.New("claim sub atau email terverifikasi tidak tersedia")
	}
	return claims, realm, nil
}

func (v *Verifier) realmByIssuer(issuer string) (Realm, bool) {
	issuer = strings.TrimRight(issuer, "/")
	for _, realm := range v.realms {
		if realm.Issuer == issuer {
			return realm, true
		}
	}
	return Realm{}, false
}

func (v *Verifier) publicKey(ctx context.Context, realm Realm, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	cached, ok := v.cache[realm.Name]
	if ok && time.Now().Before(cached.expiresAt) {
		if key := cached.keys[kid]; key != nil {
			v.mu.RUnlock()
			return key, nil
		}
	}
	v.mu.RUnlock()

	keys, err := v.fetchKeys(ctx, realm)
	if err != nil {
		return nil, err
	}
	key := keys[kid]
	if key == nil {
		return nil, errors.New("kunci token tidak ditemukan")
	}
	return key, nil
}

func (v *Verifier) fetchKeys(ctx context.Context, realm Realm) (map[string]*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, realm.Issuer+"/protocol/openid-connect/certs", nil)
	if err != nil {
		return nil, err
	}
	response, err := v.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("ambil JWKS: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS mengembalikan status %d", response.StatusCode)
	}
	var set keySet
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(&set); err != nil {
		return nil, fmt.Errorf("baca JWKS: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, item := range set.Keys {
		if item.Kty != "RSA" || item.Use != "sig" || item.Alg != "RS256" {
			continue
		}
		key, err := rsaKey(item.N, item.E)
		if err == nil {
			keys[item.Kid] = key
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("JWKS tidak berisi kunci RS256")
	}
	v.mu.Lock()
	v.cache[realm.Name] = cachedKeys{keys: keys, expiresAt: time.Now().Add(time.Hour)}
	v.mu.Unlock()
	return keys, nil
}

func rsaKey(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(modulus)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(exponent)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, value := range eBytes {
		e = e<<8 + int(value)
	}
	if e < 3 {
		return nil, errors.New("eksponen RSA tidak valid")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}
