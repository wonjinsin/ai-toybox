# ADR 0006: Prepare Records for User-Operated CLI Runs

Date: 2026-09-24

Status: Accepted

## Context

The user will execute benchmarks through the relevant model's CLI, such as Codex CLI. The user wants the prompt, response, configuration, and execution records structured before those runs begin.

## Decision

Prepare a reusable record template around four files: `prompt.md`, `response.txt`, `run.json`, and `errors.log`.

The operator initiates CLI execution. Record preparation and collection occur outside the benchmark model's generation context.

## Consequences

- Preserve the submitted prompt, captured response, declared settings, available measurements, and diagnostics with each run.
- Record the CLI name and version in addition to model and effort settings.
- Unavailable or uncollected values must remain distinguishable from measured zero values.
- The record template does not itself enforce isolation or permitted capabilities.
- [ADR 0007](0007-predeclare-non-interactive-cli-experiments.md) selects non-interactive execution with predeclared inputs. CLI-specific collection commands and isolation controls require further agreement.

## Supporting Material

- [CLI run record format](../benchmarking/run-record-format.md)
- [Run metadata template](../benchmarking/templates/run/run.json)
