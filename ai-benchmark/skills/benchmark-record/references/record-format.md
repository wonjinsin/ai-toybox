# Minimal Solar System Record

Use [run.json](../assets/run.json), format `2.1`. Each attempt contains exact bundled
`prompt.md`, `output/index.html` with local resources, and `run.json`.
Never migrate historical records. Update only the active attempt through the publication CLI.

## Task identity

The current [fixed prompt](../assets/solar-system-prompt.md) declares
`solar-system-observatory-2` and contains 15 requirements. Copy its complete text
unchanged into each new attempt's `prompt.md`. The original unversioned
11-requirement prompt is frozen in
[solar-system-prompt-v1.md](../assets/solar-system-prompt-v1.md) for historical
review routing. Resuming an existing attempt retains its saved prompt.

Task identity comes from the saved prompt, never from a directory name, model,
timestamp, or record format. `runs/v1/` is the storage root, not a task-version
declaration. Record format remains `2.1`, with no added task-version field.
Do not migrate saved attempts when the bundled prompt changes. Review versions
are separate from the record schema.

## Run directory

Under the project root's `runs/v1/` or requested output root, exclusively create
`<agent>-<model>-<effort>-<run-id>` for each invocation:

- `agent`: actual host (`codex`, `claude-code`, etc.), never inferred from model or
  discovery entrypoint; `unknown` if unavailable.
- `model`, `effort`: resolve separately per [settings resolution](settings-resolution.md).
  Missing identity, settings or usage does not block reservation or construction.
  Use `unknown` only in directory components, never in JSON metadata. Configuration-derived
  settings are allowed when runtime verification is unavailable; note source/limits.
- Normalize these three components: lowercase; replace runs outside `a-z`, `0-9`,
  `.` and `-` with `-`; collapse repeated hyphens; trim leading/trailing dots/hyphens;
  empty becomes `unknown`. Preserve original known model/effort strings in JSON.
- `run-id`: real invocation-start clock from setup, converted to `Asia/Seoul`
  (UTC+09:00), formatted `YYYYMMDD-HHmmss`. If no reliable clock is available,
  use `undated-<random-UUID>` for collision-safe reservation and explain that this
  is not a timestamp. Work measurement still starts at step 2; do not fabricate time.
- On collision, append `-02`, `-03`, etc. to the full name until exclusive creation
  succeeds. Keep the original timestamp; never reuse paths or rename old runs.

## Fields

| Field | Meaning |
| --- | --- |
| `format_version` | `2.1`, adds `session_id` to the `2.0` work-interval format. |
| `status` | `running`, `completed`, `failed`, `interrupted`. Only fully identified and measured finished attempts are `completed`. |
| `session_id` | Exact invoking session/thread ID as a concrete nonempty string. Null is allowed only for non-completed records, with a reason in notes. No UUID shape is required. |
| `model`, `effort` | Concrete nonempty strings from runtime, user declaration, then configuration, resolved separately with source/limits in notes. Null is allowed only for non-completed records. |
| `duration_ms` | Finite nonnegative implementation/check elapsed milliseconds, or null only for non-completed records. Fractional milliseconds are valid. |
| `input_tokens`, `output_tokens` | Nonnegative safe integer totals for the same interval, or null only for non-completed records. Booleans, numeric strings and fractional counts are invalid. |
| `notes` | Array of nonempty strings for skill context, sources, gaps and failures. Explain each null value; at least one note is required when any value is null. No nested audit data/log copies. |

The template is the complete schema. No extra fields, hashes, response snapshots,
cache/reasoning or per-child breakdowns, child/turn IDs, evidence archives or manifest
backups. Missing values stay null, never zero/estimates; no second request needed.
Zero is valid only as a real measured value. Reject metadata placeholders
`unknown`, `null`, `undefined`, `unavailable`, `n/a`, `default`, `auto`,
`unspecified`, `unset`, `not available`, `not applicable`, `tbd`, `todo`, and
`placeholder`, case-insensitively after trimming, and unexpanded `${...}`
expressions. `none` is invalid for session ID and model; a sourced effort of
`none` may describe an explicit no-reasoning selection, not missing information.
Known model aliases remain exactly as sourced. Notes cannot waive required values.
Populate the template's values and explanatory notes before saving a record.

## Status and resume

- A finished attempt is `completed` only when `session_id`, `model`, `effort`,
  `duration_ms`, `input_tokens`, and `output_tokens` are all known and valid,
  with proven measurement boundaries and full parent/delegated coverage.
  Completion describes the measured attempt, not app quality or exhaustive
  verification. A mobile zoom or viewport capability gap alone does not make it
  `failed`. Note unverified requirements and remaining functional defects.
- An observed failure that prevents finishing construction is `failed`. Apply the
  [scoped continuation rule](../SKILL.md#continue-safely-and-finalize), not a
  whole-attempt abort for a check-specific limitation or recoverable defect.
- If final metadata, counts, bounds or coverage cannot be proven, preserve HTML
  and finalize `failed` with explicit recording limitations and affected values
  null. Continue construction and independent checks before finalization.
  Do not turn estimates, partial totals or placeholders into facts.
- Use `interrupted` only for unfinished interrupted attempts. Resume the same attempt
  with original measurement bounds. Finished resumed work becomes `completed` only
  when every required value is proven; otherwise use `failed`.
- Abrupt termination may leave `running` if finalization cannot occur. Running,
  failed and interrupted records may retain honest nulls with explanatory notes.

## Session identity

During setup, use the actual current-host sources in
[usage resolution](usage-resolution.md). [OMP](omp.md) uses explicit current OMP
session metadata or a runtime-linked journal, never an inferred Codex thread ID.
For Codex, use `CODEX_THREAD_ID` or concrete current-turn metadata, never a generic
MCP `sessionId`. For Claude Code, use current `CLAUDE_CODE_SESSION_ID`, substituted
`${CLAUDE_SESSION_ID}` or explicit current-session metadata; reject stale inherited
values and unexpanded placeholders. Other hosts require an explicit current-session ID.
Never infer an ID from configuration, cwd, the run directory or the newest log.

Resolve identity even if counts or boundaries are unavailable. Note the source.
Missing, invalid or conflicting identity leaves `session_id` null with a reason;
reserve the early checkpoint and continue permitted work. The attempt cannot be
`completed` without proven identity. Record the invoking parent session, not children.
When resuming the same attempt, retain its original `session_id`; if the current
session changes, note the change without replacing the original identity.

## Skill context notes

Apply only skills explicitly selected in the current request. Reading and running this skill selects it; quotes/discussion do not.
Discovery, shared skill and bundled resources are one skill.

- Injection, prior activation, resume/compaction, routing, dependencies and follow-up workflows do not authorize extras.
  Only higher-priority host requirements override this; injection alone grants no higher priority. Pass this policy to delegates.
- Record only actually applied extras: user-selected or required by higher-priority host instructions. Name each skill/reason,
  distinguish mandatory injected instructions from newly required loads, and disclose unavoidable exceptions in the reply.
- Update notes when requirements change, including after resume/compaction; retain prior unselected-skill deviations even if later suppressed.
- Do not record/announce suppressed skills as active or add boilerplate notes. Use only observed/explicitly supplied context,
  never unrelated histories, hook logs or an invented skill list.
- Suppression/incomplete visibility does not prove isolation. Notes neither authorize extras nor change measurement bounds.

## Measurement boundary

Start immediately before implementation; stop after implementation and checks,
before record processing and the final user-facing reply. Save the completed
record before replying. This is benchmark work time, not full-conversation time
or API-only latency. Capture real clock readings with ordinary tools and compute
the difference; use a monotonic clock when available or consistent wall-clock
readings. Missing, reversed or unreliable bounds leave duration null.
An unexpectedly short interval is not evidence of an unreliable clock. Preserve
valid observed deltas; separate processes alone do not invalidate readings from
the same host's monotonic clock. Require concrete clock/boundary evidence to
discard a measurement, rather than an impression of how long the work should take.

For token sources, boundaries, normalization and provenance, follow [shared usage rules](usage-resolution.md) and the current-host document.
Only when delegating, also apply the shared and current-host sections of [delegated usage](delegated-usage.md).

## Validation and publication

The independently installable skill includes a dependency-free
[validate-run.mjs](../scripts/validate-run.mjs). Its original read-only mode remains:

```sh
node <skill-root>/scripts/validate-run.mjs <run.json>
```

It validates exactly the format `2.1` fields, types and bounds for every status,
rejects null required values for `completed`, and requires nonempty notes for
nulls. Operators must explain each null and establish provenance themselves.
Read-only success prints the input unchanged. Invalid input/I/O emits a diagnostic
to stderr, exits nonzero and emits no stdout. Never redirect stdout over the input.

Use explicit write modes rather than manually replacing the record:

```sh
node <skill-root>/scripts/validate-run.mjs --checkpoint <proposal.json> <run-directory>
node <skill-root>/scripts/validate-run.mjs --finalize <proposal.json> <run-directory> <expected-sha256>
```

Exclusively reserve an empty directory first. The checkpoint requires `running`
and publishes `run.json` before copying the fixed prompt or building HTML.
Keep the proposal outside the run directory and never use an artifact as a draft.
Every write validates the proposal before changing the final record, uses a
unique sibling temporary file and atomic rename, then validates exact read-back.
The CLI prints the saved JSON to stdout and its SHA-256 to stderr. Retain this hash
as the next write's `expected-sha256`; it covers the exact existing record bytes.
If recovering an existing attempt, calculate that hash from the inspected file,
not an unrelated or newest record. Do not refresh a stale hash just to force a write.

Finalization requires a terminal proposal and an existing `running` record with
the matching hash. A `completed` proposal additionally requires nonempty regular
`prompt.md` and `output/index.html` under a real output directory. The run directory,
proposal and record cannot be symlinks. Files read by publication cannot be
hard-linked or special files. Canonicalize parent directories so host aliases
such as macOS `/var` are allowed. Check proposal containment against the canonical
run directory. The run directory cannot be a filesystem root.
Failed or interrupted publication does not inspect the artifact tree. An unreadable
or unsafe artifact must not prevent recording the failure in writable JSON.
Invalid proposals, missing completed output and stale hashes preserve the old record.

For an explicitly resumed `interrupted` attempt, or an active `running` checkpoint
update, use `--checkpoint <proposal.json> <run-directory> <expected-sha256>`.
The proposal must be `running`. Preserve the original nonnull `session_id`; the
CLI rejects replacement or erasure. An originally null identity can be filled only
with evidence of the original invoking session, not the resuming session.
Keep the saved prompt, notes and original measurement bounds. Record changed
session context in notes. `completed` and `failed` records are immutable through
these modes. Resume support is not permission to modify saved historical runs
during maintenance.

Use one publisher per attempt. Parallel attempts must own separate directories.
The expected hash detects stale input but is not an atomic compare-and-swap across
concurrent publishers. Do not change target directories during publication.
The CLI is not a sandbox against hostile concurrent filesystem mutation.
Normal errors remove the CLI's temporary file. Abrupt process death may leave
that file; it is not a lock and does not prevent a resumed attempt from finalizing.
Never retry denied writes. If the final write/read-back is unavailable, report
the last observed state and preserve output.

Structure is not truth. The CLI cannot prove measurement bounds, usage coverage,
exact prompt bytes or functional correctness. Check these separately; no notes
waive a completed record's required identity and measurements. Historical records
remain untouched even when the current validator rejects them.

The implementation uses write modes in the existing validator rather than a
second finalizer. This keeps schema validation in one standalone-copyable file,
avoids duplicated rules/import dependencies, and preserves the read-only interface.
