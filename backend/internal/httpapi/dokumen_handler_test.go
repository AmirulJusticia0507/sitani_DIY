package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/storage"
)

// objekPenyimpananPalsu merekam unggahan dan mengembalikan URL tiruan.
type objekPenyimpananPalsu struct {
	objek map[string][]byte
}

func newObjekPenyimpananPalsu() *objekPenyimpananPalsu {
	return &objekPenyimpananPalsu{objek: map[string][]byte{}}
}

func (s *objekPenyimpananPalsu) Unggah(_ context.Context, objectKey string, reader io.Reader, _ int64, _ string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.objek[objectKey] = data
	return nil
}

func (s *objekPenyimpananPalsu) URLUnduh(_ context.Context, objectKey string) (string, error) {
	if _, ok := s.objek[objectKey]; !ok {
		return "", errors.New("objek tidak ada")
	}
	return "https://storage.example.test/unduh/" + objectKey, nil
}

func (s *objekPenyimpananPalsu) Periksa(_ context.Context) error { return nil }

func newDokumenHandlerUji(t *testing.T) (*Handler, *objekPenyimpananPalsu) {
	t.Helper()
	st := newPalsuStore()
	pen := newObjekPenyimpananPalsu()
	h := &Handler{
		mux: http.NewServeMux(),
		svc: nil,
	}
	h.dokumen = newDokumenHandler(st, pen, storage.Config{UkuranMaks: 1 << 20})
	h.dokumen.register(h.mux)
	return h, pen
}

func newDokumenHandlerTanpaPenyimpanan() *Handler {
	h := &Handler{mux: http.NewServeMux()}
	h.dokumen = newDokumenHandler(newPalsuStore(), nil, storage.Config{})
	h.dokumen.register(h.mux)
	return h
}

// unggahDokumenUji mengirim permintaan multipart dengan berkas berisi isi.
// Content-Type part diset eksplisit meniru klien nyata.
func unggahDokumenUji(t *testing.T, h *Handler, userID, role, nama, jenis, isi string) *httptest.ResponseRecorder {
	t.Helper()
	var tubuh strings.Builder
	writer := multipart.NewWriter(&tubuh)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename=%q`, nama))
	header.Set("Content-Type", jenis)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("buat part: %v", err)
	}
	if _, err := io.WriteString(part, isi); err != nil {
		t.Fatalf("tulis isi: %v", err)
	}
	_ = writer.WriteField("tujuan", "uji")
	if err := writer.Close(); err != nil {
		t.Fatalf("tutup writer: %v", err)
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/dokumen", strings.NewReader(tubuh.String()))
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("X-User-ID", userID)
	r.Header.Set("X-User-Role", role)
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	return w
}

func TestUnggahDokumenSukses(t *testing.T) {
	h, pen := newDokumenHandlerUji(t)
	w := unggahDokumenUji(t, h, "10", "pelanggan", "proposal.pdf", "application/pdf", "%PDF-1.7 isi-proposal")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, ingin %d; body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode respons: %v", err)
	}
	if got["object_key"] == nil || !strings.Contains(got["object_key"].(string), "dokumen/10/") {
		t.Fatalf("object_key tidak memisah per pemilik: %v", got)
	}
	if got["ukuran"] != float64(len("%PDF-1.7 isi-proposal")) {
		t.Fatalf("ukuran tidak sesuai: %v", got)
	}
	if len(pen.objek) != 1 {
		t.Fatalf("berkas tidak tersimpan di object storage")
	}
}

func TestUnggahDokumenJenisTidakDiizinkan(t *testing.T) {
	h, _ := newDokumenHandlerUji(t)
	w := unggahDokumenUji(t, h, "10", "pelanggan", "virus.exe", "application/octet-stream", "data")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400; body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "tidak diizinkan") {
		t.Fatalf("pesan tidak menjelaskan jenis berkas: %s", w.Body.String())
	}
}

func TestUnggahDokumenTanpaPenyimpanan503(t *testing.T) {
	h := newDokumenHandlerTanpaPenyimpanan()
	w := unggahDokumenUji(t, h, "10", "pelanggan", "a.pdf", "application/pdf", "isi")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, ingin 503; body=%s", w.Code, w.Body.String())
	}
}

func TestUnduhDokumenOlehPemilik(t *testing.T) {
	h, _ := newDokumenHandlerUji(t)
	_ = unggahDokumenUji(t, h, "10", "pelanggan", "a.pdf", "application/pdf", "%PDF-1.7 x")

	r := httptest.NewRequest(http.MethodGet, "/v1/dokumen/1/unduh", nil)
	r.SetPathValue("id", "1")
	r.Header.Set("X-User-ID", "10")
	r.Header.Set("X-User-Role", "pelanggan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body=%s", w.Code, w.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["url"] == nil || !strings.Contains(got["url"].(string), "dokumen/10/") {
		t.Fatalf("URL unduh tidak memuat kunci objek: %v", got)
	}
}

func TestUnduhDokumenPelangganLainDitolak(t *testing.T) {
	h, _ := newDokumenHandlerUji(t)
	_ = unggahDokumenUji(t, h, "10", "pelanggan", "a.pdf", "application/pdf", "%PDF-1.7 x")

	r := httptest.NewRequest(http.MethodGet, "/v1/dokumen/1/unduh", nil)
	r.SetPathValue("id", "1")
	r.Header.Set("X-User-ID", "77")
	r.Header.Set("X-User-Role", "pelanggan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, ingin 403; body=%s", w.Code, w.Body.String())
	}
}

func TestUnduhDokumenOlehPetugasInternal(t *testing.T) {
	h, _ := newDokumenHandlerUji(t)
	_ = unggahDokumenUji(t, h, "10", "pelanggan", "a.pdf", "application/pdf", "%PDF-1.7 x")

	r := httptest.NewRequest(http.MethodGet, "/v1/dokumen/1/unduh", nil)
	r.SetPathValue("id", "1")
	r.Header.Set("X-User-ID", "20")
	r.Header.Set("X-User-Role", "pengawas")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin 200; body=%s", w.Code, w.Body.String())
	}
}

func TestLuangkanDokumenHanyaMilikUntukPelanggan(t *testing.T) {
	h, _ := newDokumenHandlerUji(t)
	_ = unggahDokumenUji(t, h, "10", "pelanggan", "milik10.pdf", "application/pdf", "%PDF-1.7 s10")
	_ = unggahDokumenUji(t, h, "11", "pelanggan", "milik11.pdf", "application/pdf", "%PDF-1.7 s11")

	// Pelanggan 10 hanya melihat dokumennya sendiri.
	r := httptest.NewRequest(http.MethodGet, "/v1/dokumen", nil)
	r.Header.Set("X-User-ID", "10")
	r.Header.Set("X-User-Role", "pelanggan")
	w := httptest.NewRecorder()
	h.mux.ServeHTTP(w, r)
	var got struct {
		Dokumen []map[string]any `json:"dokumen"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if len(got.Dokumen) != 1 {
		t.Fatalf("pelanggan melihat %d dokumen, ingin 1", len(got.Dokumen))
	}

	// Admin melihat seluruh dokumen.
	r2 := httptest.NewRequest(http.MethodGet, "/v1/dokumen", nil)
	r2.Header.Set("X-User-ID", "1")
	r2.Header.Set("X-User-Role", "admin")
	w2 := httptest.NewRecorder()
	h.mux.ServeHTTP(w2, r2)
	var got2 struct {
		Dokumen []map[string]any `json:"dokumen"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &got2)
	if len(got2.Dokumen) != 2 {
		t.Fatalf("admin melihat %d dokumen, ingin 2", len(got2.Dokumen))
	}
}

func TestKunciObjekUnikAntarUnggahan(t *testing.T) {
	h, pen := newDokumenHandlerUji(t)
	w1 := unggahDokumenUji(t, h, "10", "pelanggan", "a.pdf", "application/pdf", "%PDF-1.7 x")
	w2 := unggahDokumenUji(t, h, "10", "pelanggan", "a.pdf", "application/pdf", "%PDF-1.7 x")
	if w1.Code != http.StatusCreated || w2.Code != http.StatusCreated {
		t.Fatalf("unggahan gagal: %d=%s, %d=%s", w1.Code, w1.Body.String(), w2.Code, w2.Body.String())
	}
	if len(pen.objek) != 2 {
		t.Fatalf("jumlah objek = %d, ingin 2 (kunci unik per unggahan)", len(pen.objek))
	}
}
