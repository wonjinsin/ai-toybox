---
name: benchmark-record
description: "Use when the user explicitly invokes benchmark-record to run the fixed Solar System benchmark."
disable-model-invocation: true
---

# Run the Solar System Benchmark

Use the current model and ordinary host tools to build the [fixed task](assets/solar-system-prompt.md), save its record, then reply in this invocation.
No second collection request, analysis/comparison, separate model launcher, collector, hook or persistent background service.
Resolve resources relative to this file. Browser startup and temporary servers follow the verification rules below.

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
blocker is resolved. Merely absent optional metadata or token counters without
an access denial still follows the existing null/unknown rules.

## Workflow

1. Set up
   - Read the invocation-start clock; resolve the current session ID, usage source and model/effort, preferring runtime settings.
   - Prepare a permitted browser check and reserve a new directory under default `runs/v1/` or the requested output root. Never overwrite existing results.
   - Copy the fixed task exactly to `prompt.md`; do not request/substitute another task. Save `run.json` as `running` with settings/source notes.

2. Start measurement
   - Immediately before implementation, capture a real clock, usage baseline and issuing request ID in a separate boundary call.
   - Missing injected counters does not end lookup. Unavailable usage must not block work.

3. Build and check
   - Save `output/index.html` and local resources under `output/`; preserve server-free `file://` execution.
   - Follow browser verification and record the method/gaps. Finish all delegated work and checks before stopping measurement.

4. Stop measurement
   - Immediately after implementation/checks, capture the end clock, matching usage endpoint and issuing request ID in a separate boundary call.
   - Exclude recording/reply. Calculate same-interval token totals per the references; unprovable values stay `null` with a reason.

5. Finalize
   - Save final `run.json` now with the fixed schema, settings/source and interruption/resume notes. Follow record status and retention rules.
   - Observations remain working data. Report write failures; never fabricate artifacts or measurements.

6. Read back and reply
   - Verify JSON, session ID, exact prompt, output path and known metrics.
   - Reply with HTML/record paths, available timing/tokens and important verification gaps, outside measurement.
