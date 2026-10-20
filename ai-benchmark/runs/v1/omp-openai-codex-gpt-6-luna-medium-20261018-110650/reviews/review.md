# Benchmark review

## Conclusion

Status: partial.
Mandatory verdict: fail.
Quality total: 78.70 / 100. Conservative lower bound. Range: 78.70 to 80.70.
Quality is uncapped. The mandatory verdict is independent of the quality score.
Mandatory failures (3): T4, T5, D6.
Mandatory unverified (0): None.
Missing checks (1): D3.
Missing reviewer metadata: reviewer.effort.
Format: 5&#46;0. Rubric: solar&#45;system&#45;5&#46;0.
Reviewed at: 2026&#45;10&#45;20T12:43:16&#46;102Z.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: Unavailable.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; determined score; full category denominator |
| selection | 3.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 4.40 | 15 | U1, U2, U3, U4, U5; determined score; full category denominator |
| visual | 2.50 | 15 | V1, V2; determined score; full category denominator |
| responsive | 5.00 | 15 | L1, L2; determined score; full category denominator |
| delivery | 3.00 to 4.00 | 10 | D2, D3, D4, D5, D6; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |

## Findings

### Confirmed failures

#### T4

Status: fail. Rating: 0. Method: http.

Action: Return from tracking to default overview at desktop size with snapshots present&#46; Expected: Full&#45;system overview with preserved live fields&#46; Observed: Time/speed/selection/playback are preserved, but the outer orbital path is visibly cut by both horizontal canvas boundaries rather than showing the full system&#46; User zoom&#45;out can fit it, but the overview action alone does not&#46; The failure is framing, not permitted vertical page scrolling&#46;

#### T5

Status: fail. Rating: 0. Method: http.

Action: Restore the captured 2x tracking snapshot after changing the slider to 0&#46;25, and restore the imported 1&#46;5x null overview after unrelated edits&#46; Expected: All displayed/restored state agrees and playback pauses with standard camera framing&#46; Observed: Time/selection/mode/pause and repeated camera framing restore, but the speed slider retains 0&#46;25 &#40;later 0&#46;95/3&#46;95&#41; while the adjacent label reports the restored 2&#46;00x or 1&#46;50x&#46; Source update&#40;&#41; never synchronizes the speed input&#46;

#### U3

Status: fail. Rating: 2. Method: http.

Action: Inspect live, snapshot and JSON feedback throughout restore and storage&#45;failure sequences&#46; Expected: Consistent understandable state&#46; Observed: Labeled sections and validation messages are clear, but restored speed label disagrees with the slider and failed import saving reports success&#46; These repeated feedback contradictions match qualitative anchor 2&#46;

#### V1

Status: fail. Rating: 2. Method: http.

Action: Inspect default overview, tracking and zoomed views at both sizes&#46; Expected: Consistently readable scene and selection&#46; Observed: Spherical bodies and selected brightening are readable, especially tracking, but default overview repeatedly clips outer orbital geometry and the smallest planet is visually weak&#46; The recurring observation ambiguity matches anchor 2; user zoom&#45;out improves it&#46;

#### D6

Status: fail. Rating: 0. Method: http.

Action: Reload after import/delete/capture; replace owned persisted data with malformed/version&#45;invalid values; fill a unique quota key and perform UI writes&#46; Expected: Persistent data or visible saving&#45;unavailable feedback with usable in&#45;memory controls&#46; Observed: All normal mutations and derived text survive reload; invalid storage recovers visibly to an empty list&#46; A real failed capture write warns correctly, but a larger valid import fails to change durable storage while replacing the warning with Snapshot list imported; live simulation unchanged&#46; Memory&#45;only import is incorrectly reported as success&#46; Original raw storage was restored and quota key removed&#46;

### Missing proof

#### D3

Status: unverified. Rating: Unverified. Method: http.

Known console&#46;error and asynchronous thrown&#45;exception markers on the reviewer&#45;owned wrapper produced no events from direct REPL console/pageerror subscriptions&#46; Required diagnostic coverage is therefore unavailable&#46; Rendering and source inspection cannot substitute for a clean&#45;console claim&#46;

### Verified passes

#### S1

Status: pass. Rating: 5. Method: http.

Action: Observe the loaded scene in motion, seek 0/30/120/17/73 and select each rendered body&#46; Expected: One sun, three size/color&#45;distinct planets and fixed 12/24/48&#45;second circular phases&#46; Observed: Gold sun, small tan Mercury, blue Earth and larger orange Jupiter render with curved perspective orbits&#46; At t=30 native scene picks identify Mercury/Earth/Jupiter at their predicted 180/90/225&#45;degree positions; t=17&#46;25 displays 157&#46;5/258&#46;8/129&#46;4 degrees&#46; Visible positions change with time and repeat after reload, matching the inspected fixed&#45;phase equations&#46;

#### S2

Status: pass. Rating: 5. Method: http.

Action: Open and reload the original HTTP page and observe subsequent motion&#46; Expected: Time starts at zero, 1x, playing, null selection and overview&#46; Observed: Reload shows t=0&#46;4 after ordinary load elapsed time, speed 1, Pause, no selection and overview; the fresh startup sample advances 5&#46;2 to 8&#46;3 over about 3&#46;1 seconds with visible orbital motion&#46;

#### S3

Status: pass. Rating: 5. Method: http.

Action: Pause at 30, wait 0&#46;5 seconds, use selection/camera and resume&#46; Expected: Frozen paused time/positions and continuous resumed motion&#46; Observed: Paused time remains 30 with unchanged angles; native selections and camera orbit still work&#46; At 2x the resumed sample advances 30&#46;5 to 36&#46;6 over 3&#46;1038 seconds without a reset or jump&#46;

#### S4

Status: pass. Rating: 5. Method: http.

Action: Set 0&#46;25x, 2x and 4x while paused and playing, then measure visible&#45;document playback away from endpoints&#46; Expected: Preserved position on setting and proportional advance&#46; Observed: Paused samples stay t=30&#46; The measured deltas are 0&#46;7 over 3&#46;1025s at 0&#46;25x, 6&#46;1 over 3&#46;1038s at 2x, and 12&#46;4 over 3&#46;1041s at 4x, all inside preregistered rounding/scheduling tolerance&#46; Playing changes preserve playback and current orbital progression&#46;

#### S5

Status: pass. Rating: 5. Method: http.

Action: Seek 0,30,120,17,73 and repeat 30/17/73 after state edits and reload with the same mobile viewport and camera framing&#46; Expected: Immediate paused positions with stable phases&#46; Observed: Each seek pauses at the requested time&#46; Decoded 335x324 scene&#45;pixel regions for 30,17,73 have zero differing color channels after reload&#46; PNG encodings differed, but decoded scene pixels did not&#46; All three bodies were visible after ordinary user zoom&#45;out&#46;

#### S6

Status: pass. Rating: 5. Method: http.

Action: Resume from 119 at 4x, wait 3&#46;1 seconds and another 0&#46;6 seconds, then seek 17 and resume&#46; Expected: Stop exactly at 120 without wrap and resume from earlier time&#46; Observed: Both endpoint reads are 120 with Resume; the earlier&#45;time run advances to 19&#46;9 with Pause&#46;

#### S7

Status: pass. Rating: 5. Method: http.

Action: Inspect CDN&#45;loaded WebGLRenderer, SphereGeometry, PerspectiveCamera and orbit geometry, then perform native drag and wheel actions&#46; Expected: Actual Three&#46;js 3D rendering with changed depth/occlusion/perspective&#46; Observed: Lit spherical surfaces, changing orbit ellipses and sun/planet occlusions respond to camera movement&#46; A fresh tracking restore returns the original perspective&#46; Resource timing records the Three&#46;js CDN request; cross&#45;origin responseStatus=0 is not treated as a load failure&#46;

#### S8

Status: pass. Rating: 5. Method: http.

Action: Capture paused t=30/2x/Mercury/tracking, edit every live field, capture while playing, round&#45;trip and reload, and import extra angle data&#46; Expected: Immutable saved fields, three derived angles and stable unique IDs&#46; Observed: Snapshot A remains 30/2/mercury/tracking with angles 180/90/225 after edits&#46; IDs and complete list text survive reload&#46; Playing capture appends without resetting live state&#46; Imported t=0 with angles=&#91;999&#93; displays 0/0/0, while fractional 17&#46;25 derives 157&#46;5/258&#46;8/129&#46;4&#46;

#### T1

Status: pass. Rating: 5. Method: http.

Action: Select Mercury, Earth and Jupiter through the native canvas and labeled select while paused and playing&#46; Expected: Shared selection and visible highlight&#46; Observed: All six scene/list routes identify the intended planet; playing canvas picks at about 30&#46;1 return the expected IDs and updated information&#46; Selected bodies brighten, including paused selection&#46; Locator offset input was inconclusive; calibrated native mouse input established the result&#46;

#### T2

Status: pass. Rating: 5. Method: http.

Action: Select each body, observe playback/seek, and restore null&#45;selection overview after tracking&#46; Expected: Name, period and angle agree with rendered selection, and null restoration clears information/highlight&#46; Observed: t=30 shows Mercury 12s/180 degrees, Earth 24s/90 degrees and Jupiter 48s/225 degrees at their actual picked positions&#46; Playback angles advance; null restore shows No planet selected and removes selected&#45;body brightening&#46;

#### T3

Status: pass. Rating: 5. Method: http.

Action: Track Mercury during a measured playback interval, seek 17/73 and switch to Earth; inspect null&#45;selection tracking&#46; Expected: Centered synchronized target, preserved live state and disabled/rejected tracking without selection&#46; Observed: Mercury remains at the rendered scene center across observed moving/seek positions; the 390x844 screenshots place the selected center approximately at scene&#45;image &#40;168,162&#41;, within the 10% tolerance&#46; Switching to Earth centers/brightens Earth and updates information without changing paused t=73&#46; Null&#45;selection tracking is disabled&#46;

#### U1

Status: pass. Rating: 5. Method: http.

Action: Native camera drags and wheel zoom in overview/tracking while paused and playing&#46; Expected: Operable camera in all four states without simulation reset&#46; Observed: All four rendered views change orientation and scale; paused time remains 30, and playing time advances naturally with selection/mode preserved&#46; Wheel delivery used the actual canvas wheel handler; pointer drag used native mouse input&#46;

#### U2

Status: pass. Rating: 5. Method: http.

Action: Begin fast and cumulative 30&#45;CSS&#45;pixel/20&#45;move drags on Mercury while Earth is selected, then exercise DOM controls&#46; Expected: Drags do not become selection clicks or leak DOM interactions into camera movement&#46; Observed: Both drags retain Earth and t=30 while moving the camera&#46; Primary DOM actions change only their intended state; a valid import preserves the scene exactly after aligning the page scroll offset&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Capture three, reject a fourth, delete the middle and capture again&#46; Expected: Visible capacity warning, immutable existing entries, unchanged live state and append&#45;only unique replacement&#46; Observed: Capacity reached message appears with byte&#45;equivalent exported list; middle deletion retains the first/third IDs and values, and a fresh distinct ID appends&#46; Live time/selection/mode/playback remain unchanged&#46;

#### U5

Status: pass. Rating: 5. Method: http.

Action: At both required sizes operate selection by typeahead plus Enter, buttons by Enter, ranges by arrows, textarea editing/import/export, and Tab/Shift&#43;Tab traversal&#46; Expected: Named visible&#45;focus controls operate without trapping focus&#46; Observed: Typeahead selects Mercury at both sizes; all primary button/range/snapshot/JSON actions produce corresponding state changes&#46; Focus outlines are solid and visible; traversal leaves and returns to the frame&#46; Home/ArrowDown&#45;only select attempts were inconclusive, not counted as failures because ordinary typeahead worked&#46;

#### V2

Status: pass. Rating: 3. Method: http.

Action: Inspect typography, spacing and simulation/planet/snapshot/JSON panels at both sizes&#46; Expected: Coherent legible hierarchy&#46; Observed: Dark bordered cards, consistent text sizes and well&#45;separated controls remain legible, including wrapped snapshot IDs and angles&#46; This is clear ordinary execution, anchor 3&#46;

#### L1

Status: pass. Rating: 5. Method: http.

Action: Exercise every primary control at measured 1440x900, including keyboard snapshot/JSON actions&#46; Expected: Usable scene/controls without horizontal scrolling or panel occlusion&#46; Observed: Document clientWidth and scrollWidth are both 1423&#46; Controls remain reachable with permitted vertical scrolling, and the camera/selection scene is operable&#46; Default full&#45;system framing is separately failed under T4&#46;

#### L2

Status: pass. Rating: 5. Method: http.

Action: Repeat full primary&#45;control operation at measured 390x844&#46; Expected: Usable scene and all controls without horizontal overflow&#46; Observed: Client/scroll widths both equal 373; stacked panels remain reachable with vertical scrolling, and selection, camera modes, sliders, playback, snapshot and JSON operations work&#46; This is embedded CSS&#45;viewport proof, not physical touch&#45;device testing&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspect index&#46;html and its references while serving output as HTTP root&#46; Expected: Complete local assets without build or repository&#45;root dependencies&#46; Observed: All local implementation is in index&#46;html; the sole external Three&#46;js CDN dependency is allowed&#46; No referenced local file is missing&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Export empty/populated lists, round&#45;trip through labeled textarea/import/export, and compare state/order/IDs&#46; Expected: Exact version&#45;1 five&#45;field schema and visible planet&#45;ID mapping with no implicit live&#45;state mutation&#46; Observed: Empty export is version 1 with an empty array; populated exports contain only id/time/speed/selectedPlanet/viewMode and preserve order, case/spaces and typed values&#46; Mercury/mercury, Earth/earth, Jupiter/jupiter mappings are visible&#46; Paused and playing round&#45;trips preserve live state apart from natural advance&#46;

#### D5

Status: pass. Rating: 5. Method: http.

Action: Run 37 invalid classes and valid empty/three&#45;entry/boundary/fractional/case&#45;sensitive/spaced&#45;ID/extra&#45;key payloads while paused and playing&#46; Expected: Atomic rejection with visible errors or list&#45;only replacement, derived angles and camera preservation&#46; Observed: Every invalid case, including mixed validity and overflow, shows an error and preserves the exact existing list/live fields&#46; Valid 0/120 and 0&#46;25/4 endpoints, 17&#46;25/1&#46;5, whitespace&#45;only ID and extra keys succeed&#46; Extra angles are ignored&#46; Import camera comparison initially had a 47&#45;image&#45;pixel scroll offset; after aligning the observed header shift, all four body centers have zero shift&#46;

## Verification and limits

Browser method: http.
Viewports: 1440x900, 390x844.
Coverage: 21 passed, 5 failed, 1 unverified.

### Notes

- Fresh review requested for every existing saved run&#46; This run is evaluated independently from unchanged output; previous reviews and scores are not evidence&#46;
- Reviewer model openai&#45;codex/gpt&#45;6&#45;astra is supplied by current host project metadata&#46; No current effort is exposed in host metadata or applicable environment; effort remains null rather than borrowing generator/default settings&#46;
- Task routing was checked against the full saved solar&#45;system&#45;observatory&#45;2 contract, R1&#45;R15, JSON schema, eight acceptance steps and six additional sequences&#46;
- Preregistered tolerances: speed samples at least3 wall seconds; max&#40;0&#46;15s,5% expected delta&#41; plus one displayed increment; paused time one displayed increment, angles1 degree; predicted angle2 degrees plus rounding; repeat framing2 CSS pixels; tracking center10% of smaller scene dimension; JSON/state identity exact except natural playing time advance&#46;
- HTTP embedded document viewports are checked at1440x900 and390x844&#46; This does not certify file://, physical touch input or excluded generation conditions&#46;
- Fresh ordered saved acceptance scenario and all six additional&#45;sequence families were exercised, followed by the fixed rate, endpoint, keyboard, import, persistence and quota matrix&#46; No original output was repaired&#46;
- Reviewer effort remains unavailable in current host metadata; status is partial independently of numeric quality&#46; Host&#45;required aside&#45;browser was loaded for direct browser safety&#46; On session resume, mandatory injected using&#45;harness&#45;flow and caveman applied; using&#45;harness&#45;flow required loading brainstorming for its read&#45;only investigation route, without feature design or code changes&#46;
- The user explicitly requested parallel review after this run&#39;s browser checks&#46; Host policy required task subagents for the other nine independent runs, overriding the skill&#39;s no&#45;other&#45;model restriction for the overall batch&#46; This run was personally observed and scored by the parent; no model consensus or isolation is claimed&#46; Shared browser access is exclusive so rate samples do not overlap other review scenes&#46;
- Native canvas coordinates were calibrated through actual picking; a locator&#45;offset mismatch was reported to tool QA rather than counted as an artifact failure&#46; Select typeahead plus Enter established keyboard operation after inconclusive Home/ArrowDown attempts&#46;
- Only transient screenshots and working&#45;memory comparisons were used&#46; A whole&#45;PNG byte mismatch was resolved by decoding image pixels; repeated scene pixels at 30/17/73 after reload matched exactly&#46; A separate import comparison was aligned for its observed page&#45;scroll/header shift before judging camera preservation&#46;
- The original app storage value was restored exactly; the quota filler was removed&#46; All owned artifact/viewport/scratch tabs were closed and the output&#45;only service stopped&#46; The parent&#45;owned shared viewport server remains available to the other review workers&#46;
