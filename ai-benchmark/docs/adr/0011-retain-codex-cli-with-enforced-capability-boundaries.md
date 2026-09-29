# ADR 0011: Retain Codex CLI with Enforced Capability Boundaries

Date: 2026-09-29

Status: Accepted

## Context

After simplifying the Go runner, the user chose to retain Codex CLI and block
execution until capability restrictions are verified. API-based model-only
execution was offered as an alternative and was not selected.

## Decision

- Keep Codex CLI as the intended execution interface for the first integration.
- Require enforcement of the explicitly allowed Skills, Tools, MCP servers,
  and agents before enabling an execution condition.
- Preserve the current execution block while that enforcement remains unverified.
- Do not substitute ordinary CLI settings or an API call for the selected condition.

## Consequences

The current command continues to save valid inputs and a blocked preflight,
then exits with code 3 without starting a model. This decision does not claim
that Codex restrictions have been verified or that live execution is supported.

The runner remains small. Any future live-execution change must establish the
controls for its supported conditions and preserve the benchmark's isolation
and evidence requirements.

## Supporting Material

- [ADR 0010: Simplify the Go Runner](0010-simplify-the-go-runner.md)
- [Execution contract](../benchmarking/execution-contract.md)
