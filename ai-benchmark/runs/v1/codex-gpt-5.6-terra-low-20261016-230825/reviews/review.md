# Solar System Review

## Conclusion

**Status:** complete. **Mandatory-requirement verdict:** fail. **Total:** 70.50 / 100.

Reviewed at 2026-10-17T08:03:54.021981+00:00 by `openai-codex/gpt-6-astra`; effort was not exposed.
The former viewport blocker is resolved. Under rubric **solar-system-2.0**, coverage is **14 passed, 5 failed, 0 unverified**. All remaining checks are scored; confirmed requirement failures remain failures.

## Scores

| Category | Score (0–5) | Weight | Reason |
|---|---:|---:|---|
| simulation | 5.00 | 25 | All six checks passed, including exact repeat-seek pixels and the 120 s stop. |
| selection | 2.50 | 20 | Scene/list selection and live information passed; tracking and overview failed. |
| usability | 1.00 | 15 | Tracking camera target and slow-drag isolation failed; control labels were clear. |
| visual | 2.50 | 15 | Scene readable; the mobile panel heading/instructions overlap the scene. |
| responsive | 5.00 | 15 | Primary controls were operable at both measured CSS document sizes. |
| delivery | 5.00 | 10 | Asset completeness (D2) and observed HTTP execution (D3) passed. |

## Findings

### T3 — Tracking changes the label, not the camera target

Select Ceres in the scene and enable tracking, then pause and seek to 30 s. Expected: Ceres remains near the center and seeking updates its camera target. Observed: the label changed to Tracking Ceres, but Ceres remained off-center. In the initial desktop observation it was about (254,489) while the scene center was (570,450); at 30 s it was around (704,391). Toggling tracking while paused did not change rendered pixels. A list switch to Boreas changed selection/information but not the target. Source: `output/index.html:16,19,25`.

### T4 — Overview does not restore the full system

At paused time 12 s, select Ceres, zoom in with seven upward wheel actions and orbit toward a top-down view until Ceres is below the viewport. Enable tracking and choose Return to overview. Expected: the full system becomes visible without changing simulation state. Observed: Ceres remained offscreen and the outer orbit was clipped. Selection, 12 s, 4x and paused state were preserved, but the overview framing was not restored. Source: `output/index.html:25`.

### U1 — Tracking camera input still uses the sun as its pivot

Drag and wheel input were exercised in overview/tracking and paused/playing states. The controls changed orientation/zoom, but while tracking Ceres they still rotated/zoomed around the sun, rather than the selected target. This independently violates the current-target camera-control condition. Source: `output/index.html:16,27`.

### U2 — A slow camera drag selects another planet

At paused time 30 s with Ceres selected, drag from (438,475) to (450,475) using twelve one-pixel movements. Expected: Ceres remains selected. Observed: mouse release selected Aurelia; time stayed 30 s. The pointer handler checks each incremental movement against 4 pixels and resets the reference point, rather than accumulating displacement from pointer-down. Source: `output/index.html:27`.

### V2 — Mobile scene overlaps the panel heading and instructions

At 390x844, the scene extends to y=520 but the panel starts at y=464.1875, producing a 55.8125-pixel overlap. Fresh screenshots showed the heading hidden and the first instruction line clipped. Primary controls remained reachable, so this is scored in V2, not duplicated in L2. Source: `output/index.html:6`. V2 rating: **2**, a recurring legibility inconsistency despite coherent desktop styling.

U3 rating **3**: labeled controls and visible state values are clear and usable, with ordinary feedback. V1 rating **3**: distinct bodies, orbit tracks and selection rings are readable, with ordinary visual execution. Camera/drag defects are not deducted again from these qualitative checks.

The saved prompt also requests Three.js. Static inspection found inline Canvas 2D rendering (`output/index.html:11`) and no Three.js dependency. This confirmed prompt-level mismatch is recorded without adding a new scoring check or weight.

## Verification and limits

- Browser method: connected Aside MCP direct REPL over an unchanged, output-scoped `127.0.0.1` HTTP server. No alternate browser, raw CDP, permission change, application-state injection or artifact patch was used.
- Actual document viewports: **1440x900** and **390x844**. Fresh fixed-size iframes in review-created blank tabs provided real CSS dimensions. Desktop client/scroll widths were 1440/1440; mobile widths were 375/375 with a vertical scrollbar. Media queries matched the measured sizes. These are embedded-document checks, not top-level mobile-window or touch-device emulation.
- Review scope: direct-file execution and compatibility are excluded from rubric solar-system-2.0. The saved generation prompt is unchanged; no excluded condition is treated as a missing check.
- Coverage: **14 passed, 5 failed, 0 unverified**. The total is **70.50 / 100**, calculated from all six unchanged category weights and the 19 current checks.

### Saved acceptance scenario

1. Fresh loading showed 0.01 s, 1.00x and automatic motion of all three planets. A measured 1x interval advanced 1.02 simulation seconds over 1.051 real seconds.
2. A scene click selected Ceres and tracking changed its label/highlight/information, but the camera did not follow it (T3 failure).
3. Pause and native seek to 30 s updated the bodies and Ceres angle to 19.7 degrees while remaining paused; the tracking camera still did not update correctly.
4. Changing speed to 2x while paused preserved time and the complete rendered canvas.
5. Resume advanced 2.05 simulation seconds over 1.0287 real seconds and continued from the sought state.
6. Selecting Boreas from the list while playing switched live information/selection and kept 2x playback; the camera target did not switch correctly.
7. Overview preserved selection/time/speed/playback during the scenario. The additional zoom/orbit counterexample above demonstrated failure to restore full-system framing (T4).
8. Pause and seek to 30 s again reproduced identical canvas pixels after restoring the original selection, with the same camera and fixed document viewport.

Additional checks covered all scene/list selections, orbit repeats at 12/24/48 s, speed endpoints, 120 s stop/restart, camera input, drag isolation and all primary mobile controls. The mobile 2x interval advanced 1.04 simulation seconds over 0.5177 real seconds. Mobile selection, tracking, pause, seeking, speed, resume and overview were reachable using vertical scrolling without horizontal scrolling.

Before measurement, absolute simulation-time tolerances were set to 0.12 s at 1x, 0.06 s at 0.25x, 0.20 s at 2x and 0.30 s at 4x; displayed-angle tolerance was 0.25 degrees. The repeat-seek pixel tolerance was zero for identical camera/selection/viewport. An earlier pixel comparison across a changed outer viewport was discarded rather than classified as a simulation defect. Instrumented page-error, console-error and failed-request arrays were empty.

The corrected shared verification-guide example was executed independently in fresh 390x844 and 1440x900 frames. Both reported exact dimensions, appropriate media-query results, no horizontal overflow, and frozen visible time after actual Pause input.

Cleanup confirmed: four review-created tabs were closed, the temporary output server stopped and thirteen transient captures removed. `prompt.md`, `run.json` and `output/index.html` were confirmed unchanged. Other run reports were not re-reviewed in this attempt.

Applied skill context: benchmark-review; developer-mandated using-harness-flow and caveman; host-routed systematic-debugging and aside-browser; implement and test-driven-development loaded by the confirmed-correction handoff; writing-skills loaded for the shared verification-guide correction. Aside guide material was used for capability inspection. Skill isolation was not achieved; benchmark-record was not invoked.

### Policy migration

User-requested policy migration at 2026-10-17T08:18:43.383221+00:00: format 2.0 / solar-system-2.0 removes direct-file execution and compatibility from review. Scores and coverage were recalculated from retained observations; no new artifact browser evaluation occurred. The observation timestamp and reviewer are unchanged.
