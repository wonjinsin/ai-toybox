# Solar System Review Rubric

Rubric version: `solar-system-2.0`.

## Supported task

Use the run's saved prompt as authority. This rubric supports the task with 11
requirements: one sun/three planets; automatic time-0, 1x playback; distinct
sizes/colors and fixed-phase 12/24/48-second circular orbits; pause/resume;
0.25x–4x speed; direct-file delivery; scene/list selection and live information;
tracking/overview; 0–120-second seeking; camera/input isolation; and usable
1440x900 and 390x844 CSS-pixel layouts. It includes the eight-step acceptance
scenario ending with a repeated seek to 30 seconds. Harmless wording changes are
compatible; missing, added or changed behavioral requirements require a new
rubric version. Do not score an older or different task with this rubric.

Review scope excludes R6's direct-file/no-server execution requirement and
file-protocol compatibility restrictions. Keep the saved prompt and generation
contract unchanged, but do not test or score those conditions. Delivery scoring
covers asset completeness and observed HTTP execution only. D1 is removed;
remaining check IDs are unchanged.

`R1`–`R11` below refer to those saved requirements. Each check has one scoring
owner. A defect may explain several observations, but do not deduct arbitrary
extra points in other categories. Mark a second check failed only when its own
independent acceptance condition is violated.

## Categories and checks

### simulation — Simulation and time correctness — 25 points

- **S1 (R1, R3):** One sun and exactly three orbiting planets with visibly
  different sizes and colors; circular orbits with periods 12, 24, 48 simulation
  seconds. Verify motion and periods, not merely constants in source.
- **S2 (R2):** Initial simulation state is time 0, speed 1x, playing. Observe
  initial state and subsequent motion; elapsed loading time is not a reset bug.
- **S3 (R4):** Pause freezes time and orbital positions; resume continues without
  a jump. Camera and selection remain usable (also exercise U1 and T1).
- **S4 (R5):** Speed supports 0.25x–4x, preserves relative speeds and current
  time/positions when changed. Changing it while paused applies on resume.
  Include 0.25x, 2x and 4x, not only the default.
- **S5 (R3, R9):** Seeking within 0–120 pauses playback, updates all positions
  and visible time immediately, and revisiting 30 seconds reproduces positions
  independently of prior speed/playback/selection. Initial phases need not be zero.
- **S6 (R9):** Playback stops at 120 without wrapping or advancing beyond it;
  seeking earlier allows playback to resume there.

### selection — Selection and tracking consistency — 20 points

- **T1 (R7):** Each planet can be selected in the scene and labeled list; both
  routes share the selection and visible highlight, including while paused.
- **T2 (R7):** Name, period and current orbital angle agree with the selected
  planet and its rendered position; update during playback and seeking.
- **T3 (R8, R9):** Tracking keeps the selected moving planet near the view center;
  seeking updates the camera immediately, and selecting another planet switches
  highlight, information and target together without changing playback/time/speed.
- **T4 (R8):** Overview shows the full system and preserves selection, simulation
  time, speed and playback state.

### usability — Controls and interaction — 15 points

- **U1 (R4, R10):** Camera orbit and zoom work in overview and tracking, playing
  and paused. Test the controls, not just their presence.
- **U2 (R10):** Camera drag does not select a planet. List, sliders and buttons
  do not accidentally select the scene or move the camera.
- **U3 (quality):** Controls communicate their purpose and current playback,
  speed, selection and tracking state. Rate discoverability and feedback using
  the qualitative anchors below. Do not duplicate mobile occlusion penalties.

### visual — Visual quality — 15 points

- **V1 (quality):** The scene and selected object remain readable against their
  surroundings; visual effects support rather than obscure observation.
- **V2 (quality):** Typography, spacing, color and panel hierarchy form a coherent,
  legible interface. Judge visible results, not the amount of CSS or decoration.

### responsive — Responsive usability — 15 points

- **L1 (R11):** At 1440x900 CSS pixels, scene and all primary controls are usable,
  without horizontal scrolling or panels blocking controls.
- **L2 (R11):** At 390x844 CSS pixels, the same conditions hold. Exercise selection,
  tracking, seeking, speed and pause/resume; a resized screenshot alone is not proof.

### delivery — Execution and delivery — 10 points

- **D2 (R6, asset completeness only):** Static inspection confirms all referenced
  local assets exist under `output/` and their paths resolve when that directory
  is served as the HTTP root, without a build or repository-root dependency.
  CDN use is allowed.
- **D3 (quality):** During browser checks, no artifact-caused console/page errors
  or failed required resources prevent intended operation. Distinguish an
  environment/network restriction from a demonstrated artifact defect.

## Ratings and aggregation

There are **19 checks**. Each must appear exactly once in the record, even if
unverified. For objective checks (all except U3, V1 and V2), use rating `5` for
`pass`, `0` for `fail`, and null for `unverified`. A grouped check passes only when
all its clauses are observed satisfied. An observed violation makes it fail even
if other clauses remain unchecked; name those limits in its evidence.

For U3, V1 and V2, use the following integer anchors with concrete observations:

| Rating | Observed quality |
| --- | --- |
| 0 | Unusable or unreadable presentation. |
| 1 | Severe confusion or legibility problems dominate. |
| 2 | Usable with recurring ambiguity or inconsistency. |
| 3 | Clear, coherent and usable; ordinary execution. |
| 4 | Polished hierarchy/feedback with only minor shortcomings. |
| 5 | Exceptionally clear and consistent across the checked states. |

For these qualitative checks, ratings 0–2 are `fail`, 3–5 are `pass`; without
visual/interaction observation use `unverified` and null, not an inferred rating.
Quality-only failures do not by themselves fail mandatory requirements.

A category score is the arithmetic mean of its check ratings on a 0–5 scale.
If any owned check is unverified, its score is null. Total is the sum of
`unrounded_category_score / 5 * weight`, rounded once to two decimal places;
it is null if any category or required reviewer metadata is missing. Display
category scores to two decimal places, but calculate the total from check ratings,
not rounded displays. A metadata gap does not erase observed ratings or category
means; finalize partial with null total even when every check was observed.

`requirements_verdict` is `fail` if any R-linked check fails, otherwise
`unverified` if any R-linked check is unverified, otherwise `pass`. A complete
review can have a failed verdict. Visual polish cannot cancel a failed
requirement. Coverage counts all checks by status; it is not a quality score.
The verdict covers only the R-linked conditions in this rubric; it does not
certify the excluded direct-file generation contract.
