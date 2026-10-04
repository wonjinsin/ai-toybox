---
name: benchmark-record
description: "Use when the user explicitly invokes benchmark-record to run the fixed Solar System benchmark."
disable-model-invocation: true
---

# Run the Solar System Benchmark

One invocation builds the [fixed task](assets/solar-system-prompt.md), saves its
small record, and answers the user. Finish recording before the final reply;
never require a second collection request. This skill only runs the benchmark;
analysis and comparison are outside its scope.

Use only skills explicitly invoked in the same user request. Do not load or
invoke another skill for implementation, verification, or recording merely
because it seems relevant. Use ordinary host tools instead.

Read [the record format](references/record-format.md) and use
[the template](assets/run.json). Resolve resources relative to this file.
Use the current model and normal host tools; no separate model launcher,
collector script, hook, or background process is needed.

## Run

1. Create `runs/<agent>-<model>-<effort>-<run-id>/` under the project root using
   the [directory naming rules](references/record-format.md#run-directory).
   Use a requested output root when supplied. Never overwrite another run.
   Copy the bundled task exactly to `prompt.md`.
   Use it without requesting another prompt or substituting an unrelated task.
   Save an initial `run.json` with status `running` and the known model/effort.
2. Immediately before implementation, read a real clock. If trustworthy usage
   counters for the current execution are already accessible, capture a matching
   baseline. Keep measurement observations only as working data; do not create
   extra transcript or evidence archives. Unavailable usage must not block work.
3. Build the fixed Solar System and save its entrypoint as `output/index.html`,
   with local resources under `output/`. When a browser is available, open the
   HTML via `file://` and check automatic motion, pause/resume, and speed control
   without starting a server. Otherwise, inspect scripts and assets for `file://`
   compatibility and missing local files, then note the browser verification
   gap. Do not claim browser behavior was verified from code alone.
4. Immediately after implementation and its checks finish, capture the end clock
   and any matching usage endpoint. This closes the measured work interval.
   Determine the available duration and input/output tokens using the record
   format. Do not extend the interval through record processing or the final reply.
   Missing or unscoped metrics stay `null` with a short explanation.
5. Save the final `run.json` in this same invocation. Use status `completed` when
   the implementation attempt finishes, or `failed` when it cannot finish.
   Missing metrics do not leave a finished attempt waiting for collection.
   Keep the fixed schema and concise notes; do not save a response copy, transcript,
   evidence directory, session/turn identifiers or previous-manifest backup.
6. Read back the JSON and verify the prompt copy, output path and known metrics.
   Reply directly with the HTML and record paths, available timing/tokens and any
   important verification gap. The brief final reply is outside the measurements.

On a recoverable failure or observed interruption, finalize the available record
as `failed` or `interrupted` and explain it. An abrupt termination may leave the
initial `running` record. If files cannot be saved, report that explicitly.
Do not fabricate an artifact, measurement or successful write.
