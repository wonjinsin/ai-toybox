# Delegated Usage

Read only when delegating. Apply [shared usage rules](usage-resolution.md), Shared procedure and the current-host section.

## Shared procedure

Read this section only when delegating. Retain dispatch identities, source
mapping, parent relationship and work boundaries in memory; apply the same rules
to resumed work and descendants. Display names, shared cwd/timestamps or copied
history do not prove linkage. Missing delegation-tool counters does not end lookup.

Use only linked existing child sources. Existing children need a pre-dispatch
baseline; new/forked children need requests attributable to this dispatch, not
copied history or assumed-zero counters. Finish all delegated checks before stop.
Before dispatch, establish a permitted accounting route for linked child work.
If no route exists, do not delegate; use parent-only execution only when the task
and host permit it. Never launch a collector or assume that parent counters
include descendants.

Choose one proven complete calculation: an inclusive interval counter once;
deduplicated final parent/child request union; or disjoint interval amounts and
final-request sums. Stable provider request IDs deduplicate cross-source copies;
namespace locally unique IDs. Conflicting final copies, unknown inclusion or
coverage leave affected full totals null and require `failed`, never `completed`.
Preserve artifacts and explain the missing coverage in notes. A valid parent
subtotal is not a replacement for a full total. Parent/thread labels or equal
counts alone do not prove inclusion. No delegates means no child-coverage requirement.
Both full input and output totals must be nonnegative safe integers for success.
Apply the [record validator](record-format.md#validation-and-publication) before
publication and at read-back; it cannot establish linkage or coverage itself.

## Codex

Use the exact-ID lookup in [Codex session and tokens](codex.md#session-and-tokens).

Retain returned child thread IDs or exposed `collabToolCall` sender/new/receiver
thread IDs. Validate dispatch linkage, then use exact-ID rollout lookup above.
An already-connected API may support `parentThreadId/ancestorThreadId` filters;
do not enable features or start a server. Thread-scoped notifications alone do
not establish descendant coverage.

## Claude Code

Use the finality, normalization and deduplication rules in [Claude Code session and tokens](claude-code.md#session-and-tokens).

Prefer a dispatch-linked `agent_transcript_path`, including existing exposed
SubagentStop metadata; do not install hooks. Otherwise discover the exact
`<session-id>/subagents/agent-<agent-id>.jsonl`. Validate both parent session and
agent relationship; `sessionId` alone cannot distinguish parent/child copies.
Use the same finality, normalization and deduplication rules.

