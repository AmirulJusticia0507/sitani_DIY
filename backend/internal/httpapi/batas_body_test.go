package httpapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBacaJSONMenolakBodyBesar(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if ok, err := bacaJSON(w, r, &body); !ok || err != nil {
			respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
			return
		}
		respondJSON(w, http.StatusOK, body)
	}

	besar := []byte(`{"nama":"` + strings.Repeat("a", int(MaksBodyJSON)) + `"}`)
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/v1/uji", bytes.NewReader(besar)))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("body besar: status = %d, ingin 413 (%s)", rec.Code, rec.Body.String())
	}
	if !utf8.ValidString(rec.Body.String()) || !strings.Contains(rec.Body.String(), "melebihi batas") {
		t.Fatalf("balasan 413 tidak sesuai: %s", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/v1/uji", strings.NewReader(`{"nama":"tomat"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("body kecil: status = %d, ingin 200", rec.Code)
	}
}

func TestBacaJSONMembedakanJsonCacat(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if ok, err := bacaJSON(w, r, &body); !ok || err != nil {
			respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/v1/uji", strings.NewReader(`{"nama":`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("JSON cacat: status = %d, ingin 400", rec.Code)
	}
}

func TestBacaJSONMenjagaBodyKosongSebagaiEOF(t *testing.T) {
	// Endpoint draft memperlakukan io.EOF sebagai "tanpa isian"; balasan 413
	// tidak boleh menyamar menjadi EOF.
	handler := func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		ok, err := bacaJSON(w, r, &body)
		if !ok {
			return // balasan galat sudah ditulis
		}
		// errors.Is(err, io.EOF) berarti body kosong: sudah ditangani handler
		// draft sebagai "tanpa isian".
		if err != nil && err != io.EOF {
			respondError(w, http.StatusBadRequest, "body permintaan tidak valid")
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/v1/permohonan/ajukan", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("body kosong: status = %d, ingin 200", rec.Code)
	}

	besar := []byte(`{"jenis_benih":"` + strings.Repeat("b", int(MaksBodyJSON)) + `"}`)
	rec = httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/v1/permohonan/ajukan", bytes.NewReader(besar)))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("body besar: status = %d, ingin 413 (%s)", rec.Code, rec.Body.String())
	}
}

func TestPesanBatasBodyMenyebutBatasUnggah(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &http.MaxBytesError{Limit: 21 << 20}
	if !pesanBatasBody(rec, err) {
		t.Fatal("pesanBatasBody = false, ingin true untuk *http.MaxBytesError")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, ingin 413", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "21 MB") {
		t.Fatalf("pesan = %s, ingin menyebut 21 MB", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	if pesanBatasBody(rec, io.ErrUnexpectedEOF) {
		t.Fatal("galat non-batas tidak boleh dijawab 413")
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("balasan ditulis untuk galat non-batas: %s", rec.Body.String())
	}
}
