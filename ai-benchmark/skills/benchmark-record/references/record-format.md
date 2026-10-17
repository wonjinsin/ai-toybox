# Minimal Solar System Record

Use [run.json](../assets/run.json), format `2.1`. Each attempt contains exact bundled
`prompt.md`, `output/index.html` with local resources, and `run.json`.
Never migrate or overwrite existing records.

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
- `model`, `effort`: resolve separately per [settings resolution](settings-resolution.md)
  before reservation. Missing identity, model, effort, clock, or a usable current-session
  usage source stops setup without a directory. Configuration-derived settings are
  allowed when runtime verification is unavailable; note source/limits, never guess.
- Normalize these three components: lowercase; replace runs outside `a-z`, `0-9`,
  `.` and `-` with `-`; collapse repeated hyphens; trim leading/trailing dots/hyphens;
  empty becomes `unknown`. Preserve original known model/effort strings in JSON.
- `run-id`: real invocation-start clock from setup, converted to `Asia/Seoul`
  (UTC+09:00), formatted `YYYYMMDD-HHmmss`. Work measurement still starts at step 2.
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

- Finished implementation/checks is `completed` only when `session_id`, `model`,
  `effort`, `duration_ms`, `input_tokens`, and `output_tokens` are all known and valid,
  with proven measurement boundaries and full parent/delegated coverage. Access or
  verification blockers never permit completion.
- An observed/recoverable failure that prevents finishing is `failed`; apply the [stop rule](../SKILL.md#stop-on-access-or-verification-blockers) for access/verification blockers.
- If an attempt exists and final metadata, counts, bounds, or coverage cannot be
  proven, preserve artifacts and finalize `failed` with affected values null and
  reasons in notes. Do not turn estimates, partial totals or placeholders into facts.
- Use `interrupted` only for unfinished interrupted attempts. Resume the same attempt
  with original measurement bounds. Finished resumed work becomes `completed` only
  when every required value is proven; otherwise use `failed`.
- Abrupt termination may leave `running` if finalization cannot occur. Running,
  failed and interrupted records may retain honest nulls with explanatory notes.

## Session identity

During setup, use the current-host identity sources and validation rules in
[usage resolution](usage-resolution.md). For Codex, record the current thread ID
from `CODEX_THREAD_ID` or concrete current-turn metadata, never a generic MCP
`sessionId`. For Claude Code, use the current `CLAUDE_CODE_SESSION_ID`, substituted
`${CLAUDE_SESSION_ID}` or explicit current-session metadata; reject stale inherited
values and unexpanded placeholders. Other hosts require an explicit current-session ID.
Never infer an ID from configuration, cwd, the run directory or the newest log.

Resolve identity even if the usage file, counts or boundaries are unavailable.
Add a short source note. Missing, invalid or conflicting identity stops fresh setup
without a directory. If an attempt already exists, retain honest nulls with reasons,
preserve artifacts and do not mark it `completed`. Record the invoking parent
session, not delegated child sessions.
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

The standalone skill includes [validate-run.mjs](../scripts/validate-run.mjs).
Run it with Node.js on the saved record before publishing and again at read-back:

```sh
node <skill-root>/scripts/validate-run.mjs <run.json>
```

It requires exactly the format `2.1` fields, validates types and bounds for every
status, and rejects null required values for `completed`. Null values in other
statuses require notes. Each note must be a nonempty string; operators must ensure
that the notes actually explain missing values and source limits.

On success, the CLI prints the validated input unchanged and exits zero. On invalid
input or an I/O error, it emits a diagnostic to stderr, exits nonzero, and emits no
stdout. It never writes files. Do not redirect stdout over the input file.
Validation checks structure, not measurement truth, provenance, or artifact quality.
Do not publish a record as successful unless both validation and source checks pass.

Historical records remain untouched. A historical `completed` record with missing
required values now fails validation. There is no legacy bypass, automatic metadata
fill, or migration.
