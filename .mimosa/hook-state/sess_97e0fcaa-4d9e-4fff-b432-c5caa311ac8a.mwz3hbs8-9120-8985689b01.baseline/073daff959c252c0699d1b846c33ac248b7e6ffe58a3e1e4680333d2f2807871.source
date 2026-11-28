#!/usr/bin/env python3
"""hive-local-llm.py — verify local Ollama, ensure the model is present,
and print (or apply) the HIVE LLM profile settings for it.

Usage:
  python3 hive-local-llm.py                  # check + print settings
  python3 hive-local-llm.py --apply          # also create the profile in HIVE (prompts for admin password)
  python3 hive-local-llm.py --pull           # ollama pull if model missing (may download GBs)

Requires: ollama installed and serving on 127.0.0.1:11434 (ollama serve / systemd).
"""
import argparse
import getpass
import json
import shutil
import subprocess
import sys
import urllib.request
import urllib.error

MODEL = "hf.co/taddymason/UMBASA-gemma-4-E2B-it-ultra-uncensored-heretic-GGUF:Q4_K_M"
OLLAMA = "http://127.0.0.1:11434"
HIVE_BASE = "http://127.0.0.1:8787"


def ollama(*args):
    return subprocess.run(["ollama", *args], capture_output=True, text=True)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--pull", action="store_true", help="ollama pull the model if missing")
    ap.add_argument("--apply", action="store_true", help="create the LLM profile in HIVE via API")
    ap.add_argument("--base", default=HIVE_BASE, help="HIVE base URL")
    args = ap.parse_args()

    if not shutil.which("ollama"):
        sys.exit("ollama CLI not found. Install ollama first (see wiki Part 3).")

    # 1. is the API up?
    try:
        with urllib.request.urlopen(OLLAMA + "/api/version", timeout=5) as r:
            print(f"ollama API: up (version {json.load(r).get('version','?')})")
    except Exception as e:
        sys.exit(f"ollama API not reachable at {OLLAMA}: {e}\nStart it: systemctl start ollama (or: ollama serve)")

    # 2. is the model present?
    ls = ollama("ls")
    if MODEL not in ls.stdout:
        if not args.pull:
            print(f"model NOT installed: {MODEL}")
            print("re-run with --pull to download it (several GB).")
            sys.exit(2)
        print(f"pulling {MODEL} ... (this can take a while)")
        p = ollama("pull", MODEL)
        if p.returncode != 0:
            sys.exit(f"ollama pull failed: {p.stderr.strip()}")
    print(f"model installed: {MODEL}")

    # 3. quick inference sanity check (loads model into RAM)
    req = urllib.request.Request(OLLAMA + "/v1/chat/completions", method="POST")
    req.add_header("Content-Type", "application/json")
    body = json.dumps({"model": MODEL, "messages": [{"role": "user", "content": "reply with the single word: ok"}],
                       "max_tokens": 8}).encode()
    try:
        with urllib.request.urlopen(req, body, timeout=180) as r:
            out = json.load(r)
        print(f"inference check: {out['choices'][0]['message']['content'][:40]!r}")
    except Exception as e:
        sys.exit(f"inference check failed: {e}")

    # 4. the HIVE settings block
    settings = f"""
HIVE  ->  Settings -> LLM -> Add profile
  Name:      Ollama-Gemma4-E2B
  Format:    openai
  Base URL:  {OLLAMA}/v1
  Model:     {MODEL}
  API key:   ollama          (any non-empty string; Ollama ignores it)
  Streaming: on
  Context:   8k recommended on this box (set OLLAMA_CONTEXT_LENGTH=8192)
Mark as default, or put first in the task LLM config chain.
"""
    print(settings)

    if not args.apply:
        return

    password = getpass.getpass("HIVE admin password: ")
    def call(method, path, token=None, body=None):
        r = urllib.request.Request(args.base + path, method=method)
        r.add_header("Content-Type", "application/json")
        if token:
            r.add_header("Authorization", "Bearer " + token)
        try:
            with urllib.request.urlopen(r, json.dumps(body).encode() if body is not None else None) as resp:
                return resp.status, json.loads(resp.read().decode() or "{}")
        except urllib.error.HTTPError as e:
            sys.exit(f"HTTP {e.code} on {method} {path}: {e.read().decode()[:300]}")

    _, login = call("POST", "/api/auth/login", body={"username": "HIVE", "password": password})
    tok = login.get("token")
    if not tok:
        sys.exit("login failed")
    _, profs = call("GET", "/api/llm/profiles", token=tok)
    items = profs if isinstance(profs, list) else profs.get("items", profs.get("profiles", []))
    if any(p.get("model") == MODEL for p in items):
        print("profile already exists in HIVE. Done.")
        return
    call("POST", "/api/llm/profiles", token=tok, body={
        "name": "Ollama-Gemma4-E2B", "format": "openai",
        "base_url": OLLAMA + "/v1", "model": MODEL, "api_key": "ollama",
        "streaming": True, "is_default": False,
    })
    print("profile created in HIVE. Go to Settings -> LLM to set it default or chain it.")


if __name__ == "__main__":
    main()
