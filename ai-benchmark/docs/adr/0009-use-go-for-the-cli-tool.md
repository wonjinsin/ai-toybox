# ADR 0009: Use Go for the CLI Tool

Date: 2026-09-28

Status: Accepted

Scope update: [ADR 0010](0010-simplify-the-go-runner.md) supersedes the command,
experiment-input compatibility, and generic collector portions of this decision.
The Go language choice and run-record format remain in effect.

## Context

The first CLI preparer was implemented in Python while its language choice remained provisional in [ADR 0008](0008-use-a-stdlib-cli-preparation-tool.md). The user explicitly selected Go when answering the language question.

## Decision

Use Go for the benchmark preparation and collection tool. Replace the provisional Python implementation and tests with Go code and tests, using the standard library without third-party dependencies.

Preserve the experiment and run JSON formats, command purposes, exact input snapshots, error handling, and refusal of unreviewed live execution conditions. Embed the canonical run template in the executable so a built binary can prepare records outside the checkout.

This decision applies to the CLI tool. The comparison UI stack, generated artifact constraints, model/effort selections, and evaluation approach remain separate decisions.

## Consequences

- Building requires a Go toolchain; the resulting executable does not require Python or Go at runtime.
- Unit and subprocess tests use Go and local fixture processes.
- The collector supports process-group cleanup on macOS and Linux; other platforms explicitly reject collection until their controls are implemented.
- The `run` command still rejects every live execution condition until a reviewed adapter can enforce it.
- Existing preparation records retain their JSON format versions. A prepared archive is still not a portable replay package or evidence of benchmark compliance.

## Supporting Material

- [Usage guide](../../README.md)
- [Execution contract](../benchmarking/execution-contract.md)
