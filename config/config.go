// Package config loads runtime configuration from a JSON file, with environment
// variables taking precedence. Currently it carries the PostgreSQL connection.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Database is the PostgreSQL connection config. Either set DSN directly, or set
// the component fields and a DSN is assembled from them.
type Database struct {
	DSN      string `json:"dsn"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// Config is the on-disk config file shape.
type Config struct {
	Database Database `json:"database"`
	SkillDir string   `json:"skill_dir"`
}

// BaseDir is the directory that anchors all runtime artifacts (config.json and
// the data/ store). It is the directory the running binary lives in, so a
// distributed executable keeps its files next to itself on any OS (Windows,
// Linux, …) regardless of the working directory it is launched from.
//
// When launched via `go run`, the binary is throwaway: it sits either in a temp
// build dir (cache miss → fresh link) OR straight inside the Go build cache
// (cache hit → run from $GOCACHE/.../...-d). We detect both and fall back to the
// current working directory so dev artifacts (data/, transcripts) and config
// resolve against the project dir, not the throwaway binary's location.
func BaseDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	dir := filepath.Dir(exe)
	if isGoRunDir(dir) {
		return "." // throwaway `go run` binary → use CWD
	}
	return dir
}

// KeyDir holds jwt.key and secret.key: HIVE_JWT_KEY_DIR (the Docker image points
// it at the /app/keys volume) or the base dir. Never the browsable data dir.
func KeyDir() string {
	dir := strings.TrimSpace(os.Getenv("HIVE_JWT_KEY_DIR"))
	if dir == "" {
		dir = BaseDir()
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return dir
}

// isGoRunDir reports whether dir is where `go run` parked its executable: under
// the system temp dir (cache miss), or anywhere inside a Go build cache
// (…/go-build/…, the cache-hit case — NOT under os.TempDir(), which is why the
// old temp-only check failed intermittently). In both cases the binary is
// throwaway, so config/data must resolve against the CWD.
func isGoRunDir(dir string) bool {
	if tmp := os.TempDir(); tmp != "" {
		if rel, err := filepath.Rel(tmp, dir); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	for _, seg := range strings.Split(filepath.ToSlash(dir), "/") {
		if seg == "go-build" {
			return true
		}
	}
	return false
}

// Path returns the config file path. Resolution order:
//  1. env HIVE_CONFIG (explicit override)
//  2. ./config.json in the current working directory (running from the project
//     dir — robust no matter where `go run` placed the temp/cached binary)
//  3. config.json next to the executable (a distributed binary keeps it beside)
//
// The first existing file wins. If none exist, the CWD path is returned so the
// "not found" message points at the project dir the user most likely expected.
func Path() string {
	if v := strings.TrimSpace(os.Getenv("HIVE_CONFIG")); v != "" {
		return v
	}
	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "config.json"))
	}
	candidates = append(candidates, filepath.Join(BaseDir(), "config.json"))
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

// Load reads and parses the config file. A missing file yields a zero Config (so
// callers fall back to defaults). A file that exists but cannot be read or parsed
// is an error: ignoring it would run with defaults the operator did not choose.
func Load() (Config, error) {
	var c Config
	path := Path()
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, fmt.Errorf("read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf("parse config file %s: %w", path, err)
	}
	return c, nil
}

// SkillDir returns the skill root directory with precedence:
//
//	env HIVE_SKILL_DIR  >  config file (skill_dir)  >  BaseDir()/skills
//
// The directory is created if it does not exist.
func SkillDir() (string, error) {
	var d string
	if v := strings.TrimSpace(os.Getenv("HIVE_SKILL_DIR")); v != "" {
		d = v
	} else {
		c, err := Load()
		if err != nil {
			return "", err
		}
		if v := strings.TrimSpace(c.SkillDir); v != "" {
			d = v
		} else {
			d = filepath.Join(BaseDir(), "skills")
		}
	}
	_ = os.MkdirAll(d, 0o755)
	return d, nil
}

// DefaultDSN is the zero-config database: a local PostgreSQL with the default
// hive/hive/hive credentials created by the quick-start install script. It lets
// a fresh install boot without writing any config, and the app then prompts the
// operator to rotate the password on first login (see server/db_security.go).
const DefaultDSN = "postgres://hive:hive@127.0.0.1:5432/hive?sslmode=disable"

// PostgresDSN resolves the connection string with precedence:
//
//	env HIVE_PG_DSN  >  config file (database.dsn, or assembled from fields)  >  built-in DefaultDSN
//
// The built-in default keeps a fresh install bootable with zero configuration;
// the UI surfaces a mandatory password-rotation prompt while it is in effect.
// source describes where the DSN came from (for startup logging + the security
// status endpoint): "environment variable HIVE_PG_DSN", "config file <path>
// (database.dsn|database fields)", or "built-in default credentials".
func PostgresDSN() (dsn, source string, err error) {
	if v := strings.TrimSpace(os.Getenv("HIVE_PG_DSN")); v != "" {
		return v, "environment variable HIVE_PG_DSN", nil
	}
	c, err := Load()
	if err != nil {
		return "", "", err
	}
	db := c.Database
	if d := strings.TrimSpace(db.DSN); d != "" {
		return d, "config file " + Path() + " (database.dsn)", nil
	}
	if db.Host != "" || db.DBName != "" || db.User != "" {
		return db.buildDSN(), "config file " + Path() + " (database fields)", nil
	}
	return DefaultDSN, "built-in default credentials", nil
}

// UsingDefaultCredentials reports whether the resolved DSN is the built-in
// hive/hive/hive default. The UI uses this to decide when to show the
// password-rotation prompt; the change endpoint refuses to run in any other
// state so it cannot be abused as a generic password changer.
func UsingDefaultCredentials() bool {
	dsn, _, err := PostgresDSN()
	if err != nil {
		return false
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return false
	}
	pw, _ := u.User.Password()
	return u.User.Username() == "hive" && pw == "hive"
}

// SaveDatabaseCreds persists a new database password (and the decomposed
// connection fields derived from the currently resolved DSN) into the config
// file, so the change survives restarts. It refuses to touch an env-var
// configuration (the operator manages that out-of-band). The file is written
// 0600 — it holds a password.
func SaveDatabaseCreds(newPassword string) error {
	if v := strings.TrimSpace(os.Getenv("HIVE_PG_DSN")); v != "" {
		return fmt.Errorf("the database is configured via the HIVE_PG_DSN environment variable; change its password in the environment, not here")
	}
	dsn, _, err := PostgresDSN()
	if err != nil {
		return err
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("parse current DSN: %w", err)
	}
	port := 5432
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	sslmode := u.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}
	dbname := strings.TrimPrefix(u.Path, "/")
	c, err := Load()
	if err != nil {
		return err
	}
	c.Database = Database{
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Password: newPassword,
		DBName:   dbname,
		SSLMode:  sslmode,
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), append(b, '\n'), 0o600)
}

func (d Database) buildDSN() string {
	host := d.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := d.Port
	if port == 0 {
		port = 5432
	}
	ssl := d.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	u := url.URL{
		Scheme: "postgres",
		Host:   host + ":" + strconv.Itoa(port),
		Path:   "/" + d.DBName,
	}
	if d.User != "" {
		if d.Password != "" {
			u.User = url.UserPassword(d.User, d.Password)
		} else {
			u.User = url.User(d.User)
		}
	}
	u.RawQuery = url.Values{"sslmode": {ssl}}.Encode()
	return u.String()
}

// String is a redacted view of the resolved DSN (password masked) for logging.
func Redact(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	if u.User != nil {
		if _, hasPw := u.User.Password(); hasPw {
			u.User = url.UserPassword(u.User.Username(), "****")
		}
	}
	return fmt.Sprintf("%s", u.String())
}
