# Salesforce Org Awareness

px0 can see the Salesforce orgs the `sf` CLI knows about. When `sf` is installed, a status-bar chip shows the workspace's default org and its connected status, and a click re-checks it live.

---

## Overview & Core Purpose

Salesforce work is org-scoped: a deploy, a query, or a retrieve all mean "against this org". The chip keeps the current target visible while you review code, the way the LSP chip keeps language-server state visible — ambient status you never have to run a command to check.

The integration is read-only. Everything `sf` knows is fetched through non-interactive `--json` subprocesses behind a timeout with telemetry off (`SF_DISABLE_TELEMETRY=true`), the same model px0 uses for coding harnesses. Nothing is cached on disk and nothing is written into the workspace.

---

## Key Capabilities

- **Status-bar chip**: Shows the default org's alias (or username when it has no alias) with a dot for its connected status — green while `Connected`, dim otherwise. Hidden entirely when `sf` is not installed.
- **Click to refresh**: Every click re-runs the `sf` calls, so a re-auth or `sf config set target-org` shows up without reloading px0. The tooltip carries the username, instance URL, and status text.
- **Org list API**: `GET /api/sf/orgs` runs `sf org list --json` with the workspace root as cwd and flattens every bucket (non-scratch orgs, scratch orgs, dev hubs, sandboxes, other) into rows of `alias`, `username`, `instanceUrl`, `isDefaultUsername`, `connectedStatus`. Secrets the CLI prints (e.g. `accessToken`) never leave the subprocess.
- **Default org API**: `GET /api/sf/default` runs `sf config get target-org --json` in the workspace root and joins the configured value back into the org list, so the reply carries its `connectedStatus` too. A configured-but-unauthenticated org still reports its name with no org row.
- **Failure as state, not faults**: A missing `sf` binary answers `{installed:false}`; a failing `sf` answers `{installed:true, error}` — both HTTP 200, since the chip renders them as UI states.

---

## Controls

| Control | Context | Action |
| :--- | :--- | :--- |
| `#st-sf` chip | Status Bar | Click (or `Enter`/`Space` while focused) to re-query `sf` |
| `GET /api/sf/orgs` | REST | List authenticated orgs |
| `GET /api/sf/default` | REST | Workspace's `target-org` plus its org-list row |
