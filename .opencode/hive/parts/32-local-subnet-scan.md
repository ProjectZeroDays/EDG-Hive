# 32 — Local subnet scan

**Goal:** Reach the LAN when hosted locally.

**Steps**
1. Detect interface and subnet mask.
2. Bind to LAN, accept RFC1918 CIDR, refuse targets outside the subnet unless overridden.

**Done when**
- The subnet is detected correctly and out-of-range targets are refused.

**Commit:** `hive(32): local subnet scan`
