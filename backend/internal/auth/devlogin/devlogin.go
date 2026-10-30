// Package devlogin menyediakan login lokal berbasis email dan password untuk
// development/UAT saat SSO Jogjaprov belum aktif. Paket ini terpisah dari
// autentikasi SSO dan hanya boleh diaktifkan di luar produksi.
//
// Keamanan yang diterapkan:
//   - Aktif hanya bila SITANI_DEV_LOGIN=true dan tidak ada penanda produksi
//     (APP_ENV/GO_ENV/ENVIRONMENT bernilai production/prod, SITANI_ENV=prod, atau
//     SENTRY_ENVIRONMENT=production). Selain itu semua endpoint menjawab 404.
//   - Hanya password ber-hash argon2id/bcrypt yang diterima. Penanda
//     "SSO-DISABLED-..." (akun hasil pemetaan SSO) selalu ditolak.
//   - Pesan gagal login seragam dan tetap menghitung hash palsu agar waktu
//     respons tidak membocorkan keberadaan akun.
//   - Sesi memakai cookie token acak yang disimpan sebagai hash SHA-256 di
//     database; token mentah tidak pernah disimpan.
package devlogin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/db"
)

// cookieToken adalah nama cookie sesi login lokal. Cookie terpisah dari cookie
// SSO agar keduanya tidak saling menimpa.
const cookieToken = "sitani_dev_session"

// realmName dipakai pada Principal dan kolom auth_sessions.realm.
const realmName = "lokal"

// sesiTTL adalah masa berlaku sesi login lokal.
const sesiTTL = 8 * time.Hour

var errKredensialSalah = errors.New("email atau password salah")

// Principal adalah identitas sesi lokal dengan bentuk field yang sama dengan
// auth.Principal. Didefinisikan di sini agar paket auth dapat memakai devlogin
// tanpa impor melingkar; pemetaan ke auth.Principal dilakukan di paket auth.
type Principal struct {
	UserID   int64  `json:"user_id"`
	Subject  string `json:"subject"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Realm    string `json:"realm"`
}

// Config menyimpan konfigurasi login lokal.
type Config struct {
	// Aktif menandakan login lokal boleh melayani permintaan.
	Aktif bool
	// Alasan berisi keterangan mengapa login lokal dinonaktifkan, untuk log.
	Alasan string
	// FrontendURL dipakai untuk memvalidasi Origin permintaan dan tujuan redirect.
	FrontendURL string
	// SecureCookie mengatur atribut Secure pada cookie sesi.
	SecureCookie bool
}

// ConfigFromEnv membaca konfigurasi login lokal dari environment.
func ConfigFromEnv() Config {
	aktif := strings.EqualFold(strings.TrimSpace(os.Getenv("SITANI_DEV_LOGIN")), "true")
	alasan := ""
	if !aktif {
		alasan = "SITANI_DEV_LOGIN tidak diset true"
	} else if lingkunganProduksi() {
		aktif = false
		alasan = "lingkungan terdeteksi produksi"
	}
	return Config{
		Aktif:        aktif,
		Alasan:       alasan,
		FrontendURL:  strings.TrimRight(os.Getenv("FRONTEND_PUBLIC_URL"), "/"),
		SecureCookie: strings.EqualFold(os.Getenv("SSO_COOKIE_SECURE"), "true"),
	}
}

func lingkunganProduksi() bool {
	for _, nama := range []string{"APP_ENV", "GO_ENV", "ENVIRONMENT", "SITANI_ENV"} {
		switch strings.ToLower(strings.TrimSpace(os.Getenv(nama))) {
		case "production", "prod":
			return true
		}
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("SENTRY_ENVIRONMENT")), "production")
}

// Manager menangani login, logout, dan pembacaan sesi lokal.
type Manager struct {
	config   Config
	database *db.DB
}

// NewManager membuat manager login lokal. Database boleh nil; saat nil semua
// permintaan dijawab 503.
func NewManager(config Config, database *db.DB) *Manager {
	return &Manager{config: config, database: database}
}

// Aktif mengembalikan true bila login lokal boleh melayani permintaan.
func (m *Manager) Aktif() bool { return m != nil && m.config.Aktif }

// Alasan mengembalikan keterangan mengapa login lokal tidak aktif.
func (m *Manager) Alasan() string {
	if m == nil {
		return "manager tidak tersedia"
	}
	return m.config.Alasan
}

// RegisterRoutes mendaftarkan endpoint login lokal. Rute selalu terdaftar agar
// respons tetap seragam; setiap handler menolak dengan 404 bila login lokal
// tidak aktif.
func (m *Manager) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/dev/status", m.status)
	mux.HandleFunc("POST /auth/dev/login", m.login)
	mux.HandleFunc("POST /auth/dev/logout", m.logout)
}

// Resolve mengembalikan Principal dari cookie atau bearer token sesi lokal
// yang valid. Nilai false berarti tidak ada sesi lokal yang cocok.
func (m *Manager) Resolve(ctx context.Context, r *http.Request) (Principal, bool) {
	if !m.Aktif() || m.database == nil {
		return Principal{}, false
	}
	raw := bearerToken(r)
	if raw == "" {
		if cookie, err := r.Cookie(cookieToken); err == nil {
			raw = cookie.Value
		}
	}
	if raw == "" {
		return Principal{}, false
	}
	var principal Principal
	err := m.database.QueryRowContext(ctx, `
		SELECT u.id, u.email, COALESCE(u.username,''), u.role,
		       COALESCE(NULLIF(u.username,''), u.email)
		FROM auth_sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ? AND s.realm = ? AND s.expires_at > CURRENT_TIMESTAMP AND u.is_active = TRUE`,
		hashToken(raw), realmName,
	).Scan(&principal.UserID, &principal.Email, &principal.Username, &principal.Role, &principal.Name)
	if err != nil {
		return Principal{}, false
	}
	principal.Realm = realmName
	principal.Subject = "local:" + strconv.FormatInt(principal.UserID, 10)
	return principal, true
}

func (m *Manager) status(w http.ResponseWriter, _ *http.Request) {
	if !m.Aktif() {
		writeError(w, http.StatusNotFound, "login lokal tidak tersedia")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"aktif": true})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// maksBodyLogin = batas body permintaan login lokal. Email dan password jauh di
// bawahnya; nilai ini hanya mencegah body raksasa ditampung di memori.
const maksBodyLogin = 4 << 10

// cobaBacaBody membaca body dengan batas maksBodyLogin dan melaporkan apakah
// body melewati batas. Hanya dipakai saat Content-Length sudah menyatakan body
// terlalu besar, sehingga body kecil tidak ikut termakan.
func cobaBacaBody(r *http.Request) bool {
	terbatas := http.MaxBytesReader(nil, r.Body, maksBodyLogin)
	_, err := io.Copy(io.Discard, terbatas)
	var terlaluBesar *http.MaxBytesError
	return errors.As(err, &terlaluBesar)
}

func (m *Manager) login(w http.ResponseWriter, r *http.Request) {
	if !m.Aktif() {
		writeError(w, http.StatusNotFound, "login lokal tidak tersedia")
		return
	}
	// Login memakai cookie sehingga Origin harus cocok untuk mencegah CSRF.
	// Diperiksa sebelum hal lain agar permintaan lintas situs selalu ditolak
	// tanpa membuka informasi tentang keadaan server.
	if !m.originDiizinkan(r) {
		writeError(w, http.StatusForbidden, "origin permintaan tidak diizinkan")
		return
	}
	// Batas body JSON berlaku sebelum pemeriksaan database: permintaan raksasa
	// tidak boleh ditampung utuh, dan penyebabnya harus jelas (413, bukan 503).
	if r.ContentLength > maksBodyLogin && cobaBacaBody(r) {
		writeError(w, http.StatusRequestEntityTooLarge, "body permintaan melebihi batas 4 KB")
		return
	}
	if m.database == nil {
		writeError(w, http.StatusServiceUnavailable, "database identitas belum tersedia")
		return
	}

	var payload loginRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maksBodyLogin))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		var terlaluBesar *http.MaxBytesError
		if errors.As(err, &terlaluBesar) {
			writeError(w, http.StatusRequestEntityTooLarge, "body permintaan melebihi batas 4 KB")
			return
		}
		writeError(w, http.StatusBadRequest, "permintaan login tidak valid")
		return
	}
	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if email == "" || payload.Password == "" {
		writeError(w, http.StatusBadRequest, "email dan password wajib diisi")
		return
	}

	principal, err := m.verifikasi(r.Context(), email, payload.Password)
	if err != nil {
		// Pesan seragam untuk email tidak dikenal maupun password salah.
		writeError(w, http.StatusUnauthorized, "email atau password salah")
		return
	}

	token, err := randomToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membuat sesi")
		return
	}
	kedaluwarsa := time.Now().Add(sesiTTL)
	if _, err := m.database.ExecContext(r.Context(), `
		INSERT INTO auth_sessions (user_id, token_hash, realm, expires_at)
		VALUES (?, ?, ?, ?)`,
		principal.UserID, hashToken(token), realmName, kedaluwarsa,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "gagal menyimpan sesi")
		return
	}
	// Bersihkan sesi kedaluwarsa milik pengguna ini agar tabel tidak menumpuk.
	_, _ = m.database.ExecContext(r.Context(),
		`DELETE FROM auth_sessions WHERE user_id = ? AND expires_at <= CURRENT_TIMESTAMP`, principal.UserID)

	http.SetCookie(w, &http.Cookie{
		Name: cookieToken, Value: token, Path: "/", HttpOnly: true,
		Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode,
		MaxAge: int(sesiTTL.Seconds()),
	})
	writeJSON(w, http.StatusOK, principal)
}

func (m *Manager) logout(w http.ResponseWriter, r *http.Request) {
	if !m.Aktif() {
		writeError(w, http.StatusNotFound, "login lokal tidak tersedia")
		return
	}
	if !m.originDiizinkan(r) {
		writeError(w, http.StatusForbidden, "origin permintaan tidak diizinkan")
		return
	}
	if cookie, err := r.Cookie(cookieToken); err == nil && cookie.Value != "" && m.database != nil {
		_, _ = m.database.ExecContext(r.Context(),
			`DELETE FROM auth_sessions WHERE token_hash = ?`, hashToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieToken, Path: "/", HttpOnly: true,
		Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) originDiizinkan(r *http.Request) bool {
	origin := strings.TrimRight(r.Header.Get("Origin"), "/")
	return origin != "" && strings.EqualFold(origin, m.config.FrontendURL)
}

// verifikasi mencocokkan email dan password dengan pengguna aktif yang memiliki
// hash password yang sah.
func (m *Manager) verifikasi(ctx context.Context, email, password string) (Principal, error) {
	var principal Principal
	var hash string
	err := m.database.QueryRowContext(ctx, `
		SELECT id, email, COALESCE(username,''), role, password_hash,
		       COALESCE(NULLIF(username,''), email)
		FROM users
		WHERE LOWER(email) = ? AND is_active = TRUE`,
		email,
	).Scan(&principal.UserID, &principal.Email, &principal.Username, &principal.Role, &hash, &principal.Name)
	if err != nil {
		// Tetap jalankan perbandingan hash palsu agar waktu respons setara.
		_ = bcrypt.CompareHashAndPassword(hashBcryptPalsu, []byte(password))
		return Principal{}, errKredensialSalah
	}
	if !verifikasiPassword(hash, password) {
		return Principal{}, errKredensialSalah
	}
	principal.Realm = realmName
	principal.Subject = "local:" + strconv.FormatInt(principal.UserID, 10)
	return principal, nil
}

// hashBcryptPalsu dipakai untuk menyamakan waktu respons saat email tidak ada.
var hashBcryptPalsu = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("sitani-tidak-ada"), bcrypt.DefaultCost)
	if err != nil {
		// Cost default tidak pernah gagal untuk password pendek; fallback tetap
		// berupa hash valid agar CompareHashAndPassword tidak error format.
		return []byte("$2a$10$7890abcdefghijklmnopqu0123456789012345678901234567890123456")
	}
	return hash
}()

// verifikasiPassword mendukung bcrypt dan argon2id. Nilai "SSO-DISABLED-..."
// sengaja tidak dianggap sah sehingga akun hasil pemetaan SSO tidak dapat
// dipakai lewat login lokal.
func verifikasiPassword(hash, password string) bool {
	switch {
	case strings.HasPrefix(hash, "$2a$"), strings.HasPrefix(hash, "$2b$"), strings.HasPrefix(hash, "$2y$"):
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	case strings.HasPrefix(hash, "$argon2id$"):
		return verifikasiArgon2id(hash, password)
	default:
		return false
	}
}

// verifikasiArgon2id memeriksa format PHC hasil argon2id.
// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func verifikasiArgon2id(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var mem uint32
	var iter uint32
	var par uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iter, &par); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	harapan, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(harapan) == 0 {
		return false
	}
	hasil := argon2.IDKey([]byte(password), salt, iter, mem, par, uint32(len(harapan)))
	return subtle.ConstantTimeCompare(hasil, harapan) == 1
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if len(value) > 7 && strings.EqualFold(value[:7], "Bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
