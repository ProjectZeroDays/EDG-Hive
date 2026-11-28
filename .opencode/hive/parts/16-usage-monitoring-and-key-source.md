# 16 — Usage monitoring and key source

**Goal:** Visibility and cost control.

**Steps**
1. Per-child requests, tokens in/out, model, cost, daily rollup, date filter, filtered logs.
2. Spend threshold flag.
3. Key source `byok` (encrypted at rest) or `hosted` (grant plus cap); neither blocks prompts with a named reason; every request tagged with its source.

**Done when**
- A `byok` child with no keys and no grant is refused.

**Commit:** `hive(16): usage monitoring and key source`
