# Hive Build Notes

## 2026-11-24 — Part 02 shipped

Commit on `main` — `hive(02): rebrand product surface to Hive`.

**Scope decision:** product-surface rename only. User-visible strings (app config/title/copyright, login + setup pages incl. Terms of Use dialog, sidebar alt text, update toast, logs sample, notify help/placeholders, report template title, guard/intercept platform-control prefix `[Hive platform control, not target defense]`, startup log line, README/SECURITY prose) now say **Hive**. Login/setup also switched from the old `logo.png` silhouette to the `cia2.png` seal already used in the sidebar.

**Deliberately left as ARTEX (contract, not brand):**
- Login username `ARTEX` + JWT subject (`server/auth.go`) — auth contract, part 11 reworks it.
- Storage keys: `artex_token` cookie/localStorage, `artex_*` preference keys.
- Env vars (`ARTEX_CONFIG`, `ARTEX_PG_DSN`, `ARTEX_KEY_DIR`, `ARTEX_NOTIFY_ALLOW_LOCAL`, budgets, etc.).
- Binary name `artex`, systemd unit, docker container `artex-pg`, repo paths, Go module path.
- Upstream URLs: BaaRai/ARTEX badges, releases feed, clone/install commands, SECURITY advisory channel.
- `server/intercept_live_test.go` fixture string (live-gated, wording differs, tests pass-through).

**Verified:** `tsc --noEmit` clean; biome findings on login page are pre-existing `useSortedClasses` nursery noise on untouched lines; `go test ./guard/ ./intercept/` green (tests assert the renamed constant).

**Ops:** the llama-server runner wedged again (99% CPU, 3.8GB RSS, swap 3/4 full) and OOM-killed `tsc`. Fix per part-01 notes: `kill -9 <llama-server pid>`; ollama.service stays active and respawns a runner on next request. Build needs the freed memory.

## 2026-11-23 — Part 01 shipped

Commit `a63a225` — `hive(01): repair model pool resolution and fallback`.

**Hardcoded model id audit (spec step 1):**
- `agent/provider.go` (8 sites): env-overridable defaults `gpt-4o`, `gpt-5`, `llama3.2`, `claude-opus-4-8` in `ConfigFromEnv`/`ConfigFrom`. Not pool ids; left.
- `ollama-setup.sh`, `/root/init_install.sh`: `OLLAMA_MODEL` default — already the corrected full hf.co id.
- `llmpool/catalog.go` `LegacyAliases`: the fix itself.
- No planner/pool code hardcodes a model id; profiles come from the DB.

**Verified live:**
- Boot pass surfaced the OpenRouter BYOK restriction as an actionable error naming `https://openrouter.ai/workspaces/default/byok` (journal, 14:59).
- Live pool run (`LIVE_POOL=1`, gated test `llmpool/live_test.go`): BYOK 404 → breaker trip with profile/model/status → failover to ollama → real completion: `reply="OK" via=end_turn tokens={in:21 out:2}` in 264.7s (cold runner load dominates; CPU-bound box).

**Environment notes:**
- 7 GB RAM box; llama-server (pid 95588, port 43357) holds ~3.1 GB and spins a core at ~98% continuously. Go builds OOM unless `-p 1` and orphaned `compile` procs are killed first.
- `go build` needs `-tags embedui` for a usable binary; `./build.sh` does it end-to-end.
- UI login is JWT (`POST /api/auth/login`, username hardcoded `ARTEX`); the DB DSN + key dir for the gated live test: `ARTEX_PG_DSN` from config.json database block, `ARTEX_KEY_DIR=/root/ARTEX` (secret.key lives there).

**Left for later parts:**
- openrouter/free profile is account-side BYOK-restricted — user must attach provider keys or switch profile; error now says so.

**Ops lessons:**
- llama-server can wedge into a 97%-CPU spin with an expired slot; symptom: ollama accepts but never dispatches. Fix: kill the runner, and if the ollama queue itself is stale, `systemctl restart ollama.service`.
- `pkill -f llmpool.test` also matches the wrapping bash -c (the pattern string is in its own cmdline) and SIGKILLs your own exec shell. Use `pkill -f 'llmpool[.]test'`.
- Model slot keep-alive is ~5 min (`--load-mode none`); idle pools pay a multi-minute cold load on this box.
- 02 (deep rename, second pass): module path → github.com/ProjectZeroDays/edg-hive; binary/service/dir/docker/db all hive-named; ARTEX_* env mirrored by server/legacy.go init; artex_* browser storage copied forward by web/src/lib/legacy.ts (wired in api.ts getToken); legacy artex_token cookie still accepted. Product grep for artex now returns only those two migration files. Box: /root/ARTEX → /root/HIVE, hive.service (artex.service removed), container hive-pg, DB user+name hive (269 tables preserved). Suite green, build clean, service live on 8787.
