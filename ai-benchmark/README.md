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

During benchmark execution and recording, select only skills explicitly invoked
in the current request. `$benchmark-record` or `/benchmark-record` selects that
skill alone unless other skills are also selected. Quoted examples or discussion
do not select a skill. Reading the entrypoint, shared skill, and bundled resources
counts as using the same skill.

Instructions already injected and active through the host or hooks continue to
apply, including after resume or compaction. They do not authorize automatic
routing to additional skills, dependencies, or follow-up workflows. Delegated
benchmark work follows the same restriction and uses ordinary host tools.

Hook injection alone does not give an instruction higher priority. Higher-priority
host requirements may still require additional skills. The agent records their
names and reasons in `run.json.notes`, separately from known active injected skill
instructions, and discloses required additions in the final reply. Check these
notes when comparing runs; they do not establish full environment isolation.
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
Unknown model/effort values appear as `unknown` in the name and null in JSON.
JSON preserves known settings as originally reported. See the complete
[directory naming rules](skills/benchmark-record/references/record-format.md#run-directory).

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

## Measurement Boundary

Measurement starts immediately before implementation and ends after its checks.
Record processing and the final user-facing reply are outside that interval.
Time includes the model's implementation work, tools and verification, not only
API latency. The task does not pin a Three.js version; delivery must work via
`file://`.

The agent uses normal clock and host inspection tools. It uses scoped final
usage or matching cumulative counter differences when available. Input totals
include cache categories and output totals include reasoning when the source
reports them, without double counting. No maintained provider parser is needed.

If the host exposes no exact usage for this interval, the affected values are
`null` with a brief reason. The skill still finishes the record and replies;
it does not ask you to collect measurements later or estimate missing numbers.
A finished implementation can have status `completed` with unknown token counts.
That status does not claim every browser requirement has been verified.

An observed failure is saved as `failed`; an observed interruption as
`interrupted`. An abrupt process termination can leave an initial `running`
record because there is no background service to finalize it.

## Shared Skill Setup

| Host | Project entrypoint |
| --- | --- |
| Codex | `.agents/skills/benchmark-record/SKILL.md` |
| Claude Code | `.claude/skills/benchmark-record/SKILL.md` |

These are ordinary files containing discovery metadata and a relative link to
[the shared skill](skills/benchmark-record/SKILL.md). Resources resolve from the
shared skill directory. There are no symlinks or duplicated workflow rules.

For another project or a personal installation, copy the complete
`skills/benchmark-record/` folder, including its assets and reference, into the
host's skill directory. Do not copy only a discovery entrypoint. The official
[Codex guide](https://learn.chatgpt.com/docs/build-skills) and
[Claude Code guide](https://code.claude.com/docs/en/skills) describe discovery
locations and invocation.

This project currently provides only execution and automatic minimal recording.
Analysis and comparison are not included. It does not switch models, launch a
second model, install hooks, or run a separate benchmark CLI.

- [Current architecture](docs/adr/0001-use-a-shared-recording-skill.md)
- [Record fields and accounting](skills/benchmark-record/references/record-format.md)
