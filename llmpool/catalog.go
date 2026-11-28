package llmpool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// catalog.go resolves configured model ids against the provider's LIVE catalog
// instead of trusting the string in the DB. A stale id (renamed upstream model,
// shortened local name like "llama3" vs "llama3:8b-instruct-q4_K_M") 404s on
// every request until a human edits it; resolving at boot repairs the profile
// once and the fix persists.

// LegacyAliases maps ids that USED to be valid to their current canonical form.
// Entries exist only for models this codebase has historically shipped as
// defaults (e.g. the init_install Ollama bundle); everything else resolves
// through the catalog by normalized/suffix matching.
var LegacyAliases = map[string]string{
	"umbasa-gemma-4-e2b-it-ultra-uncensored-heretic":
	"hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M",
}

// NormalizeModelID canonicalizes a model id for comparison: lower case, provider
// prefixes stripped ("ollama/", "openrouter/"), and all separator runs
// ("/", ":", "-", ".", whitespace) collapsed to a single "-". "HF.co/Foo Bar:Q4"
// and "foo-bar-q4" normalize identically.
func NormalizeModelID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if i := strings.IndexByte(id, '/'); i > 0 {
		switch id[:i] {
		case "ollama", "openrouter", "anthropic", "openai", "provider":
			id = id[i+1:]
		}
	}
	var b strings.Builder
	dash := false
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// ResolveModelID maps a requested id to the best catalog entry.
//
// Order: exact match (case-insensitive) → legacy alias (verified against the
// catalog when one is available) → unique normalized match → unique suffix
// match (so "umbasa-gemma-4" resolves when exactly one catalog model carries
// that prefix). ok=false means the catalog could not improve on the requested
// id — callers surface the provider's own 404 with an actionable error rather
// than guessing.
func ResolveModelID(requested string, catalog []string) (resolved string, ok bool) {
	if requested == "" {
		return "", false
	}
	for _, c := range catalog {
		if strings.EqualFold(c, requested) {
			return c, true
		}
	}
	if alias, hit := LegacyAliases[NormalizeModelID(requested)]; hit {
		if len(catalog) == 0 {
			return alias, true
		}
		for _, c := range catalog {
			if strings.EqualFold(c, alias) {
				return c, true
			}
		}
		if r, found := uniqueNormalizedMatch(alias, catalog); found {
			return r, true
		}
		return "", false // alias no longer exists upstream either
	}
	return uniqueNormalizedMatch(requested, catalog)
}

// uniqueNormalizedMatch returns the single catalog entry whose normalized form
// equals the request, starts with it, or contains it as a complete token run
// ("-"-bounded). Two or more candidates mean the id is ambiguous; we refuse to
// pick one.
func uniqueNormalizedMatch(requested string, catalog []string) (string, bool) {
	want := NormalizeModelID(requested)
	if want == "" {
		return "", false
	}
	norm := make(map[string]string, len(catalog))
	for _, c := range catalog {
		norm[NormalizeModelID(c)] = c
	}
	if exact, hit := norm[want]; hit {
		return exact, true
	}
	var match string
	for n, c := range norm {
		bounded := strings.Contains(n, "-"+want+"-") || strings.HasSuffix(n, "-"+want) || strings.HasPrefix(n, want)
		if !bounded {
			continue
		}
		if match != "" {
			return "", false // ambiguous
		}
		match = c
	}
	return match, match != ""
}

// catalogEntry mirrors the OpenAI /models wire shape ("data": [{"id": ...}]),
// which Ollama and OpenRouter-compatible endpoints share.
type catalogWire struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// FetchCatalog lists the models a provider actually serves, via GET {base}/models.
// format selects the auth scheme ("anthropic" uses x-api-key, everything else
// Bearer). Ollama answers the same shape on its OpenAI-compatible /models. A
// provider that does not implement /models yields an error — treated as "no
// catalog" (resolution is skipped, requests go through unmodified).
func FetchCatalog(ctx context.Context, baseURL, apiKey, format string) ([]string, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return nil, fmt.Errorf("empty provider base URL")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/models", nil)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		if format == "anthropic" {
			req.Header.Set("x-api-key", apiKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		} else {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	var w catalogWire
	if err := json.Unmarshal(body, &w); err != nil {
		return nil, fmt.Errorf("catalog: decode: %w", err)
	}
	out := make([]string, 0, len(w.Data))
	for _, d := range w.Data {
		if d.ID != "" {
			out = append(out, d.ID)
		}
	}
	return out, nil
}

// catalogTTL is how long a fetched catalog stays fresh. Model lists change on
// the scale of days, not seconds; a stale cache never breaks a request (the
// provider still validates the id itself), it only delays the next repair.
const catalogTTL = time.Hour

type catalogEntry struct {
	models []string
	at     time.Time
}

// Catalogs caches per-endpoint model lists so the boot repair pass and the
// mid-flight 404 re-resolve share one fetch. A failed fetch is NOT cached —
// the next caller retries it.
type Catalogs struct {
	mu   sync.Mutex
	m    map[string]catalogEntry
	fetch func(ctx context.Context, baseURL, apiKey, format string) ([]string, error)
}

// NewCatalogs builds an empty cache; fetch may be nil (FetchCatalog is used).
func NewCatalogs(fetch func(ctx context.Context, baseURL, apiKey, format string) ([]string, error)) *Catalogs {
	if fetch == nil {
		fetch = FetchCatalog
	}
	return &Catalogs{m: map[string]catalogEntry{}, fetch: fetch}
}

// key identifies one provider endpoint; different keys under one base URL
// share the catalog.
func catalogKey(baseURL, format string) string { return format + "|" + baseURL }

// Get returns the (possibly cached) model list for an endpoint. An error means
// no catalog is available; callers must proceed with the configured id.
func (c *Catalogs) Get(ctx context.Context, baseURL, apiKey, format string) ([]string, error) {
	k := catalogKey(baseURL, format)
	c.mu.Lock()
	if e, ok := c.m[k]; ok && time.Since(e.at) < catalogTTL {
		c.mu.Unlock()
		return e.models, nil
	}
	c.mu.Unlock()
	models, err := c.fetch(ctx, baseURL, apiKey, format)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.m[k] = catalogEntry{models: models, at: time.Now()}
	c.mu.Unlock()
	return models, nil
}
