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
4. If still unresolved, record null with a reason, publish the early checkpoint
   and continue permitted implementation. Finalize `failed` if a required value
   remains unavailable. Do not discard HTML or leave the attempt `running`.

In one or two short notes, name sources/unverified overrides. Keep the schema;
no extra provenance fields, config copies or evidence files are needed. Resolve
`session_id` separately per the [record format](record-format.md#session-identity),
never from settings configuration.

Prefer TOML/JSON parsers. If none is available, inspect exact scalar keys with
section/scope context. Ambiguous structure is unusable. No broad search mixing unrelated/inactive
keys, whole config/environment/credential/transcript dumps, or unrelated sessions/histories.
Briefly note missing, malformed or unsupported sources and try permitted fallbacks.
Unresolved settings are recording limitations, not construction blockers.
An access/security denial stops that lookup operation under the
[scoped continuation rule](../SKILL.md#continue-safely-and-finalize).
Do not retry or bypass denied access; use only independently permitted sources.

Preserve configured identifiers/aliases exactly in JSON; note unresolved exact
runtime models, never guess versions. Empty/reset/default selectors (model
`default`, effort `auto`) are not concrete: resolve from current-run sources or
record null and the unresolved selector in notes. Never infer effort from model,
thinking budget, account tier or built-in defaults. Metadata placeholders
`unknown`, `null`, `undefined`, `unavailable`, `n/a`, `default`, `auto`,
`unspecified`, `unset`, `not available`, `not applicable`, `tbd`, `todo`, and
`placeholder` are not concrete values, regardless of case or surrounding whitespace.
Model `none` is not concrete. Effort `none` is valid only when sourced as an
explicit no-reasoning selection, never as a substitute for unknown effort.
Token availability does not affect settings resolution, but both settings and
usage are required for success. Never rename existing run directories.

## Host settings

Read only the actual current host's document: [OMP](omp.md#settings),
[Codex](codex.md#settings) or [Claude Code](claude-code.md#settings). A provider
or model named Codex does not make an OMP session a Codex host session.

## Sources

Authoring references only, not required runtime network requests:

- [Codex configuration and precedence](https://learn.chatgpt.com/docs/config-file/config-basic)
- [Codex App Server configuration RPC](https://learn.chatgpt.com/docs/app-server)
- [Claude Code model and effort configuration](https://code.claude.com/docs/en/model-config)
- [Claude Code settings precedence](https://code.claude.com/docs/en/settings)
- [Claude Code runtime and configuration environment variables](https://code.claude.com/docs/en/env-vars)
