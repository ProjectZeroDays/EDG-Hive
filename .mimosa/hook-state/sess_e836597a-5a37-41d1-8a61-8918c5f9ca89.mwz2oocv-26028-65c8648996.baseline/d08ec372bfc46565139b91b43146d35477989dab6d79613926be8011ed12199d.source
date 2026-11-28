package llmpool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeModelID(t *testing.T) {
	cases := map[string]string{
		"hf.co/taddymason/UMBASA-gemma-4-E2B-it-GGUF:Q4_K_M": "hf-co-taddymason-umbasa-gemma-4-e2b-it-gguf-q4-k-m",
		"ollama/llama3":                                      "llama3",
		"openrouter/anthropic/claude-3.5-sonnet":             "anthropic-claude-3-5-sonnet",
		"  Llama3:8B-Instruct  ":                             "llama3-8b-instruct",
		"gpt-4o":                                             "gpt-4o",
	}
	for in, want := range cases {
		if got := NormalizeModelID(in); got != want {
			t.Errorf("NormalizeModelID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveModelID(t *testing.T) {
	catalog := []string{
		"hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M",
		"gemma-4-e2b-it-qat-Q4_K_M",
		"llama3:8b-instruct-q4_K_M",
	}
	tests := []struct {
		name      string
		requested string
		catalog   []string
		want      string
		wantOK    bool
	}{
		{"exact", "gpt-4o", []string{"gpt-4o"}, "gpt-4o", true},
		{"case-insensitive", "GPT-4O", []string{"gpt-4o"}, "gpt-4o", true},
		{"normalized separators", "llama3:8b-instruct", catalog, "llama3:8b-instruct-q4_K_M", true},
		{"suffix on full id", "umbasa-gemma-4-e2b-it-ultra-uncensored-heretic", catalog,
			"hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M", true},
		{"legacy alias verified against catalog",
			"umbasa-gemma-4-e2b-it-ultra-uncensored-heretic", catalog,
			"hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M", true},
		{"legacy alias without catalog falls back to alias",
			"umbasa-gemma-4-e2b-it-ultra-uncensored-heretic", nil,
			"hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M", true},
		{"ambiguous suffix", "gemma-4-e2b-it", catalog, "", false},
		{"absent model", "gpt-9-ultra", catalog, "", false},
		{"empty request", "", catalog, "", false},
		{"exact beats alias", "gemma-4-e2b-it-qat-Q4_K_M", catalog, "gemma-4-e2b-it-qat-Q4_K_M", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ResolveModelID(tt.requested, tt.catalog)
			if ok != tt.wantOK || (ok && got != tt.want) {
				t.Fatalf("ResolveModelID(%q) = %q,%v — want %q,%v", tt.requested, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestFetchCatalog(t *testing.T) {
	t.Run("openai shape with bearer", func(t *testing.T) {
		var gotAuth string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/models" {
				t.Errorf("path = %q", r.URL.Path)
			}
			gotAuth = r.Header.Get("Authorization")
			json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "gpt-4o"}, {"id": "gpt-4o-mini"}}})
		}))
		defer srv.Close()
		models, err := FetchCatalog(context.Background(), srv.URL, "sk-test", "openai")
		if err != nil {
			t.Fatal(err)
		}
		if gotAuth != "Bearer sk-test" {
			t.Errorf("auth = %q", gotAuth)
		}
		if len(models) != 2 || models[0] != "gpt-4o" {
			t.Errorf("models = %v", models)
		}
	})
	t.Run("anthropic header", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("x-api-key") != "ak-test" {
				t.Errorf("x-api-key = %q", r.Header.Get("x-api-key"))
			}
			json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "claude-sonnet-4"}}})
		}))
		defer srv.Close()
		models, err := FetchCatalog(context.Background(), srv.URL, "ak-test", "anthropic")
		if err != nil || len(models) != 1 {
			t.Fatalf("models=%v err=%v", models, err)
		}
	})
	t.Run("non-200 is an error, not a crash", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()
		if _, err := FetchCatalog(context.Background(), srv.URL, "", "openai"); err == nil {
			t.Fatal("expected error for 401")
		}
	})
	t.Run("empty base url", func(t *testing.T) {
		if _, err := FetchCatalog(context.Background(), "  ", "", "openai"); err == nil {
			t.Fatal("expected error for empty base URL")
		}
	})
}

func TestCatalogsCache(t *testing.T) {
	calls := 0
	c := NewCatalogs(func(ctx context.Context, base, key, format string) ([]string, error) {
		calls++
		return []string{"m1"}, nil
	})
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		models, err := c.Get(ctx, "http://x", "k", "openai")
		if err != nil || len(models) != 1 {
			t.Fatalf("Get: %v %v", models, err)
		}
	}
	if calls != 1 {
		t.Fatalf("fetch called %d times, want 1 (cache miss)", calls)
	}
}
