# 14 — Disclosure gate

**Goal:** Liability accepted before use.

**Steps**
1. Versioned file, not a template string.
2. Modal, scroll-to-bottom before accept enables, blocks the composer.
3. Record user, version, timestamp, IP; re-prompt on bump.
4. Admin page shows who accepted.

Copy v1: child account under one admin; own keys and own cost; hosted usage metered and capped; prompts and output are the user's responsibility; admin sees usage and logs; no warranty; user accepts responsibility for agent actions and confirms permission to test targets.

**Done when**
- Composer blocked until accepted; bump re-prompts.

**Commit:** `hive(14): liability disclosure gate`
