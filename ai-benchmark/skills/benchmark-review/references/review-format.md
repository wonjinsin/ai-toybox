# Review record format

Use [task routing](rubric.md#route-the-saved-task) before selecting a format.
Current task `solar-system-observatory-2` uses format `5.0`, rubric
`solar-system-5.0`, and the [27-check template](../assets/review.json).
Saved format 4 records use the [frozen format 4 schema](#frozen-format-4-schema)
and [template](../assets/review-4.0.json).
The original task uses [frozen format 2.0](review-format-2.0.md) and its
[historical template](../assets/review-2.0.json). Never convert the original task.

Each run has only `reviews/review.json` and `reviews/review.md` as the latest
attempt. Re-review replaces those two files, including on a blocker. Leave
other user files untouched. Apply the [continuation and access rules](../SKILL.md#continue-independent-permitted-checks)
and safe-output rules before writing. An artifact defect is a finding, not an
infrastructure blocker. Never repair the artifact to finish a review.

## JSON fields

The current record has exactly these fields. No additional fields are allowed.

- `format_version` is `5.0`.
- `rubric_version` is `solar-system-5.0`.
- `reviewed_at` is an actual valid ISO 8601 timestamp with timezone. Only partial
  or blocked records may use null with a note.
- `reviewer` has exactly `model` and `effort`, each a concrete nonempty string.
  Only partial or blocked records may use null with a note. Placeholder values
  such as `unknown`, `null`, `undefined`, `unavailable`, `n/a`, `default`, or `auto`
  and unexpanded `${...}` values are not metadata.
  A sourced effort of `none` may name an explicitly selected no-reasoning mode;
  it is not a substitute for missing effort. Model `none` is a placeholder.
- `status` is `complete` only when every check is verified and all required
  reviewer metadata and the timestamp are known. It is `partial` when evidence
  or required metadata is incomplete and usable findings remain. `blocked`
  means no usable evaluation or an unresolved/incompatible task contract.
- `requirements_verdict` is `pass`, `fail`, or `unverified`, derived from mandatory
  checks. A complete review may fail. A partial review may have confirmed failures.
- `verification` has exactly `browser_method`, either `http` or `none`, and
  `viewports`, unique actually observed CSS-pixel sizes such as `1440x900`.
  Method `none` requires an empty viewport list.
- `categories` contains exactly the six registry IDs in registry order with
  `id`, `weight`, `score`, and `score_upper_bound`. Weights remain 25, 20, 15, 15,
  15, and 10. Scores are finite lower and upper bounds in 0..5, rounded to two
  decimals. Every registered check stays in the denominator.
- `checks` contains all 27 registry checks exactly once in registry order. Each
  has exactly `id`, `status`, `rating`, `method`, and `evidence`.
  Status is `pass`, `fail`, or `unverified`. Ratings follow the rubric.
  Method is `http`, `static`, or `none`; verified D2 requires `static`, and every
  other verified check requires `http`. HTTP check evidence requires the review's
  browser method to be `http`. L1 and L2 passes require their actual viewports.
- `total_score` and `total_score_upper_bound` are finite lower and upper bounds on
  uncapped quality, rounded to two decimals. The mandatory verdict never caps
  either endpoint. See the [exact formulas](rubric.md#ratings-and-totals).
  Equal endpoints establish a determined rounded score, not complete evidence
  or satisfaction of mandatory requirements.
- `coverage` has integer `passed`, `failed`, and `unverified` counts totaling 27.
- `notes` is an array of nonempty strings. Record metadata sources and their
  limits, missing metadata, task mismatch, preregistered tolerances, environment
  limits, unobserved grouped clauses, applied skill context, storage fixture
  setup/restoration, and cleanup gaps. Notes cannot waive completion requirements.

Evidence is a nonempty string, not a Boolean or an invented measurement. A
verified format 5 or 4 check uses `Action: ... Expected: ... Observed: ...` in that
order. Describe the actual input, expected outcome, and observed result with
values where available. For failures provide reproducible steps; for qualitative
checks explain the chosen anchor. Put any incomplete clauses after the actual
observed violation. An `Observed: Not observed` placeholder cannot support pass
or fail. Unverified evidence states what was not observed and why. The template
starts every check unverified and explicitly says no observation occurred.

The scorer checks evidence shape, not truth. Labels, source constants, earlier
reviews, and the scorer's own output cannot prove motion, state transitions,
keyboard usability, or persistence. Only D2 is static-only. Source inspection
supports S7 but does not prove rendered Three.js operation without HTTP evidence.

Before source or browser inspection, resolve reviewer model and effort from
current host metadata, then explicit user declaration, then configuration that
applies to the current review session, in that precedence. Generic defaults and
`auto` do not establish the selected model or effort. Obtain the actual timestamp
from the available clock or a trustworthy current-session time source. Record the
sources and limits in notes. Do not use generator metadata, folder names, earlier
review identities, or invented values. Do not launch another model to fill a gap.
Unresolved required metadata remains null with reasons. Continue independently
permitted inspection and finalize `partial`. Missing metadata does not change
format 5 arithmetic or authorize inventing an identity. The same rule applies if
metadata becomes unavailable during review. Save only at a resolved safe target.

For `blocked`, all checks remain unverified with null ratings. Category bounds
are 0..5 and quality total bounds are 0..100.
The headline 0 is a lower bound with no established credit, not observed zero
quality. Verdict is unverified and coverage is 0/0/27. Explain why no usable
evaluation was possible. An unresolved task's range establishes no task identity.
Missing/malformed inputs still produce this record when the target is safe.
Do not reset same-attempt verified findings after a later capability gap or
denial; preserve them in a partial record and stop only affected operations.
A fresh review must not copy earlier findings. If target resolution, safe output,
or a write is blocked, report each output's known saved, unchanged, missing, or
unknown state.

No generation token/time metrics, hashes, copied source, screenshots, or historical
reviews belong in the record. Source-preservation hashes stay only in working
memory. Maintenance alone does not authorize rewriting saved reviews.
Current-policy projections reuse existing observations without rewriting saved
reviews. They are policy-only comparisons, not fresh browser reviews.
Format 3.0 is unsupported rather than accepted as a silent alias.
Frozen formats 4.0 and 2.0 retain their schemas and arithmetic. Missing completion
metadata still fails validation. Never fabricate metadata or convert the old task.

## Frozen format 4 schema

Format `4.0` pairs only with rubric `solar-system-4.0`. It uses the same fields,
27 checks, category bounds, and evidence rules as format 5, plus exactly these
three fields:

- `raw_total_score` and `raw_total_score_upper_bound` hold finite uncapped bounds.
- `score_cap` is 69 for a confirmed mandatory failure, otherwise null.

Its `total_score` and `total_score_upper_bound` are capped published bounds,
not format 5 quality bounds. A confirmed failure caps both endpoints at 69.
An unknown mandatory check caps only the lower endpoint when none has failed.
The [frozen policy](rubric.md#frozen-format-4-policy) defines the arithmetic.
Blocked format 4 records have raw and published ranges 0..100 and no cap.
The scorer and finalizer retain this schema. They never change a supplied
version to the current version. Format 5 rejects all three obsolete fields.

## Read-only scoring command

Run from the repository root with the path to the proposed or saved JSON:

```sh
node skills/benchmark-review/scripts/score-review.mjs <review.json>
```

The dependency-free CLI reads JSON and the check registry, then prints normalized
JSON to stdout. It writes no files and does not alter the input bytes. The module
exports the pure `normalize(review, registry)` function used by the finalizer;
the CLI runs only when invoked directly. Invalid input prints a diagnostic on
stderr, exits nonzero, and emits no scored output.
Do not redirect stdout onto its input file.
Use a host-permitted temporary working file outside the run when validating a
proposed report, then remove that working file. Do not retain extra evidence files
in the run.

It rejects unsupported or mismatched format/rubric pairs, unknown or duplicate
IDs, wrong order, weights, fields, ratings, statuses, evidence shape, methods,
missing viewport proof, and incompatible partial/blocked records. All supported
formats reject complete records without concrete reviewer model, effort, and a
valid timestamp, even if notes explain the missing values. The CLI validates
metadata structure, not its truth or provenance. For otherwise valid records it
derives category scores, coverage, verdict, status, and totals.
Well-formed stale derived values are replaced in stdout; it does not treat those
values as evidence. A partial record requires missing evidence or metadata and
retains all observed ratings and the derived mandatory verdict. Format 5 and 4
scores and upper bounds must be finite, including in partial and blocked drafts.
Historical format 2 accepts its original nonempty evidence strings and 19-check
shape, retains nullable uncapped arithmetic, and never gains new fields.
Use the publication command to generate Markdown from normalized JSON and
validate the saved pair.

One declarative registry owns check membership and weights. The normalizer owns
all versioned scoring policies and evidence validation. Browser judgments remain
in the rubric and evidence matrix because arithmetic cannot prove them.

## Read-only current-policy projection

The scorer also exports `projectCurrentPolicy(review, registry)`. It validates
through `normalize` and returns one of these shapes:

- Format 5 returns `{kind:'current', rubric:'solar-system-5.0', quality}`.
- Format 4 returns `{kind:'policy-only', source_format:'4.0',
  source_rubric:'solar-system-4.0', target_rubric:'solar-system-5.0',
  basis:'existing-observations', quality}`.
- Format 2 returns `{kind:'historical', source_format:'2.0',
  reason:'Different task and check set; no current-policy projection.'}`.

`quality` contains `lower`, `upper`, `categories`, `requirements_verdict`,
`coverage`, `mandatory_failed`, and `mandatory_unverified`. The last two fields
are arrays of check IDs in registry order. Categories retain their numeric bounds.
Format 4 projections use normalized uncapped bounds, not trusted saved totals.
The check set and quality arithmetic are identical to format 5.

The function does not mutate the input or create a review with a new timestamp,
reviewer, or observations. Malformed sources throw. Consumers must expose the
validation issue instead of falling back to an unvalidated numeric score.
The projection is not a finalizer input. Retain original published scores and
source provenance separately when displaying a policy-only comparison.


## Publication command

```sh
node skills/benchmark-review/scripts/finalize-review.mjs <run-directory> <draft-review.json>
```

Both arguments are required. The target must be an existing non-symlink directory.
The draft must be a regular, non-symlink, single-link file outside the entire target
run, including after canonical path resolution. The command preserves the draft.
The caller removes its working draft after publication. Missing or malformed run
inputs do not prevent publication of an honest blocked draft at a safe target.
Complete publication additionally requires regular nonempty `prompt.md`,
`run.json`, and `output/index.html`, with a real output directory and a parseable
run JSON object. Partial and blocked publication do not require playable output.

The command validates and normalizes before creating or replacing final outputs.
It generates every Markdown score, finding, coverage count, and note from that
normalized JSON and escapes user-controlled text as literal content. It rejects a
symlinked `reviews/`, symlinked or hard-linked named outputs, nonregular outputs,
and escaping output paths. It does not touch source artifacts or unrelated files.
It stages both files with exclusive sibling temporary files, flushes and closes
them, then renames JSON followed by Markdown. It cleans up its own temporary files.
Do not run concurrent publishers or change the target directories during publication.
Path checks detect changed directories but are not a sandbox against hostile
concurrent filesystem mutation.

Each rename is atomic. The pair is not transactionally atomic, and the command
does not attempt rollback. A failure after the JSON rename can leave new JSON
beside unchanged or missing Markdown. Readback parses and normalizes saved JSON
and compares both saved files against the generated bytes. Exit 0 means the saved
pair passed readback. Exit 1 means validation, publication, readback, or cleanup
failed; stderr gives the diagnostic. Stdout always reports a JSON object with
`ok`, per-file `outputs`, resolved `paths` when available, and an `error` or
`cleanup_errors` when applicable. Each output state is `saved`, `unchanged`,
`missing`, or `unknown`. `saved` records a successful rename, not a successful
pair readback. `unknown` means safe inspection did not establish the state.
On a readback failure, report the error alongside those operation states rather
than claiming the current contents are validated. Never retry a denied operation
or claim successful rollback.

Publish an all-unverified checkpoint before inspection, with notes that evaluation
has not occurred. Format 5 shows a 0..100 score range at that point. Publish the
final record after all independently permitted checks, including partial and
blocked outcomes. Existing saved reviews are not fresh evidence.


## Markdown report

Write the report in English with these sections:

1. **Conclusion.** Evidence status, mandatory verdict, uncapped quality bounds,
   mandatory failed and unverified IDs, timestamp, and reviewer. Identify a
   determined total or a conservative lower bound independently of evidence
   completeness. Frozen format 4 retains published/raw bounds and cap explanation.
2. **Scores.** One table of six category bounds, weights, and short reasons.
   Unknown ratings span 0..5 without becoming observed failures. Include all
   registered checks rather than renormalizing only observed ones.
3. **Findings.** Check IDs, reproducible actions, expected and observed outcomes,
   and qualitative anchor reasons. Separate confirmed failures from missing proof.
4. **Verification and limits.** Actual browser method and viewports, coverage,
   saved acceptance sequence outcomes, fixed evidence-matrix coverage, tolerances,
   unexecuted clauses, storage fixture restoration, and cleanup failures.
   State explicitly when no browser execution occurred.

Use JSON as the only numerical source. The publication command generates these
sections, including every check's full evidence and every note, and validates the
saved pair before reporting success.
