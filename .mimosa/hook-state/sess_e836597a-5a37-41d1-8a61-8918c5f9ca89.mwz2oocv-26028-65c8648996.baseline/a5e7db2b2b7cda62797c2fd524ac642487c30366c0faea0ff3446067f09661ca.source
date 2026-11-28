package llmpool_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ProjectZeroDays/edg-hive/agent"
	"github.com/ProjectZeroDays/edg-hive/db"
	"github.com/ProjectZeroDays/edg-hive/llmpool"
	"github.com/Autumn-27/norma/llm"
)

// TestLiveOllamaPool exercises the real configured pool against the live
// provider (the local Ollama on this host). Gated behind LIVE_POOL=1 so the
// suite stays hermetic:
//
//	LIVE_POOL=1 go test ./llmpool/ -run TestLiveOllamaPool -v
//
// It replicates the server's profile loading (agent.ConfigFrom per enabled
// DB profile) without constructing a full Server.
func TestLiveOllamaPool(t *testing.T) {
	if os.Getenv("LIVE_POOL") != "1" {
		t.Skip("set LIVE_POOL=1 to run against the live provider")
	}
	dsn := os.Getenv("HIVE_PG_DSN")
	if dsn == "" {
		t.Skip("set HIVE_PG_DSN to the live DSN")
	}
	pg, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer pg.Close()
	if kd := os.Getenv("HIVE_KEY_DIR"); kd != "" {
		if err := pg.InitSecrets(kd); err != nil {
			t.Fatalf("init secrets: %v", err)
		}
	}
	profs, err := pg.PoolProfiles()
	if err != nil {
		t.Fatalf("pool profiles: %v", err)
	}
	reg := llmpool.NewRegistry(nil, nil)
	var members []*llmpool.Member
	for i, p := range profs {
		cfg := agent.ConfigFrom(p.Format, p.Model, p.BaseURL, p.APIKey, p.Proxy)
		cfg.RatePerSecond, cfg.RatePerMinute = p.RatePerSecond, p.RatePerMinute
		cfg.ContextWindowK = p.ContextWindowK
		cfg.Stream = p.Streaming
		if cfg.APIKey == "" && cfg.NeedsAPIKey() {
			t.Logf("skip profile %d (%s): needs API key", p.ID, p.Name)
			continue
		}
		prov, err := cfg.NewProvider()
		if err != nil {
			t.Logf("skip profile %d (%s): %v", p.ID, p.Name, err)
			continue
		}
		rank := p.Priority
		if p.IsDefault {
			rank = llmpool.RankActive
		}
		members = append(members, &llmpool.Member{
			ID: p.ID, Name: p.Name, Model: cfg.Model, Format: string(cfg.Format),
			Priority: p.Priority, Active: p.IsDefault, Rank: rank,
			WindowTokens: cfg.CompactionWindow(), Prov: prov,
		})
		_ = i
	}
	if len(members) == 0 {
		t.Fatal("no usable profiles in the live DB")
	}
	pool := llmpool.New(members, reg)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	msg, via, usage, err := pool.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: []llm.ContentBlock{{Type: llm.BlockText, Text: "Reply with exactly: OK"}}}},
		MaxTokens: 8,
	})
	if err != nil {
		t.Fatalf("pool complete: %v", err)
	}
	text := ""
	for _, b := range msg.Content {
		if b.Type == llm.BlockText {
			text += b.Text
		}
	}
	t.Logf("reply=%q via=%s model=%s tokens=%+v", text, via, poolMemberModel(members, via), usage)
	fmt.Println("LIVE_POOL_OK")
}

func poolMemberModel(members []*llmpool.Member, name string) string {
	for _, m := range members {
		if m.Name == name {
			return m.Model
		}
	}
	return "?"
}
