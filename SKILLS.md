# HIVE Skills — format, compatibility, and upload rules

A skill is a directory under `skills/` (the `skills.dir` from `config.json`).
Each skill bundles instructions the planner/worker agents load when the
skill's description matches the task. Bundled examples: `api-recon`,
`bbot-recon`, `codeaudit`, `playwright-cli`.

## Directory layout

```
skills/<skill-name>/
  SKILL.md        required
  reference.md    optional deep-dive the agent reads on demand
  scripts/        optional helper programs (python, node, bash)
```

## SKILL.md format

YAML frontmatter, then markdown body:

```markdown
---
name: my-skill
description: Invoke this skill when <the situation that triggers it>.
mcps: SomeMcpName        # optional; MCP server the skill relies on
---

# My Skill (one-line purpose)

## Boundaries and prohibitions
| Scope | Allowed | Prohibited |
|---|---|---|
| Target | ... | ... |

### Hard process prohibitions
| Prohibited | Do this instead |
|---|---|
| ... | ... |

## Phases
1. ...
2. ...

## Definition of done
- ...
```

### Required elements

1. **Frontmatter** — `name` must equal the directory name; `description` is
   what the router matches on, so write the *trigger condition*, not marketing.
2. **Boundaries table** — what the agent may and may not do under this skill.
   This is the scope-of-authority contract; keep it explicit.
3. **Hard prohibitions** — irreversible actions, real credentials, destructive
   operations, and the *approved alternative* for each.
4. **Definition of done** — the observable condition that ends the skill.

### Optional but recommended

- `mcps:` frontmatter entry naming an MCP server (must be configured in the
  app) for tools the scripts call.
- `reference.md` for long material the agent should load lazily.
- Phase-by-phase process with explicit gates (see `api-recon`).

## What is compatible to upload

| Compatible | Notes |
|---|---|
| Markdown docs (`SKILL.md`, `reference.md`) | primary content |
| Python 3 scripts | stdlib preferred; pin deps in a header comment |
| Node/Bun scripts | state the runtime in a header comment |
| Shell scripts | bash; must be non-interactive |
| Config templates (yaml/json/txt) | no secrets in them |

## What is NOT compatible — will be rejected or must never be included

- Secrets: API keys, tokens, session cookies, credentials of any kind.
  Use placeholders (`OUTDIR`, env vars) and document where the value comes
  from at runtime.
- Target-specific data: hosts, IPs, findings from engagements. Skills are
  reusable tooling, not engagement records.
- Compiled binaries. Scripts only, human-readable.
- Hardcoded absolute paths to a developer's machine. Parameterize
  (`OUTDIR`, `$1`, env vars).
- Interactive/prompting scripts. Agents run these headless.
- Evasive or destructive tooling without an explicit authorized-engagement
  preamble and boundaries table. Every skill states its authorization frame.

## How agents consume scripts

Scripts in `scripts/` are templates. The agent copies them to its working
directory (`OUTDIR`), adapts them to the target, and runs them under the
task's intercept/approval rules. Never instruct agents to run originals
in-place from `scripts/` — copies keep the repo skill pristine and auditable.

## Testing a skill before committing

1. Drop it in `skills/<name>/` on a dev HIVE instance.
2. Create a task whose goal triggers the description; confirm the planner
   loads the skill (task detail shows the skill binding).
3. Run once in a lab target inside the declared scope; check the boundary
   tables actually gated behavior.
4. Remove any runtime state from the skill dir before committing
   (`git status` should show only your source files).
