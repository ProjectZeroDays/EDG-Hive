# 29 — DNS resolver policy

**Goal:** No-logging resolution only.

**Steps**
1. Allowlist: DNSSEC-validating, DoH, no-logging, fast.
2. Prefer Google and Cloudflare DoH.
3. Reject others with the reason shown.

**Done when**
- A non-compliant resolver is refused.

**Commit:** `hive(29): DNS resolver policy`
