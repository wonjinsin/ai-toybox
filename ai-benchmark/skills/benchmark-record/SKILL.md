---
name: benchmark-record
description: "Use when the user explicitly invokes benchmark-record to run the fixed Solar System benchmark."
disable-model-invocation: true
---

# Run the Solar System Benchmark

Use the current model and ordinary host tools to build the [fixed task](assets/solar-system-prompt.md), save its record, then reply in this invocation.
No second collection request, analysis/comparison, separate model launcher, collector, hook or persistent background service.
Resolve resources relative to this file. Browser startup and temporary servers follow the verification rules below.

The current task is `solar-system-observatory-2`, with 15 requirements. Its version
identity comes from the exact prompt, not the `runs/v1/` directory name. Keep record
format `2.1` and that storage root unchanged. The unversioned 11-requirement task is
retained unchanged as [the historical prompt](assets/solar-system-prompt-v1.md).
Do not rewrite saved prompts or rerun historical attempts during skill maintenance.

## Read first

- [Record format](references/record-format.md): directory, session ID, skill selection, status/resume and retention. Use the [JSON template](assets/run.json).
- [Shared settings rules](references/settings-resolution.md) and [shared usage rules](references/usage-resolution.md).
- Only the actual current host's settings/session/token rules: [OMP](references/omp.md), [Codex](references/codex.md) or [Claude Code](references/claude-code.md). A model/provider name does not identify the host.
- [Browser verification](references/browser-verification.md): permitted methods and gaps.

Before delegating, also read the shared and current-host sections of [delegated usage](references/delegated-usage.md).
Pass the skill selection policy to delegates. Skip other-host sections and authoring Sources.

## Continue safely and finalize

Treat artifact quality, available verification and record publication separately.
Missing metadata, usage, browser transport or a check-specific capability does not
stop construction or independent permitted checks. Keep unknown values null with
specific reasons. Missing required identity/measurements prevents `completed`, not
implementation or saving the artifact. A finished, fully identified/measured
attempt may be `completed` with functional defects or unverified browser checks
documented in notes. This status does not certify a bug-free or fully verified app.

Investigate observed functional defects and fix them during generation when
feasible. Distinguish an application bug from an unsupported browser operation.
Do not loop on unsupported operations or claim static inspection proves behavior.
Continue independent checks and record unresolved defects and gaps.

An explicit access/security denial stops the affected operation. Do not retry it,
change permissions, switch tools/protocols to bypass it, or ask a delegate to do
so. Continue only independently permitted work unless the denial's policy scope
forbids all work. Unavailable local-file capability is not itself a security
denial; apply the [documented transport rules](references/browser-verification.md).
A hard failure accessing the required prompt, target or output location prevents
the affected construction/publication. Report it rather than silently choosing a
different target. Preserve existing artifacts and finalize `failed` if the record
remains writable; a denied record write must not be retried.

Always attempt terminal finalization in this invocation after finishing permitted
work, including partial/failed outcomes. Capture an end boundary when possible
and clean up only attempt-created tabs/servers through permitted operations.
Report the exact observed error or missing capability, affected work/checks and
saved/unsaved record state. If finalization itself is unavailable, explicitly say
which record remains `running` or was never created. Do not invent a saved result.

## Workflow

1. Set up
   - Read the invocation-start clock. Resolve current session ID, model/effort and a trustworthy current-session usage source through permitted, scoped lookups. Record missing sources without delaying construction indefinitely.
   - Exclusively reserve a new empty directory under default `runs/v1/` or the requested output root. Never overwrite existing results. Directory placeholders are not JSON metadata; use honest nulls in the record.
   - Prepare a schema-valid `running` proposal outside the run directory, with known settings and notes explaining all nulls. Publish the early checkpoint with `node <skill-root>/scripts/validate-run.mjs --checkpoint <proposal.json> <run-directory>`. Retain its reported SHA-256 for finalization. Failure to reserve/publish is a hard output blocker, not permission to use a different destination.
   - Copy the current fixed task exactly to `prompt.md`, including its version line and JSON contract. Do not paraphrase, append instructions, or substitute another task. A resumed attempt keeps its original saved prompt, session identity and task version. Follow the [resume publication rules](references/record-format.md#validation-and-publication).
   - Discover permitted browser capabilities, but do not require them to start implementation.

2. Start measurement
   - Immediately before implementation, capture a real clock, usage baseline and issuing request ID in a separate boundary call.
   - If any boundary cannot be proven, record the limitation, leave the affected final measurement null, and continue permitted implementation. Missing injected counters does not end scoped lookup.

3. Build and check
   - Save `output/index.html` and local resources under `output/`; preserve server-free `file://` execution.
   - Follow browser verification and record methods/gaps. Investigate and fix observed functional errors when feasible. Finish all delegated work and available checks before stopping measurement.

4. Stop measurement
   - Immediately after implementation/checks, capture the end clock, matching usage endpoint and issuing request ID in a separate boundary call.
   - Exclude recording/reply. Calculate same-interval token totals per the references. Unprovable values stay `null` with a reason and require `failed`, never `completed`.

5. Finalize
   - Prepare terminal JSON outside the run directory with the fixed schema and settings/source, defect, verification and interruption/resume notes. Only finished, fully identified and measured attempts may be `completed`. Missing final identity, counts, bounds or full delegated coverage requires `failed`; preserve HTML. Use `interrupted` only for unfinished interrupted work.
   - Publish with `node <skill-root>/scripts/validate-run.mjs --finalize <proposal.json> <run-directory> <expected-sha256>`. This validates before touching `run.json`, rejects unsafe paths and stale/historical writes, atomically replaces it and reads it back. See the [publication contract](references/record-format.md#validation-and-publication). Correct invalid proposals without inventing facts.
   - Observations remain working data. Report publication errors and any still-running record; never fabricate artifacts or measurements.

6. Read back and reply
   - Use the CLI's validated read-back; separately verify session identity, unchanged prompt, output path and measurement provenance. Structural validation cannot prove source truth or coverage.
   - Reply with terminal status, HTML/record paths, measured timing/tokens or honest nulls, unresolved defects and verification gaps, outside measurement. Distinguish a completed measured attempt from a verified application.
