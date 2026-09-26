# ADR 0008: Use a Standard-Library CLI Preparation Tool

Date: 2026-09-26

Status: Proposed (provisional implementation; language choice awaits confirmation).

## Context

The user requested progress on prestructured, explicitly configured CLI benchmark runs. Actual model, effort, and capability selections remain open. The installed CLI's strict isolation controls have not been verified.

## Proposed Decision

Use a small repository-local Python standard-library tool to validate declarations, preview exact inputs, and prepare per-run records. Keep real execution gated until a reviewed adapter can enforce the selected condition. Keep subprocess collection separate from experiment validation and event interpretation.

The implementation uses Python 3.9-compatible code and requires no runtime packages. This choice does not determine the comparison UI stack or generated artifact stack.

## Consequences

- Input preparation and raw collection can be exercised without paid model calls.
- A prepared archive does not imply a launched or compliant benchmark.
- The initial live execution command rejects every isolation profile.
- Exact CLI controls, runnable task details, and experiment combinations remain future decisions.
- The Python implementation can be replaced if the user selects another launcher language.

## Supporting Material

- [CLI preparation guide](../benchmarking/cli-preparation.md)
- [Execution contract](../benchmarking/execution-contract.md)
