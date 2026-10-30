package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/service"
)

// permohonanHandlerUji menyiapkan handler dengan store palsu dan mengembalikan
// store-nya agar tes dapat memeriksa data yang tersimpan.
func permohonanHandlerUji() (*Handler, *palsuStore) {
	st := newPalsuStore()
	h := &Handler{mux: http.NewServeMux(), svc: service.NewService(st)}
	newCertificationHandler(h.svc).register(h.mux)
	return h, st
}

// jsonInt membentuk id permohonan untuk body/path JSON uji.
func jsonInt(id int64) string {
	return strconv.FormatInt(id, 10)
}

// validNomorPermohonan memeriksa format resmi SPB-YYYYMMDD-XXXX.
func validNomorPermohonan(nomor string) bool {
	return regexp.MustCompile(`^SPB-\d{8}-\d{4,}$`).MatchString(nomor)
}

func panggilJSON(t *testing.T, h *Handler, method, path, body string, userID, role string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
		req.Header.Set("X-User-Role", role)
	}
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, req)
	return w
}

func decodePermohonan(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var hasil map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &hasil); err != nil {
		t.Fatalf("decode respons: %v (body=%s)", err, w.Body.String())
	}
	return hasil
}

func TestSimpanDraftMengizinkanIsianKosong(t *testing.T) {
	h, _ := permohonanHandlerUji()

	w := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{}`, "10", "pelanggan")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body=%s", w.Code, w.Body.String())
	}
	got := decodePermohonan(t, w)
	if got["Status"] != "draft" {
		t.Fatalf("status permohonan = %v, ingin draft", got["Status"])
	}
	if nomor, _ := got["NomorPermohonan"].(string); strings.HasPrefix(nomor, "SPB-") {
		t.Fatalf("draft tidak boleh punya nomor resmi: %q", nomor)
	}
}

func TestSimpanDraftMemperbaruiDraftYangSama(t *testing.T) {
	h, st := permohonanHandlerUji()

	pertama := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{"jenis_benih":"Tomat"}`, "10", "pelanggan")
	if pertama.Code != http.StatusOK {
		t.Fatalf("draft pertama: %d %s", pertama.Code, pertama.Body.String())
	}
	draft := decodePermohonan(t, pertama)
	id := int64(draft["ID"].(float64))

	kedua := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft",
		`{"draft_id":`+jsonInt(id)+`,"jenis_benih":"Tomat","varietas":"Rama","luas_lahan":0.5,"lokasi_kebun":"Sleman"}`, "10", "pelanggan")
	if kedua.Code != http.StatusOK {
		t.Fatalf("draft kedua: %d %s", kedua.Code, kedua.Body.String())
	}
	if len(st.permohonan) != 1 {
		t.Fatalf("permohonan = %d baris, ingin 1", len(st.permohonan))
	}
	if decodePermohonan(t, kedua)["Varietas"] != "Rama" {
		t.Fatalf("varietas tidak tersimpan: %s", kedua.Body.String())
	}
}

func TestSimpanDraftBukanPemilikDitolak(t *testing.T) {
	h, _ := permohonanHandlerUji()

	draft := decodePermohonan(t, panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{"jenis_benih":"Tomat"}`, "10", "pelanggan"))
	id := jsonInt(int64(draft["ID"].(float64)))

	w := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{"draft_id":`+id+`,"jenis_benih":"Padi"}`, "11", "pelanggan")
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, ingin 404; body=%s", w.Code, w.Body.String())
	}
}

func TestAjukanPermohonanMemberiNomorResmi(t *testing.T) {
	h, st := permohonanHandlerUji()

	draft := decodePermohonan(t, panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{"jenis_benih":"Tomat"}`, "10", "pelanggan"))
	id := int64(draft["ID"].(float64))

	w := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/ajukan",
		`{"draft_id":`+jsonInt(id)+`,"jenis_benih":"Tomat","varietas":"Rama","luas_lahan":0.5,"lokasi_kebun":"Sleman"}`, "10", "pelanggan")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body=%s", w.Code, w.Body.String())
	}
	got := decodePermohonan(t, w)
	if got["Status"] != "diajukan" {
		t.Fatalf("status = %v, ingin diajukan", got["Status"])
	}
	nomor, _ := got["NomorPermohonan"].(string)
	if !validNomorPermohonan(nomor) {
		t.Fatalf("nomor = %q, ingin format SPB-YYYYMMDD-XXXX", nomor)
	}
	if st.permohonan[id].NomorPermohonan != nomor {
		t.Fatalf("nomor respons berbeda dari yang tersimpan")
	}

	// Pengajuan kedua atas draft yang sama ditolak dengan 409.
	ulang := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/ajukan",
		`{"draft_id":`+jsonInt(id)+`,"jenis_benih":"Tomat","varietas":"Rama","luas_lahan":0.5,"lokasi_kebun":"Sleman"}`, "10", "pelanggan")
	if ulang.Code != http.StatusConflict {
		t.Fatalf("ajukan ulang: status = %d, ingin 409; body=%s", ulang.Code, ulang.Body.String())
	}
}

// Pengajuan tanpa draft tersimpan (draft_id 0) langsung membuat dan mengajukan
// permohonan; pengajuan kedua saat masih berjalan ditolak 409.
func TestAjukanPermohonanLangsungTanpaDraft(t *testing.T) {
	h, st := permohonanHandlerUji()

	w := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/ajukan",
		`{"jenis_benih":"Tomat","varietas":"Rama","luas_lahan":0.5,"lokasi_kebun":"Sleman"}`, "10", "pelanggan")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body=%s", w.Code, w.Body.String())
	}
	got := decodePermohonan(t, w)
	if got["Status"] != "diajukan" {
		t.Fatalf("status = %v, ingin diajukan", got["Status"])
	}
	if nomor, _ := got["NomorPermohonan"].(string); !validNomorPermohonan(nomor) {
		t.Fatalf("nomor = %q, ingin format SPB-YYYYMMDD-XXXX", nomor)
	}
	if len(st.permohonan) != 1 {
		t.Fatalf("permohonan = %d baris, ingin 1", len(st.permohonan))
	}

	ulang := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/ajukan",
		`{"jenis_benih":"Padi","varietas":"IR64","luas_lahan":1,"lokasi_kebun":"Bantul"}`, "10", "pelanggan")
	if ulang.Code != http.StatusConflict {
		t.Fatalf("ajukan saat masih berjalan: status = %d, ingin 409; body=%s", ulang.Code, ulang.Body.String())
	}
}

func TestAjukanPermohonanIsianTidakLengkapDitolak(t *testing.T) {
	h, _ := permohonanHandlerUji()

	draft := decodePermohonan(t, panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{"jenis_benih":"Tomat"}`, "10", "pelanggan"))
	id := jsonInt(int64(draft["ID"].(float64)))

	w := panggilJSON(t, h, http.MethodPost, "/v1/permohonan/ajukan", `{"draft_id":`+id+`,"jenis_benih":"Tomat"}`, "10", "pelanggan")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400; body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "varietas wajib") {
		t.Fatalf("pesan = %s", w.Body.String())
	}
}

func TestPermohonanAktifKosongMengembalikan204(t *testing.T) {
	h, _ := permohonanHandlerUji()

	w := panggilJSON(t, h, http.MethodGet, "/v1/permohonan/aktif", "", "10", "pelanggan")
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, ingin 204; body=%s", w.Code, w.Body.String())
	}
}

func TestPermohonanAktifMengembalikanDraft(t *testing.T) {
	h, _ := permohonanHandlerUji()

	draft := decodePermohonan(t, panggilJSON(t, h, http.MethodPost, "/v1/permohonan/draft", `{"jenis_benih":"Tomat"}`, "10", "pelanggan"))
	w := panggilJSON(t, h, http.MethodGet, "/v1/permohonan/aktif", "", "10", "pelanggan")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body=%s", w.Code, w.Body.String())
	}
	if decodePermohonan(t, w)["ID"] != draft["ID"] {
		t.Fatalf("draft aktif tidak sesuai: %s", w.Body.String())
	}

	// Pelanggan lain tidak melihat draft tersebut.
	if lain := panggilJSON(t, h, http.MethodGet, "/v1/permohonan/aktif", "", "11", "pelanggan"); lain.Code != http.StatusNoContent {
		t.Fatalf("pelanggan lain: status = %d, ingin 204", lain.Code)
	}
}

func TestPermohonanDraftMenolakTanpaIdentitas(t *testing.T) {
	h, _ := permohonanHandlerUji()

	for _, jalur := range []struct{ method, path string }{
		{http.MethodPost, "/v1/permohonan/draft"},
		{http.MethodPost, "/v1/permohonan/ajukan"},
		{http.MethodGet, "/v1/permohonan/aktif"},
	} {
		w := panggilJSON(t, h, jalur.method, jalur.path, `{}`, "", "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: status = %d, ingin 401", jalur.method, jalur.path, w.Code)
		}
	}
}

// Peran internal yang membuka halaman sertifikasi tetap memanggil endpoint
// alur pelanggan; jawabannya harus 403 (bukan 400) agar frontend membedakan
// peran yang salah dari isian yang tidak valid.
func TestAlurDraftDitolakUntukRoleInternal(t *testing.T) {
	h, _ := permohonanHandlerUji()

	for _, jalur := range []struct{ method, path string }{
		{http.MethodPost, "/v1/permohonan/draft"},
		{http.MethodPost, "/v1/permohonan/ajukan"},
		{http.MethodGet, "/v1/permohonan/aktif"},
	} {
		body := ""
		if jalur.method == http.MethodPost {
			body = `{"jenis_benih":"Tomat"}`
		}
		w := panggilJSON(t, h, jalur.method, jalur.path, body, "3", "analis")
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s %s: status = %d, ingin 403; body=%s", jalur.method, jalur.path, w.Code, w.Body.String())
		}
	}
}
