# Benchmark review

## Conclusion

Status: partial.
Mandatory verdict: fail.
Quality total: 86.40 / 100. Conservative lower bound. Range: 86.40 to 88.40.
Quality is uncapped. The mandatory verdict is independent of the quality score.
Mandatory failures (2): T4, T5.
Mandatory unverified (0): None.
Missing checks (1): D3.
Missing reviewer metadata: reviewer.effort.
Format: 5&#46;0. Rubric: solar&#45;system&#45;5&#46;0.
Reviewed at: 2026&#45;10&#45;20T21:39:35&#43;09:00.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: Unavailable.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; determined score; full category denominator |
| selection | 3.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 4.80 | 15 | U1, U2, U3, U4, U5; determined score; full category denominator |
| visual | 4.00 | 15 | V1, V2; determined score; full category denominator |
| responsive | 5.00 | 15 | L1, L2; determined score; full category denominator |
| delivery | 4.00 to 5.00 | 10 | D2, D3, D4, D5, D6; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |

## Findings

### Confirmed failures

#### T4

Status: fail. Rating: 0. Method: http.

Action: Returned from tracking to overview at desktop and mobile sizes and sought 30 seconds in the mobile overview&#46; Expected: Full&#45;system visibility while preserving selection, time, speed and playback&#46; Observed: Desktop returns preserved state and showed the full system&#46; At actual 390x844, canvas about 347x363 CSS pixels, the standard 30&#45;second overview visibly cut off the right portion of Earth at the right canvas border; the outer orbit also reaches/crosses that border&#46; Time 30, speed 1&#46;5x, Earth selection and paused playback were preserved, but full&#45;system framing failed&#46; Cause supported by index&#46;html:139&#45;146 and 222&#45;225: resize changes camera aspect while standard overview distance remains fixed at 21, independent of narrow aspect ratio&#46;

#### T5

Status: fail. Rating: 0. Method: http.

Action: Restored tracking and null&#45;selection overview snapshots after unrelated time/speed/selection changes and manual orbit/zoom, repeated desktop restores, and restored null overview time 30 at mobile size&#46; Expected: Atomic paused state and deterministic standard full&#45;system overview or selected&#45;planet tracking framing&#46; Observed: Fields, angles, information clearing and target were correct&#46; Repeated desktop overview and tracking renders had 0 changed pixels after manual camera/live edits&#46; However, the mobile null&#45;selection 30&#45;second overview restore again visibly clipped Earth at the right canvas edge, so its required standard full&#45;system framing was not met&#46; The fixed distance 21 in standardCamera is used by restoreSnapshot at index&#46;html:293&#45;305; narrow&#45;aspect resize does not compensate&#46;

### Missing proof

#### D3

Status: unverified. Rating: Unverified. Method: http.

HTTP rendering and the required local Three&#46;js resource operated throughout the observed sequences without a visible fatal artifact failure&#46; Complete console/page&#45;error coverage was unavailable: the current shared Aside transport positive control did not deliver deliberate console&#46;error or asynchronous thrown&#45;error events&#46; Empty subscriptions therefore cannot prove this quality check&#46; No alternative browser, raw CDP or source&#45;only substitute was used&#46;

### Verified passes

#### S1

Status: pass. Rating: 5. Method: http.

Action: Observed the moving desktop scene, selected each rendered planet, and sought 0, 17, 30, 73 and 120 seconds while reading each planet information panel&#46; Expected: One sun and exactly three distinct 3D planets on circular, fixed&#45;phase 12/24/48&#45;second orbits&#46; Observed: One gold sun, small tan Mercury, medium amber Venus and larger blue Earth travel along three circular orbit paths seen in perspective&#46; At t=0 their angles were 10&#46;3/64&#46;2/124&#46;9 degrees; at 17, 160&#46;3/319&#46;2/252&#46;4; at 30, 190&#46;3/154&#46;2/349&#46;9; at 73, 40&#46;3/79&#46;2/312&#46;4; at 120, 10&#46;3/64&#46;2/304&#46;9&#46; These match phases 0&#46;18/1&#46;12/2&#46;18 radians and periods 12/24/48 within preregistered rounding tolerance&#46; Predicted projected locations were corroborated by successful native scene clicks and visible motion, not just constants&#46;

#### S2

Status: pass. Rating: 5. Method: http.

Action: Opened the page and performed same&#45;origin load&#45;event&#45;barrier reloads after import, deletion and capture&#46; Expected: Time zero, speed 1x, running, no selection and overview, followed by motion&#46; Observed: Reload samples showed 0&#46;0 s, 1x, OVERVIEW LIVE, empty selected set and explicit No planet selected; subsequent visible time/positions advanced&#46; Initial navigation snapshots at about 5 seconds were ordinary load elapsed time, not treated as reset failures&#46; Stored snapshots did not auto&#45;restore the live scene&#46;

#### S3

Status: pass. Rating: 5. Method: http.

Action: Paused motion, sought 30, waited 3&#46;1 real seconds, selected planets and orbited/zoomed while paused, then resumed&#46; Expected: Frozen time and all positions with usable camera/selection and continuous resumption&#46; Observed: Paused readouts remained unchanged; a same&#45;canvas decoded screenshot comparison reported 0 changed pixels out of 414594&#46; Each paused selection/camera action worked&#46; Resume advanced continuously from the paused state; the ordered 2x sample advanced 6&#46;2 simulation seconds over 3&#46;1083 real seconds without a jump&#46;

#### S4

Status: pass. Rating: 5. Method: http.

Action: Changed speed to 0&#46;25x, 2x and 4x while paused and playing, resumed each paused speed, and measured uninterrupted visible&#45;page samples with only this review scene running&#46; Expected: Preserved current positions/time, proportional rates and fixed relative periods&#46; Observed: Paused changes at 17 seconds produced 0 changed canvas pixels for all three speeds&#46; Samples were 0&#46;7/3&#46;1025 and 0&#46;8/3&#46;1014 simulation/real seconds at 0&#46;25x; 6&#46;2/3&#46;1030 and 6&#46;3/3&#46;1006 at 2x; 12&#46;5/3&#46;1006 and 12&#46;4/3&#46;0998 at 4x&#46; All fit preregistered tolerance&#46; Paused waits had zero simulation delta; playing changes preserved ongoing motion and the orbital angle relationships&#46;

#### S5

Status: pass. Rating: 5. Method: http.

Action: Used the real time slider at 0, 17, 30, 73 and 120, revisited 30/17/73 after playback, speed and selection changes, and repeated after reload&#46; Restored a valid fractional 17&#46;25&#45;second import&#46; Expected: Immediate paused deterministic time, positions and information including tracking updates&#46; Observed: All integer seeks displayed the exact requested paused time and phase&#45;correct angles; 17&#46;25 restored as rounded 17&#46;3 s with angles 167&#46;8/322&#46;9/254&#46;3 degrees&#46; Six repeat&#45;seek/reload canvas comparisons at identical overview/Earth selection/desktop framing each had 0 changed pixels of 414594&#46; Tracking seeks immediately recentered the selection&#46;

#### S6

Status: pass. Rating: 5. Method: http.

Action: Sought 119, set 4x, resumed across the endpoint, waited over three seconds, then sought 17 and resumed&#46; Expected: Stop at 120 without overshoot/wrap and allow earlier resumption&#46; Observed: Time reached 120&#46;0 and playback became paused, with no wrap or overshoot during the wait; the endpoint sample advanced only the remaining 0&#46;1 second&#46; Earlier resumption then advanced 12&#46;4 simulation seconds over 3&#46;1018 real seconds at 4x&#46;

#### S7

Status: pass. Rating: 5. Method: http.

Action: Inspected the loaded local Three&#46;js r160 resource and renderer/geometry/camera source, then dragged and zoomed the rendered scene in both modes&#46; Expected: Actual Three&#46;js 3D geometry with changing depth, occlusion and perspective&#46; Observed: The HTTP resource entry loaded three&#46;min&#46;js; source uses WebGLRenderer, PerspectiveCamera, SphereGeometry, TorusGeometry and circular 3D line loops&#46; Rendered shaded spheres, white selection torus and orbit ellipses changed perspective and relative occlusion under camera drag/zoom, including enlarged tracking views&#46; This was not inferred from the library label alone&#46;

#### S8

Status: pass. Rating: 5. Method: http.

Action: Captured paused A at 30 s/2x/Venus/tracking and captured while playing; changed live time, speed, selection and mode; exported, imported and reloaded a captured ID; restored fractional imported data containing bogus extra angles&#46; Expected: Stable unique IDs and immutable captured fields with all three derived displayed angles and no live&#45;state mutation&#46; Observed: A ID obs&#45;mvgourvf&#45;1&#45;zdds58 retained time 30, speed 2, Venus, tracking and angles 190&#46;3/154&#46;2/349&#46;9 degrees through live edits&#46; A playing capture exported precise time 73&#46;54919999990464 and matching rounded display/angles&#46; Captured&#45;ID export/import/reload text was exactly equal&#46; Fractional 17&#46;25 yielded 167&#46;8/322&#46;9/254&#46;3, unchanged after reload; supplied 999&#45;degree angles were ignored&#46; Paused capture preserved state and aligned canvas pixels exactly&#46;

#### T1

Status: pass. Rating: 5. Method: http.

Action: Executed all twelve combinations of Mercury/Venus/Earth, labeled&#45;list/native&#45;canvas selection, and paused/playing states; changed to another list selection before each canvas click&#46; Expected: Both routes share selection, information and visible highlight&#46; Observed: Every combination selected the requested stable ID, showed the corresponding information, and visibly placed the white torus on that planet&#46; All six paused cases remained OVERVIEW PAUSED and all six playing cases remained OVERVIEW LIVE&#46; Native mouse coordinates used observed canvas bounds; no locator&#45;coordinate assumption was used&#46;

#### T2

Status: pass. Rating: 5. Method: http.

Action: Inspected each selection during playback and the 0/17/30/73/120 seek matrix, then restored the null&#45;selection overview entry after tracking&#46; Expected: Correct name, period, angle and rendered selection; explicit empty information and no highlight when unselected&#46; Observed: Mercury/Venus/Earth showed periods 12/24/48 and angles matching the fixed&#45;phase calculation and clicked rendered spheres&#46; Readouts advanced during playback and immediately on seek&#46; Null restore showed No planet selected, cleared the selected buttons and scene torus, and disabled tracking&#46;

#### T3

Status: pass. Rating: 5. Method: http.

Action: Tracked Earth while playing for three samples spanning more than 3&#46;2 seconds, switched to Mercury then Venus, sought 30 and 73 while tracking, and examined no&#45;selection tracking availability&#46; Expected: Centered target, synchronized highlight/info and preserved playback/speed/time except requested seeking; tracking unavailable without selection&#46; Observed: In the approximately 826x621 CSS canvas, each selected sphere remained visually centered at about canvas&#45;relative &#40;413,310&#41;, within the 62&#46;1&#45;pixel threshold, over motion and immediately after seeks&#46; List switches moved target/highlight/info together and continued live playback at the current speed/time&#46; Null&#45;selection startup/restore disabled Track planet without changing other state&#46;

#### U1

Status: pass. Rating: 5. Method: http.

Action: Dragged and wheel&#45;zoomed in all four overview/tracking and playing/paused combinations&#46; Expected: Responsive orbit/zoom around the current target without changing simulation state&#46; Observed: Each drag changed ellipse orientation and 3D sphere/torus perspective, and wheel zoom changed apparent size&#46; The selected target stayed centered in tracking&#46; Paused time stayed fixed and playing motion continued; no camera operation changed selected ID, speed or playback&#46;

#### U2

Status: pass. Rating: 5. Method: http.

Action: Began a fast drag on Mercury while Earth was selected, then reset framing and performed 25 native 2&#45;CSS&#45;pixel moves from Mercury before pointer&#45;up&#46; Exercised every list/slider/button/snapshot/JSON control&#46; Expected: Drags do not select planets and DOM interactions do not leak scene selection or unintended camera movement&#46; Observed: Both fast and cumulative 50&#45;pixel drags left Earth selected while rotating the view&#46; All named DOM operations had only their intended effects&#46; Properly aligned canvas comparisons across paused import, capture, deletion and export each reported 0 changed pixels; speed changes likewise left pixels unchanged&#46; A preliminary unaligned import image comparison was discarded after observing different scroll positions, then repeated with each image cropped using its own observed canvas bounds&#46;

#### U3

Status: pass. Rating: 4. Method: http.

Action: Used simulation, selection, tracking, snapshot and JSON controls across normal, capacity, malformed&#45;data and quota states at both sizes&#46; Expected: Understandable purpose, current state and feedback&#46; Observed: Named grouped panels, explicit LIVE/PAUSED and overview/tracking pills, time/speed readouts, stable planet&#45;ID mapping, per&#45;snapshot values/angles, capacity rejection, validation errors and saving&#45;unavailable notices make the workflow clear&#46; Long IDs and dense snapshot details require reading and vertical scrolling, so the execution is polished with minor shortcomings rather than exceptional &#40;4/5&#41;&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Captured three distinct entries, attempted a fourth, deleted the middle entry, then captured again while playing and exported&#46; Expected: Visible capacity rejection, stable existing values/IDs/order, append behavior, no collision and no unintended live changes&#46; Observed: The fourth attempt displayed Capacity reached: delete a snapshot before capturing another and retained all three entries&#46; Deleting the middle left the first and third unchanged in order; the new unique obs&#45;mvgov7cv&#45;4&#45;5ibqeo appended after them&#46; Paused state stayed fixed during capacity/deletion checks; playing capture preserved playback with natural advancement&#46;

#### U5

Status: pass. Rating: 5. Method: http.

Action: At each measured viewport, navigated with Tab and Shift&#43;Tab and activated selection, pause/resume, tracking/overview, capture/restore/delete, export/import with Enter; adjusted time/speed with arrows; selected and typed textarea JSON using keyboard input&#46; Expected: Meaningful names, visible focus, normal activation and no keyboard trap for every primary DOM control&#46; Observed: Twenty&#45;eight retained control activations across the two sizes produced the requested states, including a mobile fractional tracking import/restore/delete&#46; Observed focus outlines and accessible names identified the active controls; Tab reached each target and Shift&#43;Tab left the final control&#46; Native keyboard operation remained usable with vertical scrolling and did not leak scene actions&#46;

#### V1

Status: pass. Rating: 4. Method: http.

Action: Visually inspected overview/tracking, selected/unselected, camera rotations and both viewport sizes&#46; Expected: Readable scene and selection with supportive effects&#46; Observed: Distinct blue/amber/tan shaded spheres, restrained sun glow, quiet orbit guides and a clear white selection torus support inspection&#46; Dark hemispheres and thin small&#45;screen orbit lines are modest readability limits&#46; The scene is polished and legible &#40;4/5&#41;; the concrete mobile full&#45;system framing failure is reported under T4/T5 rather than compensated by an extra visual penalty&#46;

#### V2

Status: pass. Rating: 4. Method: http.

Action: Inspected desktop side&#45;by&#45;side panels and mobile stacked simulation, selection, snapshot and JSON panels including focus/error states&#46; Expected: Coherent legible typography, spacing, color and hierarchy&#46; Observed: Navy cards, cyan accents, clear headings, consistent borders and spacing, tabular state values and distinct error/success colors form a polished hierarchy&#46; Compact snapshot/JSON text and long IDs are denser than the main controls but remain readable with vertical scrolling &#40;4/5&#41;&#46;

#### L1

Status: pass. Rating: 5. Method: http.

Action: Measured the embedded desktop document and exercised selection, tracking/overview, seek, speed, pause/resume, capture/restore/delete and textarea import/export by pointer and keyboard&#46; Expected: Usable scene and all primary controls at 1440x900 without horizontal scrolling or panel occlusion&#46; Observed: innerWidth/innerHeight were 1440/900, clientWidth and scrollWidth both 1423&#46; Scene and side controls were usable; lower snapshot/JSON controls were reached by allowed vertical scrolling&#46; No panel blocked a primary control&#46;

#### L2

Status: pass. Rating: 5. Method: http.

Action: Repeated all primary controls and the full keyboard sequence in the embedded mobile document&#46; Expected: Usable controls and scene at 390x844 without horizontal scrolling or control&#45;blocking panels&#46; Observed: innerWidth/innerHeight were 390/844; clientWidth and scrollWidth both 373&#46; Stacked panels kept selection, tracking, seeking, speed, playback, capture/restore/delete and JSON import/export reachable by vertical scrolling, including a fractional tracking restore&#46; No horizontal page overflow or blocked control was observed&#46; The separate standard overview framing defect is documented under T4/T5; manual scene camera controls remained usable&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspected output/index&#46;html, its local script reference, and complete output directory containing index&#46;html and three&#46;min&#46;js&#46; Expected: Every required local resource resides beneath the output&#45;only HTTP root without build or repository dependencies&#46; Observed: The sole referenced asset &#46;/three&#46;min&#46;js exists, is the bundled Three&#46;js r160 classic runtime; geometry, styling and app logic are inline&#46; No application external asset, build, fetch or repository&#45;root dependency exists&#46; Direct&#45;file execution is excluded&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Exported empty and populated lists through the labeled textarea, imported/exported captured and user&#45;supplied records, compared exact identity/order/types, and repeated while playing&#46; Expected: Only version and the five required snapshot fields, stable IDs and visible planet mapping, no list/live/camera mutation and actual UI round&#45;trip without an OS picker&#46; Observed: Empty export was version 1 with snapshots &#91;&#93;; populated exports contained only id/time/speed/selectedPlanet/viewMode in current order&#46; IDs mercury/venus/earth are shown beside names and JSON controls&#46; Captured&#45;ID export/import/reload text matched exactly; imported space/case IDs and fractional values survived export&#46; Angles, playback and camera offsets were absent&#46; Paused aligned export pixels and state were unchanged; playing export preserved live playback&#46;

#### D5

Status: pass. Rating: 5. Method: http.

Action: Exercised 38 separate invalid inputs against a nonempty existing list, including malformed JSON, root null/array, missing root/all entry fields, null/array entries, string/unsupported version, wrong numeric types, range errors, 1e400, non&#45;array/four&#45;entry snapshots, duplicate/empty/nonstring IDs, unknown/numeric/empty/string&#45;null planets, invalid modes and mixed valid&#45;first/invalid&#45;later entries&#46; Then imported valid empty/two/three&#45;entry payloads paused and playing, with endpoints 0/120 and 0&#46;25/4, time 17&#46;25/speed 1&#46;5, exact IDs including a single space and case&#45;distinct A/a, null/selected overview and tracking, plus extra keys and bogus angles&#46; Expected: Atomic rejection or list&#45;only replacement with exact IDs/order and derived angles&#46; Observed: All 38 rejections showed a specific visible error and preserved the full list and live state&#46; All valid payloads restored/exported correct values, ignored extra keys/angles, preserved exact IDs/order and changed only the list; aligned paused import canvas comparison was pixel&#45;identical&#46; Playing import continued tracking/playback with natural time advancement; empty import cleared only snapshots&#46;

#### D6

Status: pass. Rating: 5. Method: http.

Action: Reloaded after valid import, deletion and capture; compared IDs/order/values/derived angles and live defaults; replaced only the app snapshot key with each of the 38 malformed/schema&#45;invalid classes; exercised a real quota&#45;write failure with a unique bounded filler key&#46; Expected: Same versioned persistence/validation, empty visible recovery, defaults on reload, and visible saving&#45;unavailable with usable memory after a failed write&#46; Observed: Each successful mutation survived reload with exact records/angles while live state reset to 0/1x/playing/null/overview&#46; All 38 invalid saved inputs yielded an empty list, visible recovery notice and running defaults; controls remained usable&#46; A bounded 8&#45;MiB search filled 5242544 characters, encountered eight QuotaExceededError results, and a subsequent real capture left persisted snapshots &#91;&#93; while the new snapshot existed in memory and Saving unavailable; snapshots remain in memory only was visible&#46; Playback, selection and restore still worked&#46; This establishes the allowed failed&#45;write branch; storage read denial was not exercised or required as an additional clause&#46; Original key absence was restored and the filler removed in finally and final cleanup&#46;

## Verification and limits

Browser method: http.
Viewports: 1440x900, 390x844.
Coverage: 24 passed, 2 failed, 1 unverified.

### Notes

- Saved prompt independently confirms solar&#45;system&#45;observatory&#45;2, all 15 requirements, eight ordered steps and six additional sequences; format 5&#46;0 and rubric solar&#45;system&#45;5&#46;0 apply&#46; Generation status is not quality evidence&#46; Existing reviews and generator verification notes are not evidence for this fresh attempt&#46;
- Reviewer model openai&#45;codex/gpt&#45;6&#45;astra is the current host project&#45;context selection&#46; Host environment exposes no concrete effort and no applicable current&#45;session selection API was supplied; effort remains null, preventing complete status&#46; Timestamp is from host date &#45;Iseconds&#46; Generator metadata was not used to infer reviewer identity&#46;
- Applied agents/benchmark&#45;review shared resources and mandatory aside&#45;browser guide/repl instructions&#46; User explicitly requested parallel worker review; host dispatch overrides the skill prohibition on another&#45;model review orchestration&#46; This is not isolated single&#45;reviewer orchestration&#46; No descendants or additional models launched by this worker&#46;
- Preregistered before browser observations: rate sample &#62;=3 real seconds, tolerance max&#40;0&#46;15 simulation seconds, 5% expected delta&#41;&#43;0&#46;1&#45;second display increment; paused allowance 0&#46;1 seconds and 1 degree; predicted angles within 2 degrees plus 0&#46;1&#45;degree rounding; repeat centers within 2 CSS pixels with identical viewport/framing; tracking center within 10% of smaller canvas dimension over &#62;=3 seconds and after seeking&#46; Exact identity/order/state preservation except natural playing advance&#46;
- Browser diagnostics shared capability positive control failed to deliver console&#46;error and async throw events&#46; Empty streams therefore cannot establish D3; no raw CDP or alternative browser will be used&#46;
- Fresh runtime review used only connected Aside direct REPL after explicit BROWSER&#95;GRANTED, the assigned output&#45;only http://127&#46;0&#46;0&#46;1:8923 origin and parent&#45;owned viewport wrapper at port 8890&#46; Actual child viewports were 1440x900 &#40;client/scroll width 1423&#41; and 390x844 &#40;373&#41;; both secure contexts&#46; This is embedded CSS&#45;viewport verification, not top&#45;level viewport or touch&#45;device proof&#46; Only one review scene ran during timing&#46;
- Ordered saved steps completed: automatic motion; Earth canvas selection/tracking; pause/30&#45;second seek; paused 2x change; resume measured 6&#46;2 simulation seconds in 3&#46;1083 real seconds; Venus list switch while tracking; overview with state preserved; pause/reseek 30&#46; Six additional sequences followed in order: immutable A and repeated restore; capacity/delete/recapture/export; rejection matrix; valid null/fractional tracking imports/playing/empty/round&#45;trip; persistence/recovery/quota; full keyboard operation at desktop then mobile&#46; Remaining fixed matrix checks followed, including all twelve planet selection cases&#46;
- Image evidence was transient&#46; Decoded canvas crops compared repeat seeks after reload, paused stability, speed changes, list&#45;only operations and repeated restores&#46; Repeats with matching framing gave zero changed pixels&#46; One preliminary import comparison used different scroll alignments and was invalid; an aligned repeat using each observed canvas rectangle gave zero changed pixels&#46; No screenshot, log, collector or evidence file is retained in the run&#46;
- Confirmed shared cause for T4 and T5: the mobile aspect ratio changes without fitting the fixed 21&#45;unit standard overview camera distance&#46; At 30 seconds both overview and a null&#45;selection overview restore visibly clip Earth at the right edge&#46; Desktop and tracking framing worked&#46; This is not inferred solely from source and is not a generic CSS horizontal&#45;overflow failure&#46;
- Cleanup completed before BROWSER&#95;RELEASED: exact original absence of solar&#45;system&#45;observatory&#45;2&#45;snapshots restored at the isolated origin, unique &#95;&#95;review&#95;luna&#95;max&#95;20261020&#95;quota key absent, all worker&#45;owned tabs closed, transient image references cleared, and rr&#45;luna&#45;max&#45;parallel stopped&#46; Shared viewport service and unrelated tabs/storage were untouched&#46; Prompt, generation record and output were only read/served, never edited&#46; Parent owns source&#45;hash preservation verification&#46;
- D6 is credited for its explicitly allowed actual failed&#45;write recovery branch, not withheld for unexercised storage&#45;read denial&#46; Browser diagnostic transport is the sole remaining unverified check &#40;D3&#41;&#46; Reviewer effort remains null independently of quality bounds&#46;
