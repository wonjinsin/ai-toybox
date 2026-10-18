# Benchmark review

## Conclusion

Status: partial.
Mandatory verdict: fail.
Published total: 20.33 / 100. Conservative lower bound. Range: 20.33 to 27.33.
Raw total: 20.33 to 27.33.
Confirmed cap: 69.
The confirmed 69 cap is existing review policy for a mandatory failure, not a calibrated quality measurement.
Missing checks (3): U2, D3, D5.
Missing reviewer metadata: None.
Format: 4&#46;0. Rubric: solar&#45;system&#45;4&#46;0.
Reviewed at: 2026&#45;10&#45;18T07:23:01&#46;065514&#43;00:00.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: high.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 0.63 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; determined score; full category denominator |
| selection | 0.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 2.40 to 3.40 | 15 | U1, U2, U3, U4, U5; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |
| visual | 2.00 | 15 | V1, V2; determined score; full category denominator |
| responsive | 0.00 | 15 | L1, L2; determined score; full category denominator |
| delivery | 2.00 to 4.00 | 10 | D2, D3, D4, D5, D6; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |

## Findings

### Confirmed failures

#### S1

Status: fail. Rating: 0. Method: http.

Action: Open the unchanged artifact over HTTP and observe the scene for 3&#46;207 wall seconds&#46; Expected: One sun and three visibly distinct planets in circular motion with 12/24/48&#45;second periods&#46; Observed: The canvas was blank and remained blank while time stayed 0&#46;0s&#46; No sun or planet was visible&#46; Orbital period, shape and phase measurements cannot be made on the absent scene; source constants are not substituted for observation&#46;

#### S2

Status: fail. Rating: 0. Method: http.

Action: Open a fresh HTTP page and measure a visible&#45;tab interval&#46; Expected: 0/1x/running/null/overview initialization followed by automatic motion&#46; Observed: Initial labels were 0&#46;0s, 1&#46;00x, Pause, No planet selected, Overview&#46; During 3&#46;207 wall seconds at 1600x1000 with visibility visible before and after, time remained 0&#46;0s and the canvas was blank&#46; This is stopped playback, not ordinary load elapsed time&#46;

#### S3

Status: fail. Rating: 0. Method: http.

Action: Pause, seek 30, change speed to 2x and resume using the UI&#46; Expected: Frozen paused state followed by motion continuing from 30 without a jump&#46; Observed: Pause/Resume labels toggled and time read 30&#46;0s, but after Resume it stayed 30&#46;0s throughout a 3&#46;206&#45;second visible&#45;tab interval&#46; No orbital positions rendered, so position freeze/jump continuity could not be compared&#46;

#### S4

Status: fail. Rating: 0. Method: http.

Action: Set 0&#46;25x, 2x and 4x using the range control while paused at 30 and again after Resume; measure the saved&#45;scenario 2x interval&#46; Expected: Time is preserved on changes and resumed time advances at the selected rate&#46; Observed: Readouts changed to 0&#46;25x/2x/4x without moving 30&#46;0s, but 2x playback stayed 30&#46;0s during 3&#46;206 wall seconds, versus 6&#46;412s expected advance&#46; The larger tolerance was 0&#46;4206s including 0&#46;1s display precision&#46; Endpoint controls were exercised in both playback&#45;label states; independent rate measurements at 0&#46;25x/4x and relative orbital speeds remain unobservable on the failed scene and are not claimed&#46;

#### S5

Status: fail. Rating: 0. Method: http.

Action: Seek 0, 30, 120, 17, 73, then 30/17/73 again after speed, playback and selection edits; select each planet for its angle&#46; Expected: Seeking pauses and immediately updates visible orbital positions with stable phase&#46; Observed: All seeks paused and showed requested time&#46; At 17 the three angle labels were 150/323/265 degrees; at 73 they were 30/83/325; repeats matched&#46; The desktop scene stayed blank, so seeking never displayed all three positions&#46; After reload, imported 17&#46;25s derived angles remained 157&#46;5/327&#46;5/266&#46;9 degrees, but rendered phase/position stability and 2&#45;pixel repeat comparisons were unavailable&#46;

#### S6

Status: fail. Rating: 0. Method: http.

Action: Seek 119, set 4x, Resume and observe one uninterrupted 3&#46;203&#45;second interval, then seek 17 and Resume&#46; Expected: Playback reaches 120 and pauses without wrap/overshoot, and an earlier seek permits resumed motion&#46; Observed: With visibility visible and document 1600x1000 before and after, time remained 119&#46;0s and the button remained Pause throughout the crossing interval&#46; It never reached/stopped at 120&#46; Seeking earlier changed time to 17&#46;0 and Resume toggled the label, but no functioning resumed animation was observed&#46; No&#45;wrap behavior cannot be separately established because the clock never crossed the boundary&#46;

#### S7

Status: fail. Rating: 0. Method: http.

Action: Load the unchanged artifact over HTTP, inspect its Three&#46;js resource/source and attempt orbit/zoom&#46; Expected: Actual Three&#46;js&#45;rendered 3D geometry changes perspective and occlusion under camera movement&#46; Observed: The Three&#46;js CDN URL appeared in resource timing and source contains PerspectiveCamera/SphereGeometry/WebGLRenderer, but desktop rendered no scene&#46; After DOM interactions, mobile showed only clipped flat orbit arcs at the bottom of the canvas, not a full 3D solar system&#46; Drags and wheel input did not produce 3D perspective/depth&#46; Import presence alone is not a pass&#46; Source line 11 acquires a 2D context before trying WebGLRenderer on the same canvas; that is a source&#45;supported cause hypothesis, not a captured exception&#46;

#### T1

Status: fail. Rating: 0. Method: http.

Action: Click the scene, then use the planet list during the saved scenario&#46; Expected: Both routes select planets and show the same visible scene highlight&#46; Observed: Clicking the blank scene selected nothing&#46; Venus, Earth and Mercury list buttons updated information and aria&#45;pressed selection, but no planet or scene highlight appeared&#46; Selecting rendered planets in playing/paused states was unavailable because there were no rendered planets&#46;

#### T2

Status: fail. Rating: 0. Method: http.

Action: Select all three planets, seek and resume, then restore a null&#45;selection overview snapshot&#46; Expected: Name/period/angle correspond to rendered selection and update during playback; null restore clears information/highlight&#46; Observed: Mercury/Venus/Earth showed periods 12/24/48 and seek&#45;derived angles&#46; Null restore showed No planet selected and all three list aria&#45;pressed values false&#46; However no rendered selected planet existed for the information to agree with, and playback at 2x stayed at 30&#46;0s with static information&#46; The visible&#45;scene and live&#45;motion portions fail despite usable text updates on seeks&#46;

#### T3

Status: fail. Rating: 0. Method: http.

Action: Enable tracking with a list selection, seek, switch Earth to Venus and try tracking without selection&#46; Expected: A visible moving/seeked selection stays near scene center and target/highlight/info change together without state changes&#46; Observed: Tracking labels and information switched together and DOM time/speed/playback labels stayed intact; no&#45;selection tracking did nothing and became disabled after an update&#46; The desktop remained blank and mobile showed only clipped flat arcs, with no selected body centered or visible target/highlight&#46; Center offsets over motion could not be measured because no tracked body rendered&#46; A tracking label does not establish genuine centering&#46;

#### T4

Status: fail. Rating: 0. Method: http.

Action: Select Venus, enable tracking, then choose Full&#45;system overview after edits&#46; Expected: The full system becomes visible while state is preserved&#46; Observed: The label changed between Tracking Venus and Overview and DOM selection/time/speed/playback labels were preserved, but the scene remained blank instead of showing the full system&#46; Visible geometry and camera state cannot be established from the labels&#46;

#### T5

Status: fail. Rating: 0. Method: http.

Action: Restore selected tracking A after unrelated live edits and orbit/zoom attempts; repeat; import/restore null&#45;overview and fractional tracking entries&#46; Expected: Paused captured state plus synchronized positions, information and deterministic standard framing&#46; Observed: Restore set text fields and cleared selected information for null overview&#46; Tracking 17&#46;25/0&#46;25/Earth restored readouts and derived angles&#46; No planet positions or usable tracking/full&#45;system framing appeared, so the required visible restored scene was absent&#46; The speed slider also stayed at its previous value 4 while the restored readout showed 0&#46;25x &#40;source restore handler does not update &#35;speed&#41;&#46; Deterministic geometry framing could not be compared on the missing scene&#46;

#### U1

Status: fail. Rating: 0. Method: http.

Action: Attempt pointer orbit and wheel zoom in overview and tracking with both Resume and Pause button states, including after snapshot restore&#46; Expected: Camera movement changes the visible geometry view while preserving simulation state&#46; Observed: Desktop remained blank through drags/zoom&#46; On 390x844, flat clipped orbit arcs near the bottom of the canvas did not change into a centered 3D view after a 30px/20&#45;move drag and wheel input&#46; Tracking and overview labels changed, but usable camera orbit/zoom did not result&#46; Actual moving&#45;playback camera behavior was unavailable because animation was stalled&#46;

#### U3

Status: fail. Rating: 2. Method: http.

Action: Inspect startup, pause/resume, selection, tracking, snapshot capacity/rejection, JSON import feedback and restore&#46; Expected: Understandable current state and feedback&#46; Observed: Headings, ID mapping, capacity and invalid&#45;JSON notices were readable, but Pause represented a simulation whose clock never advanced; tracking labels named an invisible target&#46; After restoring a 0&#46;25x snapshot, the readout was 0&#46;25x while the speed slider still showed 4&#46; Rating 2 reflects recurring contradictory state information, not mobile occlusion&#46;

#### V1

Status: fail. Rating: 0. Method: http.

Action: Inspect the scene and selection in overview/tracking at desktop and mobile, including after input and camera gestures&#46; Expected: Readable scene and selection&#46; Observed: Desktop was a blank dark canvas&#46; Mobile displayed clipped flat orbit arcs at the bottom after input, with the central marker cut by the canvas edge and no readable three&#45;planet selection&#46; Rating 0 means the scene is unusable for observation, not a penalty for implementation style&#46;

#### L1

Status: fail. Rating: 0. Method: http.

Action: Verify desktop frame dimensions and operate selection, pause/resume, ranges, tracking/overview, capture/restore/delete and JSON import/export&#46; Expected: Scene and all primary controls usable at 1440x900 without horizontal overflow or blocked controls&#46; Observed: Child innerWidth/innerHeight were 1440/900; clientWidth and scrollWidth were both 1423, secure context true&#46; Every primary DOM operation was reachable with vertical scrolling, without horizontal overflow or panel blockage&#46; But the scene remained blank, so the required usable scene at this size failed&#46; This was real embedded&#45;document verification, not a screenshot crop&#46;

#### L2

Status: fail. Rating: 0. Method: http.

Action: Repeat full primary DOM and keyboard operations at the mobile frame and inspect scene/camera interaction&#46; Expected: Scene and controls usable at 390x844 without horizontal overflow or obstruction&#46; Observed: Child dimensions were 390/844, mobile media query true, clientWidth/scrollWidth both 373, secure context true and crypto&#46;randomUUID available&#46; Selection, tracking labels, seek, speed, pause/resume labels, capture/restore/delete and JSON actions were reachable in a vertically scrolling single column&#46; The 351x360 canvas instead showed an unusable clipped flat fragment after input, not a usable full solar scene&#46; No horizontal control occlusion was observed; the scene clause caused failure&#46;

#### D6

Status: fail. Rating: 0. Method: http.

Action: Reload after capture/delete/import; test the 36 invalid persisted payload classes at the exact app&#45;owned key; exhaust only an owned quota filler, then capture and import a larger list through UI&#46; Expected: Valid persistence, visible recovery and saving&#45;unavailable feedback on a real failed write, with usable in&#45;memory controls&#46; Observed: Capture/deletion/import retained IDs/order/values/derived angles across same&#45;origin reload, but live startup never animated&#46; All invalid fixtures recovered an empty list; malformed JSON said storage unavailable, schema failures said saved data invalid&#46; Quota fixture reached 5,242,810 filler characters within the 8MiB bound and observed QuotaExceededError&#46; Capture produced a memory&#45;only snapshot and an explicit Saving unavailable notice while persisted JSON stayed empty&#46; Import then replaced the memory list but left persisted JSON empty and overwrote that notice with Snapshot list imported; live simulation unchanged&#46; Thus a real failed import write was silently presented as success without saving&#45;unavailable feedback &#40;source lines 16/20&#41;&#46; Restore/seek/list controls remained operable in memory; animation remained independently broken&#46; Filler and original app&#45;key absence were restored and observed as null; read&#45;denial was not exercised&#46;

### Missing proof

#### U2

Status: unverified. Rating: Unverified. Method: http.

A fast canvas drag and a slow cumulative 30&#45;CSS&#45;pixel drag over 20 moves were delivered; list selection stayed Earth, and DOM controls, snapshots and JSON were exercised without observed selection changes beyond their requested actions&#46; However no selectable rendered planet existed at the drag start and no functional camera framing was visible&#46; Therefore the required planet&#45;origin drag/click separation and absence of leaked camera movement cannot be established&#46; No accidental selection is inferred from source, and no pass is awarded from the absence of visible scene interaction&#46;

#### D3

Status: unverified. Rating: Unverified. Method: http.

Fresh diagnostic positive control on an owned HTTP scratch page registered console/pageerror subscriptions, then appended a main&#45;world script with console&#46;error and a deferred thrown Error&#46; Both callback arrays stayed empty after 500ms&#46; Empty listeners on the artifact are therefore not proof of error&#45;free execution&#46; Post&#45;load noninterfering error/unhandledrejection listeners recorded no later errors before reload, but miss startup and console&#46;error&#46; The CDN URL appeared in actual resource timing; the absent scene is observed elsewhere, but no qualifying diagnostic error was captured to establish D3&#46; No pass is inferred from source or isolated&#45;world globals&#46; A second scratch control set rlExecuted=yes and rlThrowReached=yes in main&#45;world DOM markers, confirming execution while both event arrays still stayed empty&#46;

#### D5

Status: unverified. Rating: Unverified. Method: http.

Thirty&#45;six separate invalid UI imports against a nonempty three&#45;entry list each showed Invalid snapshot JSON; list and simulation unchanged, with exact pre/post list and DOM live&#45;state equality&#46; Covered malformed/root null/array, missing root and every entry field, null/array entries, version type/value, numeric strings/booleans, time/speed ranges, overflow 1e400, non&#45;array snapshots, four entries, duplicates, empty/non&#45;string IDs, unknown/numeric/empty/string&#45;null planets, unknown mode, null tracking and mixed valid&#45;prefix/invalid&#45;suffix atomicity&#46; Valid empty/two/three&#45;entry imports preserved case&#45;distinct and whitespace IDs, endpoints, 17&#46;25 time and 1&#46;5 speed, ignored extra keys and supplied angles, and restored entries via UI&#46; Paused and running&#45;labelled imports preserved observed fields&#46; However preservation of an actually moving simulation and usable camera framing could not be verified because playback and rendering failed before import&#46; The full grouped D5 pass is withheld; no validation failure was found&#46;

### Verified passes

#### S8

Status: pass. Rating: 5. Method: http.

Action: Capture A at paused time 30, speed 2x, Venus/tracking; change time, speed, selection, mode and playback label; capture another in the running&#45;labelled state; export/import/reload and restore fractional imported snapshots&#46; Expected: Unique stable IDs, immutable captured fields, all three correctly derived degree angles, and no capture&#45;induced live&#45;state change&#46; Observed: A retained obs&#45;8c30497c&#45;4569&#45;4913&#45;a03b&#45;0b4c6d105b56 and 30/2/venus/tracking with angles 180&#46;0/158&#46;8/2&#46;5 after edits&#46; A running&#45;labelled capture at 0/1/null/overview retained its ID and angles 0&#46;0/68&#46;8/137&#46;5 across reload&#46; All visible live fields were unchanged by capture&#46; Fractional time 17&#46;25 exported exactly and displayed 157&#46;5/327&#46;5/266&#46;9 after import/reload, ignoring supplied angles&#46; Planet IDs mercury/venus/earth remained stable&#46; This pass establishes snapshot data behavior; the independent animation failure means running&#45;labelled captures had no moving scene&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Capture three distinct entries, attempt a fourth, delete the middle entry and capture again while checking every list/live field&#46; Expected: At most three entries, visible capacity rejection, append order, stable remaining values/IDs and no live&#45;state mutation&#46; Observed: Fourth capture showed Snapshot capacity reached &#40;3&#41;&#46; Delete one to capture another, with the same three entries&#46; Deleting obs&#45;2a200bb3&#45;95dd&#45;44ad&#45;8fd7&#45;ea2e8200e2c6 kept A and obs&#45;72467f3d&#45;7369&#45;4126&#45;9b60&#45;70adb89b7705 in order&#46; The new obs&#45;66028cba&#45;91bf&#45;4bbc&#45;95ec&#45;5732636fa94d appended without collision&#46; Time 17&#46;0, speed 2x, Mercury selection, Overview and Resume state stayed unchanged; retained IDs/values/angles stayed exact&#46;

#### U5

Status: pass. Rating: 5. Method: http.

Action: At secure 1440x900 and 390x844 iframe viewports, use Tab/Shift&#43;Tab, Enter on native buttons, arrows on both ranges, and keyboard textarea editing to exercise all primary DOM controls&#46; Expected: Named controls, visible focus, working native activation and no focus trap&#46; Observed: At each size, Enter selected Mercury/Venus/Earth, toggled Pause/Resume state, switched Tracking/Overview, captured/restored/deleted a snapshot, exported JSON and imported the keyboard&#45;entered payload&#46; ArrowRight changed speed 1 to 1&#46;25 and time 0 to 0&#46;1 with seek pause&#46; The complete Tab traversal reached speed, time, all planets, tracking, overview, capture, restore, delete, textarea, export, import, exited to body, then returned to play&#46; Every primary focused control matched :focus&#45;visible with solid yellow outline; meaningful labels and snapshot&#45;specific names were observed&#46; Shift&#43;Tab returned export to textarea&#46; This keyboard&#45;operability result does not certify the independently failed animation or camera rendering&#46;

#### V2

Status: pass. Rating: 4. Method: http.

Action: Inspect the visible interface at 1440x900 and 390x844, including snapshot and JSON states&#46; Expected: Coherent legible typography, spacing, colors and hierarchy&#46; Observed: Dark navy panels, cyan section headings, muted secondary text, consistent rounded controls and strong yellow keyboard focus formed a clear hierarchy&#46; The desktop two&#45;column arrangement became a single vertically scrollable mobile column; snapshot values wrapped and JSON used a distinct monospace area&#46; Rating 4 reflects polished ordinary panel presentation with dense small snapshot/ID text as a minor limitation&#46; Scene absence is scored under V1, not used to lower this independent panel rating&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspect all files under output and every asset reference in index&#46;html&#46; Expected: Referenced local assets exist beneath the output HTTP root without build or repository dependencies&#46; Observed: output contains only index&#46;html; all CSS and application JavaScript are inline&#46; The sole script dependency is the permitted https://cdn&#46;jsdelivr&#46;net/npm/three@0&#46;160&#46;0/build/three&#46;min&#46;js CDN URL&#46; There are no referenced local dependency files or build steps&#46; Direct&#45;file execution is excluded&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Export empty, captured and imported lists through the labeled textarea; round&#45;trip via explicit import/export controls while paused and in the running&#45;labelled state&#46; Expected: Exactly version and snapshots with each required id/time/speed/selectedPlanet/viewMode field, exact types/order/IDs, visible name&#45;ID mapping, no OS picker and unchanged list/live state&#46; Observed: Empty export parsed as &#123;version:1,snapshots:&#91;&#93;&#125;&#46; Captured three&#45;entry order and IDs matched the list&#46; Imported Case/case/space IDs, endpoints 0/120 and 0&#46;25/4, and 17&#46;25/1&#46;5 exported exactly with only required fields; angles, extra keys, playback and camera keys were omitted&#46; Visible mapping Mercury=mercury, Venus=venus, Earth=earth was present&#46; Export preserved all observed live fields and list values&#46; The running&#45;labelled state remained Pause at time 30/2x; actual animation was independently stalled, not claimed to work by this export pass&#46;

## Verification and limits

Browser method: http.
Viewports: 1600x1000, 1440x900, 390x844.
Coverage: 6 passed, 18 failed, 3 unverified.

### Notes

- User&#45;requested scoring&#45;only migration to solar&#45;system&#45;4&#46;0 at 2026&#45;10&#45;18T08:19:55&#46;883Z&#46; All check observations, ratings, statuses, methods, reviewer identity, and observation timestamp are unchanged&#46; Numeric scores are guaranteed lower endpoints with explicit upper endpoints, not newly observed passes or failures&#46; This recalculation is not a fresh browser review&#46;
- Reviewer model comes from current host metadata&#46; Effort high comes from applicable &#46;omp/agents/benchmark&#45;astra&#45;review&#46;md thinking&#45;level configuration, not independent execution telemetry&#46; Timestamp comes from the host UTC clock&#46;
- Preregistered before browser observation&#46; Speed intervals at least 3 wall seconds, tolerance max&#40;0&#46;15 simulation seconds, 5% expected delta&#41; plus 0&#46;1s display increment&#46; Paused time tolerance 0&#46;1s and rounded angles 1 degree; predicted angles 2 degrees plus rounding&#46; Repeat&#45;seek/restore planet centers 2 CSS pixels at identical viewport/framing&#46; Tracking center within 10% of smaller scene dimension&#46; JSON identity, order, types and non&#45;time state preservation exact&#46;
- Only the explicitly assigned low run is reviewed&#46; Its saved prompt matches all fifteen requirements and eight&#45;step plus six additional sequences of solar&#45;system&#45;observatory&#45;2&#46; No earlier review is used as evidence&#46;
- Applied benchmark&#45;review and mandatory aside&#45;browser with direct MCP REPL only&#46; User&#45;selected poteto&#45;mode and its Prove It Works principle changed the review decision to require fresh actual artifact observations and validated diagnostic transport, not prior clues or source&#45;only proof&#46; No descendants, model launches, benchmark&#45;record, code changes or general tests/build/lint/formatters&#46;
- Saved acceptance scenario executed in order&#46; Startup showed a blank static scene; scene click could not select a planet; tracking without selection did nothing&#46; Pause and seek 30 updated DOM time; 2x preserved it&#46; Resume stayed at 30&#46;0 over 3&#46;206 wall seconds instead of approximately 6&#46;412 simulation seconds&#46; Venus list selection updated text but no scene; Overview remained blank; repeat seek 30 updated text with no positions to compare&#46;
- Source inspection suggests the startup defect is canvas context ownership: output/index&#46;html line 11 acquires a 2D context before constructing THREE&#46;WebGLRenderer on that same canvas, and line 22 calls init without a catch&#46; This is a source&#45;supported explanation, not a captured startup exception&#46; Actual evidence is the blank scene and frozen clock&#46; No runtime patch or alternate renderer was injected&#46;
- All eight saved acceptance steps and six additional sequences were attempted in order before expanded matrix probes&#46; Additional sequence 1 retained captured values and restored text state but not a rendered scene&#46; Sequence 2 verified capacity, deletion, IDs and export&#46; Sequence 3 exercised invalid import atomicity&#46; Sequence 4 verified valid/null/fractional imports and text restore; moving playback/camera preservation was unavailable&#46; Sequence 5 verified same&#45;origin persistence and recovery&#46; Sequence 6 exercised every primary keyboard DOM control at both required secure frame sizes&#46;
- Actual required CSS viewports were 1440x900 and 390x844 in fresh HTTP wrapper iframes at 127&#46;0&#46;0&#46;1:9221 loading unchanged output at 127&#46;0&#46;0&#46;1:9211&#46; This verifies embedded CSS layout and keyboard/pointer interactions, not top&#45;level viewport sizing or touch hardware&#46; The initial top&#45;level observation was 1600x1000&#46; Frame widths had ordinary 17px vertical scrollbars and no horizontal overflow&#46; Both secure contexts exposed crypto&#46;randomUUID; no insecure about:blank fixture was used&#46;
- Diagnostic positive control was repeated with DOM markers proving main&#45;world script execution and reaching the deferred throw&#46; Markers were yes, while console/pageerror callback arrays remained empty after 500ms&#46; D3 remains unverified&#46; No raw CDP, alternate browser, permission changes, artifact code/DOM/state patching or library&#45;global inference was used&#46; Noninterfering post&#45;load error listeners covered only later error/unhandledrejection events and were removed by navigation or explicit cleanup&#46;
- Unobserved clauses caused by artifact failure are identified per check&#46; The absent 3D scene prevents orbital image measurements, actual centered tracking, deterministic camera comparisons and planet&#45;origin drag separation&#46; Startup and 2x playback were explicitly measured in visible tabs; endpoint crossing at 4x also stalled over 3&#46;203 seconds&#46; UI endpoints were exercised without claiming additional exact 0&#46;25x/4x rate proof&#46; The browser was held for parent&#45;reserved timing windows; each reported interval was one uninterrupted REPL call using the explicit owned tab&#46;
- Storage fixtures touched only solar&#45;observatory&#45;v2 and the previously absent owned rl&#45;quota&#45;20261018&#45;9211 filler key on the isolated review origin&#46; The original app key was absent&#46; After invalid&#45;data and quota tests, both keys were observed absent again&#46; No global clear or unrelated storage access occurred&#46; Write&#45;quota failure does not establish storage read&#45;denial behavior&#46;
- Cleanup completed before final publication&#46; The top&#45;level artifact tab, both viewport tabs and both diagnostic scratch tabs were closed&#46; Five tool&#45;generated transient screenshot files were removed&#46; Both owned servers rl&#45;output and rl&#45;viewport were stopped&#46; No screenshot, log, evidence directory or fixture remains in the run&#46; Source preservation is checked separately using working&#45;memory before/after digests; no digest is included in this record&#46;
