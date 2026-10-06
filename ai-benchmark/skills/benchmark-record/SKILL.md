---
name: benchmark-record
description: "Use when the user explicitly invokes benchmark-record to run the fixed Solar System benchmark."
disable-model-invocation: true
---

# Run the Solar System Benchmark

Build the [fixed task](assets/solar-system-prompt.md) and finalize its record before
replying in this invocation. No second collection request, analysis or comparison.

## Skill selection

Apply only skills explicitly selected in the current request. A request to read and run
this skill selects it; quotes/discussion do not. Discovery, shared skill and bundled
resources are one skill. Injection, prior activation, resume/compaction, routing,
dependencies and follow-up workflows do not authorize extras. Only higher-priority
host requirements override this; injection alone grants no higher priority. Pass this
policy to delegates. Follow the [skill context rules](references/record-format.md#skill-context-notes)
for actually applied extras, reasons and injected-versus-required distinctions;
disclose unavoidable exceptions in the reply. Never claim isolation or announce
suppressed skills as active/add boilerplate notes for them.

## Run

Use the current model and ordinary host tools, without a separate model launcher,
collector, hook or persistent background service. Browser startup/temporary servers
follow [browser verification](references/browser-verification.md).
Read the [record format](references/record-format.md) and use its [template](assets/run.json).
For [settings](references/settings-resolution.md) and [usage](references/usage-resolution.md),
read shared setup/selection, boundary/provenance and current-host sections only;
delegated sections only if delegating. Skip authoring sources. Resolve resources
relative to this file. Claude native `${CLAUDE_SESSION_ID}` is an identifier only
when substituted.

1. **Set up.** Read the invocation-start clock; resolve and record `session_id`
   per the record format, locate the current-session usage source, resolve
   model/effort from runtime before configuration, and prepare a
   permitted browser check. Reserve `runs/<agent>-<model>-<effort>-<run-id>/` under
   the project/requested output root per naming rules, never overwriting. Copy
   the fixed task exactly to `prompt.md`; do not request/substitute another prompt.
   Save `run.json` as `running` with settings/source notes. Missing injected
   counters does not end lookup.
2. **Start measurement.** Immediately before implementation, capture a real clock
   and matching usage baseline with its issuing request identity in a separate
   boundary call per the usage reference. Unavailable usage must not block work.
3. **Build and check.** Save `output/index.html` and local resources under `output/`.
   Follow browser verification and record method/gaps; preserve server-free `file://`
   execution. Before delegating, read [delegated work](references/usage-resolution.md#delegated-work)
   and its current-host subsection. Finish delegated work/checks before stopping.
4. **Stop measurement.** Immediately after implementation/checks, capture the
   end clock and matching usage endpoint with its issuing request identity in a
   separate boundary call. Exclude record processing and the reply. Resolve
   interval input/output totals per the references; missing/unscoped metrics
   stay `null` with a short reason.
5. **Finalize.** Save final `run.json` now with the fixed schema, settings/source
   notes and any interruption/resume. Finished attempts are `completed` even with
   missing metrics; `failed` if the attempt cannot finish. Follow record retention limits; observations
   remain working data.
6. **Read back and reply.** Verify JSON, `session_id`, exact prompt, output path
   and known metrics.
   Reply with HTML/record paths, available timing/tokens and important verification
   gaps, outside measurement.

Use `interrupted` only for unfinished interrupted attempts. Resume the same attempt
and original measurement bounds; finished work is `completed`, unprovable metrics
null. Recoverable failures are `failed`; abrupt termination may leave `running`.
Report write failures; never fabricate artifacts or measurements.
