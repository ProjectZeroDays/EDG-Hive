<div align="center">

  <p align="center">
  <img src="/web/public/logo.png" alt="Hive Logo" width="150" height=150">
</p>
  
<p align="center">
  <img src="/web/public/hive-README.png" alt="Hive Logo" width="400">
</p>

---

**Autonomous penetration testing that keeps its receipts.**

----

[![Release](https://img.shields.io/github/v/release/ProjectZeroDays/EDG-Hive?label=release&color=6b34db)](https://github.com/ProjectZeroDays/EDG-Hive/releases)
[![License](https://img.shields.io/github/license/ProjectZeroDays/EDG-Hive?color=6b34db)](LICENSE)
[![CI](https://github.com/ProjectZeroDays/EDG-Hive/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/ProjectZeroDays/EDG-Hive/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/ProjectZeroDays/EDG-Hive?color=6b34db)](go.mod)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20windows%20%7C%20macos-6b34db)](https://github.com/ProjectZeroDays/EDG-Hive/releases)
[![Security](https://img.shields.io/badge/security-AGPL--3.0%20%2B%20scope%20gates-6b34db)](SECURITY.md)
[![Maintained](https://img.shields.io/badge/maintained-yes-6b34db)](https://github.com/ProjectZeroDays/EDG-Hive/commits/main)

<a href="#quick-start">Quick Start</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#installation">Installation</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#api-documentation">API</a> ·
  <a href="#security-considerations">Security</a> ·
  <a href="#troubleshooting">Troubleshooting</a> ·
  <a href="https://github.com/ProjectZeroDays/EDG-HIVE/wiki/HIVE">Wiki</a> 
</div>

---

## Table of Contents
-
- [Executive Summary](#executive-summary)
- [Features](#features)
- [Architecture](#architecture)
- [Repository Structure](#repository-structure)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Dependencies](#dependencies)
- [Configuration](#configuration)
- [Usage](#usage)
- [CLI Reference](#cli-reference)
- [API Documentation](#api-documentation)
- [Reporting](#reporting)
- [Logging](#logging)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [Glossary](#glossary)
- [Project Statistics](#project-statistics)
- [Performance Notes](#performance-notes)
- [Roadmap](#roadmap--future-updates)
- [Contributing](#contributing-guidelines)
- [Code of Conduct](#code-of-conduct)
- [Support](#support)
- [Author](#author)
- [License](#license)
- [Acknowledgments](#acknowledgments)
- [Disclaimer](#disclaimer)

---

## Executive Summary

Hive is a self-hosted, LLM-driven penetration testing platform. It is not a chat wrapper around a scanner: a **planner** decomposes an engagement goal into intents, **workers** execute those intents with real tools, and results are written back into **two linked graphs** — one describing what is known about the target, one describing what has been tried and what it proved. Every finding can carry the captured HTTP request/response that justifies it, and every autonomous action passes through scope rules, approval gates, and budgets.

- **Language:** Go (backend, single static binary) · TypeScript / Next.js (embedded frontend)
- **Database:** PostgreSQL 16
- **Models:** bring your own — Anthropic, any OpenAI-compatible endpoint, or a local Ollama
- **Deployment:** one binary + one database container; Docker Compose optional
- **License:** AGPL-3.0

---

## Features

- **Closed-loop autonomy.** Goal decomposition → intent dispatch → worker execution → write-back. The loop advances until the goal is proven or budgets are exhausted.
- **Dual-graph knowledge model.** A shared asset graph (what exists) and a per-task exploration graph (what was tried, what it produced, why), linked by anchors that power asset coverage.
- **Evidence-bound findings.** A recording MITM proxy and a browser extension capture real traffic; findings, evidence quotes, and retests stay tied to the exact request/response.
- **Autonomy with brakes.** Human-in-the-loop approvals, argv-prefix interception rules, scope/deny gates, token and turn budgets, a target noise governor, and prompt-injection framing on untrusted tool output.
- **Skills system.** Bundle instructions, references, and helper scripts per capability; the planner loads them when their description matches the task.
- **Model-agnostic LLM pool.** Priority-ordered failover chain across providers, per-profile breakers, rate limits, and usage accounting.
- **Full audit trail.** Every LLM call, tool invocation, and planning round is recorded and reviewable in the UI.

<details>
<summary>Feature deep-dive (click to expand)</summary>

| Area | Capabilities |
|---|---|
| Planning | Multi-round planning with quiet-mode edge suppression, heartbeat supervision, goal-metric judgment |
| Workers | Multi-worker execution, intent claims with TTL heartbeats, steer/kill mid-flight, retry and rerun |
| Evidence | Traffic recording proxy, per-finding evidence binding, retest workflow, evidence quotes |
| Control | Intercept approvals (single/batch), target gate, noise governor, run budgets |
| Extensibility | Custom tools, MCP servers, task templates, notification channels, skills upload |
| Operations | Systemd deployment, structured logs, LLM health tracking, coverage ledger |

</details>

---

## Architecture

```text
               ┌────────────────────────────────────────┐
               │               Operator                 │
               │  Web UI (Next.js, embedded in binary)  │
               └───────────────────┬────────────────────┘
                                   │ HTTPS / JWT
             ┌─────────────────────▼─────────────────────┐
             │              Hive (Go binary)             │
             │                                           │
             │  ┌─────────┐   ┌────────┐  ┌───────────┐  │
             │  │ Planner │─▶│ Engine │─▶│  Workers  │  │
             |  └────┬────┘   └───┬────┘  └─────┬─────┘  │
             │       │            │             │        │
             │  Exploration    LLM pool     Tools/MCP    │
             │     graph      (failover)     agents      │
             └────────┬─────────────────────────┬────────┘
                      │                         │
          ┌───────────▼──────────┐  ┌───────────▼──────────┐
          │     PostgreSQL 16    │  │ Targets (authorized) │
          │  assets · findings · │  │ MITM proxy :8788     │
          │  traffic · activity  │  │ browser extension    │
          └──────────────────────┘  └──────────────────────┘
```

- **Planner.** One per active task. Receives triggers (worker done, finding, heartbeat), judges goal metrics, and spawns/steers/kills intents.
- **Workers.** Claim intents, run real tools (and MCP-served tools), and write facts/findings back to the graphs.
- **Dual graphs.** `goal / intent / fact / finding / hint` nodes with `spawns / derived_from / yields / proves` lineage, anchored to the shared `root_domain → subdomain → ip → service → app → endpoint` asset graph.
- **Interception.** Risky tool invocations pause at the intercept gate pending operator approval (single or batch).
- **Traffic evidence.** A recording proxy captures full request/response pairs; findings bind to traffic by ID.

---

## Repository Structure

```text
hive/
├── cmd/hive/            # application entrypoint
├── server/               # HTTP API, engine, planner, workers, intercept, LLM pool
├── agent/                # agent runtime, prompts, tool dispatch
├── skills/               # bundled skill packs (SKILL.md format)
├── web/                  # Next.js frontend (exported & embedded at build time)
├── db/                   # migrations, store layer
├── intercept/            # interception & approval engine
├── traffic/              # recording MITM proxy
├── llmpool/              # provider failover, breakers, usage accounting
├── extension/            # browser extension for traffic capture
├── docs/                 # design and technology notes
├── scripts/              # operational helpers
├── install.sh            # guided installer (all-Docker or local build)
├── init_install.sh       # one-shot Ubuntu/Debian bootstrap
├── init_install.ps1      # one-shot Windows bootstrap
├── build.sh              # cross-platform release builder (embeds frontend)
├── docker-compose.yml    # reference deployment (PostgreSQL + app)
└── config.example.json   # database + skill_dir configuration template
```

---

## Quick Start

```bash
git clone https://github.com/ProjectZeroDays/EDG-Hive.git
cd HIVE
cp .env.example .env            # set POSTGRES_PASSWORD (and optionally an LLM key)
docker compose up -d --build    # builds from this checkout, starts PostgreSQL
```

Open **http://localhost:8787**. On first visit, `/setup` asks for the one-time setup token printed at startup:

```bash
docker compose logs hive | grep "Setup token"
```

> There is no prebuilt image to pull; the image is always built from your checkout.

---

## Installation

<details open>
<summary><strong>Docker Compose (recommended)</strong></summary>

```bash
git clone https://github.com/ProjectZeroDays/EDG-Hive.git && cd HIVE
cp .env.example .env
docker compose up -d --build
```

</details>

<details>
<summary><strong>Install script (all-Docker or local build)</strong></summary>

```bash
./install.sh
```

The script detects Docker, offers an all-Docker deployment or a local build-and-run path, and installs a systemd unit on Linux.

</details>

<details>
<summary><strong>One-shot bootstrap (Ubuntu/Debian, as root)</strong></summary>

```bash
curl -fsSL https://raw.githubusercontent.com/ProjectZeroDays/EDG-Hive/main/init_install.sh | bash
```

Performs, in order: dependency install → clone/build + systemd unit → PostgreSQL container → optional Cloudflare quick tunnel → admin setup → optional local Ollama with a default model → engagement target creation.

A PowerShell equivalent is provided as `init_install.ps1`.

</details>

<details>
<summary><strong>From source</strong></summary>

```bash
cd web && npm ci && npm run build:static
rsync -a --delete web/out/ server/webui/dist/
GOTOOLCHAIN=auto ./build.sh     # produces dist/hive-<os>-<arch>/hive
```

The release binary embeds the exported frontend (`-tags embedui`); the result is a single self-contained executable.

</details>

---

## Dependencies

| Component | Requirement | Notes |
|---|---|---|
| Go | 1.26+ (toolchain auto-downloads) | backend build |
| Node.js + npm | 24 LTS recommended | frontend build |
| PostgreSQL | 16+ | required at runtime |
| Docker / Compose | any recent release | optional but recommended |
| LLM provider | Anthropic, OpenAI-compatible, or Ollama | user-supplied |
| OS | Linux, Windows, macOS | amd64/arm64 |

---

## Configuration

Hive reads `config.json` (override with `HIVE_CONFIG`). Only database and skill root live here; LLM profiles and operational settings are managed in the UI.

```json
{
  "database": {
    "host": "127.0.0.1",
    "port": 5432,
    "user": "hive",
    "password": "change-me",
    "dbname": "hive",
    "sslmode": "disable"
  },
  "skill_dir": "/opt/hive/skills"
}
```

**Environment overrides** (take precedence over the file):

| Variable | Purpose |
|---|---|
| `HIVE_CONFIG` | path to config file |
| `HIVE_PG_DSN` | full PostgreSQL DSN |
| `HIVE_SKILL_DIR` | skill root directory |

LLM profiles (provider, model, base URL, keys, rate limits, failover priority) are created under **System → LLM** in the UI and stored in PostgreSQL. A one-time setup token gates the first-run admin account; it is printed once at startup.

**Database defaults.** With no `HIVE_PG_DSN` and no `database` section in the config file, the app connects to the built-in default `postgres://hive:hive@127.0.0.1:5432/hive?sslmode=disable` (the user/password/database the quick-start `init_install.sh` creates). While those default credentials are in effect, every login shows a prompt to rotate the database password; the new password is written into `config.json` (`0600`). Set `HIVE_PG_DSN` or a `database` section to opt out of the defaults.

---

## Usage

1. **Log in** at `http://localhost:8787` with the admin account created during setup.
2. **Add an LLM provider** under System → LLM (test the profile before saving).
3. **Create a task**: name it, set the goal, define scope (domains/IPs/CIDRs/URLs).
4. **Approve or auto-rule** intercepts as workers propose risky actions.
5. **Review**: exploration graph, findings with evidence, coverage graph, retests.
6. **Report**: export findings for the deliverable.

---

## CLI Reference

Hive ships as a single binary; day-to-day operation is via the UI and API.

```bash
hive                      # start server (config from HIVE_CONFIG or ./config.json)
hive -config /path/config.json   # explicit config path
```

| Surface | Command / path |
|---|---|
| Service (systemd) | `systemctl status hive` |
| Reset admin password | `./reset-password.sh` |
| Update | `./update.sh` |
| Build release | `./build.sh [--release]` |

---

## API Documentation

REST API served under `/api` (JWT bearer auth; `/api/auth/*` and `/api/health` are public).

<details>
<summary><strong>Representative endpoints</strong></summary>

| Group | Endpoints |
|---|---|
| Auth | `POST /api/auth/login` · `POST /api/auth/init` · `GET /api/auth/status` |
| System | `GET /api/health` · `GET/PATCH /api/settings` |
| Tasks | `GET/POST /api/tasks` · `GET /api/tasks/{id}` · `POST /api/tasks/{id}/chat` · `POST /api/tasks/{id}/intents/{iid}` · `GET /api/tasks/{id}/lessons` |
| Agents | `GET/POST /api/agents` · `GET/PUT /api/agents/planner/prompt` |
| LLM | `GET /api/llm/profiles` · `POST /api/llm/profiles` · `POST /api/llm/test` · `GET /api/llm/models` |
| Companies & scope | `GET/POST /api/companies` · `GET /api/companies/{id}/scope` |
| Exploration | `GET /api/exploration/findings/{id}/traffic` · `POST /api/exploration/findings/{id}/deepen` |
| Intercept | `GET /api/intercept/pending` · `POST /api/intercept/pending/batch-decide` · `GET /api/intercept/history` |
| Traffic | `GET /api/traffic` · `POST /api/traffic/ingest` |
| Tools & MCP | `GET /api/tools` · `POST /api/tools/custom` · `GET/POST /api/mcp` |
| Skills | `POST /api/skills/upload` |
| Notifications | `GET/POST /api/notify/channels` |
| Task templates | `GET/POST /api/task-templates` |

**Example — authenticate:**

```bash
curl -s -X POST http://localhost:8787/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"***"}'
```

```json
{ "token": "eyJhbGciOi...", "user": { "username": "admin", "role": "admin" } }
```

**Example — create a task:**

```bash
curl -s -X POST http://localhost:8787/api/tasks \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"External pentest — ACME","goal":"Map and test the external attack surface","scope":["acme.example","192.0.2.0/24"]}'
```

</details>

---

## Reporting

Findings accumulate per task with severity, evidence bindings, and retest state. The UI provides:

- Findings list with filters and detail pages (evidence quotes, linked traffic)
- Coverage graph per asset (tested vs. untouched)
- Exploration graph for reasoning lineage
- Activity/audit log for every planning round and tool call

Export via the API (`/api/tasks/{id}/outcome`, findings endpoints) for inclusion in client deliverables.

---

## Logging

- Backend: structured logs to stdout/journald (`journalctl -u hive -f`).
- LLM records: every call logged with tokens, latency, and profile (UI: System → LLM records).
- Traffic: full request/response pairs recorded by the proxy, reviewable per finding.

---

## Security Considerations

Hive is an offensive tool. Treat deployment accordingly.

- **Authorization only.** Use exclusively against systems you own or are explicitly authorized to test. See [Disclaimer](#disclaimer).
- **Bind locally.** Default deployment listens on loopback; put it behind authenticated reverse proxy/tunnel for remote access.
- **Setup token** is single-use and printed once; protect startup logs.
- **Secrets** live in PostgreSQL and config files — restrict file permissions, rotate LLM keys, never commit `.env`.
- **Intercept gates** and scope/deny rules are your safety rails; review defaults before unattended runs.
- **Untrusted output framing:** tool output is delimited and framed to reduce prompt-injection risk from target content.
- Vulnerability reporting: see [SECURITY.md](SECURITY.md).

---

## Troubleshooting

<details>
<summary><strong>Startup</strong></summary>

| Symptom | Cause | Fix |
|---|---|---|
| `Frontend is not embedded in this binary` | binary built without `-tags embedui` | rebuild via `./build.sh` |
| `connection refused` on 8787 | service not running / wrong bind | `systemctl status hive`; check `OLLAMA_HOST`-style env and config |
| Setup token not shown | logs rotated | restart once to print a fresh token |

</details>

<details>
<summary><strong>Database</strong></summary>

| Symptom | Fix |
|---|---|
| `database ... does not exist` | create role/db or point `HIVE_PG_DSN` at a valid DSN |
| migrations fail after downgrade | restore from backup; downgrade is not supported |

</details>

<details>
<summary><strong>LLM pool</strong></summary>

| Symptom | Fix |
|---|---|
| All profiles unavailable | check `llm_profiles` rows, API keys, base URLs; test via System → LLM |
| Ollama 404 model not found | use the exact model name from `ollama list` |
| `exceeds the available context size` | raise `OLLAMA_CONTEXT_LENGTH` or use a larger-context model |
| Breaker keeps tripping | delete tripped rows in `llm_profile_health` after fixing the root cause |

</details>

<details>
<summary><strong>Workers / planning</strong></summary>

| Symptom | Fix |
|---|---|
| Task stuck in `created` | normal while rounds run; check activity feed and planner logs |
| Round takes very long | local small models on CPU are slow; prefer GPU or cloud provider |
| Intercept queue grows | review pending approvals; add auto-rules where appropriate |

</details>

---

## FAQ

<details>
<summary><strong>Is Hive a scanner?</strong></summary>
No. Scanners execute fixed checks; Hive plans against a goal, chooses tools, and keeps an auditable reasoning trail bound to evidence.

</details>

<details>
<summary><strong>Which LLMs are supported?</strong></summary>
Anthropic, any OpenAI-compatible endpoint (including OpenRouter), and local Ollama. Profiles form a priority failover chain.

</details>

<details>
<summary><strong>Does it run on Windows?</strong></summary>
Yes — release archives include Windows binaries, and `init_install.ps1` bootstraps a native install. Linux is the primary target.

</details>

<details>
<summary><strong>Where does my data live?</strong></summary>
Entirely in your PostgreSQL instance and on-disk config. No telemetry, no cloud dependency.

</details>

---

## Glossary

| Term | Meaning |
|---|---|
| **Goal** | The engagement objective a task must prove |
| **Intent** | A planned unit of work toward a goal |
| **Fact** | An observation written back by a worker |
| **Finding** | A security-relevant conclusion, evidence-bound |
| **Exploration graph** | Per-task lineage of goals, intents, facts, findings |
| **Asset graph** | Shared model of target entities and relationships |
| **Anchor** | Link between exploration nodes and assets |
| **Intercept** | Operator approval gate for risky actions |
| **Coverage** | Which assets were exercised by which intents |
| **Skill** | Bundled capability (instructions + helpers) loaded by the planner |
| **LLM pool** | Failover chain of provider profiles |
| **Quiet mode** | Suppression of redundant planning edges while work is in flight |

---

## Project Statistics

| Metric | Value |
|---|---|
| Backend | Go (single embedded-binary deployment) |
| Frontend | Next.js / TypeScript |
| Database tables | 60+ (assets, findings, traffic, activity, …) |
| API surface | 20+ resource groups under `/api` |
| Bundled skills | api-recon, api-security, bbot-recon, codeaudit, edr-bypass-re, patch-diff, php2shell, playwright-cli, rea, scopesentry, + community packs |
| Platforms | linux/amd64 · linux/arm64 · windows/amd64 · darwin |

---

## Performance Notes

- **Planner rounds** cost one LLM call each; local 7B-class models on CPU can take tens of minutes for large prompts — use GPU or a cloud profile for latency-sensitive work.
- **Workers** parallelize across intents; per-task budgets and the noise governor prevent target overload.
- **Traffic recording** is bounded by disk; prune old captures on long engagements.
- **PostgreSQL** is the bottleneck at scale; give it fast disks and keep `llm_records`/`traffic` retention policies.

---

## Roadmap / Future Updates

- Additional provider formats and streaming improvements
- Richer report generation and client-ready exports
- Expanded skill marketplace and versioning
- Enhanced multi-target campaign orchestration
- Performance work on large exploration graphs

Track progress via [GitHub issues](https://github.com/ProjectZeroDays/EDG-Hive/issues) and [releases](https://github.com/ProjectZeroDays/EDG-Hive/releases).

---

## Contributing Guidelines

1. Fork and branch from `main`.
2. Keep changes focused; open an issue first for large design shifts.
3. Backend: `gofmt`, keep the embed build working (`./build.sh`).
4. Frontend: `npm run check` (Biome) before committing.
5. Update docs and this README with behavior changes.
6. Never commit secrets, real targets, or captured traffic.

---

## Code of Conduct

Be professional, specific, and respectful. Security work attracts strong opinions; critique ideas and evidence, not people. Harassment and disclosure of others' confidential data are grounds for removal.

---

## Support

- Issues: [github.com/ProjectZeroDays/EDG-Hive/issues](https://github.com/ProjectZeroDays/EDG-Hive/issues)
- Security issues: see [SECURITY.md](SECURITY.md) (do not open public issues)
- Documentation: [docs/](docs)

---

## Author

**ProjectZeroDays** — [github.com/BaaRai](https://github.com/ProjectZeroDays)

---

## License

Distributed under the [GNU AGPL v3.0](LICENSE). Commercial licensing inquiries: contact the author.

---

## Acknowledgments

- The llama.cpp and Ollama communities for local inference
- The MITM proxy and browser-extension ecosystems
- Contributors and early testers

---

## Disclaimer

Hive is provided for **authorized security testing only**. Running it against systems without explicit written authorization is illegal in most jurisdictions and may result in criminal penalties. The authors accept no liability for misuse. You are solely responsible for complying with all applicable laws and engagement rules of scope. If you are unsure whether a target is in scope, do not test it.

[hive](https://github.com/ProjectZeroDays/EDG-HIVE/blob/master/web/public/hive-README.png)
