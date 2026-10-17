# Review Record Format

Use the [template](../assets/review.json), format `2.0`, with no additional fields.
The target directory contains one `reviews/` directory with `review.json` and
`review.md`. These are the latest review, not a history. Replace only those two
files on re-review; do not delete other existing user files.

Each new attempt replaces the previous review, including a `blocked` attempt
under the [stop rule](../SKILL.md#stop-on-access-or-verification-blockers).
Save the current blocker instead of retaining old scores or continuing evaluation.
An observed artifact defect is a review finding, not an infrastructure blocker.

## JSON fields

- `format_version`: `2.0`.
- `rubric_version`: `solar-system-2.0`.
- `reviewed_at`: actual ISO 8601 timestamp with timezone, or null with a note.
- `reviewer`: current `model` and `effort`, each a known string or null with a note.
- `status`: `complete` when all checks are verified; `partial` when some remain
  unverified without a stop condition; `blocked` for an incompatible task,
  invalid required inputs, or an access/verification blocker.
- `requirements_verdict`: `pass`, `fail` or `unverified`, derived from the rubric.
- `verification`: `browser_method` (`http` or `none`) and `viewports` (only
  actually checked CSS-pixel sizes, expressed as strings such as `1440x900`).
- `categories`: the six template IDs/weights in template order; `score` is a
  number from 0 to 5, rounded to two decimals, or null. Do not change weights.
- `checks`: exactly the rubric's 19 checks in rubric order, excluding removed D1.
  Each object has only `id`, `status`, `rating`, `method`, `evidence`. `method` is
  `http`, `static` or `none`. `evidence` is a nonempty string stating observations
  or the missing evidence. For failures include steps, expected and observed
  behavior; for subjective ratings explain the relevant anchor. Do not claim
  precise measurements, screenshots or source lines that were not obtained.
- `total_score`: calculated 0–100 score rounded to two decimals, or null.
- `coverage`: integer `passed`, `failed`, `unverified` check counts, summing to 19.
- `notes`: concise strings for missing metadata, prompt mismatch, environment
  limits, partial grouped-check coverage, applied skill context and cleanup gaps.

For a blocked review, set all checks to `unverified` with null ratings, all scores
to null, `requirements_verdict` to `unverified`, and coverage to
0 passed, 0 failed, 19 unverified. Record only this attempt's observed browser
method and viewports, or `none` and an empty list if none were observed. Explain
the blocker and unperformed checks in notes and Markdown; do not carry over old
scores or findings. Missing required inputs or malformed run JSON also produce
a blocked replacement when the target directory is resolved and output is safe.
An unresolved target, unsafe output or denied write must instead be reported
without bypassing the restriction or claiming the replacement succeeded.

No generation token/time metrics, hashes, copied source, screenshots or historical
reviews belong in this record. During a real review, source preservation can be
checked using before/after hashes held only in working memory. Do not infer any
identity or metric from an unrelated run or session. Do not copy generator notes
as if they were observations made by this review.

A user-requested rubric migration is not a new browser review. Retain the
observation timestamp and reviewer, preserve supported findings, and explain
the migration and absence of new execution in notes. Remove excluded checks
and fields, then recompute category scores, total, coverage, status and verdict.
Do not turn a previously blocked artifact into a verified one without evidence.

## Markdown report

Write `review.md` in English, consistent with repository documentation. Use:

1. **Conclusion:** review status, mandatory-requirement verdict, total out of 100
   or `Unscored`, review time and reviewer when known.
2. **Scores:** one table of six categories, 0–5 scores (or `Unverified`), weights
   and short reasons. Do not display a partial sum as an overall score.
3. **Findings:** confirmed defects, affected check IDs, reproduction steps,
   expected versus observed behavior, and qualitative scoring reasons.
4. **Verification and limits:** browser method, actual viewports, check coverage,
   acceptance-scenario outcomes and unexecuted steps.
   Explicitly state when no browser execution occurred. Include cleanup failures.

Use the JSON as the single numerical source for Markdown. Do not create a second
independent set of scores. Read both files back before claiming they were saved.
