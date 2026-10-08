# Claude Code Settings and Usage

Apply [shared settings rules](settings-resolution.md) and [shared usage rules](usage-resolution.md) alongside this document.

## Settings

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

## Session and tokens

- Read `CLAUDE_CODE_SESSION_ID` and `CLAUDE_CONFIG_DIR` from the current shell.
  An MCP server's inherited ID may be stale after resume. Native skill
  substitution `${CLAUDE_SESSION_ID}` or exposed hook/status-line session/path
  metadata also works; literal unexpanded placeholders do not.
- Prefer an explicit current `transcript_path`. Otherwise discover the exact
  `<id>.jsonl` below `<CLAUDE_CONFIG_DIR>/projects/` or `~/.claude/projects/`.
  Directory names vary; do not derive them from cwd punctuation. Validate
  `sessionId` and workspace; do not traverse unrelated agent files.
- Use completed `type: assistant` rows' `message.usage`. Deduplicate by
  `requestId` plus `message.id`, or `message.id` when sufficient. One response
  may span several rows. Use its identified final snapshot, not streaming or
  summed copies. Conflicting copies require proven finality; a non-null
  `stop_reason` helps, but a usage object alone does not prove completion.
- Claude input is `input_tokens + cache_creation_input_tokens +
  cache_read_input_tokens`. These are disjoint. Output is `output_tokens`,
  already including thinking. Do not also add nested cache durations or thinking
  text. Missing cache categories mean zero only when the schema establishes it.
- Status-line context-window totals/current usage, /cost, limits and a headless
  result from another session are not work-interval spend.

