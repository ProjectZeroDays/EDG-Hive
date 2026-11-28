# ARTEX Enterprise GUI — Mockup Pack

Open `index.html` in any browser. Use the pills at the top to switch between the six
screens; the **◐ Theme** toggle previews the light variant, **⇱ Density** toggles a
compact mode. Numbered chips on each screen map to the feature index below.

This pack is a *structural* proposal (layout, hierarchy, workflow). Visual design
tokens (color, type, spacing) would be finalized in a design-system pass on top of
these structures.

## Screens

| # | File section | Replaces / extends | Enterprise capabilities shown |
|---|--------------|--------------------|-------------------------------|
| 1 | `s1` Command Dashboard | sparse home page | cross-engagement posture, risk trend, coverage, operator queue with SLAs, live agent feed |
| 2 | `s2` Engagement Workspace | per-task tabs + graph + transcript | 3-pane workbench, node inspector, steering, side-question inbox, run control, run diff, ⌘K palette, context meter |
| 3 | `s3` Findings Triage | findings table | kanban lifecycle, dedup, assignment, SLAs, retest windows, tracker sync, CVSS 4.0 + ATT&CK, dual control |
| 4 | `s4` Approvals Center | approval records list | risk-scored queue, explainable blocks, argv diff, decision verbs, policy presets, guarded bulk, break-glass |
| 5 | `s5` Traffic Workbench | traffic table | split panes, Lucene search, flow diff, replay, annotation → evidence, redaction lens, compare |
| 6 | `s6` Governance & Admin | system settings | tenancy, RBAC matrix, JIT elevation, scoped API keys, hash-chained audit, LLM cost, residency |

## Feature index (chip → feature)

**Dashboard — D**
D1–D5 KPI strip · D6 risk-exposure trend · D7 severity mix · D8 coverage by asset group · D9 live agent feed · D4 unified "needs a human" queue (approvals + side questions + sign-offs)

**Workspace — W**
W1 graph v2 (layouts/minimap/outcome color/why-node) · W2 node inspector (rationale, route, evidence, cost) · W3 steering with receipts · W4 side-question inbox · W5 run comparison · W6 live run control (pace/egress/budget) · W7 ⌘K command palette · W8 per-agent context meter

**Findings — F**
F1 saved views · F2 triage board · F3 dedup/merge · F4 Jira/GitHub/ServiceNow sync · F5 retest windows · F6 detail drawer (CVSS 4.0, ATT&CK, blast radius, chain) · F7 one-click retest · F8 dual-control transitions

**Approvals — A**
A1 explainable blocks · A2 argv diff vs baseline · A3 decision verbs (once/pin/modify/deny) · A4 hash-chained audit trail · A5 policy presets (Paranoid/Balanced/Aggressive) · A6 guarded bulk approve · A7 break-glass

**Traffic — T**
T1 Lucene search + saved searches · T2 request/response split panes · T3 flow diff · T4 replay · T5 annotation → evidence · T6 PII redaction lens · T7 flow compare

**Governance — G**
G1 tenancy · G2 RBAC matrix (server-enforced) · G3 JIT elevation · G4 scoped API keys · G5 immutable audit + SIEM · G6 LLM pool health · G7 cost governance · G8 residency & retention

## Implementation notes

- Graph v2 and the workbench reuse the installed `@antv/g6` / `@xyflow/react` and `cmdk` (palette) dependencies — no new heavy deps required.
- D6/D8 read data the server already computes (findings, coverage ledger); the dashboard is mostly a new aggregation endpoint + widgets.
- A1–A7 surface the existing `intercept` judge/risk/pin machinery instead of adding a parallel system.
- G2 requires an authorization middleware layer in the Go mux (~270 routes) plus a `roles`/`permissions` table — the largest single backend item in this pack.
