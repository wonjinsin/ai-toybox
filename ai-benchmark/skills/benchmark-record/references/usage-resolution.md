# Resolve Input and Output Tokens Automatically

Find usage during setup, before measurement, with ordinary read-only host tools
(inline Python or jq queries allowed). Do not ask the user to collect tokens,
launch another model/app server, install a hook, or add a maintained collector.
Settings cannot supply usage. Missing injected usage alone does not end lookup.

## Source selection and scope

1. Prefer final request usage or cumulative counters already exposed for the
   current execution by host metadata or an already-connected session API.
2. Otherwise locate the host's existing current-session file as described below.
   Use an observed current thread/session identifier or an explicit runtime path.
   Validate the identifier before using it in a filename pattern; do not treat
   unresolved placeholders, path separators, or wildcard characters as an ID.
3. If neither source can be identified/read, continue with null token fields and
   a short reason. Do not substitute configured values, user estimates, account
   limits, cost conversions, or a different session's usage.

Discover filenames only for the exact current identifier or a validated linked
child identifier from [delegated work](#delegated-work). Never choose a file
because it is newest, the only file, or shares the working directory. Do not open
unrelated transcripts to find a match. Confirm identity inside the file and its
applicable project/workspace. Unexplained multiple matches or conflicting
identities are unusable. An explicitly observed current path can support a moved
workspace when runtime metadata explains the difference.

Inspect only identity, boundary/completion metadata, and usage. Do not print
message text, tool arguments, credentials, whole environments, or whole logs.
Keep file positions, request identities, and snapshots in working memory only;
the run retains no transcripts, usage archives, or session/request identifiers.
Missing, malformed, unreadable, or unsupported sources require a short note;
try another available current-run source without failing the implementation.

## Capture comparable boundaries

- Use separate tool calls for setup/start, implementation/checks, and stop.
  The start call captures the clock and baseline only; do not also implement in
  that request. Start counting after the completed model request that issued it.
  The stop call captures the clock and endpoint only; defer record processing and
  final-reply composition until afterward. Include the completed work request
  that issued the stop call. A request that also does recording cannot be split.
- At each boundary, retain a source cursor and enough request/completion metadata
  to establish which completed requests the snapshot covers. A latest counter
  that omits the request issuing that boundary call is a stale snapshot.
- If final usage for a boundary or delegated work request is flushed after the
  tool returns, allow one read-back per selected source before finalization.
  Recover only final usage
  tied to that boundary request or delegated requests proven completed within the
  original work interval; freeze the endpoint at the original stop. Never take
  a newer session total that includes the read-back, recording or later child work.
  If finality or the exact cutoff remains unprovable, affected totals stay null.
- Subtract comparable cumulative snapshots from the same session, counter scope,
  and accounting scheme, or sum deduplicated final requests wholly after the
  baseline through the stop request. A valid delta is nonnegative; never clamp a
  reset/decrease to zero. Include billed retries and attributable child requests
  once. If known child/retry usage is missing, do not label parent-only usage as
  the full work total. Unknown coverage makes the affected total null.
- Resolve input and output independently. Require nonnegative integers, excluding
  booleans. Missing/invalid usage is not zero. Ignore an incomplete trailing JSONL
  line until the read-back; malformed usage in the measured range leaves affected
  totals unknown. Preserve genuinely observed zero usage.

These are request-accounting boundaries within the
[work interval](record-format.md#measurement-boundary). A log timestamp or file
modification time alone does not establish a request's start, finality, or scope.
Time remains the original stop clock minus start clock; a read-back does not
extend it. An available session total without a matching baseline is not enough.

## Delegated work

Do not stop lookup merely because the delegation tool omits token counts. Check
the existing linked usage sources below before declaring child usage unavailable.

1. At dispatch, retain the returned child handle, concrete thread/agent identity
   or transcript path, its parent relationship, and the delegated work boundary
   in working memory. A display name or opaque handle is not necessarily a file
   identifier; require an observed mapping. Track resumed work and descendants
   the same way; ask delegates to return identity/source metadata already exposed
   to them, without loading another skill or collecting transcripts.
2. Locate only that child's existing source using the host guidance below.
   Validate identities and the dispatch relationship; shared cwd, timestamps or
   copied parent history alone do not establish a link. Inspect only structured
   identity, delegation/completion and usage metadata, never prompts or arguments.
3. Capture a baseline before sending work to an existing child, then its endpoint
   after completion. For a new or forked child, prefer final requests attributable
   to this dispatch; do not count copied history or assume its counter starts at
   zero. Include resumed requests, retries and descendants once. Finish delegated
   implementation/checks before the original stop; the bounded read-back may
   recover final usage but must not extend the interval or add later child work.
4. Establish coverage from the actual source's documented accounting scope or
   complete request identities. A parent/thread label or equality of token values
   alone does not prove whether child usage is included. Choose one calculation:
   - An interval counter proven to include all measured parent and descendant
     requests is sufficient; use it once even if separate child logs are absent.
   - Otherwise sum final requests from the linked sources, deduplicating copies
     across sources by stable provider request/response identity. Namespace IDs
     that are only locally unique. Conflicting final copies make the affected
     field unknown. Prove that the union covers every measured request.
   - Combine parent and descendant interval amounts only when their scopes are
     proven disjoint and cover the measured work. Each amount may be a comparable
     counter delta or a deduplicated final-request sum; mixing those sources is
     allowed when their boundaries and accounting match. Never add a descendant
     again when an ancestor amount already covers it.
5. If coverage remains incomplete or overlap cannot be resolved, keep only the
   affected total null. Preserve independently valid parent interval counts in
   one concise note, explicitly labeled partial with the coverage gap. Do not
   present partial counts as the full total or discard a known output merely
   because input accounting is unknown. Follow [provenance](#record-provenance).

No delegated requests means no child-coverage requirement. Do not require child
files for a proven inclusive source, or infer zero usage from a missing child file.

## Codex

- Prefer an already-connected App Server's `thread/tokenUsage/updated` for the
  current thread. Its `tokenUsage.total` is cumulative;
  `tokenUsage.last` is one request, not the whole work interval. Use
  `inputTokens` and `outputTokens` from matching snapshots, with the boundary
  checks above. Do not start `codex app-server` or `codex exec` to obtain usage.
- Otherwise read only `CODEX_THREAD_ID` and `CODEX_HOME` from the current tool
  environment. Use the host's concrete current thread identity when available;
  `CODEX_THREAD_ID` is a fallback exposed by supported local hosts. Locate a
  matching rollout filename under `<CODEX_HOME>/sessions/`, or
  `~/.codex/sessions/` when unset. Confirm the `session_meta` payload's `id` and
  applicable `cwd`; no identifier means no filename-based session lookup.
- When the installed rollout exposes `type: "token_usage_record"`, use
  `payload.usage.input_tokens` and `.output_tokens` for final request usage,
  keyed by `payload.response_id`, or comparable
  `payload.thread_token_usage.input_tokens` and `.output_tokens` snapshots.
  Confirm `payload.thread_id` matches the current thread.
  `payload.turn_token_usage` has a different scope; do not mix it with thread
  totals. Duplicate records for one response count once.
- Older rollouts may expose `type: "event_msg"` with
  `payload.type: "token_count"`. The cumulative values are
  `payload.info.total_token_usage.input_tokens` and `.output_tokens`;
  `last_token_usage` is not cumulative. Repeated counter notifications do not
  represent new requests. A missing/null `info` is unavailable, not zero.
  Verify that both snapshots include their boundary requests: these events may
  be written after tool output. If event ordering cannot prove the cutoff, do
  not subtract whichever two counters happen to be latest.
- Codex `input_tokens`/`inputTokens` already includes cached input, and
  `output_tokens`/`outputTokens` already includes reasoning. Do not add
  `cached_input_tokens`, `cache_write_input_tokens`, `reasoning_output_tokens`,
  or their camelCase breakdowns again. Use a different provider accounting
  scheme only when its source explicitly establishes disjoint categories.
- Retain child thread IDs from dispatch results or already-exposed App Server
  `collabToolCall` metadata (`senderThreadId`, `newThreadId`, `receiverThreadId`).
  Match the sender to the delegating thread and the child to this work. Use the
  child's existing usage notifications or exact-ID rollout lookup and validation
  described above. If the connected API already supports experimental
  `parentThreadId` or `ancestorThreadId` filters, use them separately to resolve
  linked descendants and include supported subagent source kinds. Do not enable
  features or start a server for collection. Inspect the installed schema, not
  guessed fields. Thread-scoped notifications alone do not prove descendant coverage.

The rollout layouts are host-version dependent. Inspect the actual selected
file's usage structure; an unsupported layout is not permission to guess fields.

## Claude Code

- Read `CLAUDE_CODE_SESSION_ID` from the current Bash/PowerShell tool environment,
  where supported, along with `CLAUDE_CONFIG_DIR`. Do not rely on an MCP server's
  inherited session ID after resume; it can retain a startup identity.
- A natively loaded Claude skill can also provide the prompt substitution
  `${CLAUDE_SESSION_ID}`. The project discovery entrypoint passes this expanded
  value into the shared workflow. A reference read from disk is not substituted;
  a literal placeholder is unusable. This substitution is not a Bash variable.
  Direct file reads can still use `CLAUDE_CODE_SESSION_ID`. An already-exposed
  hook/status-line `session_id` and `transcript_path` are other identity sources;
  do not install or reconfigure those features to obtain them.
- Prefer an explicit current `transcript_path`. Otherwise find the exact
  `<session-id>.jsonl` filename beneath `<CLAUDE_CONFIG_DIR>/projects/`, or
  `~/.claude/projects/` when unset. Directory names can be customized; do not
  assume the current path's punctuation-to-hyphen encoding. Confirm relevant
  records' `sessionId` and project metadata. Do not traverse child-agent files
  unless their relationship to this measured work is established.
- For `type: "assistant"` records, inspect `message.usage` and completed request
  identity. Deduplicate by `requestId` plus `message.id` when both exist, or by
  `message.id` when it alone identifies the response. One response can occupy
  several transcript rows; sum responses, not rows. Use its confirmed final
  usage snapshot, not partial streaming counts or the sum of copies. If copies
  differ and the final snapshot cannot be identified, affected totals stay null.
  A non-null `message.stop_reason` can support completion, but a null value or
  the presence of a `usage` object alone does not prove finality. Use documented
  host completion/tool ordering when available; otherwise note the gap.
- Normalize each final response as:

  ```text
  input = input_tokens + cache_creation_input_tokens + cache_read_input_tokens
  output = output_tokens
  ```

  These three Claude input categories are disjoint. Do not also add the nested
  `cache_creation` duration breakdown. Thinking is included in `output_tokens`;
  do not add thinking text or `output_tokens_details.thinking_tokens` again.
  Missing cache fields can mean zero only when the source's documented schema
  establishes that convention; unknown accounting leaves input null even when
  output is known. Reject inconsistent or placeholder usage rather than repairing
  it with guesses.
- Do not subtract status-line `context_window.total_input_tokens`,
  `total_output_tokens`, `current_usage`, or context percentages. Despite the
  names, these describe current context/latest-response usage, not cumulative
  session spend. `/cost` dollars, account rate limits, and a headless result
  obtained by launching another session are not work-interval usage sources.
- Retain the agent ID from the dispatch result. Prefer an already-exposed
  `agent_transcript_path` linked to that dispatch; existing SubagentStop metadata
  can supply `session_id`, `agent_id` and that path. Do not install a hook. For a
  confirmed parent transcript and agent ID, look for the exact
  `<session-id>/subagents/agent-<agent-id>.jsonl` beneath its project directory.
  Validate the parent/session and agent relationship in available metadata;
  `sessionId` alone does not distinguish parent and child response rows. Apply
  the same finality, input normalization and request deduplication to child rows.
  If this host uses a different layout, use an explicit linked path or note the
  unsupported layout; never scan neighboring agents.

## Record provenance

Keep the schema unchanged. Use one short note for successful collection, such as
`Tokens from Codex current-session final request usage; deduplicated work requests;
cache and reasoning already included.` or
`Tokens from Claude Code current-session final response usage; cache input added
once; duplicate response rows excluded.` If only one field is known, record it
and explain the other field's gap. If collection fails, name the actual missing
identity, permission, schema, finality, or boundary; do not merely say that the
prompt lacked token metadata. Finalize the record in the same invocation.

When totals are unknown but parent interval counts are valid, retain a note such
as `Partial parent work usage: input=1823140, output=9819; full totals unknown
because child usage and parent inclusion could not be verified after linked-source
lookup.` Include only known fields; never substitute zero for an unknown one.
This is a diagnostic partial measurement, not a total or a detailed breakdown.
Successful delegated collection instead names its coverage and calculation, such
as `Tokens from linked parent/child final requests; complete work coverage;
duplicate response copies excluded.` Keep identities and paths out of notes.

## Sources

These are authoring references, not required network requests during a run:

- [Codex App Server usage notifications](https://learn.chatgpt.com/docs/app-server)
- [Claude Code session environment](https://code.claude.com/docs/en/env-vars)
- [Claude Code skill session substitution](https://code.claude.com/docs/en/skills#available-string-substitutions)
- [Claude Code transcript identity](https://code.claude.com/docs/en/hooks)
- [Claude Code status-line context counters](https://code.claude.com/docs/en/statusline#context-window-fields)
- [Claude input usage and caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching#tracking-cache-performance)
- [Claude thinking output accounting](https://platform.claude.com/docs/en/build-with-claude/thinking-steering-and-cost)
