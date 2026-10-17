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
- Only the current host's settings/session/token rules: [Codex](references/codex.md) or [Claude Code](references/claude-code.md).
- [Browser verification](references/browser-verification.md): permitted methods and gaps.

Before delegating, also read the shared and current-host sections of [delegated usage](references/delegated-usage.md).
Pass the skill selection policy to delegates. Skip other-host sections and authoring Sources.

## Stop on access or verification blockers

First apply the [browser transport rules](references/browser-verification.md):
unavailable `file://` access alone permits the scoped HTTP fallback when documented
and host-permitted. This exception never overrides an explicit access/security policy.

Stop the entire attempt immediately if a required file cannot be accessed, any
tool reports an access/security denial (including settings or usage lookup), or
Aside's direct `repl` is unavailable or cannot perform a required browser check.
Do not continue implementation, delegated work or remaining/static checks, retry
denied access, switch protocols/tools to bypass a denial or change permissions.

Only stop active work, clean up attempt-created tabs/servers, capture an end
boundary if measurement started and permitted tools remain available, and mark
an already-created `run.json` as `failed` if writable. Preserve existing artifacts;
do not create a run just to record a setup blocker. Never mark this attempt
`completed`. Do not retry a denied record write; report if the record remains
`running`. Unprovable measurements stay null.

Report the affected path/tool and operation, exact observed error (or missing
capability), work/checks not performed, and saved/unsaved record state immediately.
Distinguish current observations from earlier reported restrictions; never invent
an error or reuse earlier results as fresh evidence. Resume only after the
blocker is resolved. Missing metadata or counters without a denial still require
the setup and completion gates below. They are not optional successful-run fields.

## Workflow

1. Set up
   - Read the invocation-start clock. Resolve the current session ID and concrete model/effort, preferring runtime settings. Locate a trustworthy current-session input/output usage source with usable boundary and finality accounting.
   - Exhaust permitted fallback sources. If required setup information or a reliable clock remains unavailable, stop without reserving a directory or starting implementation. Do not fabricate values or install a collector.
   - Prepare a permitted browser check, then reserve a new directory under default `runs/v1/` or the requested output root. Never overwrite existing results.
   - For a fresh attempt, copy the current fixed task exactly to `prompt.md`, including its version line and JSON contract. Do not paraphrase, append instructions, or substitute another task. A resumed attempt keeps its original saved prompt and task version. Save `run.json` as `running` with settings/source notes.

2. Start measurement
   - Immediately before implementation, capture a real clock, usage baseline and issuing request ID in a separate boundary call.
   - Missing injected counters does not end lookup. If the start boundary cannot be proven, stop implementation and finalize the reserved attempt as `failed` with null affected measurements and reasons.

3. Build and check
   - Save `output/index.html` and local resources under `output/`; preserve server-free `file://` execution.
   - Follow browser verification and record the method/gaps. Finish all delegated work and checks before stopping measurement.

4. Stop measurement
   - Immediately after implementation/checks, capture the end clock, matching usage endpoint and issuing request ID in a separate boundary call.
   - Exclude recording/reply. Calculate same-interval token totals per the references. Unprovable values stay `null` with a reason and require `failed`, never `completed`.

5. Finalize
   - Save final `run.json` now with the fixed schema, settings/source and interruption/resume notes. Only finished, fully identified and measured attempts may be `completed`. Missing final counts, bounds, or full delegated coverage require `failed`; preserve artifacts. Unfinished interruptions retain the interruption rules.
   - Before publishing any final record, run `node <skill-root>/scripts/validate-run.mjs <run.json>` and require success. The [validator contract](references/record-format.md#validation-and-publication) covers every status. Fix invalid structure without inventing missing facts.
   - Observations remain working data. Report write failures; never fabricate artifacts or measurements.

6. Read back and reply
   - Read back the saved record and rerun the validator on that file. Verify session ID, exact prompt, output path and measurement provenance. Validation proves structure, not truth or source coverage.
   - Reply with status, HTML/record paths, measured timing/tokens and important verification gaps, outside measurement. Never report an incomplete record as successful.
