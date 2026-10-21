# Browser and State Verification

## Permitted method

Use only ego-browser (Ego Lite) through `ego-browser nodejs`. Run
`ego-browser skill` and read the complete installed guide before navigation.
Use its documented TaskSpace and Page APIs; do not assume Playwright APIs.
Do not use Aside MCP, natural-language browser tasks, another browser runtime,
or another model session. Do not install/register/upgrade tools during a review.
Documented `page.cdp()` is permitted for viewport emulation and diagnostic
events, never to change security settings or bypass an access denial.

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
startup. Create one uniquely named TaskSpace for the review and navigate its
initial `p1`, rather than creating another tab:

```sh
ego-browser nodejs -e 'const task = await taskSpace("review-unique-run"); console.log({spaceId: task.spaceId}); const page = task.page("p1"); await page.goto("http://127.0.0.1:8765/index.html"); console.log(await page.snapshot());'
```

Replace the example space name with a unique review name. Resume its printed
numeric ID in later invocations; JavaScript variables do not persist between
commands. If the port is occupied, choose another available port and use it in
both places; do not stop an unrelated process. Retain the server handle/PID.

Record actual HTTP rendering, interactions and resource-loading results.
External CDN resources are permitted; distinguish observed loading failures
from environment/network restrictions.

An explicit host/tool access or security-policy denial stops the affected
operation under the [continuation and access rules](../SKILL.md#continue-independent-permitted-checks).
Do not retry denied access, change permissions or switch protocols/tools to bypass
a denial. Continue only independent permitted work. Missing ego-browser or a required
browser capability leaves affected checks unverified; it does not prohibit
permitted static inspection for D2 or erase observed findings. A confirmed
functional failure is fail; continue other checks. Finalize partial when usable
findings remain, blocked when none do. Report the exact operation and error, not
an invented OS permission diagnosis. A past review's denial is historical context,
not a fresh attempt on this target or proof of a current restriction. Never infer
rendering or interaction success from HTML, old screenshots, or mocks.

Keep all review pages in that one TaskSpace; do not inspect unrelated tabs,
profiles, or browsing memory. On successful completion call
`await task.finish({ keep: [] })` exactly once. On an error close only owned
pages where still permitted; do not call `finish()` on a stopped/error task.
User takeover or a browser permission prompt requires handoff, not retries or
cleanup that overrides user control. Stop owned servers on exit and disclose
any cleanup gap. Inspect screenshots transiently and delete their temporary
files afterward; retain no images, videos, logs, or evidence directory.

## Verify real document viewports

Use the installed guide's documented `page.cdp()` escape hatch for CSS viewport
emulation. In the existing TaskSpace, apply the metrics before navigation:

```js
const task = await taskSpace(reviewSpaceId);
const page = task.page("p1");
await page.cdp("Emulation.setDeviceMetricsOverride", {
  width: 390, height: 844, deviceScaleFactor: 1, mobile: false,
});
await page.goto("http://127.0.0.1:8765/index.html");
console.log(await page.evaluate(() => ({
  width: innerWidth, height: innerHeight, secure: isSecureContext,
  clientWidth: document.documentElement.clientWidth,
  scrollWidth: document.documentElement.scrollWidth,
  mobileQuery: matchMedia("(max-width: 600px)").matches,
})));
console.log(await page.snapshot());
```

Supply the actual numeric space ID and server URL. Require actual 390x844 and
1440x900 document dimensions, repeating with desktop metrics. Apply and verify
metrics in each `ego-browser nodejs` invocation that exercises that size; the
override is not guaranteed to survive the command's CDP session.
Exercise real controls through `page.click`, `page.fill`, `page.press`, and the
documented mouse/keyboard APIs, then inspect a transient screenshot at each size.
If automatic scrolling leaves a control offscreen under emulation, inspect its
rectangle, scroll it into view with `page.evaluate()`, then use normal Page
input. Do not replace input with DOM `.click()` or state mutation.
CSS emulation is not touch-device proof. Never patch artifact DOM, styles,
scripts, runtime state, or reported viewport properties.

Recheck dimensions before comparing repeated positions. Clear the override
with `await page.cdp("Emulation.clearDeviceMetricsOverride")` after the checks.
A missing capability leaves affected checks unverified; a denial stops the
affected operation without fallback to another tool, protocol, or frame.
Preserve independent findings and report actual document sizes in the review.

## Focused regression automation

For artifacts matching `assets/probe-terra-xhigh.json`, use the bundled probe for
the repeated overview, persistence, malformed-storage and quota scenarios. Its
control selectors are explicit, not a universal adapter for arbitrary generated
pages. A different interface needs its own observed control mapping before use.
The probe does not replace the saved acceptance scenario or the complete rubric.

`scripts/prepare-probe.mjs` prints JSON with a `code` field for
`ego-browser nodejs`; it does not launch a browser. Supply the existing review
TaskSpace ID with `--space`; the generated code reuses `p1`, never creates a
second TaskSpace, and sets and verifies the requested top-level CSS viewport.
The host must permit supplying the owned script/profile as browser input.
Do not use it to bypass an access denial. Start the output-only server above
and a separate reviewer-owned diagnostic scratch origin:

```sh
node "<skill-root>/scripts/serve-viewport.mjs" 8766
```

This fixture server reads no repository files. Its root without query parameters
is used only as the diagnostic scratch page, not as an artifact iframe.
Then prepare one scenario:

```sh
node skills/benchmark-review/scripts/prepare-probe.mjs \
  --artifact "<run>/output" \
  --url http://127.0.0.1:8765/index.html \
  --fixture http://127.0.0.1:8766/ \
  --space 7 --size mobile --scenario quota
```

Replace `7` with the review's printed TaskSpace ID. Save the emitted `code`
unchanged to a temporary `.mjs` outside the run, then execute it:

```sh
ego-browser nodejs -e 'await import("file:///absolute/path/to/probe-round-01.mjs")'
```

This imports agent-owned Node.js code, not browser `file://` navigation.
Use a new script per round and delete it afterward. Supported scenarios are
`overview`, `persistence`, `invalid-storage`, `quota` and `diagnostics`.
`overview` also accepts `--time 5.2` and `--playing false`. Use `desktop` for
1440x900 or `mobile` for 390x844. The preparer rejects a served artifact that
differs from the selected directory. The browser checks served file hashes before
and after each scenario. These are endpoint byte checks on the owned stable
server, not captured navigation-response bodies.

The `OBSERVATORY_PROBE_JSON=` line contains the report. Read its observations,
not the fact that the script returned. Setup must reach the requested state.
Restoration is tested only after changing live state. Reloads use `page.reload()`
and wait for the actual capture control. Storage fixtures restore the exact
original raw value and remove their own quota key without a later reload changing
that restoration. The preparer closes its diagnostic scratch page and clears
viewport emulation in `finally`; it leaves `p1` in the review TaskSpace for
continued checks. An `OBSERVATORY_SCREENSHOT=` line names a temporary image to
inspect and delete. Finish the TaskSpace once after the review and stop owned
servers; the preparer does not finish the shared space.

Framing stays `unverified` in automated output. Inspect the actual rendered
screenshot separately. Diagnostics test both deliberate console and exception
signals before trusting an empty stream; an unavailable signal stays unverified.
A passing transport test alone does not establish an error-free artifact.
Do not promote any focused pass to an entire T4, D6 or D3 pass without the
remaining required observations. During a normal review, keep only the approved
two review files and incorporate supported findings into their evidence.

### Maintenance experiment

A user-requested repair comparison can retain reports outside `runs/`. The
recorded experiment is
`experiments/observatory-improvement/terra-xhigh/manifest.json`. Its candidate is
an engineering derivative of one saved artifact, not a new model generation.
To check its existing evidence and independently recalculate historical policy
effects, run from the repository root:

```sh
node skills/benchmark-review/scripts/compare-probes.mjs \
  experiments/observatory-improvement/terra-xhigh/manifest.json
```

This command does not rerun a browser. It checks the original inventory and hashes,
the current candidate and probe identities, served-byte evidence, every required
paired observation, actual viewports, and storage cleanup. It rejects stale visual
judgments by their exact report-file SHA-256. Missing observations cannot shrink
the comparison. A remaining failure, lost coverage, unverified required framing,
or absence of any repaired scenario causes a nonzero exit.

Saved experiment reports retain their original runner identity. After a transport
or probe change, they are historical evidence, not verification of the current
runner. Do not rewrite their hashes to make the comparison pass; a fresh paired
comparison requires explicit scope and new observations.

For fresh evidence, run all eight cases on both outputs. Use separate owned
output-only origins and the same top-level CSS viewport metrics:

- `mobile` with `quota`, `persistence`, `invalid-storage` and `diagnostics`.
- `mobile` with `overview --time 5.2 --playing false`.
- `desktop` with `overview --time 73 --playing false`.
- `desktop` with `overview --time 120 --playing false`.
- `mobile` with `overview --time 17 --playing true`.

Save only the parsed report JSON under the manifest's `evidence/` directory,
using `baseline-` or `candidate-`, the size, and the scenario. Overview filenames
also end in `-<time>-<playing>.json`; other filenames end in `.json`.
After inspecting each new overview screenshot, update its separate manifest
observation and `report_sha256`, obtainable with `shasum -a 256 <report.json>`.
Never update a visual hash without repeating the visual judgment.

Exit zero certifies only the listed repair scenarios. D3 transport can remain
unverified and is disclosed separately. The output's `full_benchmark_score` is
null. Policy-only projections use existing saved observations and cannot serve
as a new score for the candidate. No direct-file compatibility, touch-device
coverage, storage read-denial recovery or complete 27-check review is claimed.


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

Measure rates with only the review's target scene running. Close review-created
duplicate scene and viewport tabs before timing, and serialize timing across
reviews that share a browser. Holding other tool calls is insufficient while
their WebGL scenes still render. Keep each before/read, timed wait, and after/read
in one uninterrupted `ego-browser nodejs` invocation. Record actual wall time and visibility. If a
sample was contaminated, retain it with its context and repeat under controlled
conditions without widening the tolerance. Never close unrelated user tabs.

Use DOM/console/resource observations as supporting evidence, not replacements
for rendered motion and actual user input. Separate observed app defects from
infrastructure restrictions. Explain each failure with control actions, expected
behavior, observed behavior, and source locations when inspection establishes a
cause. Source references are supplementary; no retained screenshot is required.

### Validate diagnostic transport

Before trusting console or page-error events, exercise them on a reviewer-owned
scratch page with a deliberate console error and thrown exception. Enable
`Runtime` through documented `page.cdp("Runtime.enable")`, then read and clear
the buffered events with `page.events()`. Require both `Runtime.consoleAPICalled`
and `Runtime.exceptionThrown` signals before trusting empty arrays. For the
artifact, enable diagnostics before navigation and collect throughout the
sequences, not only after loading. A failed positive control leaves diagnostic
coverage unverified; do not switch runtimes or use undocumented transport.

Non-interfering `error` and `unhandledrejection` listeners may observe later
events. Do not prevent default handling, wrap application functions, or change
application state. Remove listeners after use and reattach after navigation.
Record their coverage limits. Post-load listeners miss startup errors and do
not capture ordinary `console.error` calls. A captured qualifying failure can
fail D3. Missing required diagnostic coverage leaves D3 unverified rather than
inventing a pass from empty arrays. Continue all independent functional checks.

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
| U5 | At both viewports use Tab/Shift+Tab and a normal activation key for each control, such as Enter for native buttons, arrows for ranges, and typeahead plus Enter for native selects. R15 says Enter, Space, **or** arrows as appropriate, not every equivalent key on every control. Operate all primary DOM controls including textarea and snapshot controls. Record accessible names, visible focus, actual action outcomes, and ability to leave controls without a trap. A transport failure of an unused equivalent key does not invalidate demonstrated normal keyboard operation. |
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

For invalid saved data, use documented `page.evaluate()` or supported
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
