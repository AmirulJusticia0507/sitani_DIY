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

// realmCookie menyimpan kunci realm alur SSO (jsp/satriya) agar permintaan
// berikutnya dapat memilih konfigurasi bahkan setelah access token kedaluwarsa.
const realmCookie = "sitani_sso_realm"

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
	manager := &Manager{
		config:   config,
		verifier: NewVerifier(config.Realms),
		database: database,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	manager.verifier.SetLeeway(config.leeway())
	return manager
}

// SetDevLocal memasang login lokal (development/UAT). Bila tidak dipasang,
// autentikasi hanya memakai SSO.
func (m *Manager) SetDevLocal(manager devLocalResolver) { m.devLocal = manager }

func (m *Manager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/login/{realm}", m.login)
	mux.HandleFunc("GET /auth/callback/{realm}", m.callback)
	mux.HandleFunc("POST /auth/logout", m.logout)
	mux.HandleFunc("POST /auth/refresh", m.refresh)
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
		if fromCookie {
			cookie, cookieErr := r.Cookie(accessTokenCookie)
			if cookieErr != nil || cookie.Value == "" {
				writeAuthError(w, http.StatusUnauthorized, "token SSO tidak valid")
				return
			}
			session, sesiErr := m.sesiDariCookie(r)
			if sesiErr != nil || session.perluDiperbarui(m.config.refreshSkew()) {
				if session.refresh == "" {
					m.clearSessionCookies(w)
					writeAuthError(w, http.StatusUnauthorized, "sesi SSO telah berakhir")
					return
				}
				diperbarui, refreshErr := m.perbaruiSesi(w, r, session)
				if refreshErr != nil {
					slog.Warn("perpanjangan sesi SSO gagal", "realm", session.realm.Name, "error", refreshErr)
					m.clearSessionCookies(w)
					writeAuthError(w, http.StatusUnauthorized, "sesi SSO telah berakhir")
					return
				}
				session = diperbarui
			}
			m.terapkanSesi(w, r, session, next)
			return
		}
		claims, realm, err := m.verifier.VerifyIssuer(r.Context(), rawToken)
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

// terapkanSesi memetakan sesi SSO yang sudah valid ke principal lalu
// meneruskan permintaan.
func (m *Manager) terapkanSesi(w http.ResponseWriter, r *http.Request, session sesiSSO, next http.Handler) {
	principal, err := m.resolvePrincipal(r.Context(), session.claims, session.realm)
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
	token, codeToken, idToken, expiresIn, err := m.exchangeCode(r.Context(), realm, code, verifierCookie.Value)
	if err != nil {
		// Dicatat agar kegagalan konfigurasi (redirect_uri, secret, PKCE) dapat
		// didiagnosis. Nilai error hanya berisi kode kesalahan OAuth.
		slog.Warn("pertukaran authorization code gagal", "realm", realm.Name, "error", err)
		writeAuthError(w, http.StatusUnauthorized, "pertukaran authorization code gagal")
		return
	}
	claims, verifiedRealm, err := m.verifier.VerifyIssuer(r.Context(), token)
	if err != nil {
		slog.Warn("verifikasi token SSO gagal", "realm", realm.Name, "error", err)
		writeAuthError(w, http.StatusUnauthorized, "token SSO hasil callback tidak valid")
		return
	}
	if m.database == nil {
		// Callback tidak boleh panik bila database identitas belum siap; token
		// tidak dapat dipetakan sehingga permintaan ditolak dengan jelas.
		writeAuthError(w, http.StatusServiceUnavailable, "database identitas belum tersedia")
		return
	}
	if _, err := m.resolvePrincipal(r.Context(), claims, verifiedRealm); err != nil {
		slog.Warn("pemetaan identitas SSO gagal", "realm", realm.Name, "error", err)
		writeAuthError(w, http.StatusUnauthorized, "identitas SSO tidak dapat dipetakan")
		return
	}
	// id_token diverifikasi secara terpisah karena aud-nya memang client id
	// (berbeda dari access token). Bila tidak valid, cookie id_token tidak
	// disimpan sehingga logout hanya memakai access token.
	if idToken != "" {
		if _, idErr := m.verifikasiIDToken(r.Context(), idToken, realm); idErr != nil {
			slog.Warn("id_token SSO tidak valid", "realm", realm.Name, "error", idErr)
			idToken = ""
		}
	}
	m.simpanSesi(w, realm, sesiSSO{token: token, refresh: codeToken, idToken: idToken}, expiresIn)
	redirect := "/"
	if cookie, err := r.Cookie("sitani_sso_redirect_" + realm.Name); err == nil {
		redirect = safeRedirect(cookie.Value)
	}
	m.clearFlowCookies(w, realm.Name)
	http.Redirect(w, r, m.config.FrontendURL+redirect, http.StatusFound)
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
	target := m.akhiriSesiSSO(r)
	m.clearSessionCookies(w)
	writeJSON(w, http.StatusOK, map[string]string{"logout_url": target})
}

// refresh adalah endpoint eksplisit untuk memperbarui sesi sebelum access
// token kedaluwarsa. Require sudah memperbarui otomatis; endpoint ini berguna
// untuk aplikasi yang perlu memastikan sesi masih hidup.
func (m *Manager) refresh(w http.ResponseWriter, r *http.Request) {
	if !m.allowedBrowserOrigin(r.Header.Get("Origin")) {
		writeAuthError(w, http.StatusForbidden, "origin permintaan tidak diizinkan")
		return
	}
	if err := m.config.Validate(); err != nil {
		writeAuthError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	session, err := m.sesiDariCookie(r)
	if session.refresh == "" {
		m.clearSessionCookies(w)
		writeAuthError(w, http.StatusUnauthorized, "sesi SSO telah berakhir")
		return
	}
	baru, refreshErr := m.perbaruiSesi(w, r, session)
	if refreshErr != nil {
		slog.Warn("perpanjangan sesi SSO gagal", "realm", session.realm.Name, "error", refreshErr)
		m.clearSessionCookies(w)
		writeAuthError(w, http.StatusUnauthorized, "sesi SSO telah berakhir")
		return
	}
	if err != nil {
		slog.Info("access token lama tidak valid, sesi diperbarui", "realm", baru.realm.Name)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"realm":      baru.realm.Name,
		"expires_at": baru.kedaluwarsa.UTC().Format(time.RFC3339),
	})
}

// akhiriSesiSSO mencabut refresh token di penyedia identitas lalu menyusun URL
// end_session_endpoint (OIDC RP-Initiated Logout) agar sesi SSO di browser juga
// berakhir. Gagal mencabut token tidak menghalangi logout lokal; kegagalannya
// dicatat dan URL end_session tetap diberikan bila dapat disusun.
func (m *Manager) akhiriSesiSSO(r *http.Request) string {
	session, _ := m.sesiDariCookie(r)
	if session.realm.Name == "" {
		return ""
	}
	realm := session.realm
	if session.refresh != "" {
		values := url.Values{
			"token":           {session.refresh},
			"token_type_hint": {"refresh_token"},
			"client_id":       {realm.ClientID},
		}
		if realm.ClientSecret != "" {
			values.Set("client_secret", realm.ClientSecret)
		}
		if err := m.revoke(r.Context(), realm, values); err != nil {
			slog.Warn("pencabutan refresh token gagal", "realm", realm.Name, "error", err)
		}
	}
	return m.logoutURL(session, m.config.FrontendURL+"/masuk")
}

// revoke memanggil revocation_endpoint. Endpoint ini tidak wajib ada di setiap
// penyedia; status selain 200 dicatat sebagai kesalahan.
func (m *Manager) revoke(ctx context.Context, realm Realm, values url.Values) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, realm.Issuer+"/protocol/openid-connect/revoke", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := m.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("revocation endpoint status %d", response.StatusCode)
	}
	return nil
}

// logoutURL menyusun URL end_session_endpoint untuk mengakhiri sesi di
// penyedia identitas. id_token_hint memakai id_token yang sudah diverifikasi;
// bila tidak ada, client_id dan post_logout_redirect_uri dipakai (Keycloak
// menerima keduanya bila redirect terdaftar pada post.logout.redirect.uris).
func (m *Manager) logoutURL(session sesiSSO, postLogout string) string {
	realm := session.realm
	values := url.Values{}
	if session.idToken != "" {
		if claims, err := m.verifikasiIDToken(context.Background(), session.idToken, realm); err == nil {
			values.Set("id_token_hint", session.idToken)
			if claims.SessionID != "" {
				values.Set("sid", claims.SessionID)
			}
		}
	}
	values.Set("client_id", realm.ClientID)
	values.Set("post_logout_redirect_uri", postLogout)
	return realm.Issuer + "/protocol/openid-connect/logout?" + values.Encode()
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
