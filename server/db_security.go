package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ProjectZeroDays/edg-hive/config"
)

// dbSecurityStatus tells the frontend whether the database is still on the
// built-in hive/hive/hive default and whether this session can rotate it.
type dbSecurityStatus struct {
	DefaultCredentials bool   `json:"default_credentials"`
	Changeable         bool   `json:"changeable"`
	Source             string `json:"source"`
}

func dbSecurityFromConfig() dbSecurityStatus {
	dsn, source, err := config.PostgresDSN()
	if err != nil {
		return dbSecurityStatus{Source: "unavailable: " + err.Error()}
	}
	u, perr := url.Parse(dsn)
	pw := ""
	if perr == nil && u.User != nil {
		pw, _ = u.User.Password()
	}
	return dbSecurityStatus{
		DefaultCredentials: perr == nil && u.User.Username() == "hive" && pw == "hive",
		Changeable:         !strings.HasPrefix(source, "environment variable"),
		Source:             source,
	}
}

// GET /api/system/db-security
func (s *Server) dbSecurity(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, dbSecurityFromConfig())
}

// POST /api/system/db-credentials  {password}
//
// Rotates the database password away from the built-in default. Only available
// while the default credentials are in effect (the first-login prompt); it is
// not a general password manager. Runs ALTER USER on the pool (the hive role
// changing its own password needs no extra privilege), then persists the new
// password into config.json so restarts keep it. Existing pool connections
// stay valid after the rotation; new connections use the stored password.
func (s *Server) dbChangeCredentials(w http.ResponseWriter, r *http.Request) {
	st := dbSecurityFromConfig()
	if !st.DefaultCredentials {
		writeErr(w, 409, "the database is not using the default credentials; this rotation is only offered while the defaults are in effect")
		return
	}
	if !st.Changeable {
		writeErr(w, 400, "the database is configured via the HIVE_PG_DSN environment variable; rotate its password in the environment")
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	pw := body.Password
	if len(pw) < 8 || len(pw) > 72 {
		writeErr(w, 400, "password must be 8–72 bytes (PostgreSQL SCRAM limit)")
		return
	}
	dsn, _, err := config.PostgresDSN()
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	u, err := url.Parse(dsn)
	if err != nil {
		writeErr(w, 500, "parse DSN: "+err.Error())
		return
	}
	user := u.User.Username()
	if user == "" {
		writeErr(w, 500, "no database user in resolved DSN")
		return
	}
	// Quote the identifier (double quotes, inner quotes doubled) — it comes from
	// the operator's own config, not a request field, but quote defensively.
	ident := `"` + strings.ReplaceAll(user, `"`, `""`) + `"`
	pg := s.pg(w)
	if pg == nil {
		return
	}
	if _, err := pg.Exec("ALTER USER "+ident+" WITH PASSWORD $1", pw); err != nil {
		writeErr(w, 500, "ALTER USER failed: "+err.Error())
		return
	}
	if err := config.SaveDatabaseCreds(pw); err != nil {
		log.Printf("[db-security] password rotated but config save failed: %v", err)
		writeErr(w, 500, fmt.Sprintf("password rotated but could not persist to %s: %v", config.Path(), err))
		return
	}
	log.Printf("[db-security] database password rotated away from default (user %s)", user)
	writeJSON(w, 200, map[string]any{"ok": true})
}
