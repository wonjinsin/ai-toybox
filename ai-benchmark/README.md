# Solar System AI Benchmark

Run one fixed Solar System task in OMP, Codex, or Claude Code. The skill builds the
HTML, saves a validated run record, and then replies with the result.
One invocation completes the workflow; there is no follow-up collection step.

The [fixed prompt](skills/benchmark-record/assets/solar-system-prompt.md) is task
`solar-system-observatory-2`. It requires a real Three.js scene with one sun and
three planets, deterministic playback and seeking, selection and camera tracking,
and responsive controls. It also requires immutable observation snapshots,
transactional JSON import/export, snapshot persistence, and keyboard access.
The saved `output/index.html` must run directly in a browser without a server or
build step. The [original task](skills/benchmark-record/assets/solar-system-prompt-v1.md)
remains available only to interpret historical runs.

## Browse the static results

Open this folder's `index.html` directly in a browser, or open the published
GitHub Pages site. The gallery loads `gallery/data.js` as a classic script.
It does not call an API, use local module imports, or require a server to read
the catalog. Application assets and output links use relative paths, including
when the site is hosted below a repository path.

The explorer keeps model navigation, search, latest and score sorting,
score-state filters, bookmarkable details, prompts, review findings, duration,
tokens, recording notes, and output previews. Missing scores are **unscored**,
not zero. Numeric zero remains a valid score. Incomplete or malformed metadata
appears with warnings. The gallery does not evaluate or modify saved runs.

### Refresh the local catalog

Run these commands from `ai-benchmark` with Node.js 22 or newer:

```sh
node scripts/generate-gallery.mjs
```

Reload the page after generation. This command reads the working tree,
including uncommitted records and reviews, and replaces only `gallery/data.js`.
The snapshot has no generation timestamp. Its revision and artifact revisions
depend on content, so unchanged input produces identical output.
Live file detection and automatic preview refresh are no longer supported.

To read another collection or write the snapshot elsewhere, use:

```sh
node scripts/generate-gallery.mjs --root /absolute/ai-benchmark --output /absolute/data.js
```

For optional HTTP preview, run:

```sh
node scripts/serve-gallery.mjs --port 4173
```

Open **http://127.0.0.1:4173/**. The server serves saved static files and output
assets, not `/api/runs`. Its `--root DIRECTORY` option selects the complete site
directory, including application assets and the generated catalog. It binds
only to loopback. Stop it with Ctrl+C.

### Install the commit hook

The Git root is the parent `ai-toybox` repository. From `ai-benchmark`, run:

```sh
node scripts/install-gallery-hooks.mjs
```

The installer changes only this repository's local `core.hooksPath`. It creates
private hooks inside the Git common directory and preserves the previous hook
chain, including the shared `~/.git_template/hooks` hooks. It does not edit
shared hooks or global Git configuration. Repeating installation retains the
original chain and refreshes the installed driver. Reinstall after changing
the hook implementation.

When staged paths under `ai-benchmark/` change, pre-commit runs the previous
pre-commit hook first, then builds a temporary tree from the effective Git
index. It runs the **staged generator and collector**, not their unstaged
working-tree versions. Partially staged records, staged deletions, and an
alternate `GIT_INDEX_FILE` therefore produce the catalog for that commit.
Untracked files and unstaged review edits do not enter the snapshot.

After successful generation, the hook updates **only**
`ai-benchmark/gallery/data.js` in the index. It leaves working-tree files
unchanged, including the local catalog. The staged catalog can therefore differ
from the catalog used by your current local preview. Generation failures stop
the commit without installing a new catalog in the index. Unrelated commits
skip generation, and deleting the entire project does not recreate it.

The hook rejects unresolved index entries, symlinks, and submodules inside
`ai-benchmark` rather than reading outside the staged tree. The installer does
not support bare repositories or worktree-specific `core.hooksPath` overrides.
The hook never pushes or deploys.

### Publish with GitHub Actions

The workflow is at the Git root,
`../.github/workflows/benchmark-gallery.yml`. After a push to `master` changes
`ai-benchmark/**` or that workflow, Actions runs the regression suite, generates
the catalog from the checked-out commit, and deploys the static site with the
official GitHub Pages actions. A manual workflow dispatch is also available.
Set the repository's **Settings → Pages → Build and deployment → Source** to
**GitHub Actions** before the first deployment.

The upload contains only `index.html`, the browser assets in `gallery/`, and
each run's `output/` files. Raw record files, repository configuration, tests,
and skills are not uploaded. Prompts, reviews, notes, and normalized run fields
are included in the public catalog because the gallery displays them.

Embedded output previews retain their iframe sandbox without same-origin
access. A standalone output opened in a new tab on Pages does not receive the
local preview server's sandbox response header. Only publish trusted artifacts.
External CDN dependencies may still require internet access. The gallery does
not rewrite benchmark outputs or repair their individual file compatibility.

### Verify changes

Run the regression suite from `ai-benchmark`:

```sh
node --test tests/gallery*.test.mjs
```

Hook regressions use temporary Git repositories and do not stage this checkout.
They cover staged-only generation, partial staging, deletions, alternate
indexes, failure handling, and existing-hook preservation.

Browser verification uses Aside's direct MCP REPL. In the static migration
check, Aside denied `file://` navigation because local file access was disabled.
Direct-file runtime compatibility remains **unverified**. Separate HTTP checks
used an output-only temporary site under `/ai-toybox/`. They exercised model
navigation, search, filters, sorting, detail sections, preview playback, new-tab
output links, keyboard navigation, zero versus unscored values, and missing-data
guidance. The normal gallery path produced no captured console or page errors
and no API requests. HTTP success does not establish `file://` compatibility.

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

The [versioned rubric](skills/benchmark-review/references/rubric.md) uses six
categories: simulation/time correctness (25), selection/tracking consistency (20),
controls and interaction (15), visual quality (15), responsive usability (15),
and delivery stability (10). Category scores range from 0 to 5.

New task reviews use `solar-system-3.0` and review format `3.0`. They preserve the
weighted result as `raw_total_score`. If a mandatory check fails, the published
`total_score` is capped at **69**, with `score_cap` recording that policy.
This threshold is a declared interpretation rule, not an empirically calibrated
model ranking. A high raw score cannot override a failed requirement.

Review the saved prompt before choosing a rubric. The original 11-requirement
task still uses its frozen `solar-system-2.0` rubric and format `2.0`, without
the new cap or requirements. Other tasks receive a blocked, unscored review.
Existing runs and reviews are not migrated. Compare scores only within the same
task and rubric version. The gallery displays rubric versions but does not
prevent cross-version score sorting.

Direct-file execution and file-protocol compatibility remain outside review
scope. Generation still requires direct-file delivery. Review checks use HTTP
and do not certify that excluded generation requirement.

Checks distinguish pass, fail and unverified. Any unverified check leaves its
category and the total unscored; verified failures are still reported separately.
Both review formats require a concrete reviewer model, effort, and review
timestamp before `status: complete` is accepted. Missing metadata stays null
with a reason and does not prevent independent checks. The result is `partial`
with null totals, even when every check was observed.
Aside's direct REPL is the only permitted browser tool. Reviews use documented,
host-permitted loopback HTTP serving only the unchanged `output/` directory.
Do not test `file://` or leave a review unscored for missing local-file access.
See the [browser preflight](skills/benchmark-review/references/verification.md)
for server setup, real CSS viewport fixtures, access-denial boundaries and cleanup.
A functional defect is a finding, not a reason to abandon the remaining checks.
Unavailable browser capabilities leave only the affected checks unverified.
Keep same-attempt findings in a `partial` review and continue permitted work.
Use `blocked` when no usable evaluation can be made or task routing is incompatible.
Never copy an earlier review's evidence into the new attempt.
An access or security denial stops the affected operation. Do not retry or bypass
it. Continue unrelated work only when policy permits it. If the target or output
is unsafe, or a write is denied, report the actual saved and unsaved state.
Past errors are historical context, not current observations. Generation time,
tokens and model identity do not affect quality scores.

Recompute a review's derived fields without modifying its file:

```sh
node skills/benchmark-review/scripts/score-review.mjs "<run>/reviews/review.json"
```

This command validates the scoring input and prints normalized JSON. It rejects
complete reviews with missing required metadata, including older saved records.
Existing files are not rewritten or grandfathered through validation. It does
not execute the artifact or certify the reviewer's observations. See the
[revision and calibration guide](docs/benchmark-v2.md) for the design choices,
verification cases, and limits on interpreting new scores.

Publish both reports from one draft outside the run directory:

```sh
node skills/benchmark-review/scripts/finalize-review.mjs "<run>" "<draft-review.json>"
```

The command validates and scores the draft, generates Markdown from that JSON,
and reads back both files. Invalid proposals do not replace existing reports.
Each file replacement is atomic, but the pair is not a filesystem transaction.
A publication error reports which files were saved. Do not claim pair completion
until both files pass read-back.

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
Here `v1` identifies the storage layout, not the task version. The saved prompt
identifies the task, and `review.json` identifies the scoring rubric.

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
in `notes`. If model, effort, or current-session identity remains unresolved,
keep the corresponding JSON values null with reasons and continue implementation.
Directory components may use `unknown`; JSON metadata may not use placeholders.
JSON preserves known identifiers or aliases without guessing exact model versions. See
[automatic settings resolution](skills/benchmark-record/references/settings-resolution.md)
and the complete
[directory naming rules](skills/benchmark-record/references/record-format.md#run-directory).

OMP uses explicit current-session runtime metadata and permitted, identity-checked
usage sources. It never treats an OMP model provider as the host or guesses a
Codex thread ID. Codex checks applicable configuration when runtime metadata is
missing. Claude Code checks its active environment and applicable settings.
No path launches another model or requires a second user request.

The [record template](skills/benchmark-record/assets/run.json) keeps:

- Concrete model and effort with source information in notes.
- The verified current invoking session/thread ID as `session_id`.
- Complete input and output token totals for the measured work.
- Elapsed implementation and verification time in milliseconds.
- A format version, execution status, and notes for sources, verification limits,
  interruptions, or failure.

Format `2.1` adds `session_id` to `2.0`. Each host records its actual invoking
session identity. Missing identity or accounting does not prevent building HTML.
Finalize such attempts as `failed` with honest nulls and reasons.
`completed` requires all metadata and measurements, but does not certify bug-free
behavior or full browser coverage. Record functional defects and verification
gaps in notes. Resuming retains original identity and measurement bounds.
See [session identity](skills/benchmark-record/references/record-format.md#session-identity).

Validate a proposed or saved record without modifying it:

```sh
node skills/benchmark-record/scripts/validate-run.mjs "<run>/run.json"
```

The validator requires `session_id`, `model`, `effort`, `duration_ms`,
`input_tokens`, and `output_tokens` for a completed run. It rejects missing
values and placeholder strings, not genuine measured zero. It validates the
record's structure, not the truth of its sources. The same gate applies to old
saved records when explicitly validated; this policy does not rewrite them.

Reserve an empty run directory and publish a running checkpoint before work:

```sh
node skills/benchmark-record/scripts/validate-run.mjs --checkpoint "<draft.json>" "<run>"
```

After work, publish a terminal draft using the SHA-256 of the current record:

```sh
node skills/benchmark-record/scripts/validate-run.mjs --finalize "<draft.json>" "<run>" "<expected-sha256>"
```

Keep drafts outside the run directory. The command validates before replacement,
rejects stale writes and terminal-record overwrites, and reads the saved JSON
back. A completed proposal also requires the saved prompt and HTML entrypoint.
Use the [publication contract](skills/benchmark-record/references/record-format.md#validation-and-publication)
for interrupted checkpoints and safe resume.

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

This fallback never overrides an explicit host or tool security policy. Stop a
denied operation without retrying it, changing permissions, or switching tools
to bypass it. Continue independent implementation and checks only where permitted.
Unavailable Aside or an unsupported input operation is a verification gap, not
a command to abandon implementation. Use available static checks without claiming
they prove browser behavior. Record the exact limitation and unperformed checks.
Finalize writable JSON before replying. A denied output write is an explicit
publication failure, never a successful save.

## Measurement Boundary

Measurement starts immediately before implementation and ends after its checks.
Record processing and the final user-facing reply are outside that interval.
Time includes the model's implementation work, tools and verification, not only
API latency. The task does not pin a Three.js version; delivery must work via
`file://`.

The agent uses normal clock and host inspection tools. It automatically checks
current-run usage and existing current-session files, following
[usage resolution](skills/benchmark-record/references/usage-resolution.md).
OMP can use a permitted current-session usage API or explicitly supplied journal;
Codex can use App Server usage or its current rollout; Claude Code can use its
current session transcript. Runtime identity selects the source. The agent never
substitutes the newest session or another conversation from the project.
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

If current-session identity or a usable usage source cannot be established during
setup, retain nulls and continue the task. A context-window counter is not interval
usage. An environment without reliable accounting can produce an artifact and a
valid finalized record, but not a `completed` format `2.1` benchmark record.

If final usage, clock bounds, or full delegated coverage cannot be established
after work starts, preserve the artifact and save the attempt as `failed`.
Keep unknown values null with precise reasons. Valid partial parent counts may
remain in notes but cannot replace full-work totals. Do not substitute estimates,
zero, or another session's values, and do not claim completion because the HTML
was built. The skill finalizes the failure and replies in the same invocation.

Use `failed` when implementation cannot finish or required metadata and
measurements remain unavailable. A functional defect or browser coverage gap
alone does not make a finished, measured attempt failed. Use `interrupted` for
unfinished work that was deliberately paused. An abrupt process termination can
still leave `running`; no background service finalizes a dead session.
These rules govern new work. Historical statuses are not relabeled.

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
[OMP](skills/benchmark-record/references/omp.md),
[Codex](skills/benchmark-record/references/codex.md), or [Claude Code](skills/benchmark-record/references/claude-code.md) document.
Read Shared procedure and the current-host section of [delegated usage](skills/benchmark-record/references/delegated-usage.md) only when delegating.
The run record schema, settings precedence and token calculations are unchanged.
The fixed task is versioned separately. Authoring Sources are not runtime inputs.

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
