# 17 — Teaming

**Goal:** Admin can join a child live.

**Steps**
1. Per-child toggle, default off.
2. Max 4 concurrent joins, enforced server-side.
3. Joined admin sees screen and task, messages in-band, takes over.
4. Child sees who joined and when.

**Done when**
- The fifth join is refused; revoke ends the stream immediately.

**Commit:** `hive(17): teaming`
