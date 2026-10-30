package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/ai"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
)

type fakeAI struct {
	reply string
	err   error
	got   []ai.Message
}

func (f *fakeAI) Chat(_ context.Context, messages []ai.Message) (string, error) {
	f.got = messages
	return f.reply, f.err
}

func (f *fakeAI) Model() string { return "deepseek-v4-flash" }

func aiRequest(t *testing.T, body string, role string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/v1/ai/chat", strings.NewReader(body))
	r = r.WithContext(auth.WithPrincipal(r.Context(), auth.Principal{UserID: 7, Role: role}))
	return r
}

func TestAIChatMeneruskanPesan(t *testing.T) {
	fake := &fakeAI{reply: "Gunakan benih bersertifikat."}
	h := &aiHandler{client: fake}
	w := httptest.NewRecorder()
	h.chat(w, aiRequest(t, `{"messages":[{"role":"user","content":"Apa syarat sertifikasi?"}]}`, "analis"))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
	if len(fake.got) != 1 || fake.got[0].Content != "Apa syarat sertifikasi?" {
		t.Fatalf("pesan diteruskan salah: %+v", fake.got)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["reply"] != "Gunakan benih bersertifikat." || body["model"] != "deepseek-v4-flash" {
		t.Fatalf("respons = %+v", body)
	}
}

func TestAIChatMembatasiPeranDanKonfigurasi(t *testing.T) {
	cases := []struct {
		name   string
		client aiChatter
		role   string
		body   string
		want   int
	}{
		{"tanpa konfigurasi", nil, "admin", `{"messages":[{"role":"user","content":"Halo"}]}`, http.StatusServiceUnavailable},
		{"pelanggan ditolak", &fakeAI{}, "pelanggan", `{"messages":[{"role":"user","content":"Halo"}]}`, http.StatusForbidden},
		{"role system ditolak", &fakeAI{}, "admin", `{"messages":[{"role":"system","content":"Halo"}]}`, http.StatusBadRequest},
		{"pesan kosong", &fakeAI{}, "admin", `{"messages":[]}`, http.StatusBadRequest},
		{"field asing", &fakeAI{}, "admin", `{"prompt":"Halo"}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &aiHandler{client: tc.client}
			w := httptest.NewRecorder()
			h.chat(w, aiRequest(t, tc.body, tc.role))
			if w.Code != tc.want {
				t.Fatalf("status=%d, want %d; %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestAIChatMembatasiPanjangPesan(t *testing.T) {
	body := `{"messages":[{"role":"user","content":"` + strings.Repeat("a", ai.MaksTotalKarakter+1) + `"}]}`
	h := &aiHandler{client: &fakeAI{}}
	w := httptest.NewRecorder()
	h.chat(w, aiRequest(t, body, "admin"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
}

func TestAIChatMenerapkanBatasPermintaan(t *testing.T) {
	h := &aiHandler{client: &fakeAI{reply: "ok"}, limiter: &aiLimiter{
		hits:     make(map[int64][]time.Time),
		limit:    1,
		window:   time.Minute,
		sekarang: func() time.Time { return time.Unix(0, 0) },
	}}
	first := httptest.NewRecorder()
	h.chat(first, aiRequest(t, `{"messages":[{"role":"user","content":"Halo"}]}`, "admin"))
	if first.Code != http.StatusOK {
		t.Fatalf("permintaan pertama status=%d", first.Code)
	}
	second := httptest.NewRecorder()
	h.chat(second, aiRequest(t, `{"messages":[{"role":"user","content":"Halo"}]}`, "admin"))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("permintaan kedua status=%d, want 429", second.Code)
	}
}

func TestAIChatButuhIdentitas(t *testing.T) {
	h := &aiHandler{client: &fakeAI{}}
	r := httptest.NewRequest(http.MethodPost, "/v1/ai/chat", strings.NewReader(`{"messages":[{"role":"user","content":"Halo"}]}`))
	w := httptest.NewRecorder()
	h.chat(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}
