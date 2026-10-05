---
name: benchmark-record
description: "Use when the user explicitly invokes benchmark-record to run the fixed Solar System benchmark."
disable-model-invocation: true
---

# Run the Solar System Benchmark

Build the [fixed task](assets/solar-system-prompt.md), save its small record, and
reply in one invocation. Finish recording before replying; never require a second
collection request. Analysis and comparison are out of scope.

Select and apply only skills explicitly invoked in the same user request. A request
to read this skill and run it also selects `benchmark-record`; quoted examples or
discussion do not select other skills.
The discovery entrypoint, this shared skill, and bundled resources are one skill.

Do not apply unselected skills, even if already injected by the host or hooks.
Keep this restriction after resume or compaction; prior activation, automatic
routing, dependencies, and follow-up workflows do not authorize another skill.
Do not load or invoke another skill merely because it seems relevant to
implementation, verification, or recording. Use ordinary host tools and pass the
same restriction to delegated benchmark work.

Follow actual instruction priority; hook injection alone grants no higher priority.
Apply an unselected skill only when a higher-priority host instruction requires it.
Record actually applied extras and reasons in `run.json.notes`, following the
record format; distinguish mandatory injected instructions from newly required
skill loads. Disclose unavoidable exceptions in the final reply; do not claim
isolation. Do not announce suppressed skills as active or add boilerplate notes
for them.

Read [the record format](references/record-format.md) and use
[the template](assets/run.json). Resolve resources relative to this file.
Use the current model and normal host tools; no separate model launcher,
collector script, hook, or background process is needed.

Claude Code session context, when substituted by its native skill loader:
`${CLAUDE_SESSION_ID}`. An unexpanded placeholder is not a session identifier.

## Run

1. Read the real invocation-start clock, then automatically
   [resolve model and effort](references/settings-resolution.md) before reserving
   a directory. Inspect available current-run sources and the current host's
   configuration; missing injected metadata alone is not a reason to stop lookup.
   Resolve fields independently and retain short source notes. Create
   `runs/<agent>-<model>-<effort>-<run-id>/` under the project root using
   the [directory naming rules](references/record-format.md#run-directory).
   Honor a requested output root. Never overwrite another run.
   Copy the bundled task exactly to `prompt.md`; do not request another prompt
   or substitute an unrelated task.
   Save an initial `run.json` with status `running`, the resolved model/effort,
   and source notes; retain them in the final record. Automatically
   [resolve a current-session usage source](references/usage-resolution.md);
   missing injected counters alone is not a reason to leave tokens unknown.
2. Immediately before implementation, read a real clock and capture a matching
   usage baseline from the resolved source. Follow the usage reference's boundary
   rules; keep observations as working data only, without transcript or evidence
   archives. Unavailable usage must not block work.
3. Build the fixed Solar System and save its entrypoint as `output/index.html`,
   with local resources under `output/`. When a browser is available, open the
   HTML via `file://` and check automatic motion, pause/resume, and speed control
   without starting a server. Otherwise, inspect scripts and assets for `file://`
   compatibility and missing local files, then note the browser verification
   gap. Do not claim browser behavior was verified from code alone.
4. Immediately after implementation and its checks finish, capture the end clock
   and matching usage endpoint, closing the measured work interval.
   Resolve input/output totals using the usage reference and record format.
   Exclude record processing and the final reply from the interval.
   Missing or unscoped metrics stay `null` with a short explanation.
5. Save the final `run.json` in this same invocation. Use status `completed` when
   the implementation attempt finishes, or `failed` when it cannot finish.
   Finalize finished attempts even with missing metrics.
   Keep the fixed schema and concise notes; do not save a response copy, transcript,
   evidence directory, session/turn identifiers or previous-manifest backup.
6. Read back the JSON and verify the prompt copy, output path and known metrics.
   Reply directly with the HTML and record paths, available timing/tokens and any
   important verification gap. The brief final reply is outside the measurements.

On a recoverable failure or observed interruption, finalize the available record
as `failed` or `interrupted` and explain it. An abrupt termination may leave the
initial `running` record. If files cannot be saved, report that explicitly.
Do not fabricate an artifact, measurement or successful write.
