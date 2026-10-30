package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenResponse menampung respons token endpoint (authorization_code dan
// refresh_token).

// Nama cookie alur SSO. Semua cookie bersifat HttpOnly dan SameSite=Lax karena
// token tidak pernah dibaca JavaScript.
const (
	// refreshTokenCookie menyimpan refresh token. Cookie ini hanya dikirim ke
	// endpoint callback dan perpanjangan sesi di backend (Path dibatasi).
	refreshTokenCookie = "sitani_sso_refresh"
	// idTokenCookie menyimpan id_token untuk logout SSO (id_token_hint dan sid).
	// Diperlukan karena cookie access token dapat sudah kedaluwarsa saat logout.
	idTokenCookie = "sitani_sso_id"
)

// Path tempat cookie refresh/id token dikirim: hanya ke backend, bukan ke
// halaman frontend.
const tokenPath = "/auth/"

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	// Error diisi bila token endpoint menolak permintaan.
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// exchangeCode menukar authorization code dengan token. Nilai balik:
// access token, refresh token, id_token, dan masa berlaku access token (detik).
func (m *Manager) exchangeCode(ctx context.Context, realm Realm, code, verifier string) (string, string, string, int, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {realm.ClientID},
		"redirect_uri":  {m.callbackURL(realm)},
		"code":          {code},
		"code_verifier": {verifier},
	}
	payload, err := m.tokenRequest(ctx, realm, values)
	if err != nil {
		return "", "", "", 0, err
	}
	return payload.AccessToken, payload.RefreshToken, payload.IDToken, payload.expiresIn(), nil
}

// refreshAccessToken memperbarui access token memakai refresh token. Sama
// seperti exchangeCode, hanya kode kesalahan OAuth yang diteruskan ke pemanggil
// (untuk log) tanpa token atau data pribadi.
func (m *Manager) refreshAccessToken(ctx context.Context, realm Realm, refreshToken string) (tokenResponse, error) {
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {realm.ClientID},
		"refresh_token": {refreshToken},
	}
	return m.tokenRequest(ctx, realm, values)
}

func (m *Manager) tokenRequest(ctx context.Context, realm Realm, values url.Values) (tokenResponse, error) {
	// Client confidential mengautentikasi dengan secret; client publik yang
	// hanya memakai PKCE tidak mengirimnya.
	if realm.ClientSecret != "" {
		values.Set("client_secret", realm.ClientSecret)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, realm.Issuer+"/protocol/openid-connect/token", strings.NewReader(values.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := m.client.Do(request)
	if err != nil {
		return tokenResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		var oauthErr tokenResponse
		_ = json.NewDecoder(io.LimitReader(response.Body, 4<<10)).Decode(&oauthErr)
		return tokenResponse{}, fmt.Errorf("token endpoint status %d: %s %s", response.StatusCode, oauthErr.Error, oauthErr.ErrorDescription)
	}
	var payload tokenResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return tokenResponse{}, err
	}
	if payload.AccessToken == "" {
		return tokenResponse{}, errors.New("access token kosong")
	}
	return payload, nil
}

func (p tokenResponse) expiresIn() int {
	if p.ExpiresIn <= 0 {
		return 300
	}
	if p.ExpiresIn < minimumCookieAgeSeconds {
		return minimumCookieAgeSeconds
	}
	return p.ExpiresIn
}

// minimumCookieAgeSeconds memastikan cookie access token tidak dikirim dengan
// MaxAge=0 (yang berarti cookie sesi dan hilang saat browser ditutup) walaupun
// penyedia identitas memberi masa berlaku sangat pendek. Validitas token tetap
// ditentukan klaim exp.
const minimumCookieAgeSeconds = 60

// refreshTokenBaru mengembalikan refresh token dari respons, atau nilai lama
// bila penyedia identitas tidak mengirim refresh token baru (rotasi opsional).
func (p tokenResponse) refreshTokenBaru(lama string) string {
	if p.RefreshToken != "" {
		return p.RefreshToken
	}
	return lama
}

// simpanSesi menulis cookie sesi hasil callback. Hanya cookie yang nilainya ada
// yang dikirim; cookie realm selalu ditulis agar refresh tetap dapat memilih
// konfigurasi setelah access token kedaluwarsa.
func (m *Manager) simpanSesi(w http.ResponseWriter, realm Realm, sesi sesiSSO, expiresIn int) {
	cookies := []*http.Cookie{m.accessCookie(sesi.token, expiresIn), m.cookie(realmCookie, realm.Name, "/", 0)}
	if sesi.refresh != "" {
		cookies = append(cookies, m.codeCookie(sesi.refresh))
	}
	if sesi.idToken != "" {
		cookies = append(cookies, m.idTokenCookie(sesi.idToken))
	}
	for _, cookie := range cookies {
		http.SetCookie(w, cookie)
	}
}

// --- Cookie -----------------------------------------------------------------

func (m *Manager) cookie(name, value string, path string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name: name, Value: value, Path: path, HttpOnly: true,
		Secure: m.config.SecureCookie, SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	}
}

func (m *Manager) accessCookie(token string, expiresIn int) *http.Cookie {
	return m.cookie(accessTokenCookie, token, "/", expiresIn)
}

// idTokenCookie menyimpan id_token sampai sesi browser berakhir (cookie sesi).
func (m *Manager) idTokenCookie(idToken string) *http.Cookie {
	return m.cookie(idTokenCookie, idToken, tokenPath, 0)
}

// codeCookie menyimpan refresh token. Memakai atribut cookie sesi (tanpa
// MaxAge) karena masa berlaku refresh token dikendalikan penyedia identitas;
// menghapusnya dilakukan lewat cookieKedaluwarsa.
func (m *Manager) codeCookie(refreshToken string) *http.Cookie {
	return m.cookie(refreshTokenCookie, refreshToken, tokenPath, 0)
}

// cookieKedaluwarsa membuat cookie yang dapat menghapus nilai di browser.
// MaxAge negatif tidak dirender sebagai atribut dalam spesifikasi cookie, jadi
// Expires juga diisi.
func (m *Manager) cookieKedaluwarsa(name, path string) *http.Cookie {
	cookie := m.cookie(name, "", path, -1)
	cookie.Expires = time.Unix(0, 0)
	return cookie
}

// clearSessionCookies menghapus seluruh cookie sesi SSO. Dipakai saat logout,
// refresh token ditolak, dan callback gagal agar browser tidak menyimpan token
// yang sudah tidak dapat dipakai.
func (m *Manager) clearSessionCookies(w http.ResponseWriter) {
	for _, cookie := range []*http.Cookie{
		m.cookieKedaluwarsa(accessTokenCookie, "/"),
		m.cookieKedaluwarsa(realmCookie, "/"),
		m.cookieKedaluwarsa(refreshTokenCookie, tokenPath),
		m.cookieKedaluwarsa(idTokenCookie, tokenPath),
	} {
		http.SetCookie(w, cookie)
	}
}

// --- Sesi tersimpan di cookie ----------------------------------------------

// sesiSSO adalah isi cookie sesi setelah diverifikasi.
type sesiSSO struct {
	token       string
	refresh     string
	idToken     string
	claims      *Claims
	realm       Realm
	kedaluwarsa time.Time
}

// parseUnverified membaca klaim tanpa memverifikasi tanda tangan. Hanya dipakai
// untuk memilih konfigurasi realm; keputusan keamanan tetap di Verify.
func parseUnverified(rawToken string, claims *Claims) error {
	_, _, err := new(jwt.Parser).ParseUnverified(rawToken, claims)
	return err
}

// sesiDariCookie memverifikasi access token dari cookie. Sisa umurnya dihitung
// dari klaim exp, bukan dari header cookie.
//
// Bila cookie realm tersedia, verifikasi memakai konfigurasi realm tersebut
// (termasuk issuer) sehingga access token dapat diverifikasi walaupun issuer di
// dalamnya sudah tidak dipakai backend — misalnya setelah host name berubah.
// Hal ini tidak melemahkan keamanan karena tanda tangan, issuer, dan azp tetap
// diperiksa; hanya sumber konfigurasi yang berbeda.
func (m *Manager) sesiDariCookie(r *http.Request) (sesiSSO, error) {
	cookie, err := r.Cookie(accessTokenCookie)
	if err != nil || cookie.Value == "" {
		return sesiSSO{}, errors.New("cookie token SSO tidak tersedia")
	}
	session := sesiSSO{token: cookie.Value}
	// Issuer di dalam token hanya dipakai untuk memilih konfigurasi; tanda
	// tangan, issuer, dan azp tetap diverifikasi pada panggilan Verify.
	if unverified := (&Claims{}); parseUnverified(cookie.Value, unverified) == nil {
		if realm, ok := m.verifier.realmByIssuer(unverified.Issuer); ok {
			session.realm = realm
		}
	}
	// Issuer lama yang sudah tidak dipakai backend (mis. host name berubah):
	// pakai cookie realm agar sesi masih dapat diperbarui lewat refresh token.
	if session.realm.Name == "" {
		if realmCookieValue, err := r.Cookie(realmCookie); err == nil {
			if realm, ok := m.config.Realms[realmCookieValue.Value]; ok {
				session.realm = realm
			}
		}
	}
	if refresh, err := r.Cookie(refreshTokenCookie); err == nil {
		session.refresh = refresh.Value
	}
	if idToken, err := r.Cookie(idTokenCookie); err == nil {
		session.idToken = idToken.Value
	}
	// Access token yang tidak valid (kedaluwarsa/rusak) bukan kegagalan sesi:
	// pemanggil dapat memperbaruinya memakai refresh token.
	var claims *Claims
	var realm Realm
	var verifyErr error
	if session.realm.Name != "" {
		claims, realm, verifyErr = m.verifier.Verify(r.Context(), cookie.Value, session.realm)
	} else {
		claims, realm, verifyErr = m.verifier.VerifyIssuer(r.Context(), cookie.Value)
	}
	if verifyErr != nil {
		return session, verifyErr
	}
	session.claims = claims
	session.realm = realm
	if claims.ExpiresAt != nil {
		session.kedaluwarsa = claims.ExpiresAt.Time
	}
	return session, nil
}

// sisaUmur mengembalikan durasi sampai token kedaluwarsa.
func (s sesiSSO) sisaUmur() time.Duration {
	return time.Until(s.kedaluwarsa)
}

// perluDiperbarui menentukan apakah access token harus diperbarui sebelum
// dipakai. Tanpa exp, token dianggap perlu diperbarui.
func (s sesiSSO) perluDiperbarui(skew time.Duration) bool {
	if s.kedaluwarsa.IsZero() {
		return true
	}
	return s.sisaUmur() < skew
}

// perbaruiSesi meminta refresh token baru lalu memverifikasi access token
// hasilnya. Cookie diperbarui hanya setelah token hasil refresh valid.
func (m *Manager) perbaruiSesi(w http.ResponseWriter, r *http.Request, sesi sesiSSO) (sesiSSO, error) {
	if sesi.refresh == "" {
		return sesiSSO{}, errors.New("refresh token tidak tersedia")
	}
	payload, err := m.refreshAccessToken(r.Context(), sesi.realm, sesi.refresh)
	if err != nil {
		return sesiSSO{}, err
	}
	claims, realm, err := m.verifier.Verify(r.Context(), payload.AccessToken, sesi.realm)
	if err != nil {
		return sesiSSO{}, fmt.Errorf("access token hasil refresh tidak valid: %w", err)
	}
	if realm.Name != sesi.realm.Name {
		return sesiSSO{}, errors.New("realm token hasil refresh berubah")
	}
	baru := sesiSSO{
		token:   payload.AccessToken,
		refresh: payload.refreshTokenBaru(sesi.refresh),
		idToken: sesi.idToken,
		claims:  claims,
		realm:   realm,
	}
	if claims.ExpiresAt != nil {
		baru.kedaluwarsa = claims.ExpiresAt.Time
	}
	http.SetCookie(w, m.accessCookie(baru.token, payload.expiresIn()))
	http.SetCookie(w, m.codeCookie(baru.refresh))
	if payload.IDToken != "" {
		baru.idToken = payload.IDToken
		http.SetCookie(w, m.idTokenCookie(baru.idToken))
	}
	slog.Info("token SSO diperbarui", "realm", realm.Name)
	return baru, nil
}

// verifikasiIDToken memvalidasi id_token hasil token endpoint: issuer, tanda
// tangan RS256, kedaluwarsa, dan client (aud/azp). sid dan sub dipakai untuk
// logout SSO.
func (m *Manager) verifikasiIDToken(ctx context.Context, raw string, realm Realm) (*Claims, error) {
	if raw == "" {
		return nil, errors.New("id_token kosong")
	}
	unverified := &Claims{}
	if _, _, err := new(jwt.Parser).ParseUnverified(raw, unverified); err != nil {
		return nil, errors.New("format id_token tidak valid")
	}
	if strings.TrimRight(unverified.Issuer, "/") != realm.Issuer {
		return nil, errors.New("issuer id_token tidak sesuai")
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("algoritma id_token harus RS256")
		}
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("header kid id_token tidak tersedia")
		}
		return m.verifier.publicKey(ctx, realm, kid)
	}, jwt.WithIssuer(realm.Issuer), jwt.WithAudience(realm.ClientID), jwt.WithExpirationRequired(), jwt.WithLeeway(m.verifier.leeway))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("verifikasi id_token gagal: %w", err)
	}
	if claims.AuthorizedParty != "" && claims.AuthorizedParty != realm.ClientID {
		return nil, errors.New("authorized party id_token tidak sesuai")
	}
	return claims, nil
}
