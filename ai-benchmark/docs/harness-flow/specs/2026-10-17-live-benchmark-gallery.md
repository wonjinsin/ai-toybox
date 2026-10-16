# Live Benchmark Gallery

## Problem and approved direction

The existing gallery freezes run records into a generated HTML snapshot. This
fails the user's requirement to see changes to the local benchmark collection as
they happen. The user selected the local filesystem, including uncommitted work,
as the source of truth and requested specification, planning, implementation, and
review in one workflow.

Use a local read-only HTTP service and automatic polling. A generated snapshot,
manual rebuild, GitHub-backed collection, database, and filesystem watcher are
not part of the solution. Polling is sufficient for this small collection and
avoids platform-dependent watcher behavior.

## Requirements

### R1. Live source of truth

Every collection request reads the current local run directories, metadata,
prompts, latest reviews, and artifact availability. No record values are embedded
in the application HTML or committed as a generated catalog. New, edited, and
deleted runs appear without rebuilding or restarting the service. An absent run
collection is an empty collection; a collection-level read failure is an error,
not a fabricated empty success.

### R2. Automatic updates

The open page loads immediately, then checks again two seconds after each
completed request. Requests do not overlap. With a responsive service and active
browser tab, a completed file save is reflected by the next poll response.
Browser throttling can delay inactive tabs. A request that has not completed
within five seconds is a connection failure; subsequent scheduled checks continue.
No manual page reload is required. Show connection state and last successful
check time. No persistent stale cache substitutes for a live read.

### R3. Preserve the user's place

Background refresh preserves the selected model, selected run, search text,
sort order, review filter, active detail section, keyboard focus, and scroll
position. An unchanged collection does not rebuild the visible DOM. A deleted
selected run shows an explicit missing-run state without silently changing the
URL; restoration makes that selection usable again. A selected model with no
remaining runs remains identifiable with an empty result list. User-initiated
navigation and the keyboard skip link retain the navigation fixes already made.

### R4. Artifact updates are independent

Metadata or review changes must not recreate an unchanged preview iframe or
restart its simulation. Changes to the selected artifact's entrypoint or local
supporting files reload its preview automatically and refresh its new-tab link.
New or deleted artifacts update availability. Changing to a different run loads
that run's result. Only the selected preview executes. The preview's scripts
cannot access the gallery's origin. Preserve original local resource paths.

### R5. Honest missing and partial data

Retain incomplete runs. Missing/malformed records and unreadable individual
files produce a visible warning for that run while other runs remain usable.
A later valid save recovers on the next successful refresh. Missing numeric
measurements remain unknown. Missing scores are unscored, while zero is a valid
score. Show review status independently of whether the total is numeric.
Collection/network failures preserve the last successful view with an explicit
stale/disconnected indication; first-load failures show a clear failure state.

### R6. Keep the explorer's existing behavior

Keep model navigation, model/effort/host/run-ID search, latest/score sorting,
score-state filtering, counts, bookmarkable run details, artifact preview and
new-tab links, category scores, review findings, original prompt, duration, token
counts, and recording notes. Keep safe text rendering, responsive layout, and
keyboard access. The gallery neither evaluates artifacts nor manufactures scores.

### R7. Local, bounded, read-only service

Use the available Node.js runtime with no database, build step, or runtime
package installation. Bind only to loopback. Serve only application assets, the
collection API, and files inside a run's output directory. Do not expose raw
repository/configuration files, directory listings, symlink escapes, or mutation
endpoints. API and artifact responses must not be served from stale browser
caches. Treat generated HTML as untrusted: sandbox both embedded and directly
served artifact documents without same-origin permission. Supporting output
assets remain loadable from those sandboxed documents. Do not change benchmark
records, reviews, or artifacts.

### R8. Clean replacement and usable startup

Replace the snapshot builder and its generated-data entrypoint rather than
keeping competing viewing modes or compatibility shims. Provide one documented
startup command, a loopback URL, an optional data-root override for isolated
checks, and an optional port override. Opening the application directly as a
file must explain that the live server is required, not show stale runs or an
endless loading indicator. Existing individual benchmark artifacts retain their
own independent delivery behavior.

### R9. Evidence

Verify an already-open gallery while creating, editing, invalidating, repairing,
and deleting records in a temporary collection. Verify numeric-zero/unscored
semantics, preview preservation for review-only edits, reload after artifact or
support-file edits, missing-run recovery, and connection failure/recovery. Use
HTTP integration checks for filesystem and serving boundaries. Browser checks
use only connected Aside MCP direct REPL; any active access denial is reported
without an alternate-tool or protocol workaround. Keep HTTP gallery checks
separate from direct-file compatibility of benchmark artifacts.

## Data contract

A successful collection response contains `revision`, `checkedAt`, and `runs`.
The revision changes when consumer-visible data or artifact revision changes,
not merely because another poll occurred. Each run retains the existing display
fields and adds `artifactRevision`, a revision covering the regular files under
its output directory. Artifact URLs are root-relative, encoded URLs ending in
`index.html`, with a revision query for cache busting. The result entrypoint and
support files are served from their original hierarchy.

## Scope boundaries

No remote/public deployment, authentication UI, editing benchmark data, scoring,
notifications, database, migration of historical artifacts, or benchmark task
execution. New dependencies or a front-end framework are unnecessary. Existing
user changes to the run collection must be preserved.
