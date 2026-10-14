# Spec: Apex test runner UX & code coverage

Mined from `salesforcedx-vscode-apex-testing` (Extension Pack, v68.x): `testDiscovery/testDiscovery.ts`, `commands/apexTestRun.ts`, `commands/apexTestSuite.ts`, `views/apexTestTreeService.ts` + `testController.ts`, `codecoverage/`, `utils/testResultProcessor.ts` + `testReportGenerator.ts`, `watchers/`.

## What it does

- **Test discovery:** the Tooling API `/services/data/v{api}/tooling/tests` endpoint enumerates every Apex test class + method in the org (paged; `testLevel=RunAllTestsInOrg` on API ≥ 68, `showAllMethods` before). Local tests come from scanning `*Test.cls`/test-suites in the project.
- **Run UX:** a test explorer tree (suites → classes → methods) plus quick-pick run targets: all tests, all *local* tests, a suite, a class, a method. CodeLens on test classes/methods runs that exact scope. Runs are async with progress; results stream into the explorer (pass/fail per method, duration, stack, message) and a results file on disk.
- **Suites:** create/edit `.testSuite-meta.xml` membership.
- **Code coverage:** a status-bar toggle paints covered/uncovered line decorations on Apex files from the last coverage-enabled run; out-of-sync results are flagged, not silently wrong.

## Reference UX flow

`sf.apex.test.run`:

1. Quick-pick builds from suites (`TestService.retrieveAllSuites`) + discovered classes, filtered to classes with ≥1 test method; "All Local" and "All" entries first.
2. `buildTestPayload` → `testService.runTestAsynchronous` → poll the `ApexTestQueueItem` until done → fetch `ApexTestResult` rows → write `test-result-*.json` under `.sfdx/tools/testresults/`.
3. `testResultProcessor` maps results onto explorer items (pass/fail icons, durations, failure message + stack on click) and a channel summary.
4. With coverage on, `codecoverage` reads the result file → `CodeCoverageResult[]` per class → line-range decorations; the status-bar toggle enables/disables; stale results (no coverage keys, file moved) map to explicit errors.
5. Watchers: file changes re-discover tests for edited classes; new result files appearing on disk refresh the explorer (covers runs triggered outside the IDE, e.g. CI or `sf apex run test` in a terminal).

## Beast mapping

| Beast piece | Endpoint / module | sf CLI underneath |
| :--- | :--- | :--- |
| Discovery | `GET /api/sf/tests` | `sf api get "/services/data/v68.0/tooling/tests?testLevel=RunAllTestsInOrg" --json` (paged via `nextUrl`) + local `*Test.cls` scan |
| Run | `POST /api/sf/tests/run` `{tests|class|suite|level, codeCoverage}` → job id | `sf apex run test --json [--tests X|--class-names|--suite-names|--test-level] --code-coverage` |
| Result | `GET /api/sf/tests/job/{id}` → per-method outcome, message, stack, duration; store result JSON in `~/.beast/testresults/` (never workspace) | `sf apex get test --json` poll / `run test --wait` |
| Explorer UI | right-hand inspector pane (`web/src/inspector.js`) or a dedicated tests panel: suite→class→method tree, per-method ✓/✗, click → failure detail | — |
| CodeLens-ish run affordances | gutter/shortcut "run this class/method" using the same file→test mapping the outline view already computes | same run endpoint |
| Coverage decorations | gutter coloring (covered=green/uncovered=red) on `.cls`/`.trigger` from `coverage.coveredLines/uncoveredLines` in the result JSON; status-bar toggle chip | `--code-coverage` result `coverage` block |
| External-run watcher | poll `sf apex get test`? No — keep it simple: re-discovery on refresh only; file watcher is optional | — |

## Effort

- Discovery + quick-pick + run + results tree — **L** (job polling + result mapping is the bulk; `.sfdx` output parsing is well-defined)
- Suites create/edit — **S** (XML edit delegated to agent harness per tenet #1, or direct manifest write as an explicit user action)
- Coverage decorations + toggle — **M**
