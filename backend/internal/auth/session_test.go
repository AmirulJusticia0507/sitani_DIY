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
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// serverSSO meniru token endpoint (authorization_code, refresh_token) dan
// revocation endpoint Keycloak, sekaligus menyajikan JWKS untuk verifikasi.
type serverSSO struct {
	*httptest.Server
	kunci *rsa.PrivateKey

	mu         sync.Mutex
	permintaan []url.Values
	// responsToken menggantikan respons token endpoint bila diisi.
	responsToken func(values url.Values) (status int, body map[string]any)
	statusRevoke int
}

func newServerSSO(t *testing.T) *serverSSO {
	t.Helper()
	kunci, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	s := &serverSSO{kunci: kunci, statusRevoke: http.StatusOK}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /protocol/openid-connect/certs", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(keySet{Keys: []jwk{{
			Kid: "ujian", Kty: "RSA", Alg: "RS256", Use: "sig",
			N: base64.RawURLEncoding.EncodeToString(kunci.PublicKey.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(kunci.PublicKey.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("POST /protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		s.mu.Lock()
		s.permintaan = append(s.permintaan, r.PostForm)
		fn := s.responsToken
		s.mu.Unlock()
		if fn != nil {
			status, body := fn(r.PostForm)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(body)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			// Umur token sengaja lebih panjang daripada ekspektasi skew di tes
			// pembaruan, sehingga tes benar-benar menguji token hasil refresh.
			"access_token":  s.tokenAkses(t, "sub-1", 5*time.Minute),
			"refresh_token": "refresh-baru", "id_token": s.idToken(t, "sub-1", "sid-1"),
			"token_type": "Bearer", "expires_in": 300,
		})
	})
	mux.HandleFunc("POST /protocol/openid-connect/revoke", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		s.mu.Lock()
		s.permintaan = append(s.permintaan, r.PostForm)
		status := s.statusRevoke
		s.mu.Unlock()
		w.WriteHeader(status)
	})
	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func (s *serverSSO) tokenAkses(t *testing.T, sub string, umur time.Duration) string {
	t.Helper()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.URL, Subject: sub, Audience: jwt.ClaimStrings{"account"},
			IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(umur)),
		},
		AuthorizedParty: "sitani-portal", Email: "analis@dev.example.invalid", EmailVerified: true,
	}
	return s.tandaTangan(t, claims)
}

func (s *serverSSO) idToken(t *testing.T, sub, sid string) string {
	t.Helper()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: s.URL, Subject: sub, Audience: jwt.ClaimStrings{"sitani-portal"},
			IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
		AuthorizedParty: "sitani-portal", Email: "analis@dev.example.invalid", SessionID: sid,
	}
	return s.tandaTangan(t, claims)
}

func (s *serverSSO) tandaTangan(t *testing.T, claims Claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "ujian"
	raw, err := token.SignedString(s.kunci)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func (s *serverSSO) realm() Realm {
	return Realm{Name: "jsp", Issuer: s.URL, ClientID: "sitani-portal", ClientSecret: "rahasia"}
}

func (s *serverSSO) terakhir(path string) url.Values {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.permintaan) - 1; i >= 0; i-- {
		if strings.HasSuffix(path, "/token") && s.permintaan[i].Get("grant_type") != "" {
			return s.permintaan[i]
		}
	}
	return url.Values{}
}

func requestDenganCookie(metode, target string, cookie ...*http.Cookie) *http.Request {
	r := httptest.NewRequest(metode, target, nil)
	for _, c := range cookie {
		r.AddCookie(c)
	}
	return r
}

// Callback yang tokennya valid tetapi database identitas belum siap harus
// menjawab 503 dengan pesan jelas, bukan panic dan bukan 401 yang menyesatkan.
func TestCallbackTanpaDatabaseMenjawab503(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{
		Realms:      map[string]Realm{"jsp": s.realm()},
		FrontendURL: "http://localhost:3001", BackendURL: "http://localhost:8080",
	}, nil)
	mux := http.NewServeMux()
	manager.RegisterRoutes(mux)

	r := requestDenganCookie(http.MethodGet, "/auth/callback/jsp?state=st&code=kode",
		&http.Cookie{Name: "sitani_sso_state_jsp", Value: "st"},
		&http.Cookie{Name: "sitani_sso_verifier_jsp", Value: "verifier"})
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin %d: %s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	if got := s.terakhir("/token").Get("grant_type"); got != "authorization_code" {
		t.Fatalf("grant_type = %q", got)
	}
	if cookies := recorder.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("callback gagal tidak boleh menulis cookie sesi, dapat %d", len(cookies))
	}
}

// Setelah pemetaan identitas berhasil, cookie sesi yang ditulis harus lengkap:
// access token di path /, sedangkan refresh dan id token dibatasi ke /auth/.
func TestSimpanSesiMenulisCookieTerbatas(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": s.realm()}, SecureCookie: true}, nil)
	recorder := httptest.NewRecorder()
	manager.simpanSesi(recorder, s.realm(), sesiSSO{
		token: s.tokenAkses(t, "sub-1", time.Minute), refresh: "refresh-1", idToken: s.idToken(t, "sub-1", "sid-1"),
	}, 300)

	perNama := map[string]*http.Cookie{}
	for _, c := range recorder.Result().Cookies() {
		perNama[c.Name] = c
	}
	if len(perNama) != 4 {
		t.Fatalf("jumlah cookie = %d, ingin 4 (access, realm, refresh, id)", len(perNama))
	}
	for _, nama := range []string{accessTokenCookie, realmCookie, refreshTokenCookie, idTokenCookie} {
		cookie, ok := perNama[nama]
		if !ok {
			t.Fatalf("cookie %s tidak ditulis", nama)
		}
		if !cookie.HttpOnly || !cookie.Secure {
			t.Fatalf("cookie %s harus HttpOnly dan Secure: %+v", nama, cookie)
		}
		inginPath := tokenPath
		if nama == accessTokenCookie || nama == realmCookie {
			inginPath = "/"
		}
		if cookie.Path != inginPath {
			t.Fatalf("cookie %s path = %q, ingin %q", nama, cookie.Path, inginPath)
		}
	}
	if perNama[realmCookie].Value != "jsp" {
		t.Fatalf("cookie realm = %q", perNama[realmCookie].Value)
	}
}

// Access token yang sudah kedaluwarsa tetap dapat diverifikasi terhadap realm
// yang diketahui (cookie realm) sehingga refresh token dapat dipakai, tetapi
// token dari issuer lain tetap ditolak.
func TestVerifierVerifyDenganRealmMenolakIssuerLain(t *testing.T) {
	server := newServerSSO(t)
	realm := server.realm()
	verifier := NewVerifier(map[string]Realm{"jsp": realm})

	segar := server.tokenAkses(t, "sub-1", time.Minute)
	if _, got, err := verifier.Verify(context.Background(), segar, realm); err != nil || got.Name != "jsp" {
		t.Fatalf("token valid ditolak: %v %v", err, got)
	}

	kedaluwarsa := server.tokenAkses(t, "sub-1", -5*time.Minute)
	if _, _, err := verifier.Verify(context.Background(), kedaluwarsa, realm); err == nil {
		t.Fatal("token kedaluwarsa seharusnya ditolak verifikasi")
	}

	lain := Realm{Name: "jsp", Issuer: "https://sso.example/realms/jsp", ClientID: "sitani-portal"}
	if _, _, err := verifier.Verify(context.Background(), segar, lain); err == nil {
		t.Fatal("token untuk issuer lain seharusnya ditolak")
	}
}

// Cookie realm dan refresh tetap terbaca walaupun access token sudah
// kedaluwarsa, sehingga sesi dapat diperpanjang tanpa login ulang.
func TestSesiDariCookieMemakaiCookieRealmSaatTokenKedaluwarsa(t *testing.T) {
	server := newServerSSO(t)
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": server.realm()}}, nil)
	r := requestDenganCookie(http.MethodGet, "/",
		&http.Cookie{Name: accessTokenCookie, Value: server.tokenAkses(t, "sub-1", -time.Minute)},
		&http.Cookie{Name: refreshTokenCookie, Value: "refresh-lama"},
		&http.Cookie{Name: realmCookie, Value: "jsp"},
	)
	session, err := manager.sesiDariCookie(r)
	if err == nil {
		t.Fatal("access token kedaluwarsa seharusnya menghasilkan kesalahan verifikasi")
	}
	if session.realm.Name != "jsp" || session.refresh != "refresh-lama" {
		t.Fatalf("realm/refresh tidak terbaca: %+v", session)
	}
	if !session.perluDiperbarui(DefaultRefreshSkew) {
		t.Fatal("sesi dengan token kedaluwarsa harus dianggap perlu diperbarui")
	}
}

func TestIssuerAliasHanyaMenerimaHostLokalSetara(t *testing.T) {
	diterima := []struct{ token, konfigurasi string }{
		{"http://localhost:8085/realms/sitani", "http://host.docker.internal:8085/realms/sitani"},
		{"http://host.docker.internal:8085/realms/sitani", "http://localhost:8085/realms/sitani"},
		{"http://127.0.0.1:8085/realms/sitani", "http://host.docker.internal:8085/realms/sitani"},
	}
	for _, kasus := range diterima {
		if !issuerAlias(kasus.token, kasus.konfigurasi) {
			t.Fatalf("alias seharusnya diterima: %s vs %s", kasus.token, kasus.konfigurasi)
		}
	}
	ditolak := []struct{ token, konfigurasi string }{
		{"http://localhost:8085/realms/jsp", "http://host.docker.internal:8085/realms/sitani"},
		{"http://localhost:8086/realms/sitani", "http://host.docker.internal:8085/realms/sitani"},
		{"https://localhost:8085/realms/sitani", "http://host.docker.internal:8085/realms/sitani"},
		{"https://sso.jogjaprov.go.id/realms/jsp", "https://sso.jogjaprov.go.id/realms/satriya"},
		{"https://jahat.example/realms/sitani", "https://sso.jogjaprov.go.id/realms/sitani"},
		{"bukan-url", "http://localhost:8085/realms/sitani"},
	}
	for _, kasus := range ditolak {
		if issuerAlias(kasus.token, kasus.konfigurasi) {
			t.Fatalf("issuer seharusnya ditolak: %s vs %s", kasus.token, kasus.konfigurasi)
		}
	}
}

func TestVerifikasiIDTokenMenerimaAudClientDanMenolakClientLain(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": s.realm()}}, nil)

	claims, err := manager.verifikasiIDToken(context.Background(), s.idToken(t, "sub-1", "sid-1"), s.realm())
	if err != nil {
		t.Fatalf("id_token valid ditolak: %v", err)
	}
	if claims.SessionID != "sid-1" {
		t.Fatalf("sid = %q", claims.SessionID)
	}

	lain := s.realm()
	lain.ClientID = "client-lain"
	if _, err := manager.verifikasiIDToken(context.Background(), s.idToken(t, "sub-1", "sid-1"), lain); err == nil {
		t.Fatal("id_token untuk client lain seharusnya ditolak")
	}
}

// Access token yang hampir kedaluwarsa diperbarui memakai refresh token, dan
// cookie baru dikirim ke browser.
func TestRequireMemperbaruiTokenHampirKedaluwarsa(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{
		Realms:      map[string]Realm{"jsp": s.realm()},
		FrontendURL: "http://localhost:3001",
		RefreshSkew: 2 * time.Minute,
	}, nil)
	// Require memerlukan database untuk memetakan principal, jadi hanya logika
	// pembaruan sesi yang diuji langsung di sini.
	session := sesiSSO{
		token: s.tokenAkses(t, "sub-1", 30*time.Second), refresh: "refresh-lama", realm: s.realm(),
	}
	recorder := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/dokumen", nil)
	baru, err := manager.perbaruiSesi(recorder, r, session)
	if err != nil {
		t.Fatalf("perbaruiSesi: %v", err)
	}
	if baru.refresh != "refresh-baru" {
		t.Fatalf("refresh token = %q, ingin hasil rotasi", baru.refresh)
	}
	if baru.perluDiperbarui(2 * time.Minute) {
		t.Fatal("token hasil refresh seharusnya masih baru")
	}
	nama := map[string]bool{}
	for _, c := range recorder.Result().Cookies() {
		nama[c.Name] = true
		if c.Name == refreshTokenCookie && c.Path != tokenPath {
			t.Fatalf("cookie refresh harus dibatasi ke %s, dapat %s", tokenPath, c.Path)
		}
	}
	for _, harus := range []string{accessTokenCookie, refreshTokenCookie, idTokenCookie} {
		if !nama[harus] {
			t.Fatalf("cookie %s tidak dikirim", harus)
		}
	}
	if got := s.terakhir("/token").Get("grant_type"); got != "refresh_token" {
		t.Fatalf("grant_type = %q", got)
	}
	if got := s.terakhir("/token").Get("client_secret"); got != "rahasia" {
		t.Fatalf("client_secret refresh = %q", got)
	}
}

func TestPerluDiperbaruiMenggunakanSisaUmur(t *testing.T) {
	skew := 60 * time.Second
	baru := sesiSSO{kedaluwarsa: time.Now().Add(5 * time.Minute)}
	lama := sesiSSO{kedaluwarsa: time.Now().Add(20 * time.Second)}
	tanpaExp := sesiSSO{}
	if baru.perluDiperbarui(skew) {
		t.Fatal("token baru tidak perlu diperbarui")
	}
	if !lama.perluDiperbarui(skew) {
		t.Fatal("token hampir kedaluwarsa harus diperbarui")
	}
	if !tanpaExp.perluDiperbarui(skew) {
		t.Fatal("token tanpa exp harus dianggap perlu diperbarui")
	}
}

// Refresh token yang ditolak penyedia identitas membuat sesi berakhir dan
// cookie dibersihkan, bukan menghasilkan 500.
func TestPerbaruiSesiGagalMenolakSuratRefreshDicabut(t *testing.T) {
	s := newServerSSO(t)
	s.responsToken = func(url.Values) (int, map[string]any) {
		return http.StatusBadRequest, map[string]any{"error": "invalid_grant", "error_description": "Token is not active"}
	}
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": s.realm()}}, nil)
	session := sesiSSO{token: "lama", refresh: "refresh-dicabut", realm: s.realm()}
	if _, err := manager.perbaruiSesi(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), session); err == nil {
		t.Fatal("refresh yang ditolak seharusnya mengembalikan kesalahan")
	}
}

// Logout mencabut refresh token, menghapus cookie, dan memberi URL
// end_session_endpoint berisi id_token_hint, sid, serta redirect kembali.
func TestLogoutMencabutTokenDanMenyusunEndSession(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{
		Realms:      map[string]Realm{"jsp": s.realm()},
		FrontendURL: "http://localhost:3001", BackendURL: "http://localhost:8080",
	}, nil)
	mux := http.NewServeMux()
	manager.RegisterRoutes(mux)

	idToken := s.idToken(t, "sub-1", "sid-9")
	r := requestDenganCookie(http.MethodPost, "/auth/logout",
		&http.Cookie{Name: accessTokenCookie, Value: s.tokenAkses(t, "sub-1", time.Minute)},
		&http.Cookie{Name: refreshTokenCookie, Value: "refresh-aktif"},
		&http.Cookie{Name: idTokenCookie, Value: idToken},
	)
	r.Header.Set("Origin", "http://localhost:3001")
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", recorder.Code, recorder.Body.String())
	}
	dicabut := false
	s.mu.Lock()
	for _, v := range s.permintaan {
		if v.Get("token") == "refresh-aktif" && v.Get("client_id") == "sitani-portal" && v.Get("client_secret") == "rahasia" {
			dicabut = true
		}
	}
	s.mu.Unlock()
	if !dicabut {
		t.Fatal("refresh token tidak dicabut di revocation endpoint")
	}

	var body struct {
		LogoutURL string `json:"logout_url"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(body.LogoutURL, s.URL+"/protocol/openid-connect/logout?") {
		t.Fatalf("logout_url = %q", body.LogoutURL)
	}
	u, err := url.Parse(body.LogoutURL)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("id_token_hint") != idToken || q.Get("sid") != "sid-9" {
		t.Fatalf("parameter logout tidak lengkap: %v", q)
	}
	if q.Get("post_logout_redirect_uri") != "http://localhost:3001/masuk" || q.Get("client_id") != "sitani-portal" {
		t.Fatalf("redirect logout = %v", q)
	}

	dihapus := map[string]bool{}
	for _, c := range recorder.Result().Cookies() {
		if c.MaxAge < 0 {
			dihapus[c.Name] = true
		}
	}
	for _, harus := range []string{accessTokenCookie, refreshTokenCookie, idTokenCookie} {
		if !dihapus[harus] {
			t.Fatalf("cookie %s tidak dihapus saat logout", harus)
		}
	}
}

func TestLogoutOriginAsingTidakMencabutToken(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": s.realm()}, FrontendURL: "http://localhost:3001"}, nil)
	r := requestDenganCookie(http.MethodPost, "/auth/logout",
		&http.Cookie{Name: refreshTokenCookie, Value: "refresh-aktif"})
	r.Header.Set("Origin", "https://jahat.example")
	recorder := httptest.NewRecorder()
	manager.logout(recorder, r)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusForbidden)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.permintaan) != 0 {
		t.Fatal("logout dengan origin asing tidak boleh mencabut token")
	}
}

func TestLogoutTanpaCookieTidakMemanggilRevoke(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": s.realm()}, FrontendURL: "http://localhost:3001"}, nil)
	r := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	r.Header.Set("Origin", "http://localhost:3001")
	recorder := httptest.NewRecorder()
	manager.logout(recorder, r)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	var body struct {
		LogoutURL string `json:"logout_url"`
	}
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	if body.LogoutURL != "" {
		t.Fatalf("logout_url = %q, ingin kosong", body.LogoutURL)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.permintaan) != 0 {
		t.Fatal("revocation tidak boleh dipanggil tanpa refresh token")
	}
}

func TestRefreshMenolakOriginAsing(t *testing.T) {
	s := newServerSSO(t)
	manager := NewManager(Config{Realms: map[string]Realm{"jsp": s.realm()}, FrontendURL: "http://localhost:3001"}, nil)
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.Header.Set("Origin", "https://jahat.example")
	recorder := httptest.NewRecorder()
	manager.refresh(recorder, r)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, ingin %d", recorder.Code, http.StatusForbidden)
	}
}

func TestConfigMembacaPengaturanRefreshDanLeeway(t *testing.T) {
	t.Setenv("SSO_REFRESH_SKEW_SECONDS", "30")
	t.Setenv("SSO_TOKEN_LEEWAY_SECONDS", "5")
	config := ConfigFromEnv()
	if config.refreshSkew() != 30*time.Second {
		t.Fatalf("refreshSkew = %v", config.refreshSkew())
	}
	if config.leeway() != 5*time.Second {
		t.Fatalf("leeway = %v", config.leeway())
	}
	// Nilai kosong memakai default; 0 tidak mematikan pembaruan token.
	t.Setenv("SSO_REFRESH_SKEW_SECONDS", "0")
	if got := ConfigFromEnv().refreshSkew(); got != DefaultRefreshSkew {
		t.Fatalf("refreshSkew 0 = %v, ingin default %v", got, DefaultRefreshSkew)
	}
}
