# Live Benchmark Gallery Plan

Source: docs/harness-flow/specs/2026-10-17-live-benchmark-gallery.md

Goal: Show current local benchmark records and artifacts in an already-open
explorer without rebuilding HTML or restarting the service.

Constraints: "Do not change benchmark records, reviews, or artifacts."
"Browser checks use only connected Aside MCP direct REPL; any active access
denial is reported without an alternate-tool or protocol workaround."
"No remote/public deployment, authentication UI, editing benchmark data, scoring,
notifications, database, migration of historical artifacts, or benchmark task
execution." Every task inherits these constraints.

The user selected the local collection and explicitly requested continuation
through implementation and review. This plan does not require repeated approval
at each phase; materially different requirements still require a decision.

## Task 1: Open the explorer against current files

Delivers: Starting one local command opens the existing explorer with freshly
read local runs, including incomplete entries, instead of a generated snapshot.

Touches: scripts/serve-gallery.mjs, gallery/data.mjs, index.html,
gallery/app.js, tests/gallery.test.mjs, scripts/build-gallery.mjs,
gallery/template.html.

Blocked by: none.

- [x] Extract and adapt collection reading from the existing builder, preserving
  model/effort/score semantics and per-run warnings.
- [x] Expose GET /api/runs with revision, checkedAt, and the run list; absent
  collection is empty and collection read errors are non-success responses.
- [x] Serve a data-free application shell and its existing styles/script from a
  loopback-only HTTP service. Default port is 4173; support --port and --root.
- [x] The page fetches data on first load and reports loading/failure accurately.
- [x] An API request after file additions, edits, or removals returns current
  values with no rebuild or restart.
- [x] Delete the old builder and template; the root HTML is source, not generated
  data. Replace snapshot-specific tests rather than retaining obsolete assertions.

## Task 2: Reflect live changes without losing the user's place

Delivers: The open explorer updates automatically while preserving navigation,
filters, focus, and the current detail section.

Touches: gallery/app.js, index.html, gallery/styles.css, tests/gallery.test.mjs.

Blocked by: Task 1.

- [x] Fetch immediately, then schedule a single next poll two seconds after each
  completion. Abort a stalled request after five seconds.
- [x] Show connection state and last successful check. Keep last good data visibly
  stale on failures; recover without a user reload.
- [x] Apply changed revisions without resetting model/run selection, query, sort,
  score filter, detail section, keyboard focus, or scroll.
- [x] Skip DOM replacement during healthy polling when data is unchanged. A missing selected model keeps
  an identifiable zero-count selection.
- [x] Selected-run deletion shows a missing state; restoration recovers the same
  selected run and detail section.
- [x] Malformed records recover after repair, and zero remains scored.

## Task 3: Update only the artifact that actually changed

Delivers: A selected result keeps running when its review changes but reloads
when its HTML or supporting files change.

Touches: gallery/data.mjs, scripts/serve-gallery.mjs, gallery/app.js,
tests/gallery.test.mjs.

Blocked by: Task 2.

- [x] Compute artifactRevision from the sorted regular output-file tree, including
  relative names, sizes and high-resolution modification/change timestamps.
- [x] Root-relative revisioned artifact links preserve the original asset tree.
- [x] Keep the same iframe DOM node during metadata/review updates when the
  artifact revision is unchanged. Reload only the selected changed output.
- [x] Changes to supporting assets, entrypoint creation/removal, and switching
  runs have the expected preview behavior.
- [x] Treat artifact documents as untrusted in embedded and new-tab contexts;
  supporting output assets remain accessible to the sandbox.

## Task 4: Bound the local server and validate real changes

Delivers: The live service reads only its intended data and remains truthful
through concurrent writes, failure, and recovery.

Touches: scripts/serve-gallery.mjs, gallery/data.mjs, tests/gallery.test.mjs,
tests/gallery-path-race.test.mjs.

Blocked by: Task 3.

- [x] Allow GET/HEAD only; unknown routes and non-output run files are not exposed.
- [x] Reject traversal, symlink escapes, and non-loopback Host values; avoid
  granting cross-origin access to the collection API.
- [x] Send no-store headers for data and served content, and sandbox output
  documents. Do not mutate any source records.
- [x] HTTP integration tests use a temporary root and an ephemeral port, exercise
  an actual running server, and leave no fixture files or listeners behind.
- [x] Smoke the real CLI and actual repository collection. Use a temporary
  collection for add/edit/delete/repair and output-change scenarios.
- [x] Use Aside direct REPL for actual HTTP UI interaction and visual checks when
  permitted; report any access boundary precisely rather than substituting tools.

## Task 5: Document, review, and deliver the single live workflow

Delivers: The documented startup behavior matches the implementation, and the
final change has independent standards and requirement reviews.

Touches: README.md, affected source/tests from review findings.

Blocked by: Task 4.

- [x] Replace snapshot/rebuild instructions with the live startup command, URL,
  data-root/port overrides, polling/failure behavior, and direct-file limitation.
- [x] Directly opening the application file shows server-start guidance rather
  than old data; standalone benchmark result delivery is unchanged.
- [x] Run the complete relevant test suite and JavaScript syntax checks after
  integration. Confirm the already-open browser observes actual file mutations.
- [x] Review the final diff in separate Standards and Spec axes. Reproduce and
  fix justified findings, then rerun the affected behavior.
- [x] Commit only this feature's source, docs, and tests. Preserve user run changes.

## Coverage

- R1: Tasks 1 and 4.
- R2: Task 2.
- R3: Task 2 and existing navigation checks in Task 4.
- R4: Task 3.
- R5: Tasks 1, 2, and 4.
- R6: Tasks 1, 2, 3, and final browser review in Task 5.
- R7: Tasks 1, 3, and 4.
- R8: Tasks 1 and 5.
- R9: Tasks 4 and 5.

## Implementation ownership and shared interface

The server/data slice owns scripts/serve-gallery.mjs and gallery/data.mjs. The UI
slice owns gallery/app.js, gallery/styles.css, and index.html, and removes the
obsolete template. Integration owns tests, documentation, obsolete builder
removal, runtime validation, and final review. Slices can be implemented in
parallel once the response contract below and failing scenarios are recorded;
no sibling edits another slice's files.

The server module exports createGalleryServer({ root }) returning an unstarted
Node HTTP server. Tests listen on 127.0.0.1 with port zero. The collector exports
collectRuns(root), returning the existing normalized run fields plus
artifactRevision. The API wraps those rows in { revision, checkedAt, runs }.
Artifact URL shape is /runs/v1/<encoded-run-id>/output/index.html?v=<revision>.
The data root affects run reads only; application assets come from this checkout.
No external dependencies are required.
