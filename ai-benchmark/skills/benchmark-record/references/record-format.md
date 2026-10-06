# Minimal Solar System Record

Use [run.json](../assets/run.json), format `2.1`. Each attempt contains exact bundled
`prompt.md`, `output/index.html` with local resources, and `run.json`.
Never migrate or overwrite existing records.

## Run directory

Under the project root's `runs/` or requested output root, exclusively create
`<agent>-<model>-<effort>-<run-id>` for each invocation:

- `agent`: actual host (`codex`, `claude-code`, etc.), never inferred from model or
  discovery entrypoint; `unknown` if unavailable.
- `model`, `effort`: resolve separately per [settings resolution](settings-resolution.md)
  before using `unknown` in names/null in JSON. Configuration-derived values are
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
| `status` | `running`, `completed`, `failed`, `interrupted`. Completion means finished, not every requirement verified. |
| `session_id` | Exact current invoking session/thread ID as a nonempty string, or null if unverified/unavailable with a reason in notes. |
| `model`, `effort` | Runtime, user-declared, then configuration-derived values, resolved separately with source/limits in notes; unresolved fields null. |
| `duration_ms` | Nonnegative implementation/check elapsed milliseconds. |
| `input_tokens`, `output_tokens` | Nonnegative integer totals for the same interval, or null if unavailable. |
| `notes` | A few short skill-context/source/gap/failure notes, including explicitly partial parent interval counts when full totals are unknown. No nested audit data/log copies. |

The template is the complete schema. No extra fields, hashes, response snapshots,
cache/reasoning or per-child breakdowns, child/turn IDs, evidence archives or manifest
backups. Missing metrics stay null, never zero/estimates; no second request needed.

## Session identity

During setup, use the current-host identity sources and validation rules in
[usage resolution](usage-resolution.md). For Codex, record the current thread ID
from `CODEX_THREAD_ID` or concrete current-turn metadata, never a generic MCP
`sessionId`. For Claude Code, use the current `CLAUDE_CODE_SESSION_ID`, substituted
`${CLAUDE_SESSION_ID}` or explicit current-session metadata; reject stale inherited
values and unexpanded placeholders. Other hosts require an explicit current-session ID.
Never infer an ID from configuration, cwd, the run directory or the newest log.

Record the validated identity unchanged even if the usage file, token counts or
measurement boundaries are unavailable. Add a short source note. If identity is
missing, invalid or conflicting, leave `session_id` null with a short reason and
continue. Record the invoking parent session, not delegated child sessions.
When resuming the same attempt, retain its original `session_id`; if the current
session changes, note the change without replacing the original identity.

## Skill context notes

Record only actually applied extras: user-selected or required by higher-priority
host instructions. Name each skill/reason; distinguish mandatory injected skills
from newly required loads. Disclose unavoidable exceptions in the reply. Update
notes when requirements change, including after resume/compaction; retain prior unselected-skill
deviations even if later suppressed.

Do not record or announce suppressed skills as active or add boilerplate notes. Use only
observed/explicitly supplied context, never unrelated histories, hook logs or an
invented skill list. Suppression/incomplete visibility does not prove isolation.
Notes neither authorize extras nor change measurement bounds.

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

For token source selection, request boundaries, normalization, delegated coverage
and provenance, follow [usage resolution](usage-resolution.md). Its shared rules
and current host section apply; delegated sections apply only when delegating.
