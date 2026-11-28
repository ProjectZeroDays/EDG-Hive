# 25 — Change identity

**Goal:** Rotate instance identity.

**Steps**
1. Regenerate instance id, hostname, session secret, TLS material, invite set.
2. Require re-auth.

**Done when**
- Old invites and sessions die; TLS reissues.

**Commit:** `hive(25): change identity`
