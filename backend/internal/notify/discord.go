// Package notify mengirim notifikasi sederhana (Discord webhook).
package notify

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
	WebhookURL string
}

func ConfigFromEnv() Config {
	return Config{WebhookURL: strings.TrimSpace(os.Getenv("DISCORD_WEBHOOK_URL"))}
}

func (c Config) Siap() bool { return c.WebhookURL != "" }

type Client struct {
	url    string
	client *http.Client
}

func New(c Config) *Client {
	if !c.Siap() {
		return &Client{url: "", client: &http.Client{Timeout: 5 * time.Second}}
	}
	return &Client{url: c.WebhookURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *Client) Kirim(ctx context.Context, pesan string) error {
	if c == nil || c.url == "" || strings.TrimSpace(pesan) == "" {
		return nil
	}
	body := map[string]string{"content": pesan}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
