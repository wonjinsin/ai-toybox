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
Before optional dispatch, establish a permitted accounting route for linked child
work. If none exists, prefer parent-only execution when the task and host permit.
If the user/host requires delegation despite unavailable accounting, perform the
permitted work and record the coverage limitation; final totals remain null and
the terminal record is `failed`. Never launch a collector or assume parent
counters include descendants. Missing child usage after dispatch does not cancel
construction or independent permitted checks. A denial stops only affected
operations under the [scoped continuation rule](../SKILL.md#continue-safely-and-finalize).

Choose one proven complete calculation: an inclusive interval counter once;
deduplicated final parent/child request union; or disjoint interval amounts and
final-request sums. Stable provider request IDs deduplicate cross-source copies;
namespace locally unique IDs. Conflicting final copies, unknown inclusion or
coverage leave affected full totals null and require `failed`, never `completed`.
Preserve artifacts and explain the missing coverage in notes. A valid parent
subtotal is not a replacement for a full total. Parent/thread labels or equal
counts alone do not prove inclusion. No delegates means no child-coverage requirement.
Both full input and output totals must be nonnegative safe integers for success.
Use the [validated atomic writer](record-format.md#validation-and-publication)
for the early checkpoint and same-invocation terminal record. It cannot establish
linkage or coverage itself.

## OMP

Use [OMP session and tokens](omp.md#session-and-tokens). Retain explicit dispatch
and parent/child identities from the current OMP tool result or runtime metadata.
Read only sources explicitly linked to those children and this dispatch. Agent
display names and shared workspace alone do not establish billing linkage.
Do not infer a Codex thread ID or search Codex/Claude histories for OMP children.
The parent journal includes child usage only when the source contract proves it.
Apply shared finality, deduplication and interval rules to the complete linked
request set. Missing child sources leave full totals null, not partial subtotals.

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

