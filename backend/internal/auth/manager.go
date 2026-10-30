package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth/devlogin"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

const accessTokenCookie = "sitani_sso_access"

type Principal struct {
	UserID   int64  `json:"user_id"`
	Subject  string `json:"subject"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Realm    string `json:"realm"`
}

type principalKey struct{}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	return principal, ok
}

// WithPrincipal menyisipkan principal terverifikasi ke context. Dipakai oleh
// middleware SSO dan pengujian handler yang tidak melewati verifikasi token.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

type devLocalResolver interface {
	Aktif() bool
	Resolve(ctx context.Context, r *http.Request) (devlogin.Principal, bool)
	RegisterRoutes(mux *http.ServeMux)
}

type Manager struct {
	config   Config
	verifier *Verifier
	database *db.DB
	client   *http.Client
	devLocal devLocalResolver
}

func NewManager(config Config, database *db.DB) *Manager {
	return &Manager{
		config:   config,
		verifier: NewVerifier(config.Realms),
		database: database,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// SetDevLocal memasang login lokal (development/UAT). Bila tidak dipasang,
// autentikasi hanya memakai SSO.
func (m *Manager) SetDevLocal(manager devLocalResolver) { m.devLocal = manager }

func (m *Manager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/login/{realm}", m.login)
	mux.HandleFunc("GET /auth/callback/{realm}", m.callback)
	mux.HandleFunc("POST /auth/logout", m.logout)
	mux.Handle("GET /auth/me", m.Require(http.HandlerFunc(m.me)))
	if m.devLocal != nil {
		m.devLocal.RegisterRoutes(mux)
	}
}

// Require memverifikasi identitas dari token SSO. Sesi login lokal
// (auth/devlogin) diterima lebih dulu karena cookie-nya berbeda; tanpa sesi
// lokal, alur SSO berjalan seperti biasa. Bila keduanya tidak tersedia,
// permintaan ditolak agar sistem gagal tertutup.
func (m *Manager) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-ID")
		r.Header.Del("X-User-Role")
		if m.devLocal != nil && m.devLocal.Aktif() {
			if lokal, ok := m.devLocal.Resolve(r.Context(), r); ok {
				principal := Principal{
					UserID: lokal.UserID, Subject: lokal.Subject, Email: lokal.Email,
					Name: lokal.Name, Username: lokal.Username, Role: lokal.Role, Realm: lokal.Realm,
				}
				r.Header.Set("X-User-ID", strconv.FormatInt(principal.UserID, 10))
				r.Header.Set("X-User-Role", principal.Role)
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, principal)))
				return
			}
		}
		if err := m.config.Validate(); err != nil {
			writeAuthError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		if m.database == nil {
			writeAuthError(w, http.StatusServiceUnavailable, "database identitas belum tersedia")
			return
		}
		rawToken := bearerToken(r)
		fromCookie := rawToken == ""
		if rawToken == "" {
			if cookie, err := r.Cookie(accessTokenCookie); err == nil {
				rawToken = cookie.Value
			}
		}
		if fromCookie && isUnsafeMethod(r.Method) && !m.allowedBrowserOrigin(r.Header.Get("Origin")) {
			writeAuthError(w, http.StatusForbidden, "origin permintaan tidak diizinkan")
			return
		}
		claims, realm, err := m.verifier.Verify(r.Context(), rawToken)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "token SSO tidak valid")
			return
		}
		principal, err := m.resolvePrincipal(r.Context(), claims, realm)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "identitas SSO tidak dapat dipetakan")
			return
		}
		// Kompatibilitas sementara untuk handler bisnis lama. Nilai dari client
		// sudah dihapus dan diganti hanya setelah token berhasil diverifikasi.
		r.Header.Set("X-User-ID", strconv.FormatInt(principal.UserID, 10))
		r.Header.Set("X-User-Role", principal.Role)
		ctx := context.WithValue(r.Context(), principalKey{}, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Manager) allowedBrowserOrigin(origin string) bool {
	return origin != "" && strings.EqualFold(strings.TrimRight(origin, "/"), m.config.FrontendURL)
}

func isUnsafeMethod(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func (m *Manager) ProtectAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/") {
			m.Require(next).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Manager) login(w http.ResponseWriter, r *http.Request) {
	realm, ok := m.config.Realms[r.PathValue("realm")]
	if !ok {
		writeAuthError(w, http.StatusNotFound, "realm SSO tidak tersedia")
		return
	}
	state, err := randomURLSafe(32)
	if err != nil {
		writeAuthError(w, http.StatusInternalServerError, "gagal membuat state SSO")
		return
	}
	verifier, err := randomURLSafe(48)
	if err != nil {
		writeAuthError(w, http.StatusInternalServerError, "gagal membuat PKCE verifier")
		return
	}
	redirect := safeRedirect(r.URL.Query().Get("redirect"))
	m.setFlowCookie(w, "sitani_sso_state_"+realm.Name, state, 600)
	m.setFlowCookie(w, "sitani_sso_verifier_"+realm.Name, verifier, 600)
	m.setFlowCookie(w, "sitani_sso_redirect_"+realm.Name, redirect, 600)

	values := url.Values{
		"client_id":             {realm.ClientID},
		"redirect_uri":          {m.callbackURL(realm)},
		"response_type":         {"code"},
		"scope":                 {"openid profile email"},
		"state":                 {state},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	}
	http.Redirect(w, r, realm.Issuer+"/protocol/openid-connect/auth?"+values.Encode(), http.StatusFound)
}

func (m *Manager) callback(w http.ResponseWriter, r *http.Request) {
	realm, ok := m.config.Realms[r.PathValue("realm")]
	if !ok {
		writeAuthError(w, http.StatusNotFound, "realm SSO tidak tersedia")
		return
	}
	stateCookie, stateErr := r.Cookie("sitani_sso_state_" + realm.Name)
	verifierCookie, verifierErr := r.Cookie("sitani_sso_verifier_" + realm.Name)
	if stateErr != nil || verifierErr != nil || stateCookie.Value == "" || stateCookie.Value != r.URL.Query().Get("state") {
		writeAuthError(w, http.StatusBadRequest, "state SSO tidak valid atau kedaluwarsa")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		writeAuthError(w, http.StatusBadRequest, "authorization code tidak tersedia")
		return
	}
	token, expiresIn, err := m.exchangeCode(r.Context(), realm, code, verifierCookie.Value)
	if err != nil {
		// Dicatat agar kegagalan konfigurasi (redirect_uri, secret, PKCE) dapat
		// didiagnosis. Nilai error hanya berisi kode kesalahan OAuth.
		slog.Warn("pertukaran authorization code gagal", "realm", realm.Name, "error", err)
		writeAuthError(w, http.StatusUnauthorized, "pertukaran authorization code gagal")
		return
	}
	claims, verifiedRealm, err := m.verifier.Verify(r.Context(), token)
	if err != nil {
		slog.Warn("verifikasi token SSO gagal", "realm", realm.Name, "error", err)
		writeAuthError(w, http.StatusUnauthorized, "token SSO hasil callback tidak valid")
		return
	}
	if _, err := m.resolvePrincipal(r.Context(), claims, verifiedRealm); err != nil {
		slog.Warn("pemetaan identitas SSO gagal", "realm", realm.Name, "error", err)
		writeAuthError(w, http.StatusUnauthorized, "identitas SSO tidak dapat dipetakan")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: accessTokenCookie, Value: token, Path: "/", HttpOnly: true,
		Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode, MaxAge: expiresIn,
	})
	redirect := "/"
	if cookie, err := r.Cookie("sitani_sso_redirect_" + realm.Name); err == nil {
		redirect = safeRedirect(cookie.Value)
	}
	m.clearFlowCookies(w, realm.Name)
	http.Redirect(w, r, m.config.FrontendURL+redirect, http.StatusFound)
}

func (m *Manager) exchangeCode(ctx context.Context, realm Realm, code, verifier string) (string, int, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {realm.ClientID},
		"redirect_uri":  {m.callbackURL(realm)},
		"code":          {code},
		"code_verifier": {verifier},
	}
	// Client confidential mengautentikasi dengan secret; client publik yang
	// hanya memakai PKCE tidak mengirimnya.
	if realm.ClientSecret != "" {
		values.Set("client_secret", realm.ClientSecret)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, realm.Issuer+"/protocol/openid-connect/token", strings.NewReader(values.Encode()))
	if err != nil {
		return "", 0, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := m.client.Do(request)
	if err != nil {
		return "", 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// Hanya kode kesalahan OAuth yang dicatat: pesannya berasal dari
		// penyedia identitas dan tidak memuat token atau data pribadi.
		var oauthErr struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		_ = json.NewDecoder(io.LimitReader(response.Body, 4<<10)).Decode(&oauthErr)
		return "", 0, fmt.Errorf("token endpoint status %d: %s %s", response.StatusCode, oauthErr.Error, oauthErr.Description)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return "", 0, err
	}
	if payload.AccessToken == "" {
		return "", 0, errors.New("access token kosong")
	}
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = 300
	}
	return payload.AccessToken, payload.ExpiresIn, nil
}

func (m *Manager) resolvePrincipal(ctx context.Context, claims *Claims, realm Realm) (Principal, error) {
	tx, err := m.database.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return Principal{}, err
	}
	defer tx.Rollback()

	var userID int64
	var role string
	err = tx.QueryRowContext(ctx, `SELECT u.id, u.role FROM sso_identities i JOIN users u ON u.id=i.user_id WHERE i.issuer=? AND i.subject=? AND u.is_active=TRUE`, realm.Issuer, claims.Subject).Scan(&userID, &role)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT id, role FROM users WHERE email=? AND is_active=TRUE`, claims.Email).Scan(&userID, &role)
		if errors.Is(err, sql.ErrNoRows) {
			username := claims.Username
			if username == "" {
				username = strings.Split(claims.Email, "@")[0]
			}
			passwordMarker, randomErr := randomURLSafe(48)
			if randomErr != nil {
				return Principal{}, randomErr
			}
			userID, err = db.InsertID(ctx, tx, `INSERT INTO users (username,email,password_hash,role,is_active) VALUES (?,?,?, ?,TRUE)`, username, claims.Email, "SSO-DISABLED-"+passwordMarker, "pelanggan")
			if err != nil {
				return Principal{}, err
			}
			role = "pelanggan"
		} else if err != nil {
			return Principal{}, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO sso_identities (user_id,issuer,subject,realm,email) VALUES (?,?,?,?,?)`, userID, realm.Issuer, claims.Subject, realm.Name, claims.Email); err != nil {
			return Principal{}, err
		}
	} else if err != nil {
		return Principal{}, err
	}
	if err := tx.Commit(); err != nil {
		return Principal{}, err
	}
	return Principal{UserID: userID, Subject: claims.Subject, Email: claims.Email, Name: claims.Name, Username: claims.Username, Role: role, Realm: realm.Name}, nil
}

func (m *Manager) me(w http.ResponseWriter, r *http.Request) {
	principal, ok := PrincipalFromContext(r.Context())
	if !ok {
		writeAuthError(w, http.StatusUnauthorized, "identitas tidak tersedia")
		return
	}
	writeJSON(w, http.StatusOK, principal)
}

func (m *Manager) logout(w http.ResponseWriter, r *http.Request) {
	// Logout memakai cookie sehingga Origin wajib cocok untuk mencegah CSRF.
	// Request API non-browser dengan bearer token tidak memerlukan endpoint ini.
	if !m.allowedBrowserOrigin(r.Header.Get("Origin")) {
		writeAuthError(w, http.StatusForbidden, "origin permintaan tidak diizinkan")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: accessTokenCookie, Path: "/", HttpOnly: true, Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode, MaxAge: -1})
	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) callbackURL(realm Realm) string {
	return m.config.BackendURL + "/auth/callback/" + realm.Name
}

func (m *Manager) setFlowCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/auth/callback/", HttpOnly: true, Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode, MaxAge: maxAge})
}

// clearFlowCookies mengakhiri cookie state/PKCE satu realm setelah callback.
// Nilai dikosongkan dan kedaluwarsa seketika (MaxAge negatif tidak dirender
// sebagai atribut dalam spesifikasi cookie, jadi tambahkan Expires juga).
func (m *Manager) clearFlowCookies(w http.ResponseWriter, realm string) {
	for _, prefix := range []string{"sitani_sso_state_", "sitani_sso_verifier_", "sitani_sso_redirect_"} {
		http.SetCookie(w, &http.Cookie{
			Name: prefix + realm, Value: "", Path: "/auth/callback/", HttpOnly: true,
			Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode,
			MaxAge: -1, Expires: time.Unix(0, 0),
		})
	}
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if len(value) > 7 && strings.EqualFold(value[:7], "Bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}

func randomURLSafe(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func safeRedirect(value string) string {
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return "/"
	}
	return value
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
