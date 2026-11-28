# PHASE 1 — Foundation

## Role
Senior full-stack engineer and product architect, full repo access. Production-grade code.

## Objective
Fix the model pool, rebrand to **Hive**, establish the design system, build auth, accounts, messaging, diagnostics and lifecycle tooling. No network stack, no release.

(Full spec: sections 1–18 as supplied by operator. Key constraints: diff reviewable, preserve existing features, no TODO stubs, ask before destructive migrations. Phase 1 done-when list governs.)

## Design tokens

```css
:root[data-theme="dark"] {
 --sidebar:#141319; --pane:#1a1920; --bg:#1c1b22;
 --card:#434042; --panel:#262530; --border:#4e4b4e;
 --accent:#ff3e3e; --accent-hover:#ff5c5c; --accent-tint:rgba(255,62,62,.12);
 --info:#6b8cff; --text:#eceaf0; --muted:#c3bfc7;
}
:root[data-theme="light"] {
 --sidebar:#ececf0; --pane:#f4f4f7; --bg:#ffffff;
 --card:#ffffff; --panel:#fafafc; --border:#dcdce4;
 --accent:#1a56db; --accent-hover:#1445b3; --accent-tint:rgba(26,86,219,.1);
 --info:#1a56db; --text:#1c1b22; --muted:#5c5a63;
}
```

## Skill format (reference for part 08)

```markdown
---
name: port-sweep
description: Fast TCP port sweep of an in-scope host before enumeration.
version: 1.0.0
tags: [recon, tcp]
---

## When to use
Before service fingerprinting, on any host in scope.

## Steps
1. Run nmap -Pn -p- --min-rate 2000 against the target.
2. Feed open ports to the fingerprint step.

## Notes
Back off to 500 if the target rate-limits.
```

## Liability disclosure (Version 1)

BEFORE YOU START — You are being given a child account on a shared instance run by one admin.

KEYS AND COST: Use your own provider keys where you have them. Keys are yours and the spend is yours. Hosted models are metered against a cap the admin sets and can be withdrawn at any time.

WHAT YOU RUN IS YOURS: Prompts, agent output, and anything the agents do to a target are your responsibility. The admin does not warrant that a scan, exploit, or generated command is correct, safe, or in scope. Verify before you point anything at production.

WHAT THE ADMIN CAN SEE: Request counts, token usage, model used, cost, your logs, and your conversations with the AI. Conversations with the admin are private to the two of you.

NO WARRANTY: Provided as-is. No guarantee of uptime, data retention, or recovery.

By continuing you accept full responsibility for the actions of any agent you start, on any system, and confirm you have permission to test what you point them at.

[ ] I have read and accept Version 1 — recorded with timestamp and IP
