# Solar System AI Benchmark

Run one fixed Solar System task in Codex or Claude Code. The skill builds the
HTML, automatically saves a small run record, and then replies with the result.
One invocation completes the workflow; there is no follow-up collection step.

The [fixed prompt](skills/benchmark-record/assets/solar-system-prompt.md) requires
one sun and exactly three planets in a Three.js scene, automatic orbits, distinct
planet sizes/colors/speeds, pause/resume, and a global speed control that preserves
relative orbital speeds. The saved `output/index.html` must also run when opened
directly in a browser without starting a server or running a build step to view it.

## Run

Open this checkout as a project and select the model and effort in your host.
These are chat messages, not terminal commands.

Codex:

```text
$benchmark-record Run the Solar System benchmark.
```

Claude Code:

```text
/benchmark-record Run the Solar System benchmark.
```

No additional task prompt is needed. The skill creates a new run directory,
copies the fixed prompt, implements the task, performs available checks, and
saves the record before answering. Use a fresh conversation for each attempt
when practical. You can name a different output root in the same request.

During benchmark execution and recording, select and apply only skills explicitly
invoked in the current request. `$benchmark-record` or `/benchmark-record` selects that
skill alone unless other skills are also selected. Quoted examples or discussion
do not select a skill. Reading the entrypoint, shared skill, and bundled resources
counts as using the same skill.

Unselected skills are not applied, even when their instructions were already
injected by the host or hooks. This restriction survives resume and compaction;
prior activation, automatic routing, dependencies, and follow-up workflows do
not authorize another skill. Delegated work follows the same restriction and
uses ordinary host tools.

Injection alone grants no higher priority. An unselected skill may still apply
when a higher-priority host instruction requires it; project rules cannot override
that instruction. The agent records actually applied extras and their reasons in
`run.json.notes`, distinguishing mandatory injected instructions from newly required
skill loads, and discloses unavoidable exceptions in the final reply. Suppressed
skills are not announced as active or given boilerplate notes. These restrictions
and notes do not establish full environment isolation.
The restriction applies to benchmark work; ordinary repository investigation and
maintenance keep normal host skill selection. `benchmark-record` itself is never
selected automatically.

If skill discovery is unavailable, ask the agent to read
`skills/benchmark-record/SKILL.md` and run the benchmark. Reopen the project or
start a new session if an existing session has not picked up the skill.

## Saved Files

```text
runs/<agent>-<model>-<effort>-<run-id>/
  prompt.md           # Exact fixed task
  output/index.html   # Open directly in a browser; local resources stay in output/
  run.json            # Minimal settings and measurements
```

The agent component identifies the host, such as `codex` or `claude-code`.
The run ID is the invocation-start time in Korea (`Asia/Seoul`), formatted as
`YYYYMMDD-HHmmss`. For example, an illustrative model named `model-a` could create
`runs/codex-model-a-high-20261002-153000/`. If that path exists, the next run uses
`-02`, then `-03`, without overwriting it.

Name components use lowercase; spaces and path separators become hyphens.
The skill automatically checks current-run metadata, explicit user selections,
and the current host's configuration before naming the run. It resolves model
and effort separately. When runtime verification is unavailable, it can use
configuration-derived values and labels their source and unverified overrides
in `notes`. Only unresolved fields appear as `unknown` in the name and null in
JSON. JSON preserves selected identifiers or aliases as originally reported;
aliases are not guessed into exact model versions. See
[automatic settings resolution](skills/benchmark-record/references/settings-resolution.md)
and the complete
[directory naming rules](skills/benchmark-record/references/record-format.md#run-directory).

Codex checks `model` and `model_reasoning_effort` in applicable configuration
when runtime metadata is missing. Claude Code also checks its active effort
environment and model/effort configuration, including user, project and local
settings. Neither path launches another model or requires a second user request.

The [record template](skills/benchmark-record/assets/run.json) keeps:

- Model and effort, when known.
- Input and output token totals, when available for the measured work.
- Elapsed implementation and verification time in milliseconds.
- A format version, execution status and a few short notes for skill context,
  sources, missing values, verification limits or failure.

There are no response copies, session/turn identifiers, hashes, transcript or
evidence archives, cache/reasoning breakdowns, or previous-record backups.
The prompt and HTML are the retained task and result. Runs and local experiments
are ignored by Git; existing archives remain unchanged. Earlier runs retain their
original prompts and may require a server, so compare them separately from runs
using the current prompt.

Browser setup checks ordinary shell Node Playwright with installed Chrome before
connected browser tools. If permitted, a temporary `playwright-core` runtime
outside `output/` supplies missing packages; no browser installation or security
changes are needed. Node REPL import failures do not establish shell unavailability.
Checks use a permitted method selected before navigation: direct
`file://` when supported, or a temporary loopback HTTP server serving only
`output/` when HTTP is permitted and direct-file navigation is unsupported.
The artifact must still run directly without a server. HTTP interaction checks
and static `file://` compatibility checks are recorded separately; they do not
prove direct-file execution. An explicit security denial stops that action,
without switching tools or protocols to bypass it. See
[browser verification](skills/benchmark-record/references/browser-verification.md).

## Measurement Boundary

Measurement starts immediately before implementation and ends after its checks.
Record processing and the final user-facing reply are outside that interval.
Time includes the model's implementation work, tools and verification, not only
API latency. The task does not pin a Three.js version; delivery must work via
`file://`.

The agent uses normal clock and host inspection tools. It automatically checks
current-run usage and existing current-session files, following
[usage resolution](skills/benchmark-record/references/usage-resolution.md).
Codex can use App Server usage or its current rollout; Claude Code can use its
current session transcript. Current runtime identity selects the file; the agent
never substitutes the newest session or another conversation from the project.
The Claude discovery entrypoint also passes its native session substitution to
the shared skill for hosts that do not expose a session ID in the tool environment.

It uses scoped final request usage or matching cumulative counter differences.
Input totals include cache categories and output totals include reasoning,
without double counting. Claude's repeated response rows are deduplicated;
status-line context-window values are not cumulative spend. A small read-only
inline JSON query is allowed; no maintained provider parser or new hook is needed.

For delegated work, the agent retains child identities at dispatch and checks
only usage sources linked to that work, including resumed work and descendants.
It uses a proven inclusive interval counter once, deduplicates complete linked
request usage, or combines proven disjoint counter deltas and final-request sums.
Missing counts in a delegation tool result alone do not end lookup. See
[delegated usage collection](skills/benchmark-record/references/usage-resolution.md#delegated-work).

If current-session identity, final usage, or exact interval boundaries cannot be
established after lookup, the affected values are `null` with a brief reason.
Existing log access does not guarantee exact accounting on every host version.
The skill still finishes the record and replies;
it does not ask you to collect measurements later or estimate missing numbers.
A finished implementation can have status `completed` with unknown token counts.
If full delegated totals remain unknown, valid parent interval counts are kept
in `notes`, explicitly labeled partial with the missing coverage; the token
fields continue to represent full work totals, never parent-only substitutes.
That status does not claim every browser requirement has been verified.

An observed failure is saved as `failed`. An interrupted, unfinished attempt is
`interrupted`; when the same attempt resumes and finishes, it is `completed`
with an interruption note and original measurement bounds. An abrupt process
termination can leave `running` because no background service finalizes it.

## Shared Skill Setup

| Host | Project entrypoint |
| --- | --- |
| Codex | `.agents/skills/benchmark-record/SKILL.md` |
| Claude Code | `.claude/skills/benchmark-record/SKILL.md` |

These are ordinary files containing discovery metadata and a relative link to
[the shared skill](skills/benchmark-record/SKILL.md). Resources resolve from the
shared skill directory. The Claude entrypoint includes current-session context;
workflow rules remain in the shared skill. There are no symlinks.

For another project or a personal installation, copy the complete
`skills/benchmark-record/` folder, including its assets and references, into the
host's skill directory. Do not copy only a discovery entrypoint. The official
[Codex guide](https://learn.chatgpt.com/docs/build-skills) and
[Claude Code guide](https://code.claude.com/docs/en/skills) describe discovery
locations and invocation.

This project currently provides only execution and automatic minimal recording.
Analysis and comparison are not included. It does not switch models, launch a
second model, install hooks, or run a separate benchmark CLI.

- [Current architecture](docs/adr/0001-use-a-shared-recording-skill.md)
- [Record fields and accounting](skills/benchmark-record/references/record-format.md)
