# Browser and State Verification

## Permitted method

Use only connected Aside MCP's direct `repl` tool. Read its current documentation
before navigation and choose a supported, permitted method. Do not use Aside
`exec`, natural-language CLI tasks, another browser runtime, raw CDP or another
model session. Do not install/register tools during a review.

Use loopback HTTP for review. Confirm that it is documented and host-permitted
before navigation. Do not navigate to `file://` or inspect file-protocol
compatibility; those generation requirements are outside this review rubric.
Unavailable local-file navigation is not a missing review check or a reason to
leave a review unscored. This policy never overrides an access/security denial.

Serve only unchanged `output/` on `127.0.0.1`, not the repository root. Run a
tracked host process:

```sh
python3 -m http.server 8765 --bind 127.0.0.1 --directory "<run>/output"
```

Replace `<run>/output` with the actual output directory (quote paths containing
spaces); use `--directory output` when already in the run directory. Wait for
startup, then use direct REPL `openTab("http://127.0.0.1:8765/index.html")`.
If the port is occupied, choose another available port and use it in both places;
do not stop an unrelated process. Retain the process handle/PID for cleanup.
Do not use Aside `exec` to start the server.

Record actual HTTP rendering, interactions and resource-loading results.
External CDN resources are permitted; distinguish observed loading failures
from environment/network restrictions.

An explicit host/tool access or security-policy denial stops the affected
operation under the [continuation and access rules](../SKILL.md#continue-independent-permitted-checks).
Do not retry denied access, change permissions or switch protocols/tools to bypass
a denial. Continue only independent permitted work. Missing Aside or a required
browser capability leaves affected checks unverified; it does not prohibit
permitted static inspection for D2 or erase observed findings. A confirmed
functional failure is fail; continue other checks. Finalize partial when usable
findings remain, blocked when none do. Report the exact operation and error, not
an invented OS permission diagnosis. A past review's denial is historical context,
not a fresh attempt on this target or proof of a current restriction. Never infer
rendering or interaction success from HTML, old screenshots, or mocks.

Create a new review tab; do not inspect unrelated tabs or browsing memory. Close
only that tab and stop the temporary server on every exit. Inspect screenshots
transiently; do not retain images, videos, logs or an evidence directory.

## Verify real document viewports

A missing `page.setViewportSize` alone is not a verification blocker. Check the
actual document dimensions; Aside's outer viewport may be fixed or change with
browser presentation. When no viewport setter is available, use direct REPL's
documented frame support to load the unchanged artifact in a fresh, fixed-size
iframe. This tests a real CSS viewport, not a screenshot crop or scaled layout.

Create the fixture only in a review-created blank tab, using the already
permitted artifact URL. Use fresh variable names and a fresh fixture for each
required size; the example uses the output-scoped HTTP server above.

```js
const viewTab = await openTab("about:blank");
await viewTab.evaluate(({ url, width, height }) => {
  document.body.style.margin = "0";
  const frame = document.createElement("iframe");
  frame.id = "review-frame";
  frame.style.cssText =
    `display:block;width:${width}px;height:${height}px;border:0`;
  frame.src = url;
  document.body.appendChild(frame);
}, { url: "http://127.0.0.1:8765/index.html", width: 390, height: 844 });
console.log((await snapshot(viewTab, { interactive: true })).tree);
```

Wait for the snapshot to show the loaded artifact, taking another snapshot if
navigation is still pending. Resolve its frame with `await viewTab.frames()`,
then measure inside that frame with `frame.evaluate`: `innerWidth`,
`innerHeight`, the applicable media query, and the document's `clientWidth` and
`scrollWidth`. Require the requested CSS dimensions before judging the layout.
Exercise actual frame controls through `frame.locator(...)` or snapshot frame
refs, and inspect a transient screenshot. Repeat at 1440x900 and 390x844.

Modify only the blank fixture, never the artifact DOM, CSS, script, runtime state,
or reported viewport properties. The current task's narrowly scoped persisted-data
and storage-quota fixtures below are the only storage-boundary exceptions.
Recheck dimensions before comparing repeated positions; a changed viewport
invalidates a pixel comparison, not the artifact.
Do not use raw CDP, another browser, permission changes or an iframe to bypass
an access denial. If no permitted frame path can perform a required viewport
check, leave that check unverified with its reason and continue independent
allowed checks. Preserve already observed passes and failures in partial JSON.

Record the actual document sizes and disclose embedded-frame verification in
notes and Markdown. It is not top-level viewport or touch-device proof.
Close the fixture tab on every exit.

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

## Current task evidence matrix

This section applies only to `solar-system-observatory-2`. Use the frozen rubric's
19 conditions for the historical task; do not require its snapshots, keyboard
coverage, null-selection defaults, or new evidence probes retroactively.

Run the saved eight steps first, then all six additional acceptance sequences in
their saved order. Record each outcome and map it to the fixed owners below.
The extra probes test the same stated requirements rather than adding hidden
features. Capture evidence as `Action: ... Expected: ... Observed: ...`.
For grouped checks, record every clause or explicitly name the missing clauses.
No observation means unverified, not a guessed pass or an automatic failure.

Before starting observations, record the following tolerances in notes. These
defaults account for display rounding and browser scheduling without forgiving
broken state transitions. If the UI has coarser documented precision, choose a
justified alternative before observing results. Never widen a tolerance after a
mismatch. If precision prevents a conclusion, record the check as unverified.

- Measure speed over at least 3 real seconds, away from the endpoint. Expected
  simulation delta is speed times measured wall delta. Allow the larger of
  0.15 simulation seconds or 5% of the expected delta, plus one displayed time
  increment. Log actual wall and simulation deltas.
- For paused states, allow one visible time increment and 1 degree for rounded
  angle readouts. Use circular angle distance, including around 0/360. Account
  for up to one rendered frame when comparing pre-pause and post-pause samples.
- Compare predicted angle against the implementation's fixed phase and period.
  Allow 2 degrees plus displayed angle rounding. If measuring through images,
  use an unchanged camera and explain the projection; raw screen angle is not
  necessarily orbital angle.
- Repeat-seek and repeated-restore image comparisons use the same viewport,
  camera mode, selection, and standard framing. Allow 2 CSS pixels for the
  same planet's center. A changed viewport or camera invalidates the comparison.
- For tracking, sample several positions over at least 3 seconds and immediately
  after a seek. The selected center should remain within 10% of the rendered
  scene's smaller dimension from the view center. Measure the actual scene
  bounds, not a panel-inclusive page center.
- State preservation and JSON identity/order/type checks are exact, not
  percentage-based. While playing, allow only natural time advance within the
  measured timing tolerance; import/export/capture/delete must not reset or
  change playback, selection, speed, or camera.

| Check | Required actions and evidence |
| --- | --- |
| S1 | Count sun and planets; inspect size/color distinctions; measure each orbit against 12/24/48 seconds and circular motion. Record fixed phases and observed positions, not just code constants. |
| S2 | Open a fresh page in the same origin; observe 0/1x/playing/null/overview initialization and subsequent motion. Do not call ordinary load elapsed time a failure. |
| S3 | Pause during motion, wait, orbit/select while paused, resume. Record time/positions before, during, and after. |
| S4 | Exercise 0.25x, 2x, and 4x while playing and paused. Measure deltas and verify no change of current positions on speed change and no resume jump. |
| S5 | Seek 0, 30, 120 and noncanonical 17 and 73 seconds using the real slider, then revisit 30, 17, and 73 after speed/playback/selection changes. Use fractional 17.25 via valid imported snapshot under S8/T5, not hidden state mutation. Record all three angles/positions and paused time. Repeat after reload for phase stability. |
| S6 | Play across 120, wait, then seek earlier and resume. Record stop, no wrap/overshoot, and resumed motion. |
| S7 | Correlate actual loaded Three.js renderer/geometry/camera source with HTTP resource observations and a rendered orbit/zoom inspection. Record depth/occlusion and changed geometry view, not just an import or canvas tag. |
| S8 | Capture while paused and playing. Record all displayed values, three angles, and ID. Change every live field; prove saved values remain fixed and capture does not alter live state. Compare derived angles after a fractional imported time and reload. |
| T1 | Select all three planets from both canvas and list while playing and paused. Record shared highlight and selection. |
| T2 | Compare name, period, and angle with rendered selection during playback and seek. Check explicit no-selection information and absent highlight at startup and after null-selection restore. |
| T3 | Track while playing, paused, and seeking; switch to another planet. Record center offsets and synchronized target/info/highlight with preserved time/speed/playback. Try tracking with no selection and observe disable/rejection without state changes. |
| T4 | Return to overview from each tracking state. Record full-system visibility and preserved selection/time/speed/playback. |
| T5 | Restore tracking and null-selection overview snapshots after live edits and manual orbit/zoom. Restore each again at the same viewport. Record paused state, exact fields, all positions, info clearing, target, and deterministic standard framing. |
| U1 | Orbit and zoom in overview/tracking, both playing/paused. Record successful camera response without simulation changes. |
| U2 | Drag beginning on a planet with a fast move, then a slow cumulative drag of at least 30 CSS pixels over at least 20 small moves before pointer-up. Confirm no accidental selection. Exercise every DOM control and confirm no leaked scene selection/camera motion. |
| U3 | Observe state labels, discoverability, capacity/errors, and interaction feedback across the exercised states. Explain the selected qualitative anchor. |
| U4 | Capture three entries, attempt a fourth, delete a middle entry, and capture again. Record visible capacity rejection, stable values/IDs, append order, preserved remaining order, no collision, and unchanged live state. |
| U5 | At both viewports use Tab/Shift+Tab and each control's normal Enter/Space/arrows. Operate all primary DOM controls including textarea and snapshot controls. Record accessible names, visible focus, actual action outcomes, and ability to leave controls without a trap. |
| V1 | Inspect rendered scene and selection across overview/tracking and both sizes. Explain readability and the qualitative anchor. |
| V2 | Inspect typography, spacing, colors, and hierarchy at both sizes, including new panels. Explain the qualitative anchor. |
| L1 | Verify actual 1440x900 document size and scrollWidth. Exercise every primary control, including snapshot/JSON operations. Record occlusion and horizontal overflow; vertical scroll is allowed. |
| L2 | Repeat full primary-control operation at actual 390x844, not only a screenshot. Record CSS dimensions, scrolling, and whether panels obstruct controls. |
| D2 | Resolve every referenced local asset beneath the served output root; record inspected paths and absence of build/root dependencies. Exclude direct-file requirements. |
| D3 | Observe page/console/resource errors throughout all browser sequences. Record artifact-caused failures separately from infrastructure restrictions. |
| D4 | Export empty and populated lists through the labeled textarea; compare exact fields/types/order/IDs, visible planet-ID mapping, and unchanged state. Round-trip through UI. No OS picker, extra exported fields, or implicit state mutation. |
| D5 | Use the rejection and valid-import matrix below, compare full pre/post list and live state, and record visible errors. Include mixed-validity payloads, nonpreset speed 1.5, playing imports, and empty-list clearing. |
| D6 | Reload after each of capture/delete/import in the same origin/context. Compare IDs/order/values/angles and R2 live defaults. Exercise invalid persisted payload and real storage-write failure using only the boundary fixtures below, then verify notice and usable controls. |

## JSON import evidence

Keep pre/post state in working memory. Start paused to avoid confusing elapsed
time with an illegal mutation, then repeat valid import/export while playing.
Use actual textarea input and explicit import/export controls throughout.

For D5, test each invalid class separately. Include malformed JSON; root null or
array; missing root or entry fields; null/array entries; a string or unsupported
version; numeric strings and booleans; time below 0 or above 120; speed below
0.25 or above 4; overflow such as `1e400`; non-array snapshots; four entries;
duplicate IDs; empty/non-string IDs; unknown or numeric planet references;
the string `"null"`; unknown view mode; and tracking without selection.
Include a valid first entry followed by an invalid entry. Every rejection must
leave the entire list and live state unchanged and show an error. Test an invalid
later entry against a nonempty existing list, not only an empty baseline.

For valid cases, use empty and three-entry lists, inclusive time/speed endpoints,
fractional time 17.25 and speed 1.5, null-selection overview, selected overview,
and selected tracking. Preserve exact imported IDs, including case-distinct IDs
and IDs with spaces; nonempty means string length at least one, not trimmed
length. Verify array order and extra-key tolerance without allowing supplied
angles to override derived values. Export must contain only required fields.
Observe list-only replacement both paused and playing. Restore each imported
entry through its UI and compare the saved fields and derived angles.

## Persisted-data and write-failure fixtures

These fixtures apply only to D6 in the new task. They do not authorize modifying
artifact source, runtime objects, browser security settings, storage prototypes,
or getter functions. They cannot establish capture/import UI success.
The access-denial rule still stops affected operations without retry or bypass.

Use the same scheme, host, port, page, and browser storage context for persistence
and reload checks. Do not switch localhost/127.0.0.1 or a port between samples.
Identify only this app's snapshot storage key from source and its actual UI
writes. Save its exact original value or absence in working memory. Do not inspect,
clear, or delete unrelated storage. Never call `localStorage.clear()`.

For invalid saved data, use documented direct REPL page evaluation or supported
storage inspection to replace only that app-owned key with malformed JSON, then
with a schema-invalid versioned payload. Reload after each and observe empty
snapshots, a visible recovery notice, and usable live controls. Restore the exact
original key value or absence afterward, including after a failed observation.
Apply the same schema-invalid classes as D5 where needed to establish that
persistence does not bypass input validation. Record what was actually exercised.

For unavailable saving, a bounded quota fixture can exercise a real write failure
without monkeypatches. Use only an isolated review-created loopback origin and
documented storage access, with no unrelated data at risk. Keep the app's original
snapshot value in memory and start the UI with room to capture another snapshot.
Choose one unique review-owned filler key and verify it is absent before use.
With a fixed upper bound of 8 MiB of string content, fill only that key; a bounded
binary search can find the largest value storage accepts. Stop the fixture on
unexpected errors, especially an access/security denial, and continue only
independent permitted work. Do not treat quota exhaustion as an access denial
when the tool executed and storage returned the expected quota exception.

Attempt capture or an import that requires a larger saved payload using the real
UI. Confirm an actual attempted write failed rather than assuming filling was
enough. Observe visible saving-unavailable feedback and usable in-memory playback,
selection, and snapshot controls. The filler alone proves no app behavior.
Remove only the filler key, restore the original app snapshot value or absence,
and record restoration/cleanup success. Never clear another key or raise the
storage quota. Do not keep the quota fixture active for other checks.

A documented browser facility that safely makes this isolated page's storage
unavailable may substitute for the quota fixture. Never claim that a quota-write
probe also exercised read denial. If no permitted fixture can exercise a required
recovery clause, retain D6 as unverified and explain the gap, unless another D6
clause already has a confirmed failure. If the required capability is unavailable,
leave only unsupported checks unverified and continue independent checks. If it
is denied, stop the affected operation without bypass and continue only permitted
work. Preserve same-attempt verified findings in a partial record. Source code
and exception-handler presence cannot fill an observation gap.
