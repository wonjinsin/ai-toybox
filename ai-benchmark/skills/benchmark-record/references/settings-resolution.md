# Resolve Model and Effort Automatically

Resolve during setup, before directory creation and measurement, using ordinary
read-only host tools. Do not ask the user to collect settings, switch
models, start another agent session, install hooks, or launch a new app server.

## Selection and provenance

Resolve `model` and `effort` separately, in this order:

1. A concrete value exposed for the current execution by host metadata, an
   already-connected session API, or a documented active-runtime environment
   variable. A generic identity such as "GPT-6" is a family, not an exact runtime
   model identifier; continue looking for a more specific source.
2. The user's explicit selection for this invocation. Label it user-declared when
   runtime verification is unavailable. It neither changes host settings nor
   overrides an observed runtime value.
3. The current host's applicable configuration, inspected automatically as below.
   Use a valid configuration-derived value even when active turn overrides cannot
   be inspected. Label it as configuration-derived, not runtime-verified.
4. `null` in JSON and `unknown` in the name only for a field with no usable value
   after these checks. State the short reason in `notes` and continue the run.

Name each value's source and unverified overrides in one or two concise `notes`.
Keep the schema; no provenance fields, configuration copies, session identifiers,
or extra evidence files are needed.

Prefer a TOML or JSON parser. If none is available, inspect only exact scalar
keys with their section/scope context; an ambiguous structure is not a usable
source. Do not derive settings from a broad text search that mixes unrelated
keys or inactive scopes. Never dump whole configuration files, process environments,
credentials, or transcripts. Do not search unrelated sessions or histories.
Missing, malformed, unreadable, or unsupported sources must be noted briefly;
continue to other available sources rather than failing the benchmark.

Preserve a configured model identifier or named alias exactly in JSON. If only an
alias is available, note that its exact runtime model is unresolved; do not map
it to a guessed version. Empty values and reset/default selectors such as model
`default` or effort `auto` are not concrete values. Resolve their actual value
from a current-run source if available; otherwise leave that field unknown and
note the unresolved selector. Never infer effort from a model name, thinking
budget, account tier, or built-in default.

## Codex

- Prefer concrete current-turn metadata. An already-connected App Server
  `config/read` can supply configuration after layering; it is still a
  configuration source, not proof that a turn has no override.
- Otherwise parse `config.toml` under the current `CODEX_HOME`, or
  `~/.codex/config.toml` when `CODEX_HOME` is unset. Read `model` and
  `model_reasoning_effort`. Inspect applicable trusted project
  `.codex/config.toml` layers from the project root toward the current directory,
  with the closest layer taking precedence over user configuration.
- Honor an active profile, startup overrides, and managed constraints when the
  current host exposes them. Follow the installed host's configuration layering;
  do not select a profile merely because a profile section exists. Inactive or
  untrusted project configuration must not override an applicable value.
- For fields still unset, check exposed cloud-managed configuration defaults,
  then the installed host's readable system configuration (on Unix,
  `/etc/codex/config.toml`). These are below user configuration; they are distinct
  from enforced managed constraints. Do not replace a higher-priority reset
  selector with a lower-priority value or guess an unexposed built-in default.
- When only user or project files can be read, use their best available values
  and note that active profile, startup, or turn overrides were not verified.

Example with no runtime metadata and only readable user defaults:

```json
{
  "model": "gpt-6.1-sol", "effort": "high",
  "notes": ["Model and effort from Codex user config (configuration-derived); active overrides were not verified."]
}
```

Prefix: `codex-gpt-6.1-sol-high-`. Token availability does not affect settings
resolution; never rename existing run directories.

## Claude Code

- Prefer concrete current-execution model metadata. For effort, inspect only
  `CLAUDE_EFFORT` in the current Bash tool environment when present: Claude Code
  exposes the active effort there when the model supports it. Treat that snapshot
  as runtime-observed; it wins over configured effort candidates.
- For configuration-derived model selection, inspect a known current-session
  `/model` choice or startup `--model` value, then `ANTHROPIC_MODEL`, then the
  applicable settings `model`. `ANTHROPIC_DEFAULT_MODEL` is a fallback only when
  no settings file sets `model`. Keep aliases such as `sonnet` verbatim and note
  that the exact runtime model was not verified.
- For configuration-derived effort, `CLAUDE_CODE_EFFORT_LEVEL` takes precedence
  over a known `--effort` or `/effort` choice and saved effort settings. Read
  applicable per-model `modelSettings` and `effortLevel` entries; respect the
  installed host's model-specific applicability and any exposed `maxEffortLevel`
  cap. Do not flatten an inactive model's entry into the active model's effort.
  If applicability cannot be established, note the gap and skip that candidate.
- Read user `settings.json` under `CLAUDE_CONFIG_DIR`, or `~/.claude` when unset,
  and project `.claude/settings.json` and `.claude/settings.local.json`. For
  scalar settings keys, precedence is managed settings, known startup
  `--settings`, project local, shared project, then user settings. Include only
  managed/startup sources exposed for this execution, and honor any known
  `--setting-sources` restriction. Current-session choices and environment
  overrides follow the field-specific rules above, not a generic merge order.
- Query only the named model/effort/location variables and relevant JSON keys.
  If startup, managed, or session-only choices are inaccessible, still use the
  best available configured values and name the unverified sources in `notes`.

Example without exact model metadata: `ANTHROPIC_MODEL=sonnet`,
`CLAUDE_CODE_EFFORT_LEVEL=high`, active `CLAUDE_EFFORT=medium`:

```json
{
  "model": "sonnet", "effort": "medium",
  "notes": [
    "Model from ANTHROPIC_MODEL (configuration-derived alias); exact runtime model and session/startup overrides were not verified.",
    "Effort from current Bash CLAUDE_EFFORT (runtime-observed)."
  ]
}
```

Prefix: `claude-code-sonnet-medium-`. Saved user/project effort never replaces
observed active effort.

## Sources

These are authoring references, not required network requests during a run:

- [Codex configuration and precedence](https://learn.chatgpt.com/docs/config-file/config-basic)
- [Codex App Server configuration RPC](https://learn.chatgpt.com/docs/app-server)
- [Claude Code model and effort configuration](https://code.claude.com/docs/en/model-config)
- [Claude Code settings precedence](https://code.claude.com/docs/en/settings)
- [Claude Code runtime and configuration environment variables](https://code.claude.com/docs/en/env-vars)
