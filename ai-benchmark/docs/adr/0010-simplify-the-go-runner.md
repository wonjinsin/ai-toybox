# ADR 0010: Simplify the Go Runner

Date: 2026-09-28

Status: Accepted

## Context

The preparation tool grew into several commands and packages before a real
benchmark configuration had been selected. The user requested a small Go
runner and explicitly retained enforcement of allowed Skills/Tools rather
than accepting the normal CLI environment.

## Decision

- Keep one command: `run [--output DIRECTORY] EXPERIMENT`.
- Use a flat experiment file containing task ID, prompt path/version, model,
  effort, and explicit Skills/Tools/MCP/Agents lists.
- Compute input hashes automatically; preserve exact prompt and settings bytes.
- Keep the four requested filenames and the existing `run.json` format `0.2`.
  Archive the input settings as one additional `experiment.json` file.
- Remove separate validation, preview, preparation, and event-inspection commands,
  the generic subprocess collector, and unconnected future-adapter machinery.
- Reject unsupported execution conditions without weakening them or invoking a
  model. Save a blocked preflight with the inputs so the refusal is explicit.

This supersedes ADR 0009's command/input compatibility and collector scope.
Go, strict capability boundaries, and the run-record format remain unchanged.

## Consequences

The operator edits two input files and uses one command. Existing archive
directories are refused, and no manually maintained input hash is needed.
The previous nested experiment format requires manual conversion; no completed
benchmark archives are migrated.

The simplified tool currently records inputs only. Codex execution remains
blocked for every combination because complete capability enforcement has not
been verified. There is no runnable model-only adapter, unsafe fallback, or API
integration. Selecting an API instead of the CLI remains a user decision.

Model/effort choices, the final benchmark prompt, capability combinations,
artifact delivery, comparison UI, and evaluation remain undecided.

## Supporting Material

- [Usage](../benchmarking/cli-preparation.md)
- [Execution boundary](../benchmarking/execution-contract.md)
- [Record format](../benchmarking/run-record-format.md)
