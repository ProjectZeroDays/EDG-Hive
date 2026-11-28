# 26 — Kill switch

**Goal:** Teardown leaving no residue.

**Steps**
1. Confirmation phrase first.
2. Optional export.
3. Wipe logs, delete accounts, drop DB, remove volumes.
4. Remove install dir, unit file, cron, tunnel.

**Done when**
- A clean VPS afterwards has no dir, unit, DB, tunnel or cron; aborting mid-way is resumable; reinstall succeeds.

**Commit:** `hive(26): kill switch teardown`
