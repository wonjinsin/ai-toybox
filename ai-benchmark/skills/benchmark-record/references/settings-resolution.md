# Resolve Model and Effort Automatically

During setup, before directory creation/measurement, resolve with ordinary
read-only host tools. Do not ask users to collect settings, switch models, start
another agent session, install hooks or launch an app server.

## Selection and provenance

Resolve `model` and `effort` separately, in order:

1. Concrete current-execution host metadata, connected session API or documented
   active-runtime environment variable. Family labels such as "GPT-6" are not
   exact model IDs; continue looking.
2. Explicit user selection for this invocation, labeled user-declared if runtime
   verification is unavailable. It does not change settings or override runtime.
3. Automatically inspected applicable host configuration, labeled
   configuration-derived. Use valid values even if active overrides are inaccessible.
4. If still unresolved, null in JSON and `unknown` in names; note why and continue.

In one or two short notes, name sources/unverified overrides. Keep the schema;
no provenance fields, config copies, session IDs or extra evidence files are needed.

Prefer TOML/JSON parsers. If none is available, inspect exact scalar keys with
section/scope context. Ambiguous structure is unusable. No broad search mixing unrelated/inactive
keys, whole config/environment/credential/transcript dumps, or unrelated sessions/histories.
Briefly note missing, malformed, unreadable or unsupported sources; try others
without failing the benchmark.

Preserve configured identifiers/aliases exactly in JSON; note unresolved exact
runtime models, never guess versions. Empty/reset/default selectors (model
`default`, effort `auto`) are not concrete: resolve from current-run sources or
leave unknown and note the unresolved selector. Never infer effort from model, thinking budget,
account tier or built-in defaults. Token availability does not affect settings;
never rename existing run directories.

## Codex

- Prefer current-turn metadata. Connected App Server `config/read` supplies layered
  configuration, not proof of absent overrides. Locate the validated current
  rollout per usage resolution before TOML fallback; read active/latest
  `turn_context.payload.model` and `.effort` even if initial environment metadata
  lacked them. Never use another session's settings.
- Otherwise parse `model`/`model_reasoning_effort` in current `CODEX_HOME/config.toml`,
  or `~/.codex/config.toml` when unset. Inspect applicable trusted project
  `.codex/config.toml` layers from project root toward cwd; closest overrides user.
- Honor exposed active profiles, startup overrides, managed constraints and host
  layering. Profile existence does not activate it; inactive/untrusted scopes
  cannot override applicable values.
- For unset fields, check exposed cloud-managed defaults, then readable host system
  configuration (`/etc/codex/config.toml` on Unix). Both are below user settings,
  distinct from enforced constraints. Never replace a higher-priority reset with
  a lower value or guess unexposed built-in defaults.
- If only user/project files are readable, use their best values and note unverified
  active-profile, startup and turn overrides.

## Claude Code

- Prefer concrete current-execution model metadata. Read only `CLAUDE_EFFORT` for
  runtime effort in current Bash when present; it exposes active supported-model
  effort, is runtime-observed and overrides configured candidates.
- Configured model precedence: known current `/model` or startup `--model`, then
  `ANTHROPIC_MODEL`, then applicable settings `model`. `ANTHROPIC_DEFAULT_MODEL`
  applies only if no settings file sets model. Preserve aliases such as `sonnet`
  and note unresolved exact runtime models.
- Configured effort: `CLAUDE_CODE_EFFORT_LEVEL` overrides known `--effort`/`/effort`
  and saved settings. Read applicable per-model `modelSettings`/`effortLevel` and
  respect exposed `maxEffortLevel` caps; skip uncertain applicability with a note. Never use
  inactive-model entries for the active model.
- Read user `settings.json` under `CLAUDE_CONFIG_DIR` or `~/.claude` when unset,
  plus project `.claude/settings.json` and `.claude/settings.local.json`.
  Scalar precedence: managed, known startup `--settings`, project local, shared
  project, user. Include only exposed managed/startup sources; honor known
  `--setting-sources` restrictions. Session/environment overrides follow the
  field-specific rules above, not a generic merge.
- Query only named model/effort/location variables and relevant JSON keys. If
  startup, managed or session-only choices are inaccessible, use best configured
  values and note gaps. Saved effort never replaces observed active effort.

## Sources

Authoring references only, not required runtime network requests:

- [Codex configuration and precedence](https://learn.chatgpt.com/docs/config-file/config-basic)
- [Codex App Server configuration RPC](https://learn.chatgpt.com/docs/app-server)
- [Claude Code model and effort configuration](https://code.claude.com/docs/en/model-config)
- [Claude Code settings precedence](https://code.claude.com/docs/en/settings)
- [Claude Code runtime and configuration environment variables](https://code.claude.com/docs/en/env-vars)
