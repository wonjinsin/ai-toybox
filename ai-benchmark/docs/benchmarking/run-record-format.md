# CLI Run Record Format

This initial template prepares records for user-operated CLI benchmarks. It defines the four requested files and leaves unknown values explicit. [ADR 0007](../adr/0007-predeclare-non-interactive-cli-experiments.md) selects non-interactive execution. The [proposed execution contract](execution-contract.md) describes prepared inputs and a possible collection workflow.

## Template Layout

```text
templates/run/
  prompt.md
  response.txt
  run.json
  errors.log
```

The three text files are intentionally empty. They are templates, not an approved benchmark prompt or evidence of a completed run. The [preparation tool](cli-preparation.md) can create a new record with validated declared inputs. Live execution and full automated record finalization remain unsupported.

An archive folder is not an isolation mechanism. The execution workspace and its read/write access must be defined separately before a benchmark is run. Do not expose other runs or evaluator material to the benchmark model.

## File Contracts

| File | Contents | Writer and timing |
| --- | --- | --- |
| `prompt.md` | Exact user input submitted to the CLI, including any declared skill invocation text, with no logging instructions appended. | Operator or external collector, before execution. |
| `response.txt` | The final assistant message as captured or exported. Intermediate messages belong in an optional transcript or event log. | CLI or external collector, after a final response is available. |
| `run.json` | Declared conditions, reported conditions, execution measurements, and capture status. | Operator or external collector, before and after execution. |
| `errors.log` | Raw CLI standard error when available. It can contain progress messages and warnings as well as errors. | External capture while the CLI runs. |

Treat source artifacts as separate evidence: a final assistant message may only summarize files written by a CLI. Artifact packaging is not established by this record format.

For any future interactive sessions without a clean export, label response capture as `manual_copy`; do not imply byte-for-byte fidelity. Record any missing output. An empty template response is not a completed empty response.

## Metadata Conventions

- `format_version` identifies this record format; it is separate from the benchmark prompt version.
- Set unique `task_id` and `run_id` values before execution. Each attempt gets its own run ID, including retries after failures.
- Record the CLI name and exact version. The first benchmark uses `non_interactive` mode.
- `experiment` references a frozen copy of the declared input settings and its hash. It is separate from observations collected during execution.
- `requested` contains settings declared before execution. Record provider-native model and effort names without translating effort labels across providers.
- Capability allowlists are `null` until decided. An explicit empty array means none are permitted. Entries should identify the selected capability and its version or content hash where available.
- `reported` contains only settings corroborated by CLI or provider evidence. Missing reported values remain `null`; do not copy requested values into them as verification.
- `prompt.sha256` hashes the exact bytes of `prompt.md`. Preserve the submitted file and version together. If skill invocation text is added, also preserve the shared task prompt (`prompt.task_snapshot_path`, `prompt.task_sha256`) and the additional invocation text (`prompt.treatment_snapshot_path`, `prompt.treatment_sha256`).
- `prompt.md` captures the operator's benchmark input, not the CLI's complete effective context. Record known additional instruction sources separately. An unavailable context inventory remains `null`.
- Evidence paths in the record are relative to the archived run directory. `requested.skills[].path` preserves the original declaration and resolves relative to the source experiment file, not the archive. The original declaration is retained in `experiment.json`; selected skill bundles are not yet copied into the archive. Do not treat a prepared archive as a portable replay package.
- Do not store credentials or a full environment dump in the record.

## Measurements and Evidence

- Use UTC timestamps. Record the timing method and interval. A CLI-process duration includes startup, provider waiting, model work, and permitted tool calls; it is not pure model inference time.
- For automated timing, use an external monotonic clock for elapsed duration. For manually observed timing, identify that method. Do not compare the two as if they had equal precision.
- Preserve a reported CLI exit code. Exit code zero does not establish that the HTML works or that the run followed the benchmark conditions.
- Record only reported token counts. Keep missing token categories `null`; record usage scope and the source event or export. Do not add cached input tokens to input tokens when cached tokens are already a subset of the reported input total.
- Capture states are `not_collected`, `captured`, `partial`, or `unavailable`. Supply a reason for partial or unavailable data. A captured empty stderr stream can be valid; an untouched empty template is not evidence of error-free execution.
- CLI failures and model/API errors can appear in structured events as well as stderr. Preserve any available raw event stream alongside `errors.log` rather than relying on stderr alone.
- Keep artifact validation separate. `artifact_validation.status` starts as `not_checked`; browser/runtime errors belong in its own referenced log when checks are later performed.
- Declared capability permissions and observed usage are separate. Missing tool events do not prove that skills, instructions, or tools were absent. Record incomplete evidence explicitly and leave compliance unverified until the relevant controls and evidence have been checked.

## Collection Lifecycle

1. Prepare a new record copy and the approved prompt; fill IDs, CLI details, requested settings, explicit allowlists, and prompt identity.
2. Establish the agreed execution isolation and collection method. Keep the archive and evaluator material outside the model's allowed context.
3. Capture actual response and diagnostics without asking the benchmark model to generate its own metrics.
4. Fill observed execution values from external measurements and CLI/provider evidence. Preserve partial output if the run fails or is interrupted.
5. Finalize the archived record. Preserve the original run evidence; store later artifact validation or evaluation results separately, referenced from the record as appropriate.

An initial record has `record_status: "template"`. A prepared copy uses `prepared`; execution uses `running`; collection finishes as `finalized`. `execution.status` independently distinguishes `not_started`, `running`, `completed`, `failed`, and `interrupted`. Finalized records can describe failed or interrupted runs.

## Codex Collection Findings

Read-only inspection found local `codex-cli 0.156.1`. Its `exec --help` lists `--json`, `--output-last-message`, and prompt input from stdin.

The [official non-interactive documentation](https://learn.chatgpt.com/docs/non-interactive-mode) describes JSONL execution events, usage in completion events, and final-response export. For a non-interactive run, an additional raw `events.jsonl` file is recommended to preserve the evidence behind normalized metadata. The four core files remain usable when a CLI cannot provide structured events; mark that limitation in the record.

The [configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference) documents `model_reasoning_effort`; supported values depend on the selected model and client. The template therefore does not impose a universal effort enumeration.

The [AGENTS.md documentation](https://learn.chatgpt.com/docs/agent-configuration/agents-md) describes global and project instruction loading. Account for that context when designing isolation. Non-interactive mode and a read-only filesystem policy are not definitions of a model-only benchmark.

These findings establish capture capabilities, not an approved execution command or verified model-only configuration.
