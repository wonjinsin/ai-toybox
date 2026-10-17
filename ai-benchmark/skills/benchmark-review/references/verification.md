# Browser and State Verification

## Permitted method

Use only connected Aside MCP's direct `repl` tool. Read its current documentation
before navigation and choose a supported, permitted method. Do not use Aside
`exec`, natural-language CLI tasks, another browser runtime, raw CDP or another
model session. Do not install/register tools during a review.

Select the transport **before the first navigation**, from current capability
documentation and verified environment constraints. Prefer the real
`output/index.html` via `file://` only when local-file navigation is explicitly
supported in the connected session. A generic `openTab(url)` API or access to
session-scoped `fs` does not establish access to host files. Do not probe
`file://` merely because its support is undocumented.

When local-file support is unavailable or unestablished, and loopback HTTP is
documented and host-permitted, start with HTTP. If a supported file navigation
fails because the browser lacks local-file capability, including
`Cannot navigate to a file URL without local file access.`, record the exact
operation/error and use this same HTTP path only when host file access and
loopback HTTP are permitted. Do not infer an OS permission failure from that
message alone. This fallback never overrides an explicit host/tool security policy.

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

Record HTTP interaction results separately from direct-file status: HTTP success
leaves D1 and `direct_file` unverified; static compatibility is D2. External CDN
resources are permitted by the task; report actual loading failures separately.

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
