// Package observability menangani pelaporan error sederhana ke Sentry bila
// dikonfigurasi. Tanpa SDK berat di go.mod, modul ini hanya mengirim event
// HTTP dasar jika SENTRY_DSN tersedia.
package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	DSN         string
	Environment string
	Release     string
}

func ConfigFromEnv() Config {
	return Config{
		DSN:         strings.TrimSpace(os.Getenv("SENTRY_DSN")),
		Environment: strings.TrimSpace(os.Getenv("SENTRY_ENVIRONMENT")),
		Release:     strings.TrimSpace(os.Getenv("SENTRY_RELEASE")),
	}
}

func (c Config) Siap() bool { return c.DSN != "" }

type Client struct {
	cfg    Config
	client *http.Client
}

func New(c Config) *Client {
	if c.Environment == "" {
		c.Environment = "development"
	}
	if c.Release == "" {
		c.Release = "dev"
	}
	return &Client{cfg: c, client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *Client) Tangkap(ctx context.Context, err error, pesan string) {
	if c == nil || !c.cfg.Siap() || err == nil {
		return
	}
	event := map[string]any{
		"event_id": "sitani",
		"level":    "error",
		"platform": "go",
		"environment": c.cfg.Environment,
		"release": c.cfg.Release,
		"message": map[string]string{"formatted": strings.TrimSpace(pesan + ": " + err.Error())},
	}
	raw, err2 := json.Marshal(event)
	if err2 != nil {
		return
	}
	req, err2 := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.DSN, bytes.NewReader(raw))
	if err2 != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err2 := c.client.Do(req)
	if err2 != nil {
		return
	}
	_ = resp.Body.Close()
}
