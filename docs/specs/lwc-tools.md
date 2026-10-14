# Spec: LWC tools — scaffolding, language server, Jest, Live Preview

Mined from `salesforcedx-vscode-lwc` (source pass) and `salesforcedx-vscode-ui-preview` (packaged-only — reconstructed with `rea analyze-javascript-application` on the `.vsix`; evidence: 102 modules, 5 renderer webviews, preview-server routes `/process/approvals`, `/process/rules`, config keys `autoRefresh`, `showQRPanel`, `debugOrgUsername`, `oauthLocalPort`, embedded `@salesforce/mnr-uem-preview-runtime` LWR preview runtime).

## What it does

Four surfaces around Lightning Web Components:

- **Scaffolding** — create an LWC from built-in templates (default JS, TypeScript, analytics dashboard) or workspace custom templates; rename a component (renames folder + files + markup references). Commands: `sf.metadata.lightning.generate.lwc`, `sf.lightning.lwc.rename`.
- **Language tooling** — `@salesforce/lwc-language-server` gives completion/hover/diagnostics for component markup (`<c-foo>` tags, `@api` props) and JS, plus component metadata from the workspace registry; status-bar LSP chip.
- **Jest unit tests** — `sfdx-lwc-jest` runner: test explorer, CodeLens run/debug per file or test case, `--watch` mode, results JSON watcher feeding the explorer.
- **Live Preview** — preview a component in a webview (or a browser/device): a local preview server renders the component; mobile pairing via QR code; auto-refresh on save; code-mode/preview-mode toggle per component.

## Reference UX flow

- **Create:** command → template quick-pick (built-ins pinned, custom templates listed) → name prompt → `<bundle>/<name>/` scaffolded (`.html`, `.js`, `.js-meta.xml`) → opened.
- **Jest:** `__tests__/*.test.js` under a bundle → CodeLens Run/Debug per file/case → `npx lwc-jest -- --runTestsByPath <file> --testNamePattern <n> --json --outputFile <tmp>` → results watcher parses the JSON → explorer updates; `--watch` toggles per file.
- **Live Preview (packaged):** `sf.lwc.local.preview.component.webview` / `sf.ui.preview` → embedded LWR server (WebPreview runtime) serves the component at a localhost origin → webview panel shows it; `autoRefresh` reloads on save; `showQRPanel` renders a QR for the Salesforce mobile app pairing flow; `debugOrgUsername` targets which org supplies live data.

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Create/rename LWC | plugin-style explicit actions: `POST /api/sf/lwc/create`, `POST /api/sf/lwc/rename` (or route through the plugin manifest per P3) | `sf lightning generate component --type lwc --name <n> --json` |
| LSP | `lwcRegistry` entry in `lspservers.go`: `lwc-language-server --stdio` for `lwc/**` (`.html`, `.js` under `lwc/`); install recipe `npm i -g @salesforce/lwc-language-server` | LSP subprocess |
| Jest | `POST /api/sf/lwc/test` `{file?, testName?, watch}` → job; results JSON → tests panel (shared with [apex-testing.md](apex-testing.md)) | `npx lwc-jest -- --runTestsByPath … --json --outputFile` subprocess |
| Preview (static pass) | `GET /api/sf/lwc/preview?component=<n>` — renders the component's HTML template + stub `@api` props in a sandboxed iframe under `/static/`; mock data first, per MISSION.md | local compile only (no org) |
| Preview (live pass) | plugin endpoint `/api/ext/lwc-preview/…` spawning an LWR server subprocess; iframe it; `~/.beast/plugins/lwc-preview/` owns the lifecycle | `sf lightning dev` / npx LWR server |
| Brand/themes (defer) | rea found the theme-preview webviews (`theme-manager-panel.js`, `themePreviewWebview.js`) — park until preview v1 lands | — |

## Effort

- Create/rename + language server registration — **S–M** (server is an npm install recipe like the existing registry entries)
- Jest run + results panel — **M** (reuses the apex-testing job/panel shape)
- Static preview (mock data iframe) — **M**; live-org preview — **L** (LWR subprocess lifecycle + plugin plumbing; split into a follow-up issue)
