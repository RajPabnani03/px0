# Spec: Conflict detection & diff

Mined from `salesforcedx-vscode-metadata` (Extension Pack, v68.x): `conflict/conflictFlow.ts`, `conflict/conflictDetection.ts`, `conflict/conflictDetectionTimestamp.ts`, `conflict/conflictView.ts` + `conflictTreeProvider.ts`/`conflictUi.ts`, `conflict/resultStorage.ts`, `shared/diff/`, `statusBar/sourceTrackingStatusBar.ts`.

## What it does

Before a deploy/retrieve/delete mutates anything, detect that local and remote copies diverged. Two mechanisms, chosen by org capability:

- **Tracking orgs** (scratch orgs, `tracksSource`): the org's source-tracking state reports which components changed on both sides — conflicts come from `SourceTrackingService.getStatus(local+remote)` where `conflict` is set; candidates are then materialized remotely and byte-compared to suppress false positives.
- **Non-tracking orgs** (prod/sandbox): a per-file timestamp heuristic — the extension stores deploy-result timestamps (`resultStorage`); a component whose org `LastModifiedDate` is newer than the stored timestamp is a *potential* conflict, again confirmed by materializing the remote copy and comparing bytes.

On conflicts: the operation aborts into a modal — **Override** (rerun with force) or **View conflicts** (open the conflicts view, cancel). The conflicts view is a tree (type → files) with per-file actions: open local, open remote, **diff**. Settings: enable/disable detection for non-tracking orgs, deploy-on-save can also ignore conflicts, source-tracking polling interval.

## Reference UX flow

`detectConflicts(componentSet, op)`:

1. Detection disabled → return.
2. `orgInfo.tracksSource` → tracking path: status stream filtered to `conflict` → URIs → local `ComponentSet` → `materializeRemoteComponents` → `filesAreNotIdentical` → `DiffFilePair[]`.
3. Else if the non-tracking setting is on → `OrgMetadataCatalog.getEntries` (refresh consistency) → `computePotentialConflictKeys` vs the stored timestamp index → materialize + byte-compare → pairs.
4. Pairs → `ConflictsDetectedError` → `ensureConflictView` + `handleConflictsModal`; `diff`/`open` commands materialize the remote copy as temp files and open the editor diff.

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Detection gate | server-side step inside `POST /api/sf/deploy` and `/api/sf/retrieve` before the real run: `{conflicts:[{type,fullName,reason}]}` response → frontend shows the modal | tracking: `sf project deploy preview --json` + `sf project retrieve preview --json` already report conflicts for tracking orgs |
| Non-tracking fallback | compare `LastModifiedDate` (Tooling API `MetadataComponent`/`listMetadata` via `sf api`/`sf org list metadata`) against deploy timestamps stored in `~/.beast/deploy-results.json`, then byte-compare by retrieving candidates to a temp dir | `sf org list metadata -m <Type> --json` + `sf project retrieve start --output-dir <tmp>` |
| Conflicts modal | small modal in `web/src`: Override → re-POST with `ignoreConflicts:true` (`--ignore-conflicts` flag); View → conflicts panel | `--ignore-conflicts` |
| Conflicts view | panel listing type→file pairs with Open/Diff actions | — |
| Diff | reuse `/api/diff` + diff view: materialize remote copy under `~/.beast/tmp/`, diff local vs remote | `sf project retrieve start --source-dir <f> --output-dir <tmp>` |
| Status chip | tracking change counts next to the org chip | `preview --json` counts |

Beast simplification worth taking: `sf project deploy/retrieve preview --json` does the tracking-side work natively (sf CLI owns source tracking); beast only needs the timestamp+materialize heuristic for non-tracking orgs.

## Effort

- Preview-based detection for tracking orgs + modal + conflicts panel — **M**
- Diff action (temp materialize + existing diff view) — **S**
- Non-tracking timestamp heuristic — **M** (own result storage + byte-compare; bounded by `~/.beast` hygiene)
