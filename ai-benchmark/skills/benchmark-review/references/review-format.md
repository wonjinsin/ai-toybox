# Review Record Format

Use the [template](../assets/review.json), format `1.0`, with no additional fields.
The target directory contains one `reviews/` directory with `review.json` and
`review.md`. These are the latest review, not a history. Replace only those two
files on re-review; do not delete other existing user files.

## JSON fields

- `format_version`: `1.0`.
- `rubric_version`: `solar-system-1.0`.
- `reviewed_at`: actual ISO 8601 timestamp with timezone, or null with a note.
- `reviewer`: current `model` and `effort`, each a known string or null with a note.
- `status`: `complete` when all checks are verified; `partial` when some remain
  unverified; `blocked` when the saved task is incompatible with the rubric.
- `requirements_verdict`: `pass`, `fail` or `unverified`, derived from the rubric.
- `verification`: `browser_method` (`file`, `http`, `none`), `direct_file`
  (`pass`, `fail`, `unverified`), and `viewports` (only actually checked CSS-pixel
  sizes, expressed as strings such as `1440x900`). Direct-file status equals D1.
- `categories`: the six template IDs/weights in template order; `score` is a
  number from 0 to 5, rounded to two decimals, or null. Do not change weights.
- `checks`: exactly the rubric's 20 checks in rubric order. Each object has only
  `id`, `status`, `rating`, `method`, `evidence`. `method` is `file`, `http`,
  `static` or `none`. `evidence` is a nonempty string stating actual observations
  or the missing evidence. For failures include steps, expected and observed
  behavior; for subjective ratings explain the relevant anchor. Do not claim
  precise measurements, screenshots or source lines that were not obtained.
- `total_score`: calculated 0–100 score rounded to two decimals, or null.
- `coverage`: integer `passed`, `failed`, `unverified` check counts, summing to 20.
- `notes`: concise strings for missing metadata, prompt mismatch, environment
  limits, partial grouped-check coverage, applied skill context and cleanup gaps.

For a blocked review, keep all checks unverified, all scores null, browser method
`none`, direct-file `unverified` and viewports empty; explain the unsupported task
in notes. Missing required input files or malformed run JSON are input errors:
report them without replacing any existing review.

No generation token/time metrics, hashes, copied source, screenshots or historical
reviews belong in this record. During a real review, source preservation can be
checked using before/after hashes held only in working memory. Do not infer any
identity or metric from an unrelated run or session. Do not copy generator notes
as if they were observations made by this review.

## Markdown report

Write `review.md` in English, consistent with repository documentation. Use:

1. **Conclusion:** review status, mandatory-requirement verdict, total out of 100
   or `Unscored`, review time and reviewer when known.
2. **Scores:** one table of six categories, 0–5 scores (or `Unverified`), weights
   and short reasons. Do not display a partial sum as an overall score.
3. **Findings:** confirmed defects, affected check IDs, reproduction steps,
   expected versus observed behavior, and qualitative scoring reasons.
4. **Verification and limits:** browser method, direct-file result, actual
   viewports, check coverage, acceptance-scenario outcomes and unexecuted steps.
   Explicitly state when no browser execution occurred. Include cleanup failures.

Use the JSON as the single numerical source for Markdown. Do not create a second
independent set of scores. Read both files back before claiming they were saved.
