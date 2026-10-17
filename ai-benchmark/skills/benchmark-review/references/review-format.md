# Review record format

Use [task routing](rubric.md#route-the-saved-task) before selecting a format.
Current task `solar-system-observatory-2` uses format `3.0`, rubric
`solar-system-3.0`, and the [27-check template](../assets/review.json).
The original task uses [frozen format 2.0](review-format-2.0.md) and its
[historical template](../assets/review-2.0.json). Do not add format 3 fields to an
old record or apply the new cap to historical scores.

Each run has only `reviews/review.json` and `reviews/review.md` as the latest
attempt. Re-review replaces those two files, including on a blocker. Leave
other user files untouched. Apply the unchanged [stop rule](../SKILL.md#stop-on-access-or-verification-blockers)
and safe-output rules before writing. An artifact defect is a finding, not an
infrastructure blocker. Never repair the artifact to finish a review.

## JSON fields

The current record has exactly these fields. No additional fields are allowed.

- `format_version` is `3.0`.
- `rubric_version` is `solar-system-3.0`.
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
  is incomplete without a stop condition, or `blocked` after a stop condition.
- `requirements_verdict` is `pass`, `fail`, or `unverified`, derived from mandatory
  checks. A complete review may fail. A partial review may have confirmed failures.
- `verification` has exactly `browser_method`, either `http` or `none`, and
  `viewports`, unique actually observed CSS-pixel sizes such as `1440x900`.
  Method `none` requires an empty viewport list.
- `categories` contains exactly the six registry IDs in registry order with
  `id`, `weight`, and `score`. Weights remain 25, 20, 15, 15, 15, and 10. Scores
  are rounded means in 0..5 or null when any owned check is unverified.
- `checks` contains all 27 registry checks exactly once in registry order. Each
  has exactly `id`, `status`, `rating`, `method`, and `evidence`.
  Status is `pass`, `fail`, or `unverified`. Ratings follow the rubric.
  Method is `http`, `static`, or `none`; verified D2 requires `static`, and every
  other verified check requires `http`. HTTP check evidence requires the review's
  browser method to be `http`. L1 and L2 passes require their actual viewports.
- `raw_total_score` is the uncapped weighted total rounded to two decimals, or
  null if any check is unverified.
- `score_cap` is 69 when at least one mandatory check fails, otherwise null.
  A confirmed failure still sets this field when raw and total are null.
- `total_score` is raw limited by the cap, or null whenever raw is null.
- `coverage` has integer `passed`, `failed`, and `unverified` counts totaling 27.
- `notes` is an array of nonempty strings. Record metadata sources and their
  limits, missing metadata, task mismatch, preregistered tolerances, environment
  limits, unobserved grouped clauses, applied skill context, storage fixture
  setup/restoration, and cleanup gaps. Notes cannot waive completion requirements.

Evidence is a nonempty string, not a Boolean or an invented measurement. A
verified format 3 check uses `Action: ... Expected: ... Observed: ...` in that
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
Unresolved required metadata takes the blocked, unscored stop path before
inspection, with null values and reasons. The same rule applies if metadata
becomes unavailable during review. Save only at a resolved safe target under
the existing output rules.

For `blocked`, reset all checks to unverified/null and all scores, raw, and cap
to null. Verdict is unverified and coverage is 0/0/27. Include the current blocker
in notes, not old findings. Browser method and viewports may describe observations
from this attempt before it stopped. If target resolution, safe output, or a write
is blocked, report exactly which files remain old, missing, or newly saved.

No generation token/time metrics, hashes, copied source, screenshots, or historical
reviews belong in the record. Source-preservation hashes stay only in working
memory. Skill maintenance does not rewrite saved reviews. Historical schema and
scoring remain supported, but old complete files with missing reviewer metadata
or timestamps now fail strict validation. There is no historical bypass and no
automatic metadata fill, migration, or rescoring. A later explicit migration
request must not fabricate metadata or observations, waive the completeness
gate, or convert the old task into the new task; strict task routing still applies.

## Read-only scoring command

Run from the repository root with the path to the proposed or saved JSON:

```sh
node skills/benchmark-review/scripts/score-review.mjs <review.json>
```

The dependency-free CLI reads JSON and the check registry, then prints normalized
JSON to stdout. It writes no files and does not alter the input bytes. It exports
no API. Invalid input prints a diagnostic on stderr, exits nonzero, and emits no
scored output. Metadata errors identify the affected field.
Do not redirect stdout onto its input file.
Use a host-permitted temporary working file outside the run when validating a
proposed report, then remove that working file. Do not retain extra evidence files
in the run.

It rejects unsupported or mismatched format/rubric pairs, unknown or duplicate
IDs, wrong order, weights, fields, ratings, statuses, evidence shape, methods,
missing viewport proof, and incompatible partial/blocked records. Both supported
formats reject complete records without concrete reviewer model, effort, and a
valid timestamp, even if notes explain the missing values. The CLI validates
metadata structure, not its truth or provenance. For otherwise valid records it
derives category scores, coverage, verdict, status, and totals.
Well-formed stale derived values are replaced in stdout; it does not treat those
values as evidence. Compare every derived field against the proposed JSON before
publishing, resolve any disagreement, and regenerate Markdown from that same
normalized record. After saving, run the command on the saved path and compare
again. Historical format 2 accepts its original nonempty evidence strings and
19-check shape, retains uncapped arithmetic, and never gains new fields.

The design uses one declarative registry rather than per-category scoring
functions. Both contracts use the same arithmetic and validation. Data keeps
historical membership explicit without compatibility wrappers. Browser judgments
remain in the rubric and evidence matrix because numeric code cannot prove them.

## Markdown report

Write the report in English with these sections:

1. **Conclusion.** Status, mandatory verdict, published total or `Unscored`, raw
   total or `Unscored`, applied cap or `None`, timestamp, and reviewer. Explain
   a 69 cap as policy rather than a calibrated quality measurement.
2. **Scores.** One table of six categories, scores or `Unverified`, weights, and
   short reasons. Never display a partial sum as a total.
3. **Findings.** Check IDs, reproducible actions, expected and observed outcomes,
   and qualitative anchor reasons. Separate confirmed failures from missing proof.
4. **Verification and limits.** Actual browser method and viewports, coverage,
   saved acceptance sequence outcomes, fixed evidence-matrix coverage, tolerances,
   unexecuted clauses, storage fixture restoration, and cleanup failures.
   State explicitly when no browser execution occurred.

Use JSON as the only numerical source. Validate both reports in memory before
replacing either. Read both saved files back before claiming the pair is complete.
