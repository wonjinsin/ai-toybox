# Small Go Runner

**Current limit: input recording works; Codex execution is blocked.** The selected
Skills/Tools restrictions cannot yet be enforced by this integration. It never
falls back to ordinary CLI settings. See [ADR 0010](../adr/0010-simplify-the-go-runner.md).

## Usage

Build with Go 1.25 or later; no dependencies are required:

```sh
go build -o bin/ai-benchmark ./cmd/ai-benchmark
./bin/ai-benchmark run experiments/selected.json
```

There is one command. To choose the new output directory, put the option before
the experiment path:

```sh
./bin/ai-benchmark run --output runs/attempt-001 experiments/selected.json
```

Copy [experiment.json](templates/experiment.json), then fill its fields:

- `task_id`, `prompt_version`: your task and prompt identifiers.
- `prompt`: a finalized UTF-8 prompt file, relative to the experiment file.
- `model`, `effort`: your explicitly selected provider-native values.
- `skills`, `tools`, `mcp_servers`, `agents`: explicit name lists; `[]` permits
  none. `null` means undecided and is rejected. A name is a declaration, not an
  installed capability or an enforcement mechanism.

Hashes are computed automatically. Unknown or duplicate settings, incomplete
declarations, and empty or draft prompts are rejected. Model/effort support is
not checked. No real configuration has been selected; the solar-system prompt
remains a draft. The previous nested experiment format is no longer accepted.

## Output

Without `--output`, a new UTC timestamped directory is created under `runs/`:

```text
runs/<timestamp>/
  prompt.md        # Exact task input bytes
  response.txt     # Empty until a supported execution path exists
  run.json         # Conditions, hashes, capture states, and preflight result
  errors.log       # Empty: no CLI stderr has been collected
  experiment.json  # Exact original settings
```

Existing directories are never reused. A blocked attempt saves its inputs and
`preflight.status: "blocked"`, returns exit code **3**, and leaves execution as
`not_started`. Empty response/log files are not evidence of a completed run.
No Skill is loaded or invoked, even when its name is declared. No skill bundle
is copied; these archives are not portable replay packages.

Exit **0** is help; **2** is invalid input or a filesystem error. There is
currently no successful live execution path. The binary does not call Codex,
an API, or a model. A future execution path requires verified restrictions;
switching to an API requires a separate user decision.

## Development

```sh
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go vet ./...
```

Tests use temporary fixture inputs and the locally built runner, with no model
calls. The binary is also exercised outside the checkout. The embedded
`run.json` template retains format `0.2`; old archives are left unchanged.
