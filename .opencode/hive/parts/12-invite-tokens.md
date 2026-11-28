# 12 — Invite tokens

**Goal:** Gated registration.

**Steps**
1. `hive_inv_` + 32 random bytes, base32; store SHA-256 only, show once.
2. Fields: label, max_uses, expires_at, revoke, created_by.
3. Redemption at `/register?token=`, validated and decremented in one transaction.
4. Timing-safe compare, generic error, 5/min/IP, audited.
5. Admin page with uses left, expiry, revoke, copy.

**Done when**
- Concurrent redemption never overshoots max_uses.

**Commit:** `hive(12): invite tokens`
