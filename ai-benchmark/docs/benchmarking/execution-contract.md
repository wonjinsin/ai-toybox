# Execution Contract

Status: strict capability conditions retained; live execution unsupported.

[ADR 0011](../adr/0011-retain-codex-cli-with-enforced-capability-boundaries.md)
records the user's decision to keep Codex CLI and block execution until the
selected capability restrictions are verified.

## Current Workflow

One command reads an experiment file, validates the explicit declarations, and
copies the prompt and configuration into a new archive. It computes both hashes
and records the unsupported execution condition. See the [README](../../README.md).

The archive preserves the requested model, effort, prompt version, and separate
Skills/Tools/MCP/Agents lists. It does not claim that any request was sent or that
those capabilities were installed, invoked, or verified. A blocked preflight
leaves response, time, token usage, reported model, and CLI version unobserved.

## Conditions for Live Execution

- Only explicitly permitted capabilities may be available to the benchmark.
- `model only` excludes model-called tools, file reads/writes, Skills, MCP,
  helpers, and other agents. The external runner may deliver the input and
  record the response; it must not solve the task or improve the result.
- Each run needs a fresh context that cannot access other runs or evaluation
  material. An output directory alone does not provide that isolation.
- Model and effort must be passed explicitly. No automatic substitution or
  undeclared retries are permitted.
- Preserve raw final output, stderr/events, unsuccessful attempts, UTC times,
  elapsed process duration, and reported usage. Never infer missing values.
- Skill comparisons must preserve the common task and any additional invocation
  text separately. The current blocked path does not add Skill mentions.

## Codex Limitation

Read-only inspection of local `codex-cli 0.156.1` confirmed model/config flags,
stdin input, JSONL events, and final-response export. The
[official configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference)
provides individual capability controls. The
[Skill documentation](https://learn.chatgpt.com/docs/build-skills) describes
discovery from multiple locations. These findings do not establish a complete
allowlist for all built-in tools and instruction sources.

Therefore every current Codex condition is rejected, including empty lists.
Read-only mode, a prompt saying “do not use tools,” and `--ignore-user-config`
are not accepted substitutes. There is no unsafe override. API execution is
not implemented or selected.
