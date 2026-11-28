# 08 — Skill format validation

**Goal:** Enforce the skill schema on upload.
Format: markdown with frontmatter `name`, `description`, `version`, `tags`; body in imperative steps; filename slug equals `name`; 64 KB cap; multi-file as `.zip` with `SKILL.md` at root.

**Steps**
1. Validate on upload, reject with the failing field named.
2. Description must be one sentence phrased "use when...".
3. Confirm the real loader keys and document them.

**Done when**
- Valid skill installs; each rule rejects a fixture.

**Commit:** `hive(08): skill format validation`
