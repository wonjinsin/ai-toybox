# ADR 0007: Predeclare Non-Interactive CLI Experiments

Date: 2026-09-24

Status: Accepted

## Context

The user selected non-interactive CLI execution, such as `codex exec`, and clarified that each invocation must use explicitly prepared model, effort, skill/tool, and prompt settings. Output record templates alone do not establish those execution conditions.

## Decision

Prepare the benchmark prompt and explicit experiment settings before launching each non-interactive CLI run. The operator starts the run with a single initial task submission, without follow-up steering during that run.

Record requested execution inputs separately from observed results. Do not infer actual settings or capability use from the declared configuration alone.

## Consequences

- A non-interactive CLI session can still contain multiple internal model turns and permitted tool calls. It is not necessarily one model inference request.
- The prompt and selected execution conditions must be reviewable before a run starts.
- Models, effort levels, and specific skills/tools remain subject to user selection.
- The proposed input format and launcher behavior are documented for review. No launcher implementation or benchmark execution is authorized by this ADR alone.

## Supporting Material

- [Proposed execution contract](../benchmarking/execution-contract.md)
- [Experiment input template](../benchmarking/templates/experiment.json)
- [Shared task prompt draft](../benchmarking/drafts/solar-system-task.md)
