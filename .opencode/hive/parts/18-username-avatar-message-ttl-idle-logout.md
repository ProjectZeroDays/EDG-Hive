# 18 — Account features

**Goal:** Identity and hygiene.

**Steps**
1. Username set and change, unique, rate-limited.
2. Avatar upload with crop, shown bottom-left.
3. Disappearing messages: off, 1h, 24h, 7d, deleted server-side by scheduler.
4. 15-minute inactivity logout for children, configurable.

**Done when**
- TTL messages are gone from the DB after expiry; idle logout fires.

**Commit:** `hive(18): account features`
