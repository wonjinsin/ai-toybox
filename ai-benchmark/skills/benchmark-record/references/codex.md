# Codex Settings and Usage

Apply [shared settings rules](settings-resolution.md) and [shared usage rules](usage-resolution.md) alongside this document.

## Settings

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

## Session and tokens

1. Read identity separately from settings parsing, with shell built-ins or Python
   `os`; a failed `tomllib` import does not prove missing environment variables:

   ```sh
   python3 -c 'import os,json; print(json.dumps({k:os.environ.get(k) for k in ("CODEX_THREAD_ID","CODEX_HOME")}))'
   ```

   Prefer the concrete current thread identity. On Node REPL hosts,
   `nodeRepl.requestMeta.threadId` or parsed `x-codex-turn-metadata.thread_id`
   can supply it; generic MCP `sessionId` is not a Codex thread ID.
2. Prefer an already-connected `thread/tokenUsage/updated` API: `tokenUsage.total`
   is cumulative, `last` is one request. Use comparable `inputTokens/outputTokens`.
3. Otherwise search `<CODEX_HOME>/sessions/`, or `~/.codex/sessions/` when unset.
   Actual names contain a timestamp prefix: `rollout-<timestamp>-<id>.jsonl`.
   With validated ID `id` and directory `sessions`, discover exact-ID candidates:

   ```sh
   rg --files --hidden "$sessions" -g "*$id*.jsonl"
   ```

   Confirm `session_meta.payload.id` and applicable `cwd`; no ID means no
   filename lookup. An unreadable environment permits the default directory
   only as a candidate, not proof of active CODEX_HOME.
4. For `type: token_usage_record`, validate `payload.thread_id`. Use final
   `payload.usage.input_tokens/output_tokens` keyed by `response_id`, or
   comparable `payload.thread_token_usage` snapshots. `turn_token_usage` has
   different scope; never mix it with thread totals. Duplicate responses count once.
   Capture clock, matching snapshot and issuing `response_id` together. When the
   selected rollout shows boundary-call → final-usage → matching-tool-output
   ordering, validate that linkage; otherwise use exposed request linkage and
   shared finality/read-back rules. A byte cursor captured
   inside the tool may already include the issuing request. Never choose the first
   request after that cursor, or a later lookup request, as the boundary.
   Freeze each boundary's `response_id`. A read-back selects those IDs, never the
   newest row, which may include the read-back request itself. In one read-back,
   retrieve both IDs from the full validated source, not a recent tail. Omitted
   tool output is not missing source usage. Calculate in code:

   ```python
   by_response = {}
   for row in validated_rows:
       if row.get('type') != 'token_usage_record':
           continue
       p = row['payload']
       if p['thread_id'] == current_thread_id and p['response_id'] in (start_response_id, stop_response_id):
           assert p['response_id'] not in by_response or by_response[p['response_id']] == p
           by_response[p['response_id']] = p
   start = by_response[start_response_id]['thread_token_usage']
   stop = by_response[stop_response_id]['thread_token_usage']
   totals = {k: stop[k] - start[k] for k in ('input_tokens', 'output_tokens')}
   ```

   Here `validated_rows` supplies final records from the selected source;
   require nonnegative integer deltas before saving. Missing boundary IDs remain
   null only after this exact lookup fails.
5. Older `event_msg` / `payload.type: token_count` uses
   `payload.info.total_token_usage`; `last_token_usage` is not cumulative.
   Repeated notifications are not requests and missing `info` is not zero.
   Confirm boundary-request inclusion even when events follow tool output.
6. Codex input already includes cached input; output includes reasoning.
   Do not add cache/reasoning breakdowns again. Inspect actual host layout;
   unsupported schemas justify null, not guessed fields or a new app server.

