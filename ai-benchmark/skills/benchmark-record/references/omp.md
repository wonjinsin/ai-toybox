# OMP Settings and Usage

Apply [shared settings rules](settings-resolution.md) and
[shared usage rules](usage-resolution.md). Select this reference when the actual
host is OMP, including an OMP session using an `openai-codex/...` model. Provider
names do not turn OMP sessions into Codex threads. Do not read Codex or Claude
private histories to fill OMP metadata gaps.

## Settings

Prefer concrete current-execution OMP metadata or an already-connected runtime
API. Resolve model and effort independently. An exposed model ID does not imply
an effort level, and a family label does not identify an exact runtime model.
Use explicit user selections or applicable OMP configuration only under the
shared precedence rules, labeling them user-declared/configuration-derived and
noting unverified runtime overrides.

Inspect only current-host settings sources explicitly exposed by runtime
metadata, the active host's documented tools or a supplied current configuration
path. Do not invent environment variable names, default configuration paths,
journal paths or source schemas. Do not dump environments, credentials, request
bodies or unrelated history. If the host exposes no usable source, retain nulls,
publish the early checkpoint and continue permitted implementation.

## Session and tokens

1. Resolve the invoking parent OMP session identity from explicit current-session
   metadata or an already-connected host API. A generic MCP/browser transport
   session ID, agent display name, cwd or model provider is not that identity.
   Reject placeholders and conflicts. Do not infer `CODEX_THREAD_ID` or select
   the newest session. Resolve identity even when no usage source is exposed.
2. Prefer an existing current-session final-usage API. Otherwise use an OMP
   runtime journal only when the host supplies its exact current-session path or
   an explicitly scoped locator. If a supplied locator requires discovery, bound
   it by the concrete session ID and current workspace before opening bodies.
   Confirm the source's session identity and workspace linkage. A moved workspace
   requires explicit runtime linkage. No filesystem-wide session search, guessed
   home-directory layout, neighboring journals or copied conversation history.
3. Inspect only the identity, runtime settings, request IDs, completion/finality,
   usage and parent/child linkage fields needed for accounting. Read the actual
   documented/exposed schema before selecting fields; OMP journal layouts are not
   assumed to match Codex rollout or Claude transcript schemas. A journal entry
   containing usage is not sufficient proof that the request is final.
4. Capture the real clock, source position and issuing request identity in
   separate start/stop calls. Count completed requests after the start request
   through the stop request, or subtract comparable cumulative endpoints with
   proven boundary-request inclusion. File timestamps, a latest counter snapshot
   and an arbitrary byte offset do not establish the cutoff. Freeze the original
   stop time/request ID before finalization work.
5. If final usage is published after a boundary call returns, allow one read-back
   of the selected source, scoped to those frozen IDs and linked work. Do not
   select the latest row or extend the interval to recording/read-back/reply
   requests. Ignore an incomplete trailing JSONL line until this read-back.
   Unknown finality, missing boundary requests or malformed measured usage leaves
   affected fields null. Do not keep waiting for unsupported telemetry.
6. Calculate totals in code from deduplicated final request records or comparable
   cumulative counters, not mental arithmetic. Deduplicate by stable provider
   request ID, namespacing source-local IDs. Conflicting final records need proven
   finality, not an arbitrary preferred copy. Count billed retries once each.
   Normalize cache and reasoning only according to the exposed provider/source
   accounting contract. Never add categories already included in totals or
   treat an absent category as zero without that contract. Unknown accounting
   leaves the affected total null. Preserve real measured zero.
7. If delegating, apply the [OMP delegated rules](delegated-usage.md#omp).
   Parent usage alone is not full coverage unless the source proves inclusion.
   Missing descendant records cannot be replaced with estimated or parent-only
   totals. On resume, preserve the original session identity and interval bounds;
   recover proven linked cross-session coverage or leave totals null.

A missing source, capability or usage field is a recording limitation, not a
construction blocker. An explicit denial stops that lookup operation without
retry or bypass; continue only independently permitted work under the
[scoped continuation rule](../SKILL.md#continue-safely-and-finalize). A finished
artifact with unprovable required identity/measurement is finalized `failed`,
with HTML preserved and explanatory notes. A fully identified/measured finished
attempt can be `completed` despite documented browser gaps or remaining defects.

Use [validated atomic publication](record-format.md#validation-and-publication)
for the early `running` checkpoint and terminal record in this invocation. Notes
name source kind, session scope, frozen-boundary calculation, finality and coverage
limits without copying private journal paths, request bodies or identifiers other
than the required `session_id`. If no journal was exposed/read, say so rather than
claiming journal-derived evidence. Structural validation does not prove usage.
