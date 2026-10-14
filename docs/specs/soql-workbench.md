# Spec: SOQL workbench — editor, completion, query builder, results

Mined from `salesforcedx-vscode-soql` (Extension Pack, v68.x): `lspClient/` (completion middleware, validation), `editor/` (custom editor instance, query runner, data view), `soql-builder-ui/` (Lit-based visual builder), plus `salesforcedx-vscode-apex/src/embeddedSoql/` for SOQL completion inside Apex.

## What it does

A `.soql` workspace for the default org:

- **Text mode** — a `.soql` document with syntax highlighting, diagnostics, and completion driven by the `@salesforce/soql-language-server` LSP.
- **Builder mode** — a visual query builder (Lit webview): pick sObject → pick fields → WHERE clause editor → ORDER BY → LIMIT → live SOQL preview. Text and builder edit the same document; a toggle flips modes.
- **Completion smarts** — the LSP returns placeholder items (`__SOBJECTS_PLACEHOLDER`, `__SOBJECT_FIELDS_PLACEHOLDER`); client middleware expands them into real sObject names and context-filtered field suggestions using org describe metadata, so completions know types, relationships (`Parent.Field`), and WHERE-appropriate fields.
- **Run** — "Run Query" executes against the default org with a row cap; results open in a data grid; an "all rows" path fetches beyond the first page; CSV export exists via the grid.
- **EXPLAIN** — "Query Plan" returns the optimizer's plan (`explain` on the query API).
- **Embedded SOQL** — inside Apex `[ SELECT … ]` blocks, a virtual document scheme bridges to the SOQL server so completion works in Apex too.

## Reference UX flow

`sf.data.query.document` / `soql.open.new.text.editor`:

1. `.soql` file → custom editor with text area + toolbar (mode toggle, org indicator, Run, Query Plan).
2. Keystrokes → SOQL LSP completion → middleware resolves placeholders via `MetadataDescribeService.describeCustomObject` (cached describe per sObject, invalidated on org change / `set_default_org` event).
3. Run → `runQuery(text, {maxRows})` → `openQueryDataView` renders records as a grid (records, `totalSize`, done flag); "All Rows" iterates `queryMore`.
4. Query Plan → `EXPLAIN` variant → plan table (cardinality, cost, notes) in the same data view.
5. Builder events (`sobjects_request`, `sobject_metadata_request`, `run_query`, `get_query_plan`, `set_default_org`) flow over the webview bridge; document stays the single source of truth (`ui_soql_changed` → document edit, debounced echo suppression).

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Query editor | `.soql` opens the existing editor + a workbench toolbar strip (new `web/src/soql.js` module); Run/Explain buttons | — |
| Run query | `POST /api/sf/query` `{soql, maxRows?, explain?}` | `sf data query --query "…" --json` / `sf data query --query "EXPLAIN …"` |
| Results grid | reuse `internal/table` + `web/src/table.js` CSV renderer — SOQL JSON → CSV conversion server-side (flatten relationship fields to dotted paths); row cap honors `MaxRows` | `--result-format json` |
| CSV export | same endpoint with `format=csv` passthrough (`sf` already emits csv) | `--result-format csv` |
| All rows | `more=true` keeps a `queryLocator`/`nextRecordsUrl` cursor; `GET /api/sf/query/next` | `sf data query --query-more` or `sf api get <nextRecordsUrl>` |
| Completion | register `soql-language-server` in `lspRegistry` (`lspservers.go`) for `.soql`; npm install recipe `@salesforce/soql-language-server` | LSP subprocess, same model as pyright/ts-ls |
| sObject/field data for completion+builder | `GET /api/sf/sobjects`, `GET /api/sf/sobjects/{name}/describe` (cached under `~/.beast/`) | `sf sobject list --json`, `sf sobject describe -s <name> --json` |
| Builder (phase 2) | form-driven panel that composes SOQL text (object picker → field multiselect → where rows → order/limit) writing back into the `.soql` editor; no Lit — plain ES module like the rest of `web/src` | — |
| Embedded SOQL in Apex (phase 3) | Apex LSP + SOQL LSP coordination is real work; defer until both servers are registered | — |

## Effort

- Query endpoint + results grid + CSV + explain — **M** (table renderer already exists; biggest piece is JSON→table flattening of nested `records`/`attributes`)
- `.soql` language registration + `soql-language-server` in `lspRegistry` + describe-backed completion cache — **M**
- Visual builder panel — **L**
- Embedded SOQL-in-Apex completion — **XL** (defer; needs Apex LSP + the placeholder-middleware dance)
