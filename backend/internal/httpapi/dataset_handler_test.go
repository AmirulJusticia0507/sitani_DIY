package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/dataset"
)

type fakeDatasetRepo struct {
	saved dataset.Import
	rows  dataset.Result
	all   bool
}

func (f *fakeDatasetRepo) Simpan(_ context.Context, userID int64, name, filename string, parsed dataset.Result) (dataset.Import, error) {
	f.rows = parsed
	f.saved = dataset.Import{ID: 9, NamaDataset: name, NamaFile: filename, JumlahBaris: len(parsed.Rows), DiunggahOleh: userID}
	return f.saved, nil
}

func (f *fakeDatasetRepo) Daftar(_ context.Context, _ int64, all bool) ([]dataset.Import, error) {
	f.all = all
	return []dataset.Import{f.saved}, nil
}

func datasetRequest(t *testing.T, role string) *http.Request {
	t.Helper()
	var body strings.Builder
	w := multipart.NewWriter(&body)
	_ = w.WriteField("nama_dataset", "Produksi pangan")
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="produksi.csv"`)
	h.Set("Content-Type", "text/csv")
	part, err := w.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(part, "komoditas,jumlah\nPadi,120\n")
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/v1/dataset/import", strings.NewReader(body.String()))
	r.Header.Set("Content-Type", w.FormDataContentType())
	return r.WithContext(auth.WithPrincipal(r.Context(), auth.Principal{UserID: 42, Role: role}))
}

func TestDatasetImportRequiresAuthorizedRole(t *testing.T) {
	repo := &fakeDatasetRepo{}
	h := &datasetHandler{repo: repo}
	w := httptest.NewRecorder()
	h.importData(w, datasetRequest(t, "pelanggan"))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
}

func TestDatasetImportCSV(t *testing.T) {
	repo := &fakeDatasetRepo{}
	h := &datasetHandler{repo: repo}
	w := httptest.NewRecorder()
	h.importData(w, datasetRequest(t, "analis"))
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
	if repo.saved.JumlahBaris != 1 || repo.saved.DiunggahOleh != 42 || repo.rows.Rows[0]["komoditas"] != "Padi" {
		t.Fatalf("data tersimpan salah: %+v rows=%+v", repo.saved, repo.rows)
	}
	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
}

func TestDatasetListScopesAdminAndAnalyst(t *testing.T) {
	for _, tc := range []struct {
		role string
		all  bool
	}{{"analis", false}, {"admin", true}} {
		repo := &fakeDatasetRepo{}
		h := &datasetHandler{repo: repo}
		r := httptest.NewRequest(http.MethodGet, "/v1/dataset/import", nil)
		r = r.WithContext(auth.WithPrincipal(r.Context(), auth.Principal{UserID: 42, Role: tc.role}))
		w := httptest.NewRecorder()
		h.listImports(w, r)
		if w.Code != http.StatusOK || repo.all != tc.all {
			t.Errorf("role %s: status=%d all=%v", tc.role, w.Code, repo.all)
		}
	}
}
