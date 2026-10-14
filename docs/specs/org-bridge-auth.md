# Spec: Org bridge — auth, org lifecycle, anonymous Apex & logs

Mined from `salesforcedx-vscode-org` + `salesforcedx-vscode-apex-log` (Salesforce Extension Pack, v68.x). Everything else in beast depends on an authenticated default org — this is the foundation the other specs assume.

## What it does

Three related surfaces:

1. **Org auth & lifecycle.** Web OAuth login for an org or Dev Hub, access-token login for CI, logout (default or all), create/delete scratch orgs, open an org in the browser, display org details, list all authenticated orgs, pick the default org per workspace, manage aliases. Declared commands in the reference: `sf.org.login.web`, `sf.org.login.web.dev.hub`, `sf.org.login.access.token`, `sf.org.logout.default/all`, `sf.org.create`, `sf.org.open`, `sf.org.display.default/username`, `sf.org.delete.default/username`, `sf.org.list.clean`, `sf.set.default.org`, `sf.alias.list`.
2. **Anonymous Apex.** Execute the open `.apex`/`.anon` document (or just the selection) against the default org; surface compile errors inline at line/column; on success show the result and offer the execution log.
3. **Debug logs & trace flags.** Set a trace flag on the current user (or another user) with a DebugLevel, show the active flag in the status bar, auto-collect new logs as they appear, list/get/open logs, scheduled cleanup of expired flags.

## Reference UX flow

- **Login:** command palette → "Authorize an Org" → pick environment (Production / Sandbox / Custom URL) → browser opens to OAuth → VS Code catches the `localhost:1717/OauthRedirect` callback → org appears in the org list, set as default if none.
- **Default org:** status-bar chip shows the current default org's alias/username; clicking opens an org picker quick-pick listing every authenticated org (alias, username, type marker for dev hub vs scratch/sandbox). The pack also gates every org-mutating command behind "is a default org set?" and prompts to authorize when not.
- **Anonymous Apex:** open a `.apex` file → CodeLens "Execute Anonymous Apex" (document or selection) → progress → compile errors land as editor diagnostics at the reported line/column; success → toast with "Open Log" action; log file is saved under the extension's log dir and opened.
- **Logs:** `sf.apex.traceFlags.createForCurrentUser` → status bar shows an animated "recording" chip while the flag is live → a poller lists `ApexLog` rows newer than the flag's start for each traced user → new logs auto-download → `sf.apex.log.get` retrieves a specific log on demand.

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Org list | `GET /api/sf/orgs` | `sf org list --json` |
| Default org for workspace | `GET /api/sf/default`, `POST /api/sf/default` | `sf config get target-org --json` / `sf config set target-org <alias>` (workspace-root cwd) |
| Org chip | status bar chip in `web/src/status.js`; click = refresh, menu = org picker | same |
| Authorize | `POST /api/sf/org/login` (returns the OAuth URL; user completes in their own browser) | `sf org login web --json` |
| Open org | `POST /api/sf/org/open` (return `instanceUrl`+sid frontdoor URL, open in new tab) | `sf org open --json --url-only` |
| Anonymous Apex | `POST /api/sf/apex/execute` `{path|text}` → `{compiled, line, column, compileProblem, success, exceptionMessage, log}` | `sf apex run --json --file/-` |
| Trace flags | `POST /api/sf/traceflags` create/delete; `GET /api/sf/traceflags` list | `sf data create record`/`sobject delete` on `TraceFlag`+`DebugLevel`, or `sf apex` helpers |
| Log list/get | `GET /api/sf/logs`, `GET /api/sf/logs/{id}` | `sf apex list log --json`, `sf apex get log --json` |
| Log tail | extend `/api/stream` SSE channel (same pattern as `/api/git/stream`) | `sf apex tail log` (long-running subprocess, bounded, cancellable) |

Subprocess rules per AGENT.md: `--json` only, `SF_DISABLE_TELEMETRY=true`, `SF_COLORIZE_LOGS=false`, bounded timeouts, `{installed:false}` when `sf` is absent — never a hard error. Workspace default-org selection persists in `~/.beast/settings.json` (never inside the workspace).

Issue #2 already covers `orgs`/`default` endpoints + the status chip; this spec extends it into the full auth/lifecycle + apex/log surface.

## Effort

- Org bridge core (orgs/default/chip) — **M** (already scoped in issue #2)
- Login + open + picker actions — **S** once the bridge exists
- Anonymous Apex (diagnostics + open-log action) — **M**
- Trace flags + status chip + log list/get — **M**; SSE tail streaming — **S** on top
