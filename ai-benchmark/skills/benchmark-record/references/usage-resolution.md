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

## Host lookup

Read only the current host's document: [Codex](codex.md#session-and-tokens) or [Claude Code](claude-code.md#session-and-tokens).

## Delegated work

Only when delegating, read Shared procedure and the current-host section of [delegated usage](delegated-usage.md).

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
