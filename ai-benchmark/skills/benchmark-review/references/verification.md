# Browser and State Verification

## Permitted method

Use only connected Aside MCP's direct `repl` tool. Read its current documentation
before navigation and choose a supported, permitted method. Do not use Aside
`exec`, natural-language CLI tasks, another browser runtime, raw CDP or another
model session. Do not install/register tools during a review.

Prefer the real `output/index.html` via `file://`. If that protocol is unsupported
and loopback HTTP is permitted, serve only unchanged `output/` temporarily on
`127.0.0.1`. Record HTTP interaction results separately from direct-file status.
Static compatibility is D2, not a substitute for D1. External CDN resources are
permitted by the task; report actual loading failures and environment limits.

An explicit access denial, including `Cannot navigate to a file URL without local
file access.`, stops that access. Do not switch protocols/tools or change security
settings to bypass it. If Aside is unavailable, continue static inspection and
record all unsupported runtime/visual checks as unverified. Never infer rendering
or interaction success from HTML, screenshots from previous reviews, or mocks.

Create a new review tab; do not inspect unrelated tabs or browsing memory. Close
only that tab and stop the temporary server on every exit. Inspect screenshots
transiently; do not retain images, videos, logs or an evidence directory.

## Observe the saved task

Run the saved eight-step scenario in order: automatic motion; scene selection
and tracking; pause and seek 30; set 2x while paused; resume at twice real-time;
select another planet from the list; return to overview; pause and seek 30 again.
Record actual results, not just a statement that the scenario was attempted.
Then cover every remaining rubric check, including the speed endpoints, 120-second
boundary, paused camera controls, drag/click separation and both viewport sizes.

For circular motion, compare observed orbital positions/angles with
`phase + 360 * simulation_time / period` modulo 360, using each implementation's
fixed initial phase and angle convention. Do not demand phase zero or confuse
camera projection with orbital angle. Compare repeat-seek states with the same
camera and selection. Observe time advancement over a measured interval at each
speed; note the interval and allow for measured frame/timer precision. Choose and
record numeric tolerances before judging the observation; never relax them after
a mismatch to produce a pass. If accessible observations cannot establish a
condition, leave it unverified rather than exposing or patching app internals.

Use DOM/console/resource observations as supporting evidence, not replacements
for rendered motion and actual user input. Separate observed app defects from
infrastructure restrictions. Explain each failure with control actions, expected
behavior, observed behavior, and source locations when inspection establishes a
cause. Source references are supplementary; no retained screenshot is required.
