# Solar System review rubric

Rubric version `solar-system-5.0`. Review format `5.0`.

## Route the saved task

The saved `prompt.md` is authoritative, not the current generation prompt or an
existing review's version. Select exactly one contract before inspecting behavior.

| Saved task | Rubric | Format and template |
| --- | --- | --- |
| `solar-system-observatory-2`, all R1 through R15 and its acceptance scenario | This rubric, `solar-system-5.0` | [Format 5.0](review-format.md), [template](../assets/review.json) |
| Original 11-requirement task and original eight-step scenario, including saved prompts without an explicit task ID | [Frozen rubric 2.0](rubric-2.0.md), `solar-system-2.0` | [Frozen format 2.0](review-format-2.0.md), [frozen template](../assets/review-2.0.json) |
| Missing, ambiguous, or incompatible behavioral contract | Blocked attempt with no established credit | Use the contract identified by an explicit supported task ID, otherwise format 5.0 as an envelope only; explain the mismatch |

An explicit ID does not override incompatible requirements. Harmless wording
changes are allowed. An old prompt does not acquire new requirements because a
new template exists. Historical task behavior, schema, and scoring are frozen,
with version-specific links where needed. The stricter metadata completion gate
applies to all supported formats; old complete files with missing metadata fail validation.
The format 2 template's empty `checks` array must be populated with its 19 registry
entries before scoring. Saved format 4 records retain the [frozen scoring policy](#frozen-format-4-policy)
and [format 4 template](../assets/review-4.0.json). Never mutate historical records
as part of skill maintenance.

Direct-file execution, no-server delivery, and file-protocol compatibility remain
excluded from review, including for the new task. Do not test `file://`. R6 is
owned only by D2 for asset completeness and D3 for observed HTTP execution.
The verdict does not certify excluded generation conditions.

## Categories and check owners

The [registry](../assets/check-registry.json) fixes IDs, order, category, weight,
mandatory status, and qualitative status. There are 27 current checks. Existing
IDs are retained, and D1 remains removed. Every mandatory clause below must be
observed for its check to pass. One observed violation fails a grouped check;
name unobserved clauses in its evidence. A defect can affect several checks,
but each failure needs its own violated acceptance condition. Do not apply extra
penalties or lower visual ratings to compensate for a functional defect.

### simulation, 25 points

- **S1, R1 and R3.** One sun and exactly three planets have visibly different
  sizes and colors. Their circular orbits have fixed phases and periods of 12,
  24, and 48 simulation seconds. Observe positions and motion, not only constants.
- **S2, R2.** Initial time is 0, speed is 1x, playback is running, selection is
  null, and view mode is overview. Observe startup and subsequent motion.
  Loading time alone is not evidence of a reset bug.
- **S3, R4.** Pause freezes time and all orbital positions. Resume continues
  without a jump. Camera and selection remain usable, exercised under U1 and T1.
- **S4, R5.** Speed covers 0.25x through 4x. Changes preserve current time and
  positions and all relative orbital speeds. Test 0.25x, 2x, and 4x both while
  playing and while paused, then resume.
- **S5, R3 and R9.** Seeking across 0 through 120 pauses and immediately updates
  visible time and all positions. Repeat 30 seconds and noncanonical times after
  changes in speed, playback, and selection. Phases are implementation-chosen,
  fixed, and stable across reload. Do not demand phase zero.
- **S6, R9.** Playback stops at 120 without wraparound or overshoot. Seek earlier
  and confirm playback can resume there.
- **S7, R1.** The artifact actually uses Three.js to render a three-dimensional
  solar scene. Inspect the loaded implementation and resource path, then observe
  rendered depth, occlusion, and perspective changes while orbiting the camera.
  A library import, label, flat SVG, 2D canvas, or prerecorded animation is not
  sufficient. Static code alone never earns a pass.
- **S8, R12.** Capture records time, speed, selected planet, view mode, and all
  three derived orbital angles. IDs are unique, nonempty, and stable for the
  lifetime of each snapshot, including export, import, and reload. Planet IDs
  are implementation-chosen and stable across reload. Change live time, speed,
  selection, and mode after capture and confirm saved fields and angles stay
  unchanged. Angles derive from saved time and fixed phases, not imported angles.
  Display every captured field and all three angles in degrees in the list or
  inspectable details. Capture must not change live state.

### selection, 20 points

- **T1, R7.** Select each planet through the scene and labeled list. Both routes
  share selection and a visible highlight, including while paused.
- **T2, R7.** Selected name, period, and orbital angle agree with the planet and
  its rendered position. Observe updates during playback and seeking.
  With no selection, show no highlight and an explicit no-selection information
  state. Restore must clear any previous selected-planet information.
- **T3, R8 and R9.** Tracking keeps the moving selection near the view center.
  Seeking updates the camera immediately. Selecting a different planet switches
  highlight, information, and target together without changing time, speed, or
  playback. Check genuine centering, not merely a tracking label.
  Without a selected planet, disable or visibly reject tracking without changing
  state.
- **T4, R8.** Overview shows the full system and preserves selection, simulation
  time, speed, and playback state.
- **T5, R12.** Restore applies the snapshot's time, speed, selection, and view mode
  together and pauses playback. All positions, information, and tracking agree.
  Restore resets the camera to deterministic standard overview or tracking
  framing. It need not serialize or restore manual camera offsets. Exercise
  selected tracking and unselected overview snapshots after unrelated live edits.

### usability, 15 points

- **U1, R4 and R10.** Camera orbit and zoom work in overview and tracking, both
  playing and paused. Presence of controls is not proof of operation.
- **U2, R10.** Camera drags do not select planets. Include slow cumulative drags,
  not only one fast move. List, slider, button, snapshot, and JSON interactions
  do not accidentally select the scene or move its camera.
- **U3, quality.** Purpose and current playback, speed, selection, tracking,
  snapshot, and JSON feedback are understandable. Use the qualitative anchors.
  Do not duplicate mobile occlusion failures here.
- **U4, R12.** At most three snapshots exist. A fourth capture shows a visible
  capacity message without replacing or changing any snapshot. Captures append
  to the list. Delete a chosen snapshot without changing live state or remaining
  order and values, then capture a new one without an ID collision.
- **U5, R15.** All primary DOM controls work by keyboard, including selection,
  pause/resume, speed, seeking, overview/tracking, capture, restore, deletion,
  and textarea import/export. Each has a meaningful accessible name and visible
  focus. There is no focus trap. Canvas camera orbit need not support a keyboard.

### visual, 15 points

- **V1, quality.** Scene and selection remain readable. Effects support rather
  than obscure observation.
- **V2, quality.** Typography, spacing, colors, and panel hierarchy are coherent
  and legible. Judge visible results, not CSS volume or decoration.

### responsive, 15 points

- **L1, R11.** At 1440x900 CSS pixels, the scene and every primary control remain
  usable without horizontal scrolling or panels blocking controls. Exercise the
  new snapshot and JSON controls too.
- **L2, R11.** At 390x844 CSS pixels, the same conditions hold. Exercise selection,
  tracking, seek, speed, pause/resume, snapshot capture/restore/delete, and JSON
  import/export. A resized screenshot alone is insufficient.
  Vertical scrolling is allowed at either viewport.

### delivery, 10 points

- **D2, R6 asset completeness only.** All referenced local assets exist under
  `output/` and resolve when it is the HTTP root, without a build or repository
  root dependency. CDN use is allowed. This is the only static-only check.
- **D3, quality.** No artifact-caused console/page errors or failed required
  resources prevent intended operation during browser checks. Distinguish
  infrastructure restrictions from observed artifact defects.
- **D4, R13.** A labeled textarea and explicit export/import controls work without
  an OS file picker. Export produces only the required fields in
  `{version:1,snapshots:[{id,time,speed,selectedPlanet,viewMode}]}`.
  Verify empty and populated lists, order, every field and type, stable IDs,
  unchanged list and live state, and round-trip through the actual UI.
  The name/planet-ID mapping is exposed beside names or JSON controls.
  Do not export angles, playback state, or camera offsets.
- **D5, R13.** Import validates the entire payload before changing anything.
  Required version is numeric `1`; snapshots is an array with at most three
  entries. Each has a unique nonempty string ID, finite numeric time in 0..120,
  finite numeric speed in 0.25..4, selectedPlanet null or a known stable planet
  ID, and viewMode `overview` or `tracking`. Tracking requires a selection.
  Missing required fields, wrong types, unsupported versions, invalid references,
  duplicates, range errors, and malformed JSON show an error and leave both
  snapshots and live state unchanged. A valid import replaces only snapshots;
  live time, playback, speed, selection, view mode, and camera stay unchanged.
  Imported angles cannot override angles derived from time. Test a mixed payload
  with an earlier valid entry and a later invalid entry to establish atomicity.
  Root and entries must be objects, not arrays or null. Reject numeric strings,
  booleans, and non-finite numbers. Fractions, including speed 1.5, are valid.
  Empty arrays clear only the list. Preserve array order and exact case-sensitive
  IDs without trimming, coercion, or regeneration. Ignore extra object keys.
  Verify valid imports both paused and playing; natural time advance is allowed.
- **D6, R14.** Snapshot capture, deletion, and valid import each persist across
  reload using the same versioned JSON object and validation rules. Retain IDs,
  order, values, and derived angles in the same origin and storage context.
  Reload restores only snapshots and starts the live scene at R2 defaults.
  Invalid persisted data yields an empty list, a visible recovery notice, and
  usable live controls. Unavailable storage or a failed write produces a visible
  saving-unavailable notice, not silent success or a false promise of persistence.
  The in-memory tool remains usable. These recovery clauses require observations,
  not exception-handler source or assumptions.

## Ratings and totals

Objective checks use 5/pass, 0/fail, or null/unverified. U3, V1, and V2 use integer
qualitative ratings with concrete observations. D3 remains objective and quality-only.

| Rating | Observed quality |
| --- | --- |
| 0 | Unusable or unreadable. |
| 1 | Severe confusion or legibility problems dominate. |
| 2 | Usable with recurring ambiguity or inconsistency. |
| 3 | Clear, coherent, usable, ordinary execution. |
| 4 | Polished hierarchy and feedback with minor shortcomings. |
| 5 | Exceptionally clear and consistent across checked states. |

Qualitative ratings 0 through 2 fail; 3 through 5 pass. Unobserved quality stays
null/unverified. Quality-only failure does not fail mandatory requirements.

Scores and evidence completeness are separate. For a category with `n` registered
checks, observed-rating sum `s`, and `u` unknown ratings, its lower mean is `s/n`
and upper mean is `(s + 5u)/n`. Keep every registered check in the denominator.
An unknown rating remains null, not an observed zero or five. Its possible
completions define the bounds.

Calculate quality bounds from the unrounded category means times weight/5. Round
category and final total endpoints to two decimals, half up. Do not calculate
totals from rounded categories. The bounds enclose possible rounded scores;
they are not probabilities or a promise that every interior value is attainable.
Store lower endpoints as `score` and `total_score`, and upper endpoints as
`score_upper_bound` and `total_score_upper_bound`.

`requirements_verdict` is fail if any mandatory check fails, otherwise unverified
if any mandatory check is unverified, otherwise pass. Quality is uncapped.
A high quality score does not establish conformance. Display the mandatory
verdict, failures, and unknown checks beside the score. Format 5 has no
`raw_total_score`, `raw_total_score_upper_bound`, or `score_cap` fields.

Equal endpoints mean the rounded quality score is determined, even if reviewer
metadata is missing. Otherwise the headline is a conservative lower bound and
must be labeled beside its range. A mandatory S7 failure and an unknown D6 with
all other ratings at 5 give quality 94.88..96.88 and a fail verdict. The failure
does not collapse the unknown interval or establish a pass.

All format 5 category and total endpoints are numeric. Missing reviewer metadata
affects completion status, not arithmetic. Coverage counts all 27 checks by
status. Complete means every check is observed and required reviewer metadata
and the timestamp are known, not that every check passes. Partial retains
usable findings. Blocked has all checks unverified, category ranges 0..5 and
total ranges 0..100. Its zero headline means no established credit, not proven
zero quality or a compatible task.

Preserve verified findings when later independent checks cannot run. Arithmetic
validates consistency, not browser facts or the truth of metadata and prose.
Frozen format 2 retains its original nullable, uncapped arithmetic and no bound
fields. Format 3 is unsupported, not silently reinterpreted.

## Frozen format 4 policy

Rubric `solar-system-4.0` and format `4.0` retain the same 27 check definitions,
weights, mandatory flags, and evidence rules as format 5. The policy differs.
Format 4 stores the uncapped bounds as `raw_total_score` and
`raw_total_score_upper_bound`. Category bounds use the same formulas above.
Its published `total_score` and `total_score_upper_bound` follow these rules:

- With a confirmed mandatory failure, cap both raw endpoints at 69.
- With no confirmed failure but an unknown mandatory check, cap only the lower
  endpoint at 69. The unknown check could fail; the upper completion passes.
- With every mandatory check passing, use the raw endpoints without a cap.

`score_cap` is 69 only for a confirmed mandatory failure; otherwise it is null.
The cap is declared policy, not a calibrated quality threshold. Equal published
endpoints can hide an uncertain raw score. An optional D3 unknown and a
mandatory failure can leave raw bounds 82.2..84.2 but published bounds 69..69.
All-unverified records retain 0..100 bounds and no confirmed cap.

Normalization and publication preserve these semantics for saved format 4.
`projectCurrentPolicy(review, registry)` exposes the uncapped bounds separately,
with the same verdict, failures, and unknowns. It labels the result policy-only
and based on existing observations. It neither rewrites a review nor establishes
new browser evidence. Format 2 measures a different task and cannot be projected.
