# 01 — LLM pool repair

**Goal:** Planner stops 404ing. Roles never run dry.

**Steps**
1. Grep every hardcoded model id, list them in the commit body.
2. Resolve ids against live provider catalogs at boot.
3. Provider prefix normalization plus alias map for legacy ids.
4. Per-role fallback chain with a validation call before use.
5. 404 re-resolves; 429/5xx retries with backoff and jitter.
6. Detect the OpenRouter BYOK restriction, return an actionable error naming key and limit.
7. Log failing id, provider, status.

**Done when**
- Planner run completes on a healthy pool; mocked 404 recovers; empty pool names a reason; BYOK restriction surfaces the fix.

**Commit:** `hive(01): repair model pool resolution and fallback`
