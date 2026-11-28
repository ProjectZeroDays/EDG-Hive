package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostgresDSNPrecedence(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{"database":{"host":"10.1.2.3","port":6000,"user":"u","password":"p","dbname":"d","sslmode":"require"}}`), 0o644)
	t.Setenv("HIVE_CONFIG", cfgPath)

	// no env DSN → built from config file fields
	t.Setenv("HIVE_PG_DSN", "")
	got, _, err := PostgresDSN()
	want := "postgres://u:p@10.1.2.3:6000/d?sslmode=require"
	if err != nil || got != want {
		t.Fatalf("from file: got %q err %v want %q", got, err, want)
	}

	// env wins over config file
	t.Setenv("HIVE_PG_DSN", "postgres://envwins/x")
	if got, _, err := PostgresDSN(); err != nil || got != "postgres://envwins/x" {
		t.Fatalf("env should win, got %q err %v", got, err)
	}

	// no env, no file → built-in default DSN (zero-config fresh install; the UI
	// then prompts to rotate away from the hive/hive default)
	t.Setenv("HIVE_PG_DSN", "")
	t.Setenv("HIVE_CONFIG", filepath.Join(dir, "nope.json"))
	got, src, err := PostgresDSN()
	if err != nil || got != DefaultDSN {
		t.Fatalf("missing config should fall back to the built-in default, got %q err %v", got, err)
	}
	if !strings.Contains(src, "default") {
		t.Fatalf("source should name the built-in default, got %q", src)
	}

	// full dsn in config file is used verbatim
	os.WriteFile(cfgPath, []byte(`{"database":{"dsn":"postgres://full/dsn"}}`), 0o644)
	t.Setenv("HIVE_CONFIG", cfgPath)
	if got, _, err := PostgresDSN(); err != nil || got != "postgres://full/dsn" {
		t.Fatalf("file dsn verbatim, got %q err %v", got, err)
	}
}

func TestMalformedConfigIsAnError(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	// What install.sh used to write for a password containing a double quote.
	os.WriteFile(cfgPath, []byte(`{"database":{"host":"h","password":"a"b","dbname":"d"},"skill_dir":"/s"}`), 0o644)
	t.Setenv("HIVE_CONFIG", cfgPath)
	t.Setenv("HIVE_PG_DSN", "")
	t.Setenv("HIVE_SKILL_DIR", "")

	if _, _, err := PostgresDSN(); err == nil || !strings.Contains(err.Error(), "parse config file") {
		t.Fatalf("PostgresDSN: want a parse error, got %v", err)
	}
	if d, err := SkillDir(); err == nil {
		t.Fatalf("SkillDir: want a parse error, got dir %q", d)
	}
	if _, err := Load(); err == nil {
		t.Fatal("Load: want a parse error")
	}
}

// The default-credentials flag must track the resolved DSN so the UI prompt
// appears exactly when the built-in hive/hive default is in effect.
func TestUsingDefaultCredentials(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	t.Setenv("HIVE_CONFIG", cfgPath)
	t.Setenv("HIVE_PG_DSN", "")

	// no config at all → built-in default → flag on
	if !UsingDefaultCredentials() {
		t.Fatal("no config should report default credentials")
	}
	// explicit non-default password → flag off
	os.WriteFile(cfgPath, []byte(`{"database":{"dsn":"postgres://hive:secret@127.0.0.1:5432/hive"}}`), 0o600)
	if UsingDefaultCredentials() {
		t.Fatal("custom password should clear the default flag")
	}
	// env DSN beats everything
	t.Setenv("HIVE_PG_DSN", "postgres://hive:hive@127.0.0.1:5432/hive")
	if !UsingDefaultCredentials() {
		t.Fatal("env DSN matching the default should report default credentials")
	}
}

// SaveDatabaseCreds persists a decomposed database section (password included)
// into the config file, refusing to touch an env-var configuration.
func TestSaveDatabaseCreds(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	t.Setenv("HIVE_CONFIG", cfgPath)
	t.Setenv("HIVE_PG_DSN", "")

	if err := SaveDatabaseCreds("new-strong-pass"); err != nil {
		t.Fatalf("save: %v", err)
	}
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatalf("written config unparseable: %v", err)
	}
	if c.Database.User != "hive" || c.Database.Password != "new-strong-pass" || c.Database.DBName != "hive" {
		t.Fatalf("unexpected database section: %+v", c.Database)
	}
	if c.Database.DSN != "" {
		t.Fatalf("dsn field should be decomposed into fields, got %q", c.Database.DSN)
	}
	fi, _ := os.Stat(cfgPath)
	if fi.Mode().Perm() != 0o600 {
		t.Fatalf("config must be 0600 (holds a password), got %v", fi.Mode().Perm())
	}

	// env-configured → refuse (operator manages out-of-band)
	t.Setenv("HIVE_PG_DSN", "postgres://hive:hive@127.0.0.1:5432/hive")
	if err := SaveDatabaseCreds("x"); err == nil {
		t.Fatal("env DSN should refuse config writes")
	}
}
