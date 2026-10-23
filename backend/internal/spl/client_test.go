package spl

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRequiresExplicitBaseURL(t *testing.T) {
	if _, err := New(Config{}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("New() error = %v, ingin ErrNotConfigured", err)
	}
}

func TestNewRejectsUnsafeBaseURL(t *testing.T) {
	for _, base := range []string{"spl.jogjaprov.go.id", "file:///tmp", "https://user:pass@example.test", "https://example.test?key=value"} {
		if _, err := New(Config{BaseURL: base}); err == nil {
			t.Errorf("New(%q) harus gagal", base)
		}
	}
}

func TestDoJSONTransport(t *testing.T) {
	type request struct {
		Nama string `json:"nama"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/gateway/v1/kontrak" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Errorf("Content-Type = %q", got)
		}
		var got request
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil || got.Nama != "uji" {
			t.Errorf("request body = %+v, err=%v", got, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"diterima"}`))
	}))
	defer server.Close()

	client, err := New(Config{BaseURL: server.URL + "/gateway", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Status string `json:"status"`
	}
	if err := client.Do(context.Background(), http.MethodPost, "v1/kontrak", request{Nama: "uji"}, &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "diterima" {
		t.Fatalf("status = %q", response.Status)
	}
}

func TestDoRejectsRouteEscape(t *testing.T) {
	client, err := New(Config{BaseURL: "https://example.test/base"})
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{"", "/absolute", "../escape", "%2e%2e/escape", "https://other.test/path", "v1/path?x=1", `v1\\path`} {
		if err := client.Do(context.Background(), http.MethodGet, route, nil, nil); !errors.Is(err, ErrInvalidRoute) {
			t.Errorf("Do(route=%q) error = %v, ingin ErrInvalidRoute", route, err)
		}
	}
}

func TestDoReturnsStatusOnlyForUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("contains private upstream details"))
	}))
	defer server.Close()
	client, err := New(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	err = client.Do(context.Background(), http.MethodGet, "resource", nil, nil)
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("error = %#v, ingin HTTPError 401", err)
	}
	if strings.Contains(err.Error(), "private") {
		t.Fatal("error membocorkan body upstream")
	}
}
