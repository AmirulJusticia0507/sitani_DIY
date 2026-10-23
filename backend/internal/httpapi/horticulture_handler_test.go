package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/horticulture"
)

type fakeHorticultureRepository struct {
	created horticulture.Application
	docs    []horticulture.Document
	items   []horticulture.Application
}

func (f *fakeHorticultureRepository) Create(_ context.Context, app horticulture.Application, docs []horticulture.Document) (horticulture.Application, error) {
	app.ID, app.Number, app.Status = 7, "RH-2026-ABC12345", "diajukan"
	app.Documents = docs
	f.created, f.docs = app, docs
	return app, nil
}

func (f *fakeHorticultureRepository) List(_ context.Context, applicantID int64, all bool) ([]horticulture.Application, error) {
	if !all {
		items := make([]horticulture.Application, 0)
		for _, app := range f.items {
			if app.ApplicantID == applicantID {
				items = append(items, app)
			}
		}
		return items, nil
	}
	return f.items, nil
}

func (f *fakeHorticultureRepository) Get(_ context.Context, id int64) (horticulture.Application, error) {
	for _, app := range f.items {
		if app.ID == id {
			return app, nil
		}
	}
	return horticulture.Application{}, horticulture.ErrNotFound
}

func makeHorticultureRequest(body string, role string, userID int64) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/v1/horticulture/applications", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r.WithContext(auth.WithPrincipal(r.Context(), auth.Principal{UserID: userID, Role: role}))
}

const validHorticultureRequest = `{"nama_organisasi":"Kelompok Tani Makmur","nama_kontak":"Siti","telepon":"081234567890","alamat":"Jalan Tani 1","kalurahan":"Sumber","kapanewon":"Sleman","kabupaten_kota":"Sleman","luas_lahan_ha":2.5,"tanaman":[{"nama":"Cabai","varietas":"Rawit","jumlah_tanaman":1000,"sumber_benih":"Bersertifikat"}],"dokumen":[{"id":1,"jenis":"proposal"},{"id":2,"jenis":"organization"},{"id":3,"jenis":"members"},{"id":4,"jenis":"farm_form"},{"id":5,"jenis":"farm_map"}]}`

func TestHorticultureCreateRequiresVerifiedApplicant(t *testing.T) {
	repo := &fakeHorticultureRepository{}
	h := &horticultureHandler{repo: repo}
	for _, tc := range []struct {
		name string
		role string
		id   int64
		want int
	}{{"no principal", "", 0, http.StatusUnauthorized}, {"wrong role", "analis", 12, http.StatusForbidden}} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/v1/horticulture/applications", strings.NewReader(validHorticultureRequest))
			if tc.id > 0 {
				r = r.WithContext(auth.WithPrincipal(r.Context(), auth.Principal{UserID: tc.id, Role: tc.role}))
			}
			w := httptest.NewRecorder()
			h.create(w, r)
			if w.Code != tc.want {
				t.Fatalf("status=%d, want=%d, body=%s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestHorticultureCreateValidatesRequiredDocumentsAndPersists(t *testing.T) {
	repo := &fakeHorticultureRepository{}
	h := &horticultureHandler{repo: repo}
	invalid := strings.Replace(validHorticultureRequest, `,{"id":5,"jenis":"farm_map"}`, "", 1)
	w := httptest.NewRecorder()
	h.create(w, makeHorticultureRequest(invalid, "pelanggan", 42))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing required document status=%d body=%s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	h.create(w, makeHorticultureRequest(validHorticultureRequest, "pelanggan", 42))
	if w.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if repo.created.ApplicantID != 42 || repo.created.Organization != "Kelompok Tani Makmur" || len(repo.docs) != 5 {
		t.Fatalf("data repository salah: app=%+v docs=%+v", repo.created, repo.docs)
	}
	var got horticulture.Application
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Number != "RH-2026-ABC12345" || got.Status != "diajukan" {
		t.Fatalf("response tidak sesuai: %+v", got)
	}
}

func TestHorticultureListOnlyExposesOwnApplicationsToApplicant(t *testing.T) {
	repo := &fakeHorticultureRepository{items: []horticulture.Application{{ID: 1, ApplicantID: 42}, {ID: 2, ApplicantID: 43}}}
	h := &horticultureHandler{repo: repo}
	r := httptest.NewRequest(http.MethodGet, "/v1/horticulture/applications", nil)
	r = r.WithContext(auth.WithPrincipal(r.Context(), auth.Principal{UserID: 42, Role: "pelanggan"}))
	w := httptest.NewRecorder()
	h.list(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var got struct {
		Applications []horticulture.Application `json:"permohonan"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Applications) != 1 || got.Applications[0].ApplicantID != 0 {
		t.Fatalf("list response should contain one sanitized owned item: %+v", got.Applications)
	}
}
