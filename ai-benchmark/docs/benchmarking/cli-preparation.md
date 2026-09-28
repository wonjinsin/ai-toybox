# CLI Preparation and Evidence Collection

## Current Scope

The repository provides a Go standard-library preparation tool, selected in [ADR 0009](../adr/0009-use-go-for-the-cli-tool.md). Build with Go 1.25 or later from the `ai-benchmark` directory. There are no third-party dependencies. The built executable embeds its record template and does not require the checkout, Go, or Python at runtime.

```sh
go build -o bin/ai-benchmark ./cmd/ai-benchmark
./bin/ai-benchmark --help
```

For local development, `go run ./cmd/ai-benchmark --help` is also available. Use the compiled binary when collecting exit codes; `go run` can wrap a program's nonzero exit code.

Implemented commands validate declared inputs, preview the intended prompt and arguments, and create a new archive. They do not invoke Codex. All `run` requests currently return exit code 3 because no reviewed isolation adapter is installed. This restriction applies to every capability configuration, including empty allowlists.

The Go choice applies to this tool; it does not select the comparison UI or artifact technology. The task prompt remains a draft, and no model, effort level, skill, or tool has been selected for a real benchmark.

## Prepare an Experiment

1. Copy [experiment.json](templates/experiment.json) to your experiment directory.
2. Put the approved shared task in a UTF-8 file. Resolve `task.prompt_path` relative to the experiment file. The checked-in solar-system draft is not executable.
3. Fill every field. Use exact model and effort identifiers, pin the CLI version, and use explicit capability arrays. `null` means undecided; `[]` permits none. Only `artifact.delivery: "final_response"` is supported by this preparer.
4. Compute the prompt's SHA-256 over its exact bytes, for example with `shasum -a 256 path/to/prompt.md`. Record it in `task.prompt_sha256`.
5. Set `status` to `ready` only after selecting and reviewing the inputs. Here, `ready` means the input declaration is complete; it does not certify isolation or model support.

Run these commands with your own paths:

```sh
./bin/ai-benchmark validate experiments/selected.json
./bin/ai-benchmark preview experiments/selected.json --output runs/attempt-001
./bin/ai-benchmark prepare experiments/selected.json --output runs/attempt-001
```

`preview` returns JSON, including `launch_ready: false`, blockers, a candidate argument array, and the intended user input. The candidate arguments are incomplete execution controls, not a command approved for benchmark use. `validate` checks input shape and content hashes; it does not probe a model or the installed CLI version.

`prepare` creates the destination exclusively. Existing directories are refused, even if empty. Use a new directory for every attempt; never reuse a prepared archive as an execution workspace.

```text
runs/attempt-001/
  experiment.json   # Exact original declaration
  task-prompt.md    # Exact shared task bytes
  treatment.md      # Explicit skill mentions, when selected
  prompt.md         # Intended complete user input
  preview.json      # Candidate arguments and unresolved controls
  response.txt      # Empty; not collected
  errors.log        # Empty; not collected
  run.json          # Prepared record; execution not started
```

Archived experiment paths preserve the original declaration and remain relative to its original location; they are not automatically rebased for replay. `run.json` references the archived prompt snapshots with archive-relative paths. Skill source bundles are hashed but not copied. Full portable replay requires a future adapter to preserve those resources and its configuration evidence.

The record stores requested settings separately from observations. Observed CLI version, reported model/effort, token counts, timing, and compliance remain unknown. Empty response/error files do not imply successful execution. Preparation neither submits the prompt nor validates the resulting artifact.

## Inspect Captured Events

```sh
./bin/ai-benchmark inspect path/to/events.jsonl
```

This read-only command emits a source path/hash and a summary. It never modifies the source or updates a run record. Files up to 64 MiB are supported. Summary references such as `events.jsonl:3` identify a line in the input identified by the outer `source` object.

Only a single valid `turn.completed` event supplies usage. Known nonnegative integer counts are retained without float conversion, missing categories remain `null`, and cached input is not added to input totals. Multiple completion events have unknown aggregation scope, so the summary does not add their counts. Malformed JSONL, duplicate keys, invalid UTF-8, unpaired Unicode surrogate escapes, excessive nesting, and truncated JSON produce diagnostics and unknown usage. JSON nesting is limited to 1,000 levels. A parseable last line without a newline is accepted; this does not prove transport completeness.

Error-event messages are retained as diagnostics. Unknown event fields are left in the original log. The summary never infers model or effort, verifies compliance, or establishes that an artifact works. Exit code 0 means inspection completed; check the returned diagnostics and capture status for evidence quality.

## Skill Declarations

A selected skill entry has exactly these fields:

```json
{
  "name": "selected-skill",
  "path": "skills/selected-skill",
  "sha256": "<complete bundle digest>",
  "invocation": "explicit"
}
```

The directory must contain `SKILL.md`. The digest covers every regular file recursively, ordered lexicographically by path components: append the UTF-8 POSIX relative path, a NUL byte, then the file's binary SHA-256 digest to a bundle SHA-256 calculation. Symbolic links are rejected. The Go helper `internal/experiment.SkillDigest` implements this calculation, retaining the earlier digest format.

The preview prefixes `$selected-skill` to the common task, separated by two newlines. Multiple mentions retain the declaration order, one per line. Shared task bytes, treatment bytes, and combined prompt bytes have separate hashes. This is a proposed invocation treatment; it does not establish CLI discovery, activation, frontmatter compatibility, resource availability, or actual use. Referenced resources outside the bundle need separate pinning in a future adapter.

## Execution Boundary

```sh
./bin/ai-benchmark run experiments/selected.json --output runs/attempt-002
```

This currently refuses execution before creating output. A profile name is not an implemented isolation policy. No override or automatic fallback is provided.

`internal/collector.Execute` is an internal subprocess collection primitive for a future trusted adapter. It is tested with local Go fixture processes only. It preserves raw stdin/stdout/stderr evidence, final-response availability, exit code, UTC timestamps, monotonic duration, timeout, and interruption status. It accepts a cancellation context; cancellation is recorded as `interrupted` with reason `context_cancelled`. On macOS and Linux, cleanup targets the original process group on every exit path; leftover children mark the attempt failed even after exit code 0. Other platforms explicitly reject collection. Children that escape the group require separate isolation controls. Its record is a partial collection record, not a complete benchmark manifest; the adapter must combine it with prepared metadata and evidence interpretation.

The primitive does not isolate capabilities, sanitize inherited environment/configuration, verify CLI identity, or hide archives from tools. Separate directories alone provide no access control. Do not call it directly to claim a compliant benchmark. Before live execution, agree and implement the model/effort support checks, instruction inventory, skill/tool/MCP/agent controls, isolated filesystem/network, and resource snapshots.

## Errors and Checks

CLI exit codes: `0` for successful preparation or inspection, `2` for invalid input or filesystem errors, `3` for unsupported execution conditions. A successful preparation exit is not a successful benchmark.

Run the local unit tests:

```sh
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go vet ./...
```

Tests use fixture prompts and fake CLI processes. They do not execute a benchmark model, call an external service, build an artifact, or evaluate the solar-system task.

The entry point is also exercised as a compiled binary from outside the checkout. Coverage reports statement coverage, not branch coverage or evidence that a real Codex adapter works. The former Python implementation and tests have been replaced; record format versions remain unchanged.

Validation on 2026-09-28 with Go 1.25.1 on macOS: race-enabled tests, `go vet`, and the binary build passed. Overall statement coverage was 90.9%. The entry-point package reports 0% because its test executes a separately compiled binary; the functional packages report 85.0–98.7%, and the embedded template package reports 100%.
