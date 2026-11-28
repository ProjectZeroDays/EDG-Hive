# 04 — Permissions and notifications

**Goal:** One permission ask, useful alerts.

**Steps**
1. Single first-run popup: files, sound, push, camera, location, clipboard, with per-item state and re-request.
2. Browser notification on agent question, agent error, and every alert type.
3. Distinct sound per type, mute toggle, DND respected, repeats deduped.

**Done when**
- Denial degrades gracefully; each alert type fires and sounds.

**Commit:** `hive(04): permissions and notifications`
