# Feature Specs (mined)

Feature specs for beast, mined from existing Salesforce tooling per MISSION.md. Each spec covers: what the feature does, the reference UX flow, which beast endpoint/UI surface it maps to, and an effort estimate.

## Method

- **Primary source:** plain-source pass over [`forcedotcom/salesforcedx-vscode`](https://github.com/forcedotcom/salesforcedx-vscode) (the Salesforce Extension Pack monorepo — currently Effect-TS based, commands and views declared per-package in `package.json` `contributes`).
- **Packaged bits:** [`morluto/rea`](https://github.com/morluto/rea) (`npx rea-agents analyze-javascript-application`) run against the published `.vsix` bundles fetched from Open VSX. The published extensions are webpack-minified single-file bundles embedding `@salesforce/core`; the shipped artifact confirms the `package.json` `contributes` command surface is the ground truth for the packaged UX. `salesforcedx-vscode-ui-preview` exists only as a bundled artifact (no monorepo source) — its surface was reconstructed via rea + embedded literals.
- **Important delta:** the real extensions call Salesforce APIs in-process via `@salesforce/core` (`Connection`, SDR/`ComponentSet`, `OrgMetadataCatalog`). Beast routes every org mutation through `sf` CLI subprocesses per the AGENT.md invariants — each spec names the `sf` command that covers the same surface.

## Specs

| Spec | Reference surface | Priority |
| :--- | :--- | :--- |
| [org-bridge-auth.md](org-bridge-auth.md) | `salesforcedx-vscode-org` — login/logout/open/display/list + anonymous Apex, logs, trace flags (`salesforcedx-vscode-apex-log`) | P1 — blocks everything |
| [deploy-retrieve.md](deploy-retrieve.md) | `salesforcedx-vscode-metadata` — project deploy/retrieve, source-path deploy, manifest ops, deploy-on-save | P1 |
| [soql-workbench.md](soql-workbench.md) | `salesforcedx-vscode-soql` — `.soql` editor, query builder, LSP completion, results grid, EXPLAIN | P1 |
| [org-browser.md](org-browser.md) | `salesforcedx-vscode-org-browser` — metadata tree, filters, per-node retrieve | P2 |
| [apex-testing.md](apex-testing.md) | `salesforcedx-vscode-apex-testing` — test discovery, run UX, suites, code coverage | P2 |
| [conflict-detection.md](conflict-detection.md) | `salesforcedx-vscode-metadata` — tracking + timestamp conflict detection, conflicts view, diff | P2 |
| [lwc-tools.md](lwc-tools.md) | `salesforcedx-vscode-lwc` + `salesforcedx-vscode-ui-preview` — scaffolding, language server, Jest, Live Preview | P3 |

Effort scale for a builder session: **S** < half a session, **M** ≈ one session, **L** ≈ two sessions, **XL** = split into follow-up issues.
