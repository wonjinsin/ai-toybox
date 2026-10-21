# Solar System AI Benchmark

Run the same interactive Three.js task in OMP, Codex, or Claude Code, save the generated artifact and execution record, then review a saved attempt separately.

**[Browse the live benchmark gallery](https://wonjinsin.github.io/ai-toybox/)** · [Fixed task](skills/benchmark-record/assets/solar-system-prompt.md) · [Saved runs](runs/v1/)

The gallery is the current score catalog. It shows model outputs, quality ranges, requirement verdicts, review findings, and available execution measurements. Scores are not duplicated here because saved reviews and scoring policies can change.

## What the benchmark measures

The current task is **`solar-system-observatory-2`**. It asks for an observation tool with:

- A genuine Three.js scene containing one sun and exactly three planets, with fixed orbital phases and periods of 12, 24, and 48 simulation seconds.
- Consistent playback, pause/resume, speed from 0.25x to 4x, and seeking from 0 to 120 simulation seconds.
- Planet selection, live information, camera tracking, and a full-system overview.
- Up to three immutable observation snapshots, restore/delete controls, transactional JSON import/export, and browser-storage persistence.
- Usable controls at 1440×900 and 390×844 CSS pixels, including keyboard access to primary DOM controls.

The [complete prompt](skills/benchmark-record/assets/solar-system-prompt.md) defines all 15 requirements, the JSON contract, and acceptance sequences. Generated `output/index.html` must open directly through `file://`, without a server or build step. External CDN assets may still need network access; direct-file delivery does not mean offline delivery.

Generation and review are separate operations. `benchmark-record` creates an attempt and records it in one invocation. `benchmark-review` evaluates one explicitly selected saved attempt without repairing it. Neither skill switches models or starts a second model session.

## Browse results

Open the **[GitHub Pages gallery](https://wonjinsin.github.io/ai-toybox/)** without installing the project, or open the local [`index.html`](index.html) in a browser with JavaScript enabled.

The explorer provides model navigation, search, latest/quality sorting, scored/unscored filters, bookmarkable run details, prompts, review findings, recording notes, and HTML previews. It reads the static `gallery/data.js` snapshot, not a live API. Local edits appear only after regenerating the catalog and reloading the page; published results update after deployment.

### Read the results correctly

| Signal | Meaning |
| --- | --- |
| Generation record status | Whether implementation and required recording/accounting finished. `failed` can still have a usable, reviewed HTML artifact. |
| Quality range | Evidence-backed score bounds out of 100. The headline is the conservative lower bound unless both endpoints coincide. |
| Requirements verdict | Whether mandatory requirements pass, fail, or remain unresolved. A high quality score does not override a failed requirement. |
| Review status and coverage | Whether the assessment and reviewer metadata are complete, and which checks passed, failed, or remain unverified. `partial` is not the same as a failed artifact. |

New current-task reviews use **format `5.0` / rubric `solar-system-5.0`**. Quality scores have no mandatory-failure cap. Unknown check ratings remain null and contribute a 0–5 range; every registered check remains in the denominator. Equal rounded endpoints do not by themselves establish complete coverage. Missing or invalid reviews are **unscored**, not zero.

The gallery can also project valid frozen format `4.0` observations through the current 5.0 policy. These are labeled **policy-only projections**: they do not rerun the artifact, add evidence, or improve its behavior. Original published scores and reviews remain available in the details. Historical format `2.0` results use a different task and are kept separate from the current-policy ranking.

See the [rubric](skills/benchmark-review/references/rubric.md) and [scoring design notes](docs/benchmark-v2.md) for weights, task routing, frozen policies, and interpretation limits. Model identity, generation time, and token counts do not affect quality scores.

> [!WARNING]
> Gallery previews execute trusted repository HTML with scripts and same-origin access. Artifacts can access gallery pages and browser storage; the iframe is not a security boundary. Do not add or publish untrusted HTML. Opening an artifact in the gallery does not certify its benchmark requirements or direct-file compatibility.

## Run a new attempt

### Prerequisites

- Open **`ai-benchmark/` as the project** in your host and choose the model and effort before starting. A fresh conversation per attempt is recommended.
- Use **Node.js 22 or newer** for the bundled record/review helpers, gallery scripts, and tests. These tools need no `npm install` step.
- For generation browser verification, install and connect **Aside MCP before the attempt**. `benchmark-record` uses its direct `repl` tool. For review, install **ego-browser (Ego Lite)** beforehand; `benchmark-review` uses `ego-browser nodejs` and the installed `ego-browser skill` guide, not Aside or a second model.
- A `completed` record needs a verified current-session identity, model, effort, elapsed work time, and full-interval input/output token totals. If accounting is unavailable, the skill still builds what it can and saves honest nulls and a `failed` record.

For Codex, register an existing [Aside CLI installation](https://docs.aside.com/help/developers), replacing the executable path:

```sh
codex mcp add aside -- /absolute/path/to/aside mcp --host local
```

Start a new Codex session after registration. For other hosts, connect Aside through the host's MCP setup. Do not install tools or change MCP configuration during the measured attempt.

### Invoke the recording skill

These are **chat messages, not terminal commands**.

Codex:

```text
$benchmark-record Run the Solar System benchmark.
```

Claude Code:

```text
/benchmark-record Run the Solar System benchmark.
```

OMP, or a host without project skill discovery:

```text
Read skills/benchmark-record/SKILL.md and run the Solar System benchmark.
```

No additional task prompt is needed. The skill reserves a fresh run directory, copies the fixed prompt, implements the HTML, performs permitted checks, and publishes the terminal record before answering. You can request a different output root in the same message; existing attempts are not overwritten.

During benchmark execution and recording, only explicitly selected skills apply unless higher-priority host instructions require others. The skill records unavoidable exceptions in `run.json.notes`; this does not establish full environment isolation. Ordinary repository maintenance keeps the host's normal skill selection. See [skill context rules](skills/benchmark-record/references/record-format.md#skill-context-notes).

### Browser verification boundaries

Generation must deliver direct-file-compatible HTML. Use `file://` verification only when supported and permitted; otherwise the recording workflow permits loopback HTTP over the unchanged `output/` directory. HTTP interaction checks do not prove direct-file execution.

Use only tabs created for the attempt and close them afterward. Do not access unrelated tabs or browser memory. An explicit access or security denial stops the affected operation: do not retry it, change permissions, or switch transports to bypass it. Unavailable capabilities remain documented verification gaps, not fabricated passes. See [generation browser verification](skills/benchmark-record/references/browser-verification.md).

## Review a saved attempt

Invoke the separate [review skill](skills/benchmark-review/SKILL.md) with one actual run directory. Replace the example path below.

Codex:

```text
$benchmark-review runs/v1/<agent>-<model>-<effort>-<run-id>
```

Claude Code:

```text
/benchmark-review runs/v1/<agent>-<model>-<effort>-<run-id>
```

OMP, or a host without project skill discovery:

```text
Read skills/benchmark-review/SKILL.md and review runs/v1/<agent>-<model>-<effort>-<run-id>.
```

The reviewer reads the saved prompt to select the compatible rubric. It does not invoke generation, repair the artifact, or modify `prompt.md`, `run.json`, or `output/`. It writes only:

- `reviews/review.json`: structured checks, scores, coverage, and verification limits.
- `reviews/review.md`: human-readable findings and reproduction steps.

**Re-review replaces both reports.** There are no timestamped review backups or retained screenshot/evidence archives; unrelated existing files are preserved. Each replacement is atomic, but the pair is not a filesystem transaction. Publication is complete only after both files pass read-back.

Reviews use **ego-browser (Ego Lite)** and permitted loopback HTTP serving only the unchanged output directory. Read `ego-browser skill`, use one TaskSpace per review, and follow the documented Page APIs. **Direct-file execution remains outside review scope**, even though generation requires it. Functional defects are findings, not reasons to abandon unrelated checks. Unavailable capabilities leave affected checks unverified; missing required reviewer metadata keeps the review partial. See [review verification](skills/benchmark-review/references/verification.md) and the [review format](skills/benchmark-review/references/review-format.md).

## Saved files and measurements

```text
runs/v1/<agent>-<model>-<effort>-<run-id>/
  prompt.md           # Exact task used for this attempt
  output/index.html   # Generated artifact; local resources stay in output/
  run.json            # Generation settings, measurements, status, and notes
  reviews/            # Added by the separately invoked review skill
    review.json
    review.md
```

`runs/v1/` is the storage layout, not the task version. The saved prompt identifies the task; the review identifies its rubric. Runs are tracked in Git. Local `experiments/` content is ignored and is not part of the published run catalog.

The host component identifies OMP, Codex, or Claude Code, not the model provider. Run IDs use invocation-start time in `Asia/Seoul` as `YYYYMMDD-HHmmss`; collisions receive `-02`, `-03`, and so on. Unknown directory components may use `unknown`, but unknown JSON metadata stays null with a reason, not a placeholder or guessed model version.

Current generation records use format **`2.1`**. Measurement starts immediately before implementation and ends after its checks and cleanup. Record processing and the final reply are excluded. Token totals must cover that same work interval, including all delegated work, without double counting. Context-window counters, estimates, and unrelated sessions are not substitutes for usage measurements.

`completed` requires valid `session_id`, `model`, `effort`, `duration_ms`, `input_tokens`, and `output_tokens`; it does not certify bug-free behavior or complete browser coverage. Missing required accounting yields `failed` even when the artifact exists. Deliberately paused unfinished work uses `interrupted`; an abruptly terminated session may leave `running`, because no background service finalizes it. Historical records are not automatically relabeled.

Detailed contracts:

- [Record fields, naming, status, and safe publication](skills/benchmark-record/references/record-format.md)
- [Settings resolution](skills/benchmark-record/references/settings-resolution.md), [usage accounting](skills/benchmark-record/references/usage-resolution.md), and [delegated usage](skills/benchmark-record/references/delegated-usage.md)
- Host-specific sources: [OMP](skills/benchmark-record/references/omp.md), [Codex](skills/benchmark-record/references/codex.md), [Claude Code](skills/benchmark-record/references/claude-code.md)

## Local gallery and maintenance

Run the terminal commands in this section from **`ai-benchmark/`**. From the parent repository:

```sh
cd ai-benchmark
```

### Refresh and preview

Regenerate the local catalog after changing records or reviews:

```sh
node scripts/generate-gallery.mjs
```

This reads the working tree, including uncommitted records, and replaces only `gallery/data.js`. Reload [`index.html`](index.html) afterward. The snapshot has no generation timestamp; unchanged inputs produce identical output. There is no live file watcher or automatic preview refresh.

To read another collection or write to another destination:

```sh
node scripts/generate-gallery.mjs --root /absolute/ai-benchmark --output /absolute/data.js
```

For optional HTTP preview:

```sh
node scripts/serve-gallery.mjs --port 4173
```

Open **http://127.0.0.1:4173/** and stop the server with Ctrl+C. It binds only to loopback and serves saved gallery assets and run outputs, not `/api/runs`. `--root DIRECTORY` selects a complete site directory, including its application assets and generated catalog. Use another port if 4173 is occupied; do not stop unrelated processes.

### Validate records and scores

These commands are read-only. Replace `<run>` with an existing run directory:

```sh
node skills/benchmark-record/scripts/validate-run.mjs "<run>/run.json"
node skills/benchmark-review/scripts/score-review.mjs "<run>/reviews/review.json"
```

The record validator checks structure and required completion values, not the truth of measurement sources. The scorer validates input and prints normalized JSON; it neither rewrites the review nor executes the artifact. Existing records do not bypass validation because they are historical.

Normal recording and review invocations handle publication. For manual checkpoint/finalization work, use the [record publication contract](skills/benchmark-record/references/record-format.md#validation-and-publication) and [review publication command](skills/benchmark-review/references/review-format.md#publication-command). Keep drafts outside the target run. Record finalization rejects stale writes and terminal-record overwrites; invalid review proposals do not replace saved reports.

### Optional commit hook

```sh
node scripts/install-gallery-hooks.mjs
```

The installer changes only this repository's local `core.hooksPath`, stores private hooks in the Git common directory, and preserves the previous hook chain. It does not edit shared hooks or global Git configuration. Reinstall after changing the hook implementation. Bare repositories and worktree-specific `core.hooksPath` overrides are unsupported.

When staged paths under `ai-benchmark/` change, pre-commit runs the previous hook first, then uses the **staged generator and collector** to build a catalog from the effective Git index. It supports partial staging, staged deletions, and alternate `GIT_INDEX_FILE` values; unstaged edits and untracked files do not enter the snapshot.

The hook updates **only `ai-benchmark/gallery/data.js` in the index**, not the working-tree catalog. Regenerate locally when you want your preview to match working-tree edits. Generation failure stops the commit. Unresolved index entries, symlinks, and submodules inside the project are rejected. Unrelated commits skip generation; the hook never pushes or deploys.

### GitHub Pages deployment

The [Benchmark gallery workflow](../.github/workflows/benchmark-gallery.yml) runs when a push to **`master`** changes `ai-benchmark/**` or the workflow itself. It also supports manual dispatch. Before the first deployment, set **Settings → Pages → Build and deployment → Source → GitHub Actions**.

Actions uses Node.js 22, runs the gallery/hook regression tests, generates the catalog from the checked-out commit, and deploys with the official GitHub Pages actions. The upload contains `index.html`, browser assets in `gallery/`, and each run's `output/` files. It does not upload raw record files, repository configuration, tests, or skills as standalone files, but **prompts, reviews, notes, and normalized run fields are embedded in the public catalog**. Check that content before publishing.

### Verification

Run all bundled regression tests:

```sh
node --test tests/*.test.mjs
```

> [!WARNING]
> The full suite currently requires local experiment evidence that is not included in a clean checkout. `tests/benchmark-probes.test.mjs` reads `experiments/observatory-improvement/terra-xhigh/manifest.json` and `evidence/*.json`, but `/experiments/` is Git-ignored. Without those files, that test file fails with `ENOENT` before its test cases run. The Pages workflow runs only `tests/gallery*.test.mjs`; it does not exercise this experiment-dependent test.

Hook tests use temporary Git repositories rather than staging this checkout. Tests cover recording, publication, scoring, gallery behavior, paths, hooks, and probe helpers; they do not certify generated artifacts against the rubric. Artifact verification requires observed browser checks under the applicable [generation](skills/benchmark-record/references/browser-verification.md) or [review](skills/benchmark-review/references/verification.md) policy.

## Shared skills and references

| Host | Project discovery files |
| --- | --- |
| Codex | `.agents/skills/benchmark-record/SKILL.md`, `.agents/skills/benchmark-review/SKILL.md` |
| Claude Code | `.claude/skills/benchmark-record/SKILL.md`, `.claude/skills/benchmark-review/SKILL.md` |
| OMP / manual loading | Read the corresponding shared `skills/benchmark-record/SKILL.md` or `skills/benchmark-review/SKILL.md` explicitly. |

The Codex and Claude Code entrypoints are ordinary files, not symlinks. They point to shared workflows under `skills/`; resources resolve from the shared directory. Claude's recording entrypoint also supplies current-session context. Reopen the project or start a new session if newly installed skills are not discovered.

For another project or a personal installation, copy the **complete** recording or review skill directory, including assets, references, and scripts. Copying only a discovery entrypoint is insufficient. See the host's [Codex skill guide](https://learn.chatgpt.com/docs/build-skills) or [Claude Code skill guide](https://code.claude.com/docs/en/skills) for installation locations.

- [Recording workflow](skills/benchmark-record/SKILL.md)
- [Review workflow](skills/benchmark-review/SKILL.md)
- [Current rubric](skills/benchmark-review/references/rubric.md) and [review schema](skills/benchmark-review/references/review-format.md)
- [Task and scoring design](docs/benchmark-v2.md)
- [Shared-skill architecture](docs/adr/0001-use-a-shared-recording-skill.md)
