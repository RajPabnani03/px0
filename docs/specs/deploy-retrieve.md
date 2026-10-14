# Spec: Deploy & retrieve flows

Mined from `salesforcedx-vscode-metadata` (Extension Pack, v68.x): `commands/projectDeployStart.ts`, `commands/retrieveStart/projectRetrieveStart.ts`, `shared/deploy/deployComponentSet.ts`, `shared/retrieve/retrieveComponentSet.ts`, `services/deployOnSaveService.ts`, `commands/deployManifest.ts`/`retrieveManifest.ts`/`generateManifest.ts`, `commands/sourceDiff.ts`, `statusBar/sourceTrackingStatusBar.ts`.

## What it does

Push local metadata source to the default org (deploy) and pull org metadata down into the project (retrieve). Surfaces in the reference:

- **Project deploy / retrieve** — one command deploys every locally-changed component ("push" semantics via source tracking when the org tracks source), another retrieves everything tracked.
- **Targeted deploy/retrieve** — from a file in the editor, a folder, or a tree selection ("deploy/retrieve this source").
- **Manifest-driven ops** — deploy or retrieve exactly what a `package.xml` describes; generate a manifest from selections.
- **Deploy on save** — optional setting; saving a metadata file deploys it (with a configurable debounce/queue).
- **Delete source** — deploy a destructive change or delete a component (local + remote), with confirmation.
- **Source diff** — diff a single file or component against its org counterpart.
- **Source-tracking status bar** — tracking orgs show a chip with pending local/remote change counts; click opens the changes view.

## Reference UX flow

`projectDeployStart` is the archetype:

1. Build a `ComponentSet` for the workspace's changed components (source-tracking diff when `tracksSource`, else the whole default source dir).
2. `ensureNonEmptyComponentSet` — nothing to do → quiet "no local changes" notice.
3. Preparation progress → **conflict detection gate** (see [conflict-detection.md](conflict-detection.md)) → on conflicts a modal offers **Override** (reruns with `--ignore-conflicts`) or **View conflicts** (opens the conflicts view; cancels the op).
4. `MetadataDeployService.deploy` runs with progress; output goes to a channel formatted as a deploy result table (component, type, status, problem).
5. Failures merge `ComponentFailures` per file → editor **diagnostics** on the failing file/line → channel auto-shows → "deploy completed with errors".
6. The deploy result is stored (`resultStorage`) — it is the baseline the next conflict check and the status bar counts compare against.

Retrieve mirrors this with the overwrite-confirm step (components already present locally get a "overwrite N components?" confirm) and a post-retrieve icon/refresh.

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Deploy workspace changes | `POST /api/sf/deploy` `{scope: project}` | `sf project deploy start --json` (add `--dry-run` for preview) |
| Deploy this file/tree | `POST /api/sf/deploy` `{paths:[…]}` | `sf project deploy start --source-dir <path> --json` |
| Deploy manifest | `POST /api/sf/deploy` `{manifest}` | `sf project deploy start --manifest <package.xml> --json` |
| Retrieve | `POST /api/sf/retrieve` `{scope|paths|manifest|metadata}` | `sf project retrieve start --json` (+ `--source-dir`, `--manifest`, `--metadata Type:Name`) |
| Preview before acting | same endpoints with `dryRun:true` | `sf project deploy/retrieve preview --json` |
| Progress + result | reuse the agent job model — `POST /api/sf/deploy` returns a job id; `/api/sf/job/{id}` polls; output appended to a log pane (like `handleAgentJob`) | `--json` polled via `--job-id` + `sf project deploy report --json` |
| Deploy on save | opt-in setting in `~/.beast/settings.json`; file watcher debounce → same deploy endpoint | same |
| Failure diagnostics | map `componentFailures[].lineNumber/columnNumber/problem` onto the existing problem/diagnostic surface used by LSP diagnostics | parsed from `--json` |
| Source diff | reuse `/api/diff` + diff view; materialize the org copy to a temp file under `~/.beast/` (never in the workspace) | `sf project retrieve start --source-dir <f> --output-dir <tmp>` then diff |
| Tracking counts | status-bar chip next to org chip | `sf project deploy preview --json` + `sf project retrieve preview --json` counts |

All mutations are explicit user actions (deploy button/menu/palette entry) — per AGENT.md px0 never writes content itself; deploy-on-save stays behind an explicit opt-in setting exactly like the reference.

## Effort

- Deploy/retrieve endpoints + job + output pane + diagnostics — **L**
- Manifest ops + source diff — **S** once the base exists (diff view already exists)
- Deploy on save — **S** (watcher + debounce + setting)
- Status-bar tracking counts — **S** (depends on preview parsing)
