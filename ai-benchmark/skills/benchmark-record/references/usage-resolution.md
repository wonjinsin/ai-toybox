# Resolve Input and Output Tokens

Locate an existing current-session source during setup, using read-only host
tools. No new model/app server, hook or maintained collector. Read only identity,
runtime settings, request boundaries/completion and usage; never print or retain
transcripts, arguments, credentials or whole environments.

## Source selection and scope

Prefer current-run final usage or an already-connected cumulative usage API.
Otherwise use the host's exact current identity or explicit runtime file path.
Validate identity, project and source scope before reading counts. Reject
unresolved placeholders, path separators/wildcards in IDs, conflicting matches
or unsupported accounting. Never select a newest/neighbor file, infer identity
from cwd, or substitute another conversation's usage. A moved workspace needs
runtime evidence linking it to this session.

Try another available current-run source after a failed lookup. Missing injected
counters, failed commands and unavailable sources are different conditions.
Use null when current-run sources or trustworthy scoped counts remain unavailable;
name the actual lookup, schema, cutoff or coverage failure.
Metrics do not block implementation or finalization.

## Capture comparable boundaries

- Use separate calls for start, work/checks and stop. Start captures only clock
  and usage; count after the completed request issuing start. Stop captures
  only clock and usage; include the completed work request issuing stop.
  Recording/read-back/reply requests are excluded. A mixed request cannot be split.
- Retain source position and completed request identity at each boundary.
  Confirm the counter includes that boundary request; a latest snapshot alone
  does not establish the cutoff. Log/file timestamps alone are insufficient.
- If final usage flushes after the tool returns, allow one read-back per selected
  source. Recover final usage tied to that boundary or linked delegated requests
  completed within the original interval. Freeze the original stop clock/cutoff;
  do not include later recording, read-back or descendant work.
- Subtract comparable cumulative counters from the same source/scope/accounting,
  or sum deduplicated final requests after start through stop. Include billed
  retries and linked work once. Require complete coverage; unknown finality,
  overlap or cutoff leaves affected totals null.
  Calculate and serialize totals in code from the identified boundary records,
  never mental arithmetic or manually transcribed totals.
- Resolve input/output independently as nonnegative integers, excluding booleans.
  Unknown/invalid values are not zero; preserve genuine zeros. Never clamp a
  counter reset or estimate from time, price, text, context occupancy or streaming.
  Ignore an incomplete trailing JSONL line until read-back; malformed measured
  usage leaves affected totals unknown.
- Normalize disjoint input cache categories and output reasoning exactly once.
  Do not add subsets already included in totals. Unknown accounting leaves the
  affected field null. A session total without a baseline is not interval usage.

Duration remains original stop minus start. Consistent clocks from separate
processes are valid; a short observed interval is not grounds to discard it.
On resume, preserve/recover the original boundaries from this source, including
interruption time; never silently restart measurement for the same attempt.

## Codex

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

## Claude Code

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

## Delegated work

Read this section only when delegating. Retain dispatch identities, source
mapping, parent relationship and work boundaries in memory; apply the same rules
to resumed work and descendants. Display names, shared cwd/timestamps or copied
history do not prove linkage. Missing delegation-tool counters does not end lookup.

Use only linked existing child sources. Existing children need a pre-dispatch
baseline; new/forked children need requests attributable to this dispatch, not
copied history or assumed-zero counters. Finish all delegated checks before stop.

Choose one proven complete calculation: an inclusive interval counter once;
deduplicated final parent/child request union; or disjoint interval amounts and
final-request sums. Stable provider request IDs deduplicate cross-source copies;
namespace locally unique IDs. Conflicting final copies, unknown inclusion or
coverage leave affected full totals null. Parent/thread labels or equal counts
alone do not prove inclusion. No delegates means no child-coverage requirement.

### Codex delegated usage

Retain returned child thread IDs or exposed `collabToolCall` sender/new/receiver
thread IDs. Validate dispatch linkage, then use exact-ID rollout lookup above.
An already-connected API may support `parentThreadId/ancestorThreadId` filters;
do not enable features or start a server. Thread-scoped notifications alone do
not establish descendant coverage.

### Claude Code delegated usage

Prefer a dispatch-linked `agent_transcript_path`, including existing exposed
SubagentStop metadata; do not install hooks. Otherwise discover the exact
`<session-id>/subagents/agent-<agent-id>.jsonl`. Validate both parent session and
agent relationship; `sessionId` alone cannot distinguish parent/child copies.
Use the same finality, normalization and deduplication rules.

## Record provenance

Keep the fixed schema. One short note names the source, scope and calculation,
for example: `Tokens from current Codex final work requests; deduplicated;
cache/reasoning already included.` On failure, name identity, permission, schema,
finality, boundary or coverage failure, not merely absent prompt metadata.
Finalize in this invocation even with null metrics.

If parent interval counts are valid but full linked coverage is unknown, keep
known fields in one note, explicitly partial: `Partial parent work usage:
input=123, output=45; full totals unknown because linked child usage unavailable.`
Never present partial as full totals or replace missing fields with zero.
Successful delegated notes name complete coverage/calculation. Keep identities
and source paths out of the saved record.

## Sources

Authoring references only, not runtime network requests:
[Codex App Server](https://learn.chatgpt.com/docs/app-server),
[Claude session environment](https://code.claude.com/docs/en/env-vars),
[Claude hooks](https://code.claude.com/docs/en/hooks),
[Claude cache accounting](https://platform.claude.com/docs/en/build-with-claude/prompt-caching).
