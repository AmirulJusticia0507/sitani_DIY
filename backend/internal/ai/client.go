// Package ai menyediakan klien chat untuk penyedia yang kompatibel dengan API
// OpenAI (mis. GripHub Router). API key hanya dibaca dari environment dan tidak
// pernah diteruskan ke browser.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout    = 90 * time.Second
	maxResponseBytes  = 2 << 20
	MaksPesan         = 20
	MaksPanjangPesan  = 4000
	MaksTotalKarakter = 16000
)

var ErrNotConfigured = errors.New("konfigurasi asisten AI belum tersedia")

// Config hanya memuat parameter transport. Nilai berasal dari environment,
// bukan dari input pengguna, sehingga model dan tujuan tidak dapat diganti
// oleh pemanggil API.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

// ConfigFromEnv membaca AI_BASE_URL, AI_API_KEY, AI_MODEL, dan
// AI_TIMEOUT_SECONDS. Nilai kosong akan ditolak oleh New.
func ConfigFromEnv() Config {
	cfg := Config{
		BaseURL: strings.TrimSpace(os.Getenv("AI_BASE_URL")),
		APIKey:  strings.TrimSpace(os.Getenv("AI_API_KEY")),
		Model:   strings.TrimSpace(os.Getenv("AI_MODEL")),
		Timeout: defaultTimeout,
	}
	if raw := strings.TrimSpace(os.Getenv("AI_TIMEOUT_SECONDS")); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			cfg.Timeout = time.Duration(seconds) * time.Second
		}
	}
	return cfg
}

// HTTPError hanya memuat status code. Body balasan penyedia tidak diteruskan
// karena dapat memuat detail akun atau data lain yang tidak perlu diketahui
// pengguna.
type HTTPError struct{ StatusCode int }

func (e *HTTPError) Error() string {
	return fmt.Sprintf("penyedia AI mengembalikan HTTP %d", e.StatusCode)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	base  *url.URL
	key   string
	model string
	http  *http.Client
}

// New memvalidasi konfigurasi dan membuat client dengan timeout terbatas.
// Pesan error tidak pernah memuat API key.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" || cfg.APIKey == "" || cfg.Model == "" {
		return nil, fmt.Errorf("%w: AI_BASE_URL, AI_API_KEY, dan AI_MODEL wajib diisi", ErrNotConfigured)
	}
	base, err := url.Parse(cfg.BaseURL)
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") ||
		base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("AI_BASE_URL harus berupa URL http(s) tanpa userinfo, query, atau fragment")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	return &Client{
		base:  base,
		key:   cfg.APIKey,
		model: cfg.Model,
		http:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

// Model mengembalikan slug model yang dipakai, bukan nilai dari klien.
func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// Chat mengirim percakapan ke {AI_BASE_URL}/chat/completions dan mengembalikan
// jawaban asisten. Role pesan divalidasi oleh pemanggil.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	if c == nil || c.base == nil {
		return "", ErrNotConfigured
	}
	if len(messages) == 0 {
		return "", errors.New("percakapan tidak boleh kosong")
	}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(chatRequest{Model: c.model, Messages: messages}); err != nil {
		return "", fmt.Errorf("encode permintaan AI: %w", err)
	}
	target := *c.base
	target.Path = strings.TrimRight(target.Path, "/") + "/chat/completions"
	target.RawPath = ""
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), &body)
	if err != nil {
		return "", fmt.Errorf("buat permintaan AI: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("permintaan AI gagal: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return "", fmt.Errorf("baca balasan AI: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &HTTPError{StatusCode: resp.StatusCode}
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("balasan AI bukan JSON yang dikenali: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("balasan AI tidak memuat pilihan jawaban")
	}
	reply := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if reply == "" {
		return "", errors.New("balasan AI kosong")
	}
	return reply, nil
}
