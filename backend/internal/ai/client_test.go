package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewMenolakKonfigurasiKosong(t *testing.T) {
	if _, err := New(Config{BaseURL: "https://contoh.invalid/v1", Model: "m"}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("error = %v, ingin ErrNotConfigured", err)
	}
	if _, err := New(Config{BaseURL: "https://contoh.invalid/v1?x=1", APIKey: "k", Model: "m"}); err == nil {
		t.Fatal("base URL dengan query semestinya ditolak")
	}
}

func TestChatMengirimPermintaanKompatibelOpenAI(t *testing.T) {
	var gotPath, gotAuth string
	var got chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":" Halo petani "}}]}`))
	}))
	defer srv.Close()

	client, err := New(Config{BaseURL: srv.URL + "/v1", APIKey: "rahasia", Model: "deepseek-v4-flash"})
	if err != nil {
		t.Fatal(err)
	}
	reply, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "Halo"}})
	if err != nil {
		t.Fatal(err)
	}
	if reply != "Halo petani" {
		t.Fatalf("jawaban = %q", reply)
	}
	if gotPath != "/v1/chat/completions" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer rahasia" {
		t.Errorf("authorization = %q", gotAuth)
	}
	if got.Model != "deepseek-v4-flash" || len(got.Messages) != 1 || got.Messages[0].Content != "Halo" {
		t.Errorf("payload = %+v", got)
	}
}

func TestChatMeneruskanStatusTanpaBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"kunci rahasia salah"}}`))
	}))
	defer srv.Close()

	client, err := New(Config{BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Chat(context.Background(), []Message{{Role: "user", Content: "Halo"}})
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(err.Error(), "rahasia") {
		t.Fatal("pesan error tidak boleh memuat body penyedia")
	}
}

func TestChatMenolakBalasanKosong(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()

	client, err := New(Config{BaseURL: srv.URL, APIKey: "k", Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "Halo"}}); err == nil {
		t.Fatal("balasan tanpa choices semestinya ditolak")
	}
}

func TestClientNilTidakPanik(t *testing.T) {
	var client *Client
	if _, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "Halo"}}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("error = %v", err)
	}
	if client.Model() != "" {
		t.Fatal("model untuk client nil harus kosong")
	}
}
