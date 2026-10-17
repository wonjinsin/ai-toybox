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

An explicit host/tool access or security-policy denial triggers the
[whole-review stop rule](../SKILL.md#stop-on-access-or-verification-blockers),
as does unavailable Aside or no permitted method for a required browser check.
Do not retry denied access, change permissions or switch protocols/tools to bypass
a denial. Do not fall back to static inspection; replace the previous review with
the current blocked result. Report the exact operation and error, not an invented
OS permission diagnosis. A past review's denial is historical context, not a fresh
attempt on this target or proof of a current restriction. Never infer rendering or
interaction success from HTML, screenshots from previous reviews, or mocks.

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

Modify only the blank fixture, never the artifact DOM, CSS, script, state or
reported viewport properties. Recheck dimensions before comparing repeated
positions; a changed viewport invalidates a pixel comparison, not the artifact.
Do not use raw CDP, another browser, permission changes or an iframe to bypass
an access denial. If the permitted frame path cannot perform a required check,
apply the existing stop rule.

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
