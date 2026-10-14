# Beast — a Salesforce-native IDE

**Mission:** turn this px0 fork into a personal "beast" IDE for Salesforce development — px0's speed and review-first UX, plus the power of a full SF dev environment and a real extension system.

## How the mission runs

- The **beast-builder** Devin automation watches this repo: every **newly opened GitHub issue** spawns a builder session that implements it and opens a PR. File an issue = give a work order.
- This file is the mission tracker. Builder sessions update the feature checklist as they land changes.
- Feature specs are mined from existing tools (VS Code Salesforce Extension Pack, Code Builder, Illuminated Cloud) using [`morluto/rea`](https://github.com/morluto/rea) — `npx rea-agents` static JS/Electron analysis, no native engine needed for JS targets. Specs land in `docs/specs/`.

## Invariants (from px0's AGENT.md — do not break)

- Single static Go binary, zero runtime deps for the user
- No workspace litter — config/plugins live in `~/.beast/` (or `~/.px0/`), never in the repo being browsed
- Nothing heavy on the request hot path; everything expensive is a subprocess with a timeout
- px0 never writes file content itself; Salesforce mutations (deploy, retrieve, anonymous Apex) are **explicit user actions** routed through `sf` CLI subprocesses — same model as the agent harnesses

## Feature map

### P1 — Org-aware (makes it a Salesforce tool)

- [ ] **sf CLI bridge** — `internal/sfdx` wrapper + `/api/sf/orgs`, `/api/sf/default` (mirror how `agent.go` wraps harness CLIs); status bar chip with default org alias/instance
- [ ] **SOQL workbench** — query editor pane, `sf data query` subprocess, results in the existing table renderer (`internal/table`), CSV export
- [ ] **Org switcher** — pick default org per workspace, remembered in `~/.beast/settings.json`

### P2 — Intelligent

- [ ] **Apex language server** — register `apex-ls` (jorje jar) in `lspservers.go` with an install recipe; Apex go-to-def/hover
- [ ] **LWC preview** — render Lightning web components in-browser (static/mock-data pass first; live org data later)
- [ ] **Metadata explorer** — Tooling API object/field tree via `sf`, searchable, plugs into the sidebar
- [ ] **Apex logs** — tail `sf apex tail log`, stream over SSE like git status

### P3 — Extensible ("extensions and everything")

- [ ] **Plugin system v1** — plugins as JSON manifests in `~/.beast/plugins/*/plugin.json`: declare `/api/ext/<name>/<path>` endpoints backed by subprocess commands, plus `web/src`-style JS/CSS mounts under `/static/plugins/`. Core stays zero-dep; plugins are more subprocesses.
- [ ] **First-party plugins** — deploy-on-save (explicit), package.xml diff view, permission-set inspector
- [ ] **Extension manager panel** — list/enable/disable plugins from the UI

### P4 — Mined parity

- [ ] **Spec mining loop** — run rea against VS Code Salesforce extensions → `docs/specs/*.md` → file issues for anything worth having
- [ ] Coverage targets: deploy/retrieve flows, org browser UX, SOQL completion, Apex test runner UX, conflict detection

## Budgets (every feature pays)

- Cold start <1 ms, idle daemon ~30 MB RSS, bounded DOM in the browser tab
- New endpoint = inside `ServeHTTP` wrapper (gzip/CSP/metrics); new UI = ES module in `web/src/`; new external tool = subprocess behind a timeout
- `make check` green + `make test` green before every PR
