# Solar System AI Benchmark

Run one fixed Solar System task in Codex or Claude Code. The skill builds the
HTML, automatically saves a small run record, and then replies with the result.
One invocation completes the workflow; there is no follow-up collection step.

The [fixed prompt](skills/benchmark-record/assets/solar-system-prompt.md) requires
one sun and exactly three planets in a Three.js scene, automatic orbits, distinct
planet sizes/colors/speeds, pause/resume, and a global speed control that preserves
relative orbital speeds. The saved `output/index.html` must also run when opened
directly in a browser without starting a server or running a build step to view it.

## Browse the Live Results

Start the read-only gallery from the repository root with Node.js 22 or newer:

```sh
node scripts/serve-gallery.mjs
```

Open **http://127.0.0.1:4173/**. No package installation, build, database, or
generated catalog is required. Opening the application `index.html` directly
shows startup guidance; the live gallery requires this server.

The server reads the current `runs/v1` collection, including uncommitted files.
The page checks immediately and then two seconds after each completed request.
New runs, deletions, status changes, prompts, and review scores update without
rebuilding, restarting, or reloading the page. Inactive browser tabs may be
throttled. A request taking longer than five seconds is treated as a connection
failure; scheduled checks continue. The connection banner marks retained data
as stale, and the footer shows the last successful check.

The explorer keeps model navigation, search, latest/score sorting, score-state
filters, bookmarkable details, original prompts, review findings, duration,
tokens, and recording notes. Background updates preserve selection, filters,
the active detail section, keyboard focus, and scroll position. A deleted
selected run shows a missing-run message and recovers automatically if restored.
The keyboard skip link does not change the selected model or run.

Metadata or review updates do not restart an unchanged preview. Changes to the
selected output HTML or its supporting files reload that preview automatically.
After reconnection, a preview opened since the last confirmed connection can
reload once even if the collection is unchanged. An established preview keeps
running through the outage.
Incomplete and malformed records remain visible with warnings and recover after
repair. Missing scores are **unscored**, not zero; numeric zero remains a valid
score, and review status is always displayed. The gallery never evaluates or
modifies benchmark records, reviews, or artifacts.

The server binds only to loopback and exposes the application, collection API,
and regular files inside each run's `output/` directory. It does not serve
directory listings, raw repository/configuration files, or symlinked paths.
Output documents are sandboxed without gallery-origin access, including when
opened with **결과물 새 탭 열기**. Original relative output assets remain loadable;
an artifact's external CDN dependencies may still require internet access.

To use another port or an isolated data root (whose collection is `runs/v1`):

```sh
node scripts/serve-gallery.mjs --port 4174 --root /absolute/data-root
```

Application assets always come from this checkout. Stop the server with Ctrl+C.
Standalone benchmark artifacts retain their own direct-file delivery contract;
successful HTTP gallery previews do not establish their `file://` compatibility.
Actual browser verification uses only Aside's direct MCP REPL.

The application source is `index.html` and `gallery/`; there is no snapshot
builder. Run the live HTTP/filesystem regression suite with:

```sh
node --test tests/gallery*.test.mjs
```

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

## Review a Saved Run

Use the separate [review skill](skills/benchmark-review/SKILL.md) with one explicit
run directory. These are chat messages:

```text
$benchmark-review runs/v1/<agent>-<model>-<effort>-<run-id>
```

In Claude Code, use `/benchmark-review` with the same directory argument.
If discovery is unavailable, ask the agent to read
`skills/benchmark-review/SKILL.md` and review that directory.
The review does not invoke the generation skill, repair the artifact, or change
`prompt.md`, `run.json` or `output/`.

It saves only `reviews/review.json` and `reviews/review.md` inside the run.
Re-review replaces those two files with the latest assessment; there are no
timestamp subdirectories, retained screenshots, evidence archives or backups.
Unrelated existing files are preserved.

The [fixed rubric](skills/benchmark-review/references/rubric.md) uses six categories:
simulation/time correctness (25), selection/tracking consistency (20), controls
and interaction (15), visual quality (15), responsive usability (15), and delivery
stability (10). Each category is scored from 0 to 5, with a weighted total out of
100. Review the saved prompt; incompatible historical tasks receive a blocked,
unscored review rather than newer requirements applied retroactively.

Checks distinguish pass, fail and unverified. Any unverified check leaves its
category and the total unscored; verified failures are still reported separately.
HTTP execution cannot verify direct-file delivery. Aside's direct REPL is the only
permitted browser tool. The review selects its transport before navigation:
use `file://` only with confirmed local-file support; otherwise use documented,
host-permitted loopback HTTP serving only the unchanged `output/` directory.
If file navigation reports missing local-file capability, continue over HTTP
only when host file access and loopback serving are permitted. Do not probe
undocumented file access. See the
[browser preflight](skills/benchmark-review/references/verification.md)
for the server command, port selection, access-denial boundaries and cleanup.
HTTP reviews leave direct-file delivery unverified and therefore the total unscored.
Outside this scoped transport fallback, a required inaccessible file, an explicit
access/security denial, unavailable Aside or an unsupported required check stops
evaluation immediately. Do not continue static inspection or score the artifact.
Clean up review-created resources and
replace both review files with this attempt's `blocked` result: no scores, the
exact blocker and unperformed checks. A new attempt replaces the previous review,
including a previously successful one. If the target or output is unsafe, or a
write is denied, report the actual saved/unsaved state without bypassing it.
Past errors are historical context, not current observations. Generation time,
tokens and model identity do not affect quality scores.

## Saved Files

```text
runs/v1/<agent>-<model>-<effort>-<run-id>/
  prompt.md           # Exact fixed task
  output/index.html   # Open directly in a browser; local resources stay in output/
  run.json            # Minimal settings and measurements
  reviews/            # Added only by benchmark-review; latest review replaces these files
    review.json       # Structured scores, checks and verification limits
    review.md         # Human-readable findings and reproduction steps
```

Benchmark results are stored under [`runs/v1/`](runs/v1/) and tracked in Git.

The agent component identifies the host, such as `codex` or `claude-code`.
The run ID is the invocation-start time in Korea (`Asia/Seoul`), formatted as
`YYYYMMDD-HHmmss`. For example, an illustrative model named `model-a` could create
`runs/v1/codex-model-a-high-20261002-153000/`. If that path exists, the next run uses
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
- The current invoking session/thread ID as `session_id`, when verified.
- Input and output token totals, when available for the measured work.
- Elapsed implementation and verification time in milliseconds.
- A format version, execution status and a few short notes for skill context,
  sources, missing values, verification limits or failure.

Format `2.1` adds `session_id` to `2.0`. Codex records the current thread ID;
Claude Code records the current session ID. Unavailable or unverified identities
stay null with a short reason in `notes`, independently of token availability.
Resuming the same attempt preserves its original session ID. See
[session identity](skills/benchmark-record/references/record-format.md#session-identity).

There are no response copies, child/turn identifiers, hashes, transcript or
evidence archives, cache/reasoning breakdowns, or previous-record backups.
The prompt and HTML are the retained task and result. Runs and local experiments
are ignored by Git; existing archives remain unchanged. Earlier runs retain their
original prompts and may require a server, so compare them separately from runs
using the current prompt.

Browser verification uses only connected Aside MCP's direct `repl` tool. The current model
controls the browser; Aside `exec` and natural-language CLI tasks must not launch
another model session. Install the [Aside CLI](https://docs.aside.com/help/developers)
and register MCP before running a benchmark. For Codex:

```sh
codex mcp add aside -- /absolute/path/to/aside mcp --host local
```

Start a new Codex session to load the registered tools. Use a new benchmark tab
and close only tabs created by the attempt. Do not access unrelated tabs or memory.
Select the transport before navigation: use `file://` only with confirmed support.
When local-file support is unavailable or unestablished, use documented,
host-permitted loopback HTTP. A file navigation failure caused only by missing
browser local-file capability permits the same fallback; it does not establish
an OS permission failure.

Serve only the actual run's unchanged `output/` directory as a tracked host process:

```sh
python3 -m http.server 8765 --bind 127.0.0.1 --directory "<run>/output"
```

Replace `<run>/output` with the real path, or use `output` from the run directory.
Wait for startup, then open `http://127.0.0.1:8765/index.html` through Aside's
direct REPL. If the port is occupied, use another available port in both the
command and URL; do not stop unrelated processes. Stop the temporary server on
every exit, including failure, before measurement ends.

The artifact must still run directly without a server. Record HTTP interaction
checks separately from static `file://` compatibility; neither proves direct-file
execution. See [browser verification](skills/benchmark-record/references/browser-verification.md).

This fallback never overrides an explicit host/tool security policy. If a required
file is inaccessible, access is explicitly denied, Aside is unavailable or no
permitted method can perform a required check, stop the entire attempt. Do not
continue implementation or remaining/static checks, switch tools or protocols to
bypass a denial, or install another browser runtime. Preserve existing artifacts;
mark an already-created, writable `run.json` as `failed`, never `completed`.
If setup stopped before a run existed, create no run. Report the exact blocker,
unperformed work and saved/unsaved state; do not retry a denied write. Missing
optional metadata alone is not an access denial.

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
[delegated usage collection](skills/benchmark-record/references/delegated-usage.md).

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

### Validating skill changes

For runtime regression comparisons, create new attempts using the current `2.1` template,
the identical fixed task, matching host/model/effort, and the same permitted Aside verification method.
Check all task requirements and its acceptance scenario, and compare scoped measurements only when their boundaries and coverage are proven.
Historical `2.0` records or Playwright observations do not establish equivalence under the current schema and browser policy; preserve them unchanged.
Fresh-context instruction evaluations and static checks are narrower evidence, not substitutes for an observed browser run.

## Shared Skill Setup

| Host | Project entrypoint |
| --- | --- |
| Codex | `.agents/skills/benchmark-record/SKILL.md`, `.agents/skills/benchmark-review/SKILL.md` |
| Claude Code | `.claude/skills/benchmark-record/SKILL.md`, `.claude/skills/benchmark-review/SKILL.md` |

These are ordinary files containing discovery metadata and a relative link to
the corresponding shared skill under `skills/`. Resources resolve from that
shared skill directory. The Claude recording entrypoint includes current-session
context; workflow rules remain in shared skills. There are no symlinks.

### Skill document layout

The shared `SKILL.md` contains the workflow; references separate the detailed rules by role.
Read the shared [record format](skills/benchmark-record/references/record-format.md),
[settings](skills/benchmark-record/references/settings-resolution.md), [usage](skills/benchmark-record/references/usage-resolution.md)
and [browser verification](skills/benchmark-record/references/browser-verification.md), then only the current host's
[Codex](skills/benchmark-record/references/codex.md) or [Claude Code](skills/benchmark-record/references/claude-code.md) document.
Read Shared procedure and the current-host section of [delegated usage](skills/benchmark-record/references/delegated-usage.md) only when delegating.
The fixed task, record schema, settings precedence and token calculations are unchanged; authoring Sources are not runtime inputs.

The review skill separates its workflow, scoring rubric, browser checks and
record format. Its versioned JSON template is independent of `run.json`; adding
a review does not migrate generation records or change their measurement bounds.

For another project or a personal installation, copy the complete
`skills/benchmark-record/` or `skills/benchmark-review/` folder, including its assets
and references, into the host's skill directory. Do not copy only a discovery entrypoint. The official
[Codex guide](https://learn.chatgpt.com/docs/build-skills) and
[Claude Code guide](https://code.claude.com/docs/en/skills) describe discovery
locations and invocation.

This project provides execution with automatic minimal recording and a separately
invoked, single-run review. Batch ranking and cross-run comparison are not included.
Neither skill switches models, launches a second model, installs hooks, or runs
a separate benchmark CLI.

- [Current architecture](docs/adr/0001-use-a-shared-recording-skill.md)
- [Record fields and accounting](skills/benchmark-record/references/record-format.md)
