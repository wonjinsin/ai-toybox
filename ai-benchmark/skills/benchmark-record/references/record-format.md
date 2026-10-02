# Minimal Solar System Record

Use [run.json](../assets/run.json), format `2.0`. One run directory contains the
exact bundled `prompt.md`, `output/index.html` with any local resources, and
`run.json`. The directory name identifies the attempt. Existing records are not
migrated or overwritten.

## Run directory

Group runs under the project root's `runs/`, or a user-specified output root.
Create one child named `<agent>-<model>-<effort>-<run-id>` for each invocation.

- `agent`: the actual host, such as `codex` or `claude-code`; do not infer it
  from the model or the discovery entrypoint. Use `unknown` if unavailable.
- `model`, `effort`: the known settings used for the record. Use `unknown` for
  missing values in the directory name and null in JSON; never guess settings.
- Normalize these three name components to lowercase. Replace runs of characters
  outside `a-z`, `0-9`, `.` and `-` with `-`, collapse repeated hyphens, and trim
  leading/trailing dots and hyphens. Use `unknown` if empty. Preserve original
  known model/effort strings in JSON; do not add fields to the record.
- `run-id`: read the real invocation-start clock during setup, convert it to
  `Asia/Seoul` (UTC+09:00), and format it as `YYYYMMDD-HHmmss`. This timestamp
  identifies the run; the measured work interval still starts at step 2.
- Reserve the directory with exclusive creation. If the name already exists,
  append `-02`, `-03`, and so on to the full name until creation succeeds. Keep
  the original timestamp, never reuse an existing path, and never rename old runs.

Example: `runs/codex-model-a-high-20261002-153000/`, followed by
`runs/codex-model-a-high-20261002-153000-02/` on a collision. `model-a` is an
illustrative model name.

## Fields

| Field | Meaning |
| --- | --- |
| `format_version` | `2.0`; distinguishes these work-interval metrics from older whole-turn records. |
| `status` | `running`, `completed`, `failed` or `interrupted`. Completion means the attempt finished, not that every requirement was verified. |
| `model`, `effort` | Known host settings or the user's explicit selection. Unknown values stay null; note a user-declared value when it was not observed. |
| `duration_ms` | Nonnegative elapsed milliseconds for implementation and its checks. |
| `input_tokens`, `output_tokens` | Nonnegative integer totals for that same work interval, or null when unavailable. |
| `notes` | A few short notes for measurement sources/gaps, verification limits or failure. Do not add nested audit data or copy logs. |

Do not add fields, hashes, response snapshots, cache/reasoning breakdowns, session
identifiers, evidence archives or manifest backups. The template is the complete
record. Missing measurements remain null, never zero or an estimate, and do not
require another user request.

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

For tokens, use final per-request usage wholly attributable to the work interval,
or subtract comparable cumulative counters captured at its boundaries. Do not
use a session total, context-window occupancy, partial streaming value or text
length. Deduplicate requests and exclude prior work, recording and the final
reply. Include retries and child usage only when attributable without double
counting. If a model request spans both work and recording and cannot be split,
or updates arrive too late, leave affected totals null instead of estimating.

Normalize input to include cache reads/writes and output to include reasoning
when the source reports those categories. Add disjoint categories once; do not
add subsets already included in a total. Unknown accounting means null. Consult
only a known current-execution source; do not search unrelated histories or build
a provider-specific parser. If the source is inaccessible, finalize with nulls
and a brief note. State the source/calculation briefly in notes for known tokens.
