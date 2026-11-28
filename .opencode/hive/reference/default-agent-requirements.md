# Default requirements for coding agents — generic template

Received 2026-11-24 from operator. Applies Hive-wide; map each item to the
queue before implementing. This repo is Go + Next.js, not Python — items
below marked (py) are template language that does not apply verbatim.

## 1. Automatic updates
Startup check of GitHub Releases API; if newer than embedded version, log a
notice, expose `/api/update/check` + `/api/update/download`, dashboard banner
with one-click download. No daemon — check on startup + manual button.
**Hive status:** already shipped upstream — `server/update_check.go`
(opt-in via `ARTEX_UPDATE_CHECK`, polls BaaRai/ARTEX releases), toast banner in
`main-content.tsx`. Revisit under part 35 (release) to point the feed at the
fork once the deep rename (part 02 decision) settles.

## 2. GitHub workers
- (py) pytest suite on every dependabot PR — Hive runs `go test ./...`, no
  dependabot config yet; add under part 35.
- Auto-merge when tests pass — part 35.
- (py) Release workflow building APK + EXE + NSIS + DEB + DMG — template
  language. Hive ships a single static Go binary (`build.sh` →
  `dist/artex-linux-amd64/`); release workflow under part 35 builds per-target
  Go binaries and publishes release notes.

## 3. (py) NSIS native installer
Template assumes a Python standalone EXE. Hive is one static binary; a Windows
installer is a part-35 decision, likely a zip + optional NSIS wrapper. The
existing `init_install.sh` one-shot bootstrap covers Linux.

## 4. Enterprise README generator requirements
Fortune-500-grade README: banner/logo, hero intro, executive summary,
shields.io badges, horizontal text-link menu (Download / Report a Bug / Wiki /
Website), features, architecture diagram, repo structure, dependencies,
installation, configuration, usage, tutorials, CLI reference, API docs,
reporting, logging, security, troubleshooting, FAQ, statistics, performance
notes, roadmap, contributing, code of conduct, support, glossary, author,
license, acknowledgments, disclaimer. GitHub markdown only; fully linked TOC;
`<details>`/`<summary>` collapsibles; ASCII diagrams; syntax-highlighted
examples.
**Hive status:** belongs to part 27 (documentation set). Current README is a
solid mid-tier rewrite; upgrade it to the full enterprise spec in that part.
