#!/bin/bash
#===============================================================================
# ARTEX one-shot bootstrap (Ubuntu/Debian server, run as root)
#
#   curl -fsSL <raw-url>/init_install.sh | bash
#   or:  bash init_install.sh
#
# Does, in order:
#   1. deps (git, docker, curl, zstd, python3)
#   2. clones/updates the repo into /root/ARTEX, installs binary + systemd unit
#   3. Postgres container (artex-pg)
#   4. cloudflared quick tunnel as a persistent systemd unit (artex-tunnel)
#   5. prints the /setup URL + one-time setup token, waits for you to create
#      the admin account
#   6. Ollama (local LLM) + UMBASA-gemma-4-E2B uncensored model, OpenAI REST on
#      127.0.0.1:11434
#   7. registers that model in ARTEX as a provider and sets it DEFAULT
#      (prompts for your admin password; never stored)
#   8. asks for the engagement target (ip, cidr/24, or domain/url) and creates
#      the task + scope row
#
# Persistence: artex, artex-tunnel, ollama are all enabled systemd units.
# Quick-tunnel URLs rotate per start; run `artex-url` anytime to see the
# current one. For a STABLE public URL use a named tunnel (docs/SKILLS.md ->
# wiki) - requires your own Cloudflare zone.
#
# Things a headless server cannot do for you (it prints instead):
#   - open YOUR browser, - put text in YOUR clipboard. Copy-paste by hand.
#===============================================================================
set -euo pipefail

REPO_URL="${ARTEX_REPO:-https://github.com/ProjectZeroDays/ARTEX.git}"
INSTALL_DIR="${ARTEX_DIR:-/root/ARTEX}"
MODEL="${OLLAMA_MODEL:-hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M}"
export HOME=/root

say()  { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }
warn() { printf '\033[1;33mWARN: %s\033[0m\n' "$*"; }
die()  { printf '\033[1;31mERROR: %s\033[0m\n' "$*" >&2; exit 1; }
[ "$(id -u)" = 0 ] || die "run as root"

#--- 1. deps -------------------------------------------------------------------
say "Checking dependencies"
export DEBIAN_FRONTEND=noninteractive
need_apt=()
for c in git curl tar zstd python3; do command -v "$c" >/dev/null || need_apt+=("$c"); done
command -v docker >/dev/null || need_apt+=(docker.io)
if [ "${#need_apt[@]}" -gt 0 ]; then
  apt-get update -y && apt-get install -y "${need_apt[@]}"
fi
systemctl enable --now docker >/dev/null 2>&1 || true

# resumable, stall-aborting downloader
dl() { # dl <url> <out>
  local url="$1" out="$2" i
  for i in $(seq 1 60); do
    echo "  download attempt $i ($(stat -c%s "$out" 2>/dev/null || echo 0) bytes)"
    curl -C - --fail --location --connect-timeout 30 --speed-time 90 --speed-limit 512 \
         "$url" -o "$out" && return 0
    sleep 15
  done
  return 1
}

#--- 2. repo + binary ----------------------------------------------------------
say "Cloning / updating ARTEX ($INSTALL_DIR)"
if [ -d "$INSTALL_DIR/.git" ]; then
  git -C "$INSTALL_DIR" pull --ff-only || warn "git pull failed; using existing checkout"
else
  git clone "$REPO_URL" "$INSTALL_DIR"
fi
BIN="$INSTALL_DIR/dist/artex-linux-amd64/artex"
[ -x "$BIN" ] || die "release binary not found at $BIN"
chmod +x "$BIN"

if [ ! -f "$INSTALL_DIR/config.json" ]; then
  cat > "$INSTALL_DIR/config.json" <<EOF
{
  "server":   { "host": "127.0.0.1", "port": 8787 },
  "proxy":    { "host": "127.0.0.1", "port": 8788 },
  "database": { "type": "postgres", "dsn": "postgres://artex:artex-local-dev-only@127.0.0.1:5432/artex?sslmode=disable" },
  "skills":   { "dir": "$INSTALL_DIR/skills" },
  "llm":      { "dir": "$INSTALL_DIR/llm_profiles" }
}
EOF
  warn "wrote config.json with a DEVELOPMENT db password - change it for real use"
fi

say "Installing systemd units"
cat > /etc/systemd/system/artex.service <<EOF
[Unit]
Description=ARTEX server
After=network-online.target docker.service
Wants=network-online.target

[Service]
Type=simple
Environment=ARTEX_CONFIG=$INSTALL_DIR/config.json
Environment=HOME=/root
WorkingDirectory=$INSTALL_DIR
ExecStart=$BIN
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

#--- 3. postgres ---------------------------------------------------------------
if ! docker ps --format '{{.Names}}' | grep -qx hive-pg; then
  say "Starting Postgres container"
  docker run -d --name hive-pg --restart unless-stopped \
    -e POSTGRES_USER=hive -e POSTGRES_PASSWORD=hive \
    -e POSTGRES_DB=hive -p 127.0.0.1:5432:5432 postgres:15-alpine
  for i in $(seq 1 30); do
    docker exec hive-pg psql -U hive -d hive -c 'select 1' >/dev/null 2>&1 && break
    sleep 2
  done
fi
systemctl daemon-reload
systemctl enable --now artex

#--- 4. cloudflare tunnel (persistent) -----------------------------------------
say "Installing cloudflared + persistent tunnel unit"
if ! command -v cloudflared >/dev/null; then
  dl https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 /tmp/cloudflared.bin \
    && install -m 0755 /tmp/cloudflared.bin /usr/local/bin/cloudflared
fi
cat > /etc/systemd/system/artex-tunnel.service <<'EOF'
[Unit]
Description=ARTEX cloudflared quick tunnel
After=artex.service
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/cloudflared tunnel --no-autoupdate --url http://127.0.0.1:8787
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now artex-tunnel
cat > /usr/local/bin/artex-url <<'EOF'
#!/bin/bash
journalctl -u artex-tunnel --no-pager -n 300 | grep -oE 'https://[a-z0-9.-]+\.trycloudflare\.com' | tail -1
EOF
chmod +x /usr/local/bin/artex-url
for i in $(seq 1 30); do TURL=$(artex-url); [ -n "$TURL" ] && break; sleep 2; done
TURL=${TURL:-<run: artex-url>}

#--- 5. setup token ------------------------------------------------------------
say "ARTEX first-run setup"
TOKEN=""
for i in $(seq 1 60); do
  TOKEN=$(journalctl -u artex --no-pager -n 200 | grep -ioE 'setup token[^ ]* [^ ]+' | tail -1 || true)
  [ -n "$TOKEN" ] && break
  sleep 2
done

cat <<EOF

  +------------------------------------------------------------------+
  |  1. Open your browser:   $TURL/setup
  |  2. Paste the setup token from the server log (see below)
  |  3. Create your admin account
  +------------------------------------------------------------------+

EOF
if [ -n "$TOKEN" ]; then
  echo "  SETUP TOKEN: $TOKEN"
else
  echo "  (no fresh token - admin likely exists already)"
  echo "  get one with:  journalctl -u artex | grep -i 'setup token' | tail -1"
fi
read -rp "  Press ENTER here once the admin account is created... " _

#--- 6. ollama + model ---------------------------------------------------------
say "Installing Ollama (local LLM, OpenAI-compatible REST on 127.0.0.1:11434)"
if ! command -v ollama >/dev/null; then
  dl https://ollama.com/download/ollama-linux-amd64.tar.zst /root/ollama-linux-amd64.tar.zst \
    || die "ollama download failed"
  tar --zstd -xf /root/ollama-linux-amd64.tar.zst -C /
fi
cat > /etc/systemd/system/ollama.service <<'EOF'
[Unit]
Description=Ollama Service
After=network-online.target

[Service]
User=root
Environment="HOME=/root"
Environment="OLLAMA_HOST=127.0.0.1:11434"
Environment="OLLAMA_CONTEXT_LENGTH=8192"
Environment="OLLAMA_MAX_LOADED_MODELS=1"
ExecStart=/usr/local/bin/ollama serve
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
EOF
systemctl daemon-reload
systemctl enable --now ollama
for i in $(seq 1 30); do curl -sf http://127.0.0.1:11434/api/version >/dev/null && break; sleep 2; done

say "Pulling model $MODEL (several GB)"
ollama pull "$MODEL"

#--- 7. ARTEX provider = default ----------------------------------------------
say "Registering the local model as ARTEX default provider"
read -rp "ARTEX admin username [ARTEX]: " AUSER; AUSER=${AUSER:-ARTEX}
read -rsp "ARTEX admin password (not stored): " APASS; echo
ARTEX_USER="$AUSER" ARTEX_PASS="$APASS" MODEL="$MODEL" python3 - <<'PYEOF'
import json, os, sys, urllib.request
base = "http://127.0.0.1:8787"
def call(method, path, body=None, token=None):
    req = urllib.request.Request(base + path, method=method)
    req.add_header("Content-Type", "application/json")
    if token: req.add_header("Authorization", "Bearer " + token)
    data = json.dumps(body).encode() if body is not None else None
    with urllib.request.urlopen(req, data) as r:
        return json.loads(r.read().decode() or "{}")
try:
    tok = call("POST", "/api/auth/login",
               {"username": os.environ["ARTEX_USER"], "password": os.environ["ARTEX_PASS"]})["token"]
    model = os.environ["MODEL"]
    prof = call("POST", "/api/llm/profiles", {
        "name": "Ollama-Gemma4-E2B", "format": "openai",
        "base_url": "http://127.0.0.1:11434/v1", "model": model,
        "api_key": "ollama", "streaming": True,
        "max_tokens": 4096, "temperature": 0.3})
    pid = prof.get("id") or prof.get("ID")
    call("POST", "/api/llm/profiles/active", {"id": pid})
    print(f"  provider registered (id={pid}) and set as DEFAULT")
except Exception as e:
    print(f"  WARN: provider setup failed: {e}", file=sys.stderr)
    print("  add manually: Settings -> LLM -> Add profile")
    print("    format=openai  base_url=http://127.0.0.1:11434/v1")
    print(f"    model={model}  api_key=ollama  (any non-empty string)")
PYEOF

#--- 8. target -----------------------------------------------------------------
say "Engagement target"
cat <<'EOF'
  Scope kinds accepted:
    ip        203.0.113.10
    cidr      203.0.113.0/24
    root_domain / subdomain / url   example.com
EOF
read -rp "  Target: " TARGET
[ -n "$TARGET" ] && python3 "$INSTALL_DIR/scripts/artex-new-target.py" "$TARGET" --base http://127.0.0.1:8787

#--- done ----------------------------------------------------------------------
say "Done. Day-2 commands"
cat <<EOF
  current tunnel URL : artex-url
  services           : systemctl status artex artex-tunnel ollama
  logs               : journalctl -u <unit> -f
  local LLM API      : curl http://127.0.0.1:11434/v1/models
EOF
echo
