# Proposed CLI Execution Contract

Status: Contract under development; preparation implemented, live execution unsupported.

[ADR 0007](../adr/0007-predeclare-non-interactive-cli-experiments.md) establishes non-interactive execution with a prepared prompt and explicit conditions. This document describes the intended launch contract. The [preparation guide](cli-preparation.md) distinguishes implemented behavior from future adapter requirements. There is no verified isolation configuration or live execution adapter yet.

## Inputs and Results

| Item | Purpose |
| --- | --- |
| Shared task prompt | The same approved task requirements for every compared configuration. |
| `experiment.json` | Predeclared model, effort, CLI version, capabilities, task prompt identity, and artifact delivery mode. |
| Per-run `prompt.md` | Exact user input delivered to the CLI, including any explicitly declared skill invocation text. |
| Per-run `run.json` | A snapshot reference to the experiment inputs plus independently collected execution observations. |
| Response, diagnostics, events, and artifacts | Evidence collected from that invocation, with capture status and provenance. |

The operator selects one experiment input. A deterministic launcher would validate it, prepare the allowed environment, invoke the selected CLI, and collect evidence. It would perform no model inference, task solving, artifact repair, or undeclared retries. Deterministic input preparation does not guarantee identical model outputs on repeated runs.

Direct CLI commands remain possible, but the operator would need to perform the same configuration and capture steps manually. A small configuration-driven preparer uses Python provisionally; see [ADR 0008](../adr/0008-use-a-stdlib-cli-preparation-tool.md). This does not establish the safety or fairness of a direct CLI invocation.

## Input Template

Use [the experiment template](templates/experiment.json) to review the proposed fields. It deliberately starts with `status: "draft"` and unselected values. An execution adapter must reject that state rather than use CLI defaults.

- `task` identifies the approved prompt file, version, and exact-byte SHA-256 hash.
- `cli` identifies the CLI, exact required version, and non-interactive mode.
- `model` and `effort` use exact provider-native values explicitly selected by the user.
- `capabilities` contains separate allowlists for skills, tools, MCP servers, and other agents. `null` is unselected; `[]` explicitly permits none.
- `artifact.delivery` is selected as `final_response` or `workspace_files` when the output contract is agreed. File writing by the model requires an explicitly allowed tool.
- `isolation_profile` identifies the reviewed execution controls. A profile name alone does not establish isolation.

Skill entries should identify their source directory, name, pinned revision or complete content digest, and invocation policy. Freeze the selected skill's referenced resources as well as its main instructions. Tool and MCP entries must identify the permitted callable surface; do not silently expand it to an entire integration.

## Proposed Launch Procedure

1. Validate the selected configuration. Reject missing fields, a draft or empty prompt, hash/version mismatches, unsupported model/effort combinations, unavailable selected skills, and unimplemented capability controls. Where support cannot be established before launch, preserve any CLI rejection as a failed attempt; do not substitute another model or effort.
2. Produce a reviewable resolved configuration and exact delivered prompt. A preparation preview may show the intended invocation without calling the model; it must distinguish controls verified locally from behavior requiring execution evidence.
3. Prepare a fresh session and isolated execution workspace. Apply the agreed instruction, skill, tool, MCP, agent, plugin, hook, memory, filesystem, and network controls. The exact mechanism requires a separate implementation decision.
4. Freeze a per-run copy of the experiment settings and the approved shared prompt. Hash and preserve the actual delivered prompt separately if it contains declared skill invocation text.
5. Start the CLI once, supplying model and effort explicitly and sending the prepared prompt through stdin. Capture the CLI's emitted events, stderr, final response, exit code, and externally measured elapsed time.
6. Populate observed fields only from evidence. Record unsupported or unavailable measurements as such. Keep unsuccessful and interrupted attempts under their own run IDs.
7. Preserve source artifacts and evidence. Run later artifact checks and evaluation under a separate stage, without feeding their results back into the recorded generation run.

## Skill Activation and Prompt Fairness

Keep the task requirements identical. For a skill treatment, preserve and hash three inputs separately: the shared task prompt, the additional invocation text, and the combined user input actually delivered. Do not add unrelated hints or workflow instructions to the shared task prompt.

Distinguish a skill being selected, available, explicitly invoked, and observed as used. Selection alone is not proof of use, and logs may provide incomplete evidence. If a selected skill needs an unapproved tool or agent, stop preparation and resolve the mismatch with the user.

## Model-Only Boundary

The user's model-only definition excludes all model-called tools, including shell commands, file reads/writes, browser tools, MCP, helpers, and other agents. In that condition, the proposed delivery mode is HTML text in the final response, which an external collector saves without rewriting it.

The model does not read the prompt file with a tool; the external launcher supplies its contents. External recording must not provide task-solving assistance. A run may be called compliant only after the relevant capability and context controls have been checked; a directory change or a read-only sandbox is insufficient.

If the installed CLI cannot enforce the selected condition, stop with an unsupported-condition result. Do not silently relabel a tool-enabled CLI run as model-only or change the requested experiment.

## Codex Mapping and Limitations

Local `codex exec --help` confirms `--model`, `--config`, stdin prompt input, `--json`, and `--output-last-message`. The [official command reference](https://learn.chatgpt.com/docs/developer-commands) describes command-line configuration overrides. The [configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference) documents `model_reasoning_effort`.

The [skills documentation](https://learn.chatgpt.com/docs/build-skills) describes explicit skill mentions, implicit activation, and skill discovery from multiple scopes. Therefore, writing a skill name in experiment metadata is not enough to configure its availability or prove its use.

There is no verified model-only launch recipe in this repository yet. Existing configuration, project/global instructions, implicit skills, and tool surfaces must be accounted for before an execution adapter can claim to enforce the experiment contract.
