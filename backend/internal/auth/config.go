package auth

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Realm struct {
	Name         string
	Issuer       string
	ClientID     string
	ClientSecret string
}

type Config struct {
	Realms       map[string]Realm
	FrontendURL  string
	BackendURL   string
	SecureCookie bool
	// RefreshSkew adalah umur minimum sisa token akses: token yang tersisa
	// kurang dari ini diperbarui lewat refresh token sebelum dipakai.
	RefreshSkew time.Duration
	// Leeway adalah toleransi jam saat memverifikasi exp/iat/nbf token.
	Leeway time.Duration
}

// DefaultRefreshSkew memberi ruang untuk perjalanan jaringan dan pemrosesan
// sebelum token akses kedaluwarsa.
const DefaultRefreshSkew = 60 * time.Second

// DefaultLeeway dipakai karena jam server penyedia identitas dan backend dapat
// berbeda beberapa detik.
const DefaultLeeway = 60 * time.Second

func ConfigFromEnv() Config {
	config := Config{
		Realms:       make(map[string]Realm),
		FrontendURL:  strings.TrimRight(os.Getenv("FRONTEND_PUBLIC_URL"), "/"),
		BackendURL:   strings.TrimRight(os.Getenv("BACKEND_PUBLIC_URL"), "/"),
		SecureCookie: strings.EqualFold(os.Getenv("SSO_COOKIE_SECURE"), "true"),
		RefreshSkew:  durationFromEnv("SSO_REFRESH_SKEW_SECONDS", DefaultRefreshSkew),
		Leeway:       durationFromEnv("SSO_TOKEN_LEEWAY_SECONDS", DefaultLeeway),
	}
	if config.FrontendURL == "" {
		config.FrontendURL = "http://localhost:3000"
	}
	if config.BackendURL == "" {
		config.BackendURL = "http://localhost:8080"
	}
	addRealm(config.Realms, "jsp", os.Getenv("SSO_JSP_ISSUER"), os.Getenv("SSO_JSP_CLIENT_ID"), os.Getenv("SSO_JSP_CLIENT_SECRET"))
	addRealm(config.Realms, "satriya", os.Getenv("SSO_SATRIYA_ISSUER"), os.Getenv("SSO_SATRIYA_CLIENT_ID"), os.Getenv("SSO_SATRIYA_CLIENT_SECRET"))
	return config
}

func addRealm(realms map[string]Realm, name, issuer, clientID, clientSecret string) {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	clientID = strings.TrimSpace(clientID)
	if issuer != "" && clientID != "" {
		// Client secret opsional: client publik memakai PKCE saja. Client
		// confidential (seperti realm lokal Keycloak) wajib mengisinya agar
		// token endpoint menerima pertukaran authorization code.
		realms[name] = Realm{Name: name, Issuer: issuer, ClientID: clientID, ClientSecret: strings.TrimSpace(clientSecret)}
	}
}

func (c Config) Validate() error {
	if len(c.Realms) == 0 {
		return errors.New("konfigurasi SSO JSP/Satriya belum tersedia")
	}
	return nil
}

// refreshSkew mengembalikan ambang sisa umur token akses. Nol berarti nilai
// default, sehingga 0 tidak bisa dipakai untuk mematikan pembaruan token.
func (c Config) refreshSkew() time.Duration {
	if c.RefreshSkew <= 0 {
		return DefaultRefreshSkew
	}
	return c.RefreshSkew
}

func (c Config) leeway() time.Duration {
	if c.Leeway < 0 {
		return 0
	}
	return c.Leeway
}

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}
