# 15 — User states and management actions

**Goal:** Admin control over accounts.

**Steps**
1. States active, suspended, disabled, pending. Disable kills live sessions within one request cycle.
2. Actions: disable, suspend, reinstate, reset passkey, force re-enroll, delete, extend invite.

**Done when**
- A disabled session fails on its next request.

**Commit:** `hive(15): user states and actions`
