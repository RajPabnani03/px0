# Spec: Org browser / metadata explorer

Mined from `salesforcedx-vscode-org-browser` (Extension Pack, v68.x): `tree/metadataTypeTreeProvider.ts`, `tree/orgBrowserNode.ts`, `services/fullOrgDiscovery.ts`, `services/metadataTypePreload.ts`, `commands/retrieveMetadata.ts`, `commands/filterMetadata.ts`, `commands/componentPresence.ts`, `services/filterState.ts`.

## What it does

A sidebar tree of every metadata type in the default org:

- **Two levels:** metadata *types* (ApexClass, CustomObject, Report…) → *components* (each class, object, report folder, report). Folder-bearing types nest one more level.
- **Presence markers:** components that exist locally get a distinct icon; local-only and org-only filtering toggle independently (`showLocal`, `showOrg`).
- **Filtering:** a filter box matches component names (plain substring or regex) across the tree.
- **Full discovery:** the org's metadata catalog is fetched lazily per type on expand; an optional "discover everything" walk prefetches all types with progress (concurrency-bounded) so filtering and counts are instant afterward.
- **Retrieve:** right-click a component (or a whole type) → retrieve into the project, with an overwrite-confirm when N of the members already exist locally; retrieved components flip their icon.

## Reference UX flow

`metadata -> sfdxOrgBrowser` view:

1. View opens → root lists metadata types from `OrgMetadataCatalog` (backed by `describeMetadata` + cached `listMetadata` calls; cache persists across sessions).
2. Expand a type → `catalog.getChildren({type})` → components (lazy); folders expand further.
3. Slow uncached org calls show a status-bar request indicator after ~300 ms so cache hits never flicker.
4. Toolbar: refresh-type, collapse-all, show-local/show-org toggles, filter input.
5. Retrieve → `getRetrieveMembers(node)` (component → its `{type, fullName}`; type → all children) → `confirmOverwrite` counts members already in the project → `OrgBrowserRetrieveService.retrieve(members)` → icon flip + success toast.

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Type list | `GET /api/sf/orgbrowser/types` | `sf org list metadata-types --json` (or `sf api get /services/data/vXX.0/tooling` discovery) |
| Components per type | `GET /api/sf/orgbrowser/children?type=ApexClass[&folder=]` | `sf org list metadata -m ApexClass --json` |
| Presence check | local scan via existing index/ignore walker (`.xml` per component) — no API needed | — |
| Retrieve | reuse `POST /api/sf/retrieve` from [deploy-retrieve.md](deploy-retrieve.md) with `{metadata:[{type,fullName}]}` | `sf project retrieve start --metadata Type:Name --json` |
| Tree UI | sidebar tree — extend `web/src/tree.js`/`panels.js` with a lazily-loaded branch provider; filter input like the existing fuzzy box | — |
| Catalog cache | `~/.beast/orgbrowser-cache.json` keyed by org id (never in workspace) | — |
| Full discovery | background job endpoint; progress over `/api/stream` SSE | fan-out of `sf org list metadata` per type, concurrency ~5 |

## Effort

- Types + lazy children + presence markers + filter — **L** (the tree provider and its caching discipline is the bulk; every org call stays a bounded subprocess)
- Retrieve integration — **S** once deploy/retrieve lands
- Full-discovery prefetch + SSE progress — **S**
