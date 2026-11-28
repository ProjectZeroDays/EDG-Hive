# 02 — Rebrand

**Goal:** No trace of "Artex".

**Steps**
1. Rename in package, metadata, titles, about, config keys, directories, strings, comments, CI, Docker, env prefixes.
2. Theme-aware logo on login.
3. Migration so existing installs and configs keep working.

**Done when**
- `grep -ri artex` returns only the migration file; existing config upgrades.

**Commit:** `hive(02): rebrand to Hive`
