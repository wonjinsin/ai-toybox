# AI Benchmark

Compare artifacts produced from the same prompt across models, reasoning effort
levels, and explicitly allowed Skills/Tools. Each attempt has its own records.
Artifact comparison and independent LLM evaluation are planned.

**Current status: the Go tool saves inputs, but does not execute Codex yet.**
Every valid configuration is blocked until its capability restrictions can be
enforced. This is the agreed [Codex CLI policy](docs/adr/0011-retain-codex-cli-with-enforced-capability-boundaries.md).

## Quick Start

Run these commands from the `ai-benchmark` directory. Go **1.25 or later** is
required to build; there are no third-party dependencies. The current input-only
workflow does not require a Codex installation or account.

### 1. Build

```sh
go build -o bin/ai-benchmark ./cmd/ai-benchmark
./bin/ai-benchmark --help
```

### 2. Prepare the two input files

```sh
mkdir -p experiments/first
cp -n docs/benchmarking/templates/experiment.json experiments/first/experiment.json
```

Create `experiments/first/prompt.md` with the finalized benchmark prompt. Keep
the same task requirements across the configurations you want to compare.
The checked-in [solar-system task](docs/benchmarking/drafts/solar-system-task.md)
is still a draft; the tool rejects empty prompts and prompts marked `Status: Draft`.
The task must be agreed and finalized before it can be used.

Edit `experiments/first/experiment.json`. Example shape:

```json
{
  "task_id": "task-001",
  "prompt": "prompt.md",
  "prompt_version": "v1",
  "model": "<your-selected-model-id>",
  "effort": "<your-selected-effort>",
  "skills": [],
  "tools": [],
  "mcp_servers": [],
  "agents": []
}
```

Replace the model and effort placeholders with your chosen provider-native
values. No real model or effort has been selected for the project yet; the tool
does not verify their provider support. The empty capability lists above are
an example declaration permitting none, not an approved benchmark configuration.

- `prompt` is relative to the actual experiment file, resolving a symlink first.
- `prompt_version` is your prompt revision identifier. Hashes are calculated automatically.
- Each capability list must be explicit: `[]` permits none; `null` means undecided
  and is rejected. Names alone do not install, enable, or enforce capabilities.
- Unknown or duplicate settings and incomplete declarations are rejected.

### 3. Save an attempt

```sh
./bin/ai-benchmark run experiments/first/experiment.json
```

The command creates `runs/<UTC-timestamp>/`. To choose a directory, put
`--output` before the experiment path:

```sh
./bin/ai-benchmark run --output runs/attempt-001 experiments/first/experiment.json
```

Use a new directory for every attempt. Existing directories are refused, even
when empty. The built binary embeds its record template and works outside this
checkout. Use the binary to observe its exit code directly; `go run` may wrap
a nonzero program exit code.

### 4. Read the saved record

For valid inputs, the current command prints the archive path and a `blocked:`
message, then exits with code **3**. This is expected: no model was started.

```text
runs/<UTC-timestamp>/
  prompt.md        # Exact input bytes
  response.txt     # Empty; response not collected
  run.json         # Requested conditions, hashes, and blocked preflight
  errors.log       # Empty; CLI stderr not collected
  experiment.json  # Exact original settings
```

In `run.json`, check:

- `preflight.status`: `blocked`, with a reason.
- `execution.status`: `not_started`.
- Reported model/effort, CLI version, timing, and token usage: `null`.

An empty response or log is not evidence of a completed run. Requested settings
are recorded separately from observations. Skill bundles are not copied or
loaded. The archived settings preserve their original paths; the archive is
not a portable replay package.

| Exit code | Meaning |
| --- | --- |
| `0` | Help displayed. |
| `2` | Invalid input or filesystem error. |
| `3` | Valid inputs saved; execution blocked by the capability policy. |

There is currently no successful live execution path. The next execution
milestone is verified enforcement of the selected Codex capability restrictions
and run isolation. Comparison UI and evaluation follow separately.

## Project Layout

```text
cmd/ai-benchmark/             # Command entry point
internal/runner/             # Input validation and record creation
assets.go                    # Embedded run.json template
docs/adr/                    # Accepted and historical design decisions
docs/benchmarking/           # Execution contract, record format, draft, templates
```

`bin/` and coverage files are generated locally and ignored by Git.
`experiments/` and `runs/` are created when preparing and recording attempts.

## Development

```sh
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go vet ./...
```

Tests use temporary fixtures and the locally built runner. They do not call a
benchmark model or external service. The original nested experiment format is
no longer accepted; existing run records keep format version `0.2`.

## Details

- [Execution contract](docs/benchmarking/execution-contract.md)
- [Run record format](docs/benchmarking/run-record-format.md)
- [Go runner scope](docs/adr/0010-simplify-the-go-runner.md)
- [Codex CLI decision](docs/adr/0011-retain-codex-cli-with-enforced-capability-boundaries.md)
