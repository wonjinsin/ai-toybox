# ADR 0001: Run the Solar System Benchmark with Automatic Minimal Recording

Date: 2026-09-30

Status: Accepted

## Context

The user wants one invocation to implement the fixed Solar System task and
finish recording automatically. The retained information should be limited to
the prompt, HTML result, model/effort, input/output tokens and elapsed time.
Analysis is out of scope for the current project. Both Codex and Claude Code
must use the same workflow without provider-specific collectors.

## Decision

- Keep one fixed [task prompt](../../skills/benchmark-record/assets/solar-system-prompt.md)
  and one canonical `benchmark-record` skill. Each host has an ordinary discovery
  `SKILL.md` that directs the agent to the shared instructions.
- A run invocation performs the task, verifies what it can, saves the final
  record, and replies. There is no operation selector or separate collection
  request. Analysis and comparison are not implemented.
- Measure implementation and its verification. Capture the start immediately
  before that work and the end immediately afterward. Record processing and the
  final user-facing reply are excluded. This allows automatic finalization in
  the same invocation without post-response hooks.
- Store only `prompt.md`, `output/index.html` with local resources, and a minimal
  `run.json`. Format `2.0` holds model, effort, duration and input/output totals,
  plus version, execution status and concise notes. Omit response copies,
  detailed usage breakdowns, session identifiers, evidence archives and backups.
- Group attempts under `runs/<agent>-<model>-<effort>-<run-id>/` by default.
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
- Preserve existing archives. Their older measurement boundaries are different;
  do not silently rewrite them as format `2.0`.

## Consequences

The user invokes only execution. There is no maintained collector script,
provider adapter, model launcher or background service. Hosts may expose no
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
