# Benchmark task and scoring revision

## Why the old total is easy to misread

The saved [Terra low review](../runs/v1/codex-gpt-5.6-terra-low-20261016-230825/reviews/review.md)
reports 82/100 while tracking is a label-only state and slow camera dragging
selects a planet. The saved [Luna low review](../runs/v1/omp-openai-codex-gpt-6-luna-low-20261010-173403/reviews/review.md)
reports 84.5/100 while tracking seeks lag and Overview fails to frame the system.
These are historical reviewer findings, not fresh browser observations.

The old rubric averages independent categories. A failed tracking check loses
only part of the selection category, while unrelated passing features keep their
points. Its separate `requirements_verdict` already reports failure, but a reader
who looks only at the total can miss it. The task requests Three.js without a
dedicated renderer check in that rubric.

These examples establish a score-interpretation problem. They do not establish
that Luna or Terra should score below a particular number, or that a working
implementation is too easy for every model.

## Task changes

Task `solar-system-observatory-2` keeps the three-planet observation tool and its
12/24/48-second circular periods. It adds behavior that must remain consistent
across multiple operations:

- Capture up to three immutable snapshots and restore their time, speed,
  selection, and view mode together. Restoration pauses playback.
- Export a versioned JSON snapshot list. Import validates the complete payload
  before replacing the list. Rejection preserves the previous list and live state.
- Persist snapshots across reload while restarting live simulation at its initial
  state. Invalid stored data and unavailable storage require visible recovery.
- Operate primary DOM controls by keyboard, including snapshot and JSON controls.

The task specifies its JSON shape and failure behavior. It does not prescribe a
framework beyond the existing Three.js requirement, an internal architecture,
extra planets, realistic celestial mechanics, or a visual theme. Angles derive
from snapshot time and fixed orbital phases instead of becoming a second source
of imported state.

## Scoring changes

Rubric `solar-system-3.0` adds explicit checks for genuine Three.js rendering and
the new behavior. Its six category weights remain unchanged. A declarative
[check registry](../skills/benchmark-review/assets/check-registry.json) owns check
membership and mandatory-versus-qualitative status.

The review records both the weighted `raw_total_score` and the published
`total_score`. A confirmed mandatory failure sets `score_cap` to 69. When all
checks are verified, the total is the lower of the raw score and that cap.
Without a mandatory failure, the cap is null. A qualitative weakness alone does
not trigger it. Missing evidence still makes both totals null, even when another
check has a confirmed failure.

The cap is a policy choice. It means that 70 or above requires all in-scope
mandatory behavior to pass. It is not a measured difficulty threshold. Keeping
the raw score makes the policy visible and preserves diagnostic information.
Simply making the task harder would not prevent another broken feature from
receiving a high weighted average. Lowering visual ratings to reach a desired
model ranking would bias the evaluation and is not part of this change.

The original saved task continues to use rubric `solar-system-2.0` and review
format `2.0`. No existing score is capped retroactively. A new task score cannot
be compared numerically with an old task score as if they measured the same work.
`runs/v1/` remains the storage layout, not the task version. The gallery can still
sort different rubric versions together, so use its displayed rubric identifier
before interpreting that order.

## Required completion values

A completed generation record must include its actual session ID, model, effort,
elapsed work time, and both full-interval token totals. Setup must establish the
identity, settings, clock, and accounting source before creating a fresh attempt.
If final boundaries or complete usage cannot be established after work starts,
preserve the artifact and finalize a failed record with honest unknowns.

A complete review must include the reviewer model, effort, and review timestamp
in addition to verified checks. Resolve reviewer metadata before browser work.
If it remains unavailable, use the safe blocked-review path, not a completion
claim with a note. This applies to both supported review formats.

The standalone [run validator](../skills/benchmark-record/scripts/validate-run.mjs)
and [review scorer](../skills/benchmark-review/scripts/score-review.mjs) enforce
these completion requirements. Notes cannot waive required values. Placeholder
strings and invented zeros are not acceptable substitutes. The tools validate
structure, not the truth of a claimed measurement or source.

Unknown fields remain null in incomplete or failed records. An unapplied
`score_cap` remains null because that means no cap, not missing evidence.
Existing saved artifacts are not migrated. An older complete record with missing
required values now fails validation instead of receiving an exception.

## Verification and calibration

Run the record-completion and scoring regression cases from the repository root:

```sh
node --test tests/benchmark-record.test.mjs tests/benchmark-score.test.mjs
```

The cases exercise completed, running, failed, interrupted, partial, blocked,
malformed, and historical inputs through the real CLIs. They include missing
required values, placeholders, genuine zero measurements, and unchanged inputs.
They do not prove browser behavior, measurement provenance, or model separation.
Reviewers must still obtain actual observations and accounting sources.

Before strict completion validation was introduced, the task-revision check
passed all 38 repository tests, including eight scorer regression cases.
A read-only CLI run on the 11 then-saved review records reproduced
their category scores, totals, coverage, status, and requirement verdicts without
changing input bytes. The current blank template remained unscored with all 27
checks unverified. These checks establish scoring behavior, not new model quality.

Strict-completion verification passed all 72 repository tests. Direct CLI smoke
checks accepted real zero measurements, rejected a completed record with a
missing token total, and accepted its honest failed-state counterpart. Both
review formats rejected missing completion metadata and retained unscored blocked
records. The saved Astra/high review still normalized without changes; an older
generation record with null tokens was rejected without being rewritten.
Checksums confirmed that the 59 protected saved-run, gallery, and prompt files
remained unchanged during this change. These checks do not prove that a future
host exposes the required accounting sources or that an agent cannot skip a
mandatory validation step.

The storage-failure fixture was exercised through Aside's direct REPL on an
isolated loopback page. A bounded, reviewer-owned filler key caused a real
`QuotaExceededError`. The failed write preserved the previous value, and writes
succeeded after removing the filler. Both fixture keys, the tab, the server,
and the temporary directory were removed. This verifies the fault-injection
method, not any generated application's recovery behavior.

Before claiming improved model discrimination, collect fresh runs with the same
host, tools, task version, and generation conditions. Keep generator identity and
generation cost out of scoring decisions. Use at least five independent attempts
per configuration and alternate configuration order. Review all attempts,
including failures and blocked attempts. Compare mandatory completion rates,
per-check failures, raw and capped score distributions, and review coverage.
Report blocked attempts separately rather than treating them as quality failures
or silently dropping them.

Re-review a subset independently with generator identity concealed where the
host permits it. Compare check disagreements before interpreting score gaps.
If comparable configurations all satisfy the new requirements, the new task may
still be too easy. If their differences are mainly reviewer disagreement or
browser failures, additional task complexity will not resolve that measurement
problem. Do not tune the cap or anchors to make a named model rank lower.

This revision does not include new model-generation trials or fresh artifact
reviews. It changes the task, review instructions, and scoring calculation.
The effect on model separation remains unmeasured.
