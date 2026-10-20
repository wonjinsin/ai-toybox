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

Rubric `solar-system-5.0` keeps the 27 current-task checks, category weights, and
evidence requirements of format 4. The declarative
[check registry](../skills/benchmark-review/assets/check-registry.json) owns check
membership and mandatory-versus-qualitative status. New current-task reviews use
format `5.0` and the [current template](../skills/benchmark-review/assets/review.json).

Format 5 separates quality, mandatory conformance, and evidence completeness.
`total_score` and `total_score_upper_bound` are uncapped quality bounds.
`requirements_verdict` is fail if any mandatory check fails, otherwise unverified
if any mandatory check is unknown, otherwise pass. A score above 69 can accompany
a fail verdict. Failures and missing proof must remain visible beside the score.

An unknown rating stays null and contributes the possible interval 0..5. Each
category uses every registered check in its denominator, not only observed
checks. Totals use unrounded means before two-decimal half-up rounding.
Objective ratings remain 0, 5, or null. The policy does not invent partial credit
for grouped clauses that old evidence did not measure separately.

Missing reviewer metadata affects completion status, not quality arithmetic.
All-unverified blocked records carry 0..100 total bounds. Their zero headline
means no established credit, not observed zero quality. A fully observed review
can be complete and fail mandatory requirements. Format 5 omits the obsolete
`raw_total_score`, `raw_total_score_upper_bound`, and `score_cap` fields.

### Policy-only projections preserve history

`projectCurrentPolicy(review, registry)` in the
[review scorer](../skills/benchmark-review/scripts/score-review.mjs) validates the
source and computes the current quality view without modifying it. Format 5
returns a current result. Format 4 returns a policy-only result labeled
`existing-observations`, with uncapped bounds, category bounds, coverage, the
mandatory verdict, and mandatory failed and unverified IDs. Malformed sources
throw instead of acquiring a trusted numeric projection.

The gallery retains the original published score and original output URL.
Its derived current-policy view records the SHA-256 of the source review bytes
separately from the review. A policy-only score difference is not new behavior
credit. No reviewer, timestamp, observation, run record, prompt, or saved output
is rewritten.

Format 2 reviews measure the original task with a different check set. They
receive no current-policy projection and remain separate from current-policy
score ordering. `runs/v1/` is the storage layout, not a scoring-policy version.

A controlled artifact repair belongs in a separate copy with its own identity
and fresh focused observations. Historical passes cannot become fresh evidence
for that copy. Passing focused checks does not establish a full 27-check score.
Keep policy-only differences separate from any demonstrated behavior changes.

### Frozen format 4 and 2 policies

Format 4 retains numeric category, raw-total, and published-total bounds.
A confirmed mandatory failure sets `score_cap` to 69 and caps both published
endpoints. An unknown mandatory check can cap only the lower endpoint without
changing the verdict to fail or setting a confirmed cap. Equal published bounds
can therefore hide uncertain uncapped quality. The cap is a declared policy,
not a calibrated difficulty or quality threshold.

Saved format 4 records retain those semantics through normalization and
publication. The [frozen template](../skills/benchmark-review/assets/review-4.0.json)
and [historical schema](../skills/benchmark-review/references/review-format.md#frozen-format-4-schema)
remain available. Original-task reviews still use rubric `solar-system-2.0` and
format `2.0`, including nullable partial scores and no bounds or cap fields.
Neither frozen format silently becomes format 5.

The earlier explicit current-task migration from format 3 to 4 preserved check
ratings, observations, reviewer identity, and observation time. Its verification
below describes that earlier migration, not validation of format 5 or a fresh
browser review.

### Format 4 verification

The numeric-score migration passed all 110 repository tests. The finalizer saved
and read back all five JSON/Markdown pairs. Every category and total endpoint is
numeric, and the regenerated five-run catalog matches the saved review JSON.
All original check observations, reviewer metadata, and observation timestamps
were preserved. Checksums confirmed 22 original source files remained unchanged.

Direct Aside REPL checks over loopback HTTP exercised all five score details,
numeric-score filters, and score sorting. The gallery showed five numeric results
and zero unscored results. Desktop rendering and a real 390x844 iframe viewport
displayed the scores, ranges, and missing-proof notices without horizontal
overflow. These checks verify score publication and presentation, not fresh
observations of the benchmark artifacts.

## Required completion values

A completed generation record must include its actual session ID, model, effort,
elapsed work time, and both full-interval token totals. Missing metadata or usage
does not block implementation. Reserve an early running checkpoint, finish
permitted work, and finalize `failed` with honest unknowns when recording values
cannot be established. A finished, measured attempt may be `completed` with
documented functional defects or browser verification gaps. That status records
execution completion, not artifact quality.

A complete review must include reviewer model, effort, and review timestamp in
addition to verified checks. Missing metadata does not block inspection. Preserve
same-attempt evidence in `partial`; formats 5 and 4 still publish numeric bounds,
while frozen format 2 keeps null totals. Missing browser capabilities leave the
affected checks unverified. Use `blocked` for no usable evaluation or incompatible
task routing. Never bypass an access denial to continue.

The standalone [run validator](../skills/benchmark-record/scripts/validate-run.mjs)
and [review scorer](../skills/benchmark-review/scripts/score-review.mjs) enforce
these completion requirements. Notes cannot waive required values. Placeholder
strings and invented zeros are not acceptable substitutes. The tools validate
structure, not the truth of a claimed measurement or source.

Unknown metadata and check ratings remain null in incomplete records. Formats 5
and 4 use numeric score bounds instead. Only frozen format 4 has `score_cap`;
format 5 keeps mandatory conformance independent of quality. Original saved
artifacts are not changed by policy projection. Older complete records missing
required metadata fail validation instead of receiving an exception.

## Reliable publication

The recording validator also publishes checkpoints and terminal records. It
validates the proposal before replacing `run.json`, rejects stale writes using
the existing record's SHA-256, and refuses to overwrite finalized histories.
Completed records require a saved prompt and HTML entrypoint. The checksum is
working state, not a new record field.

The review finalizer scores one draft and generates both reports from the same
normalized JSON. Each replacement is atomic. The pair is not a filesystem
transaction, so an error reports each output's saved state. Both commands reject
unsafe output paths and read back their writes. Structural validity never proves
browser behavior or measurement provenance.

These changes address saved attempts where a missing browser input capability
made an otherwise finished attempt fail, and where recording remained `running`.
No prompt, saved run, or historical score is rewritten. A forcibly terminated
session can still leave its last checkpoint. The workflow cannot promise a final
write after its process has stopped or the host denies output access.


## Verification and calibration

Run the record-completion and scoring regression cases from the repository root:

```sh
node --test tests/benchmark-*.test.mjs
```

The cases exercise completed, running, failed, interrupted, partial, blocked,
malformed, and historical inputs through the real CLIs and projection API. They
cover uncapped quality with mandatory failures, unknown bounds, frozen scoring,
immutable projection inputs, metadata gaps, retained findings, zero measurements,
stale writes, unsafe paths, source preservation, and JSON/Markdown agreement.
They do not prove browser behavior, measurement provenance, or model separation.
Reviewers must still obtain actual observations and accounting sources.

The continuation/publication repair passed `node --test tests/*.test.mjs` with
92 tests. A separate CLI smoke run exercised running, interrupted, resumed, and
completed publication with an explicit browser gap. Under the then-current
format 3 policy, the metadata-gap case published a partial review with null totals
and all observed checks retained. Six discovery entrypoints and 81 local skill links
passed structural checks. Checksums confirmed that 19 existing run, prompt, and
gallery files remained unchanged. These checks use synthetic publication fixtures,
not a fresh Solar System benchmark or browser quality evaluation.
A separate injected disk-error scenario failed the Markdown rename after JSON
publication. The CLI exited nonzero, reported JSON as saved and Markdown as
unchanged, preserved the previous Markdown, and removed its temporary file.

Before strict completion validation was introduced, the task-revision check
passed all 38 repository tests, including eight scorer regression cases.
A read-only CLI run on the 11 then-saved review records reproduced
their category scores, totals, coverage, status, and requirement verdicts without
changing input bytes. The then-current format 3 blank template remained unscored
with all 27 checks unverified. These checks establish historical scoring behavior,
not new model quality.

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
