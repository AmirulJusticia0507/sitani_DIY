// Package spl menyediakan transport HTTP generik untuk integrasi SPL/Cakra.
// Operasi bisnis, endpoint, dan autentikasi harus ditentukan dari kontrak
// resmi sebelum client ini dipakai untuk panggilan produksi.
package spl

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

const defaultTimeout = 10 * time.Second

var (
	ErrNotConfigured = errors.New("konfigurasi SPL belum tersedia")
	ErrInvalidRoute  = errors.New("rute SPL harus berupa path relatif yang aman")
)

// Config hanya berisi parameter transport yang tidak bergantung pada kontrak
// bisnis. BaseURL harus berasal dari konfigurasi resmi; tidak ada default host.
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// ConfigFromEnv membaca SPL_BASE_URL dan SPL_TIMEOUT_SECONDS.
func ConfigFromEnv() Config {
	cfg := Config{BaseURL: strings.TrimSpace(os.Getenv("SPL_BASE_URL")), Timeout: defaultTimeout}
	if raw := strings.TrimSpace(os.Getenv("SPL_TIMEOUT_SECONDS")); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			cfg.Timeout = time.Duration(seconds) * time.Second
		}
	}
	return cfg
}

// HTTPError hanya memuat status code; body respons upstream sengaja tidak
// diteruskan karena dapat berisi data pribadi atau informasi sensitif.
type HTTPError struct{ StatusCode int }

func (e *HTTPError) Error() string { return fmt.Sprintf("SPL mengembalikan HTTP %d", e.StatusCode) }

// Client menjalankan request JSON terhadap base URL yang sudah dikonfigurasi.
// Header autentikasi tidak dibuat otomatis: mekanismenya menunggu kontrak resmi.
type Client struct {
	base *url.URL
	http *http.Client
}

// New memvalidasi konfigurasi dan membuat client dengan timeout terbatas.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, ErrNotConfigured
	}
	base, err := url.Parse(cfg.BaseURL)
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, fmt.Errorf("SPL_BASE_URL harus berupa URL http(s) tanpa userinfo, query, atau fragment")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	return &Client{base: base, http: &http.Client{Timeout: cfg.Timeout}}, nil
}

// Do mengirim JSON opsional dan decode JSON respons bila output bukan nil.
// route harus path relatif (contoh hanya setelah ditentukan kontrak resmi:
// "v1/..."), bukan URL atau path yang meloloskan diri dari base URL.
func (c *Client) Do(ctx context.Context, method, route string, input, output any) error {
	if c == nil || c.base == nil {
		return ErrNotConfigured
	}
	u, err := c.routeURL(route)
	if err != nil {
		return err
	}
	var body io.Reader
	if input != nil {
		var data bytes.Buffer
		if err := json.NewEncoder(&data).Encode(input); err != nil {
			return fmt.Errorf("encode request SPL: %w", err)
		}
		body = &data
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf("buat request SPL: %w", err)
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if output != nil {
		req.Header.Set("Accept", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request SPL gagal: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode}
	}
	if output == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	dec := json.NewDecoder(io.LimitReader(resp.Body, 4<<20))
	if err := dec.Decode(output); err != nil {
		return fmt.Errorf("decode respons JSON SPL: %w", err)
	}
	return nil
}

func (c *Client) routeURL(route string) (*url.URL, error) {
	if route == "" || strings.HasPrefix(route, "/") || strings.Contains(route, "\\") || strings.Contains(route, "?") || strings.Contains(route, "#") {
		return nil, ErrInvalidRoute
	}
	ref, err := url.Parse(route)
	if err != nil || ref.IsAbs() || ref.Host != "" || ref.RawQuery != "" || ref.Fragment != "" {
		return nil, ErrInvalidRoute
	}
	for _, segment := range strings.Split(ref.Path, "/") {
		if segment == ".." || segment == "." || segment == "" {
			return nil, ErrInvalidRoute
		}
	}
	base := *c.base
	base.Path = strings.TrimRight(base.Path, "/") + "/" + ref.Path
	base.RawPath = ""
	return &base, nil
}
