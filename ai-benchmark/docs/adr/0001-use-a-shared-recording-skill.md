# ADR 0001: Run the Solar System Benchmark with Automatic Minimal Recording

Date: 2026-09-30

Status: Accepted

## Context

The user wants one invocation to implement the fixed Solar System task and
finish recording automatically. The retained information should be limited to
the prompt, HTML result, session identity, model/effort, input/output tokens and
elapsed time.
Analysis was outside the original recording scope; the review extension below
adds a separately invoked assessment. Both Codex and Claude Code use the same
workflow without provider-specific collectors.

## Decision

- Keep one fixed [task prompt](../../skills/benchmark-record/assets/solar-system-prompt.md)
  and one canonical `benchmark-record` skill. Each host has an ordinary discovery
  `SKILL.md` that directs the agent to the shared instructions.
- A run invocation performs the task, verifies what it can, saves the final
  record, and replies. There is no operation selector or separate collection
  request. Analysis and comparison are not implemented by the recording skill.
- Measure implementation and its verification. Capture the start immediately
  before that work and the end immediately afterward. Record processing and the
  final user-facing reply are excluded. This allows automatic finalization in
  the same invocation without post-response hooks.
- Store only `prompt.md`, `output/index.html` with local resources, and a minimal
  `run.json`. Format `2.1` adds the invoking `session_id` to the `2.0` fields:
  model, effort, duration, input/output totals, version, execution status and
  concise notes. Omit response copies,
  detailed usage breakdowns, child/turn identifiers, evidence archives and backups.
  Resolve the exact current session/thread identity independently of usage
  availability; use null with a reason if unverified. Resume retains the original
  session ID, with a note if the current session changes.
- Group attempts under `runs/v1/<agent>-<model>-<effort>-<run-id>/` by default.
  Use normalized host/model/effort names and a Korea-time invocation timestamp
  (`YYYYMMDD-HHmmss`), with numbered suffixes for collisions. Unknown name
  components use `unknown`; JSON retains original known settings or null.
- Resolve model and effort automatically before naming a run. Prefer current-run
  observations, then explicit user declarations, then host configuration. Allow
  configuration-derived values when runtime metadata is unavailable, recording
  their source and unverified overrides in existing notes. Resolve each field
  independently; use unknown only after available sources have been checked.
- Resolve usage automatically with ordinary read-only host tools. Prefer scoped
  host/API usage, then the existing local file for the observed current session:
  Codex rollouts or Claude Code transcripts. Filename discovery must match an
  exact current identity; never select the newest or a project-neighbor session.
  The Claude discovery entrypoint passes native session context to the shared
  skill, with current tool-environment identity available as another source.
- Capture matching work boundaries, deduplicate final requests, and normalize
  cache/thinking accounting per host. Inline JSON inspection is allowed without
  maintaining a provider collector. A bounded read-back may recover delayed
  boundary usage; recording/final-reply usage remains excluded. Missing identity,
  unsupported accounting, or ambiguous boundaries leave affected metrics null
  with a short reason. Finalize anyway; no further user action is required.
- For delegated work, retain linked identities and boundaries in working memory
  at dispatch. Check existing linked child sources before declaring usage missing.
  Use a proven inclusive interval total once, deduplicate complete linked final
  requests, or combine proven disjoint counter deltas and final-request sums.
  Include resumed work and descendants without copied history or duplicate
  accounting. If full coverage is unknown,
  keep affected totals null and retain valid parent interval counts in a concise
  partial note; do not add schema fields or per-child breakdowns.
- Preserve existing archives. Their older measurement boundaries are different;
  do not silently rewrite them into the current format.

## Review extension (2026-10-16)

- Add a separate, explicitly invoked `benchmark-review` shared skill with Codex
  and Claude Code discovery entrypoints. It accepts one existing run directory
  and never invokes generation or repairs the saved artifact.
- Preserve the original prompt, output and `run.json`. The recording schema and
  measurement boundaries above remain unchanged.
- Save only `reviews/review.json` and `reviews/review.md` in that directory.
  Re-review replaces those two files; do not retain timestamp directories,
  screenshots, evidence archives or previous-review backups.
- Use a versioned six-category rubric and the saved prompt. The original review
  policy left incompatible tasks blocked and unscored. Missing evidence remained
  unverified and withheld incomplete category scores and totals. The numeric-score
  revision below replaces that rule only for the current task.
- Keep browser restrictions unchanged: Aside direct REPL only, direct-file and
  HTTP verification separate, no bypass of explicit access denials. Generation
  status and prior reviews are not fresh verification evidence.

## Numeric-score policy

Current-task reviews use [format 4](../../skills/benchmark-review/references/review-format.md).
Evidence completeness and numerical certainty are independent. All category and
total scores are numeric lower bounds with explicit upper endpoints. Unknown
check ratings remain null, and confirmed failures remain failures. A determined
capped total does not make a partial review complete.

An explicitly requested scoring migration may recalculate existing observations
without pretending to run a new review. Preserve reviewer metadata and timestamps,
all check evidence, and every original artifact. Frozen format 2 retains its
original nullable arithmetic. The gallery labels score uncertainty separately
from review status instead of substituting zero for missing review data.


## Consequences

Generation and review are invoked separately. There is no maintained collector
script, provider adapter, model launcher or background service. Hosts may expose no
readable current-session file or usable work-interval usage; automatic lookup
does not guarantee all metrics are available on every version. An abrupt process
termination can leave a running record.

A completed attempt is not proof of fully working browser behavior. The record
notes unavailable verification. Model, tools, context and dependency choices can
vary; the project does not enforce an isolated model-only benchmark.

## References

- [Usage](../../README.md)
- [Execution skill](../../skills/benchmark-record/SKILL.md)
- [Minimal record format](../../skills/benchmark-record/references/record-format.md)
- [Review skill](../../skills/benchmark-review/SKILL.md)
