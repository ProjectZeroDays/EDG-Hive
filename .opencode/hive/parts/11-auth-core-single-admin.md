# 11 — Auth core

**Goal:** One admin, everyone else a child.

**Steps**
1. First account becomes admin, unique-constrained.
2. Transfer CLI-only, behind the current admin's passkey.
3. `REGISTRATION_MODE`: open, invite, closed; auto-flips to invite on non-loopback bind.
4. Mode shown on the login screen.

**Done when**
- A second admin is impossible by any path; bind detection flips the mode.

**Commit:** `hive(11): single-admin auth core`
