package auth

import (
	"errors"
	"os"
	"strings"
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
}

func ConfigFromEnv() Config {
	config := Config{
		Realms:       make(map[string]Realm),
		FrontendURL:  strings.TrimRight(os.Getenv("FRONTEND_PUBLIC_URL"), "/"),
		BackendURL:   strings.TrimRight(os.Getenv("BACKEND_PUBLIC_URL"), "/"),
		SecureCookie: strings.EqualFold(os.Getenv("SSO_COOKIE_SECURE"), "true"),
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
