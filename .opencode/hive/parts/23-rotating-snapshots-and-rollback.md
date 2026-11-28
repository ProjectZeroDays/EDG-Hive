# 23 — Snapshots

**Goal:** Recoverable state.

**Steps**
1. Every 4 hours, keep 3, newest overwrites the oldest.
2. Database, config and uploads only, compressed and checksummed, stored outside the served directory.
3. Admin rollback with confirmation and an automatic pre-rollback snapshot.

**Done when**
- Rotation holds at 3; rollback restores and verifies the checksum.

**Commit:** `hive(23): rotating snapshots`
