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
