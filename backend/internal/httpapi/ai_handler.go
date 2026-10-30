package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/ai"
	"github.com/AmirulJusticia0507/sitani_DIY/backend/internal/auth"
)

// aiChatter memisahkan handler dari transport agar bisa diuji dengan klien palsu.
type aiChatter interface {
	Chat(context.Context, []ai.Message) (string, error)
	Model() string
}

type aiHandler struct {
	client  aiChatter
	limiter *aiLimiter
}

func (h *aiHandler) register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/ai/chat", h.chat)
}

// aiAllowedRole membatasi asisten AI ke peran internal. Akses untuk pelanggan
// menunggu keputusan DPKP agar biaya dan cakupan layanan jelas.
func aiAllowedRole(role string) bool {
	switch role {
	case "superadmin", "admin", "pengawas", "analis", "penandatangan", "legalisasi", "percetakan":
		return true
	default:
		return false
	}
}

type aiChatRequest struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

func (h *aiHandler) chat(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok || principal.UserID <= 0 {
		respondError(w, http.StatusUnauthorized, "identitas pengguna tidak tersedia")
		return
	}
	if !aiAllowedRole(principal.Role) {
		respondError(w, http.StatusForbidden, "peran tidak diizinkan memakai asisten AI")
		return
	}
	if h.client == nil {
		respondError(w, http.StatusServiceUnavailable, "asisten AI belum dikonfigurasi")
		return
	}
	if h.limiter != nil && !h.limiter.allow(principal.UserID) {
		respondError(w, http.StatusTooManyRequests, "batas permintaan asisten AI tercapai, coba lagi nanti")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var payload aiChatRequest
	if err := dec.Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "body harus JSON dengan field messages")
		return
	}
	if len(payload.Messages) == 0 || len(payload.Messages) > ai.MaksPesan {
		respondError(w, http.StatusBadRequest, "jumlah pesan harus 1 sampai 20")
		return
	}
	messages := make([]ai.Message, 0, len(payload.Messages))
	total := 0
	for _, item := range payload.Messages {
		role := strings.TrimSpace(item.Role)
		if role != "user" && role != "assistant" {
			respondError(w, http.StatusBadRequest, "role pesan harus user atau assistant")
			return
		}
		content := strings.TrimSpace(item.Content)
		if content == "" {
			respondError(w, http.StatusBadRequest, "isi pesan tidak boleh kosong")
			return
		}
		if utf8.RuneCountInString(content) > ai.MaksPanjangPesan {
			respondError(w, http.StatusBadRequest, "satu pesan maksimal 4000 karakter")
			return
		}
		total += utf8.RuneCountInString(content)
		if total > ai.MaksTotalKarakter {
			respondError(w, http.StatusBadRequest, "total isi percakapan maksimal 16000 karakter")
			return
		}
		messages = append(messages, ai.Message{Role: role, Content: content})
	}

	reply, err := h.client.Chat(r.Context(), messages)
	if err != nil {
		status := http.StatusBadGateway
		message := "asisten AI tidak dapat dihubungi"
		var httpErr *ai.HTTPError
		switch {
		case errors.Is(err, ai.ErrNotConfigured):
			status, message = http.StatusServiceUnavailable, "asisten AI belum dikonfigurasi"
		case errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusTooManyRequests:
			status, message = http.StatusTooManyRequests, "kuota penyedia AI sedang penuh, coba lagi nanti"
		case errors.As(err, &httpErr):
			status, message = http.StatusBadGateway, "penyedia AI menolak permintaan"
		}
		respondError(w, status, message)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"reply": reply, "model": h.client.Model()})
}

// aiLimiter membatasi permintaan per pengguna dalam satu jendela waktu agar
// kredit penyedia AI tidak habis karena pemakaian berulang.
type aiLimiter struct {
	mu       sync.Mutex
	hits     map[int64][]time.Time
	limit    int
	window   time.Duration
	sekarang func() time.Time
}

func newAILimiterDariEnv() *aiLimiter {
	limit := 10
	if raw := strings.TrimSpace(os.Getenv("AI_MAX_REQUESTS_PER_MINUTE")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return &aiLimiter{
		hits:     make(map[int64][]time.Time),
		limit:    limit,
		window:   time.Minute,
		sekarang: time.Now,
	}
}

func (l *aiLimiter) allow(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.sekarang()
	batas := now.Add(-l.window)
	tersisa := l.hits[userID][:0]
	for _, t := range l.hits[userID] {
		if t.After(batas) {
			tersisa = append(tersisa, t)
		}
	}
	if len(tersisa) >= l.limit {
		l.hits[userID] = tersisa
		return false
	}
	l.hits[userID] = append(tersisa, now)
	return true
}
