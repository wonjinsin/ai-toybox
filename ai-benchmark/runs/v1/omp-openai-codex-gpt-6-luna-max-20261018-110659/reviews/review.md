# Benchmark review

## Conclusion

Status: partial.
Mandatory verdict: fail.
Published total: 69.00 / 100. Determined score. Range: 69.00.
Raw total: 75.40 to 77.40.
Confirmed cap: 69.
The confirmed 69 cap is existing review policy for a mandatory failure, not a calibrated quality measurement.
Missing checks (1): D3.
Missing reviewer metadata: None.
Format: 4&#46;0. Rubric: solar&#45;system&#45;4&#46;0.
Reviewed at: 2026&#45;10&#45;18T16:27:55&#46;952229&#43;09:00.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: high.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; determined score; full category denominator |
| selection | 4.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 4.80 | 15 | U1, U2, U3, U4, U5; determined score; full category denominator |
| visual | 4.00 | 15 | V1, V2; determined score; full category denominator |
| responsive | 0.00 | 15 | L1, L2; determined score; full category denominator |
| delivery | 4.00 to 5.00 | 10 | D2, D3, D4, D5, D6; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |

## Findings

### Confirmed failures

#### T4

Status: fail. Rating: 0. Method: http.

Action: Returned from tracking to overview in paused and playing states, including at the required 390x844 size, then sought 0 and pressed Overview&#46; Expected: Full system visible and selection/time/speed/playback preserved&#46; Observed: Desktop overview preserved state and showed the complete system&#46; At mobile default overview time 0, the blue outer Earth is visibly cut off by the left canvas edge while the sun and inner planets are visible; the 347x363 canvas cannot show the complete system with this standard framing&#46; Reproduction uses ordinary seek 0 and Overview, not manual zoom&#46; Source keeps overview distance 21 with vertical FOV 45 despite the narrow aspect ratio &#40;index&#46;html 95,139&#45;146,222&#45;225&#41;&#46; This fails full&#45;system overview&#46;

#### L1

Status: fail. Rating: 0. Method: http.

Action: At actual 1440x900 exercised every primary control, then imported a legal one&#45;record payload with id equal to 2000 q characters, time 17, speed 1, null selection and overview, and restored it&#46; Expected: No horizontal scrolling or horizontally hidden controls for valid snapshot data&#46; Observed: Normal records had clientWidth=scrollWidth=1423 and every control worked&#46; Restoring the valid long ID expanded scrollWidth to 15712 while clientWidth stayed 1423; the Restore status had scrollWidth 14792, cards expanded and JSON buttons extended beyond the viewport&#46; A horizontal scrollbar was visible&#46; Source writes the entire ID into snapshotMessage without overflow wrapping &#40;index&#46;html 305 and &#46;status CSS line 8&#41;&#46; This is a valid&#45;ID boundary, independently repeated after quota cleanup&#46;

#### L2

Status: fail. Rating: 0. Method: http.

Action: Verified actual secure 390x844 child viewport and operated selection, tracking/overview, seek/speed, pause/resume, capture/restore/delete and JSON by keyboard&#46; Repeated the valid 2000&#45;character&#45;ID import and restore&#46; Expected: Same usable controls without horizontal scrolling&#46; Observed: Normal records had clientWidth=scrollWidth=373 and all listed actions succeeded with vertical scrolling&#46; Long&#45;ID restore expanded document scrollWidth to 14835 with clientWidth 373; the 14792&#45;wide status forced cards/JSON controls offscreen and exposed a horizontal scrollbar&#46; Separately, default overview at time 0 clipped Earth at the left edge&#46; L2 fails despite successful ordinary keyboard/control coverage&#46;

### Missing proof

#### D3

Status: unverified. Rating: Unverified. Method: http.

Unverified: The shared fresh diagnostic positive control, including an actual main&#45;world script console&#46;error and asynchronous throw on a reviewer&#45;owned HTTP scratch page after subscriptions, returned empty console/pageerror arrays&#46; Thus those subscriptions do not establish diagnostic coverage&#46; Required Three&#46;js resource and scene render were observed, but startup errors and ordinary console&#46;error calls cannot be certified absent with this transport&#46;

### Verified passes

#### S1

Status: pass. Rating: 5. Method: http.

Action: Observed initial live motion of all three bodies, then sought 0, 12, 24, 48, 17, 30 and 73 while inspecting each selected angle and full rendered scene&#46; Expected: One sun and three differently sized/colored planets on circular 12/24/48&#45;second orbits with fixed phases&#46; Observed: Small tan Mercury, medium gold Venus and larger blue Earth moved around one sun on perspective&#45;projected circular rings&#46; Initial phases were 10&#46;3/64&#46;2/124&#46;9 degrees &#40;source phases 0&#46;18/1&#46;12/2&#46;18 radians&#41;&#46; At 12 the angles were 10&#46;3/244&#46;2/214&#46;9; at 24, 10&#46;3/64&#46;2/304&#46;9; at 48 all returned to initial angles and the complete scene image was identical to time 0&#46; Mercury returned each 12 seconds, Venus changed to the opposite side at 12 and returned at 24, and Earth traversed quarter/half/full positions at 12/24/48&#46; Live images also showed all three moving, not only changing readouts&#46;

#### S2

Status: pass. Rating: 5. Method: http.

Action: Opened a fresh HTTP scene and reloaded after importing, deleting and capturing snapshots&#46; Expected: Live initialization at time 0, speed 1x, running, no selection and overview rather than automatic snapshot restore&#46; Observed: Each fresh scene showed overview/live, 1x, explicit no&#45;selection information, disabled tracking and no highlight&#46; Early post&#45;load samples were 4&#46;7, 5&#46;0 and 4&#46;9 seconds, reflecting navigation elapsed time rather than saved 17&#46;25/30/120&#45;second states&#46; Initial source time is 0 and subsequent observed live time advanced 21&#46;9 to 25&#46;0 over 3&#46;105 wall seconds at 1x&#46; Persisted snapshots loaded independently of live defaults&#46;

#### S3

Status: pass. Rating: 5. Method: http.

Action: Paused during motion, waited 3&#46;1 seconds, selected planets and orbited/zoomed while paused, then resumed from sought states&#46; Expected: Time and every orbital position freeze; selection/camera remain usable and resume continues without reset&#46; Observed: At paused 30&#46;3 seconds the full state and scene screenshot remained identical across the wait&#46; All three canvas/list selections and camera operations worked while paused&#46; Resumed samples continued from 17 and 30 seconds at the selected rate without resetting or jumping to another orbital phase&#46;

#### S4

Status: pass. Rating: 5. Method: http.

Action: Set 0&#46;25x, 2x and 4x while paused at 17, resumed each, and changed rates during playback including 2x&#46; Expected: No seek or position change on speed adjustment and correct common time scaling&#46; Observed: Every paused adjustment retained 17&#46;0 seconds and byte&#45;identical scene images&#46; Single&#45;call visible 1440x900 measurements were 17&#46;1 to 17&#46;8 in 3&#46;109 s at 0&#46;25x, 17&#46;4 to 23&#46;6 in 3&#46;105 s at 2x, and 17&#46;8 to 30&#46;3 in 3&#46;107 s at 4x&#46; A playing change to 2x gave 17&#46;8 to 24&#46;0 in 3&#46;102 s&#46; These meet the preregistered tolerance&#46; Playing changes retained selection/mode/playback and continuous positions; all&#45;body angle/position comparisons retained 12:24:48 period relationships&#46;

#### S5

Status: pass. Rating: 5. Method: http.

Action: Sought 0, 30, 120, 17 and 73 with the real slider; revisited 30/17/73 after speed, playback and selection changes, and after reload&#46; Restored imported time 17&#46;25&#46; Expected: Immediate paused deterministic positions and matching information at any valid time&#46; Observed: All seeks paused&#46; Mercury/Venus/Earth angles were 190&#46;3/154&#46;2/349&#46;9 at 30, 160&#46;3/319&#46;2/252&#46;4 at 17, 40&#46;3/79&#46;2/312&#46;4 at 73 and 10&#46;3/64&#46;2/304&#46;9 at 120&#46; Same&#45;view repeat&#45;seek crops were identical before reload; full rendered all&#45;body layouts and angles returned after reload&#46; Fractional 17&#46;25 yielded 167&#46;8/322&#46;9/254&#46;3 degrees, with time readout rounded to 17&#46;3&#46; Scroll&#45;offset comparisons were excluded rather than treated as camera changes&#46;

#### S6

Status: pass. Rating: 5. Method: http.

Action: Sought 119, set 4x, resumed across the endpoint, waited two 3&#46;1&#45;second intervals, then sought 17 and resumed&#46; Expected: Stop exactly at 120 without wrapping or overshoot; earlier seeking permits playback&#46; Observed: Both endpoint samples were 120&#46;0, overview/paused and Earth 304&#46;9 degrees&#46; Subsequent playing 2x observation advanced 17&#46;8 to 24&#46;0 in 3&#46;102 wall seconds&#46;

#### S7

Status: pass. Rating: 5. Method: http.

Action: Inspected the local Three&#46;js dependency and renderer/mesh/camera implementation, observed loaded HTTP resources and dragged/zoomed the actual scene in both camera modes&#46; Expected: Genuine 3D Three&#46;js rendering rather than a flat proxy&#46; Observed: three&#46;min&#46;js loaded from the output&#45;only origin and the scene reported r160&#46; SphereGeometry, torus selection meshes, perspective camera and WebGLRenderer source correlate with shaded spheres, front/back ring occlusion and changing sphere/ring orientation, perspective and scale during camera orbit/zoom&#46; The sun and planets occluded each other in tracking views&#46; No claim is based on isolated&#45;world visibility of THREE&#46;

#### S8

Status: pass. Rating: 5. Method: http.

Action: Captured paused A at 30/2x/Venus/tracking, changed live time/speed/selection/view and resumed, then restored/exported it&#46; Captured while playing and checked later reloads and fractional imported angles&#46; Expected: Stable unique ID, every captured field and all three derived angles remain immutable; capture does not alter live state&#46; Observed: A kept ID obs&#45;mvdhfxgi&#45;1&#45;rnj0j2 and 30/2x/Venus/tracking with angles 190&#46;3/154&#46;2/349&#46;9 despite unrelated live edits&#46; Live capture obs&#45;mvdi3eeo&#45;1&#45;s2kmnh exported exact time 30&#46;38805000007154, speed 1&#46;5, Mercury, overview, with displayed derived angles 202&#46;0/160&#46;0/352&#46;8; playback continued 30&#46;3 to 36&#46;2 over the 3&#46;930&#45;second capture/export/delete interval&#46; Another playing capture retained ID obs&#45;mvdhrsvi&#45;1&#45;i1anze and values after reload&#46; Imported 17&#46;25 angles were derived as 167&#46;8/322&#46;9/254&#46;3, ignoring supplied angles; exact IDs and values survived export/import/reload&#46;

#### T1

Status: pass. Rating: 5. Method: http.

Action: Selected Mercury, Venus and Earth from the labeled list and by actual sphere clicks, both paused and playing&#46; Expected: Shared selection, information and visible highlight&#46; Observed: Paused canvas clicks at time 30 selected Mercury 190&#46;3, Venus 154&#46;2 and Earth 349&#46;9 degrees&#46; Repeated clicks at playing 0&#46;25x selected the corresponding bodies at about time 30&#46;1&#46; Each route updated the same list selection/info and white orbital marker&#46; Paused selection stayed usable&#46;

#### T2

Status: pass. Rating: 5. Method: http.

Action: Inspected every planet at fixed seeks and during playback, then restored an unselected overview snapshot&#46; Expected: Correct name, period, angle, rendered position and explicit empty information when unselected&#46; Observed: Mercury/Venus/Earth showed periods 12/24/48 and the angle matrix documented under S1/S5&#46; Earth playback info advanced from 354&#46;4 through 19&#46;8 to 46&#46;0 degrees while its tracked scene moved&#46; Null restore cleared the marker, planet name/period/angle, showed No planet selected and disabled tracking&#46;

#### T3

Status: pass. Rating: 5. Method: http.

Action: Tracked Earth while playing and paused, sought 30/73 and selected different planets while tracking; attempted tracking with no selection&#46; Expected: Actual centering and synchronized highlight/info/target, with preserved time/speed/playback on selection and rejected null tracking&#46; Observed: Across 3&#46;438 real seconds at 2x, Earth remained near image center &#40;about 406,497 in a rendered scene spanning x35&#46;&#46;778,y218&#46;&#46;776&#41; at live times 30&#46;6,34&#46;0,37&#46;5, well inside the 62&#45;CSS&#45;pixel tolerance for the 826x621 scene&#46; Seek 73 kept it centered immediately&#46; Mercury and Venus switched target/info/highlight together; switching selection did not seek or change speed/playback&#46; Null&#45;selection restores disabled Track planet without changing live state&#46;

#### T5

Status: pass. Rating: 5. Method: http.

Action: Restored captured Venus tracking A after editing every live field, then imported/restored null&#45;selection overview at 30/1&#46;5x and Earth tracking at 17&#46;25/1&#46;5x&#46; Dragged and zoomed tracking and restored again&#46; Expected: Atomic paused restore, correct positions/info/target, cleared null selection, deterministic standard framing&#46; Observed: A returned to 30/2x/Venus/tracking paused&#46; Null restore showed no highlight, explicit No planet selected, disabled tracking and overview&#46; Fractional tracking showed Earth 254&#46;3 degrees centered, with saved Mercury/Venus angles 167&#46;8/322&#46;9&#46; Repeated null and tracking scene screenshot buffers were identical at 1440x900 after manual camera changes&#46;

#### U1

Status: pass. Rating: 5. Method: http.

Action: Dragged 70&#45;pixel camera orbits and scrolled to zoom in overview/tracking, playing/paused&#46; Expected: Each camera operation visibly responds without unintended simulation changes&#46; Observed: Overview rings changed from shallow ellipses to higher&#45;angle ellipses and the sun/planets enlarged; tracking changed the sphere lighting/torus orientation and scale&#46; Paused before/after remained 30&#46;0/0&#46;25x/Earth in the selected mode&#46; Playing samples continued naturally around 30&#46;1 to 30&#46;2 at 0&#46;25x&#46; Both modes and both playback states responded&#46;

#### U2

Status: pass. Rating: 5. Method: http.

Action: Started a fast drag on Mercury, reset overview, then dragged 36 CSS pixels in 24 small pointer moves starting on Mercury; exercised list, sliders, buttons, snapshots and JSON&#46; Expected: Drags never become planet clicks; DOM controls do not leak input into scene selection/camera&#46; Observed: Earth remained selected after both fast and slow Mercury&#45;start drags while camera orientation changed&#46; DOM interactions changed only their intended fields; paused speed and snapshot/import comparisons retained positions and camera, and keyboard controls did not select a different canvas body or unexpectedly orbit/zoom&#46;

#### U3

Status: pass. Rating: 4. Method: http.

Action: Used the full observation, snapshot, validation, keyboard and recovery flows&#46; Expected: Understandable purpose, current state and feedback&#46; Observed: Simulation/selection/snapshot/JSON cards, live/paused mode pill, speed/time outputs, stable planet&#45;ID mapping, named snapshot actions and clear capacity/validation/storage notices make the tool easy to operate&#46; Rating 4 reflects polished hierarchy and explicit feedback, with dense snapshot details and rounded fractional time as minor limitations&#46; Responsive overflow is owned by L1/L2 rather than deducted again here&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Captured A at 30/2x/Venus/tracking, appended 17&#45;second and 73&#45;second records, attempted a fourth, deleted the middle record, and captured again&#46; Expected: Capacity rejection preserves all three records; deletion preserves remaining order and a new capture appends with a unique ID without changing live state&#46; Observed: Capacity message appeared and the list text was exactly unchanged&#46; Export order was A &#40;obs&#45;mvdhfxgi&#45;1&#45;rnj0j2&#41;, original third &#40;obs&#45;mvdhgmgs&#45;3&#45;bi22oe&#41;, new fourth &#40;obs&#45;mvdhgn3e&#45;4&#45;37zuwb&#41;&#46; Their values and IDs survived; paused 73/2x/Mercury/overview stayed unchanged&#46;

#### U5

Status: pass. Rating: 5. Method: http.

Action: At 1440x900 and 390x844 used Tab/Shift&#43;Tab, Enter on native buttons, Home/End/arrows on both ranges, and keyboard text entry plus Enter import/export&#46; Expected: Every primary DOM control has a meaningful name, visible focus, normal keyboard activation and no trap&#46; Observed: Keyboard selection, pause/resume, speed, seek, tracking/overview, capture, restore, delete and textarea export/edit/import all executed&#46; Tab traversed every control, escaped the iframe controls and returned; Shift&#43;Tab reversed normally&#46; Named active controls matched :focus&#45;visible with solid outlines, confirmed in screenshots&#46; Mobile and desktop captures/restore/deletion succeeded in secure HTTP frames&#46; No unused Space&#45;key transport failure was treated as an application failure&#46;

#### V1

Status: pass. Rating: 4. Method: http.

Action: Inspected overview, tracking, selected/unselected states and both sizes while moving and paused&#46; Expected: Readable scene with helpful rather than obscuring effects&#46; Observed: The dark background, thin orbit guides, shaded differently sized planets and white selection torus make depth and selection clear without excessive effects&#46; Rating 4 reflects clear, restrained presentation with subdued tan/gold contrast&#46; Mobile overview clipping is recorded as the concrete T4 layout/framing failure, not an extra visual penalty&#46;

#### V2

Status: pass. Rating: 4. Method: http.

Action: Inspected complete control panels, snapshots and JSON feedback at both CSS sizes&#46; Expected: Coherent legible typography, spacing, colors and hierarchy&#46; Observed: Consistent dark cards, cyan primary controls/readouts, restrained headings, aligned labels and separate green/red feedback establish a polished hierarchy&#46; Rating 4 allows for compact secondary snapshot/JSON text&#46; The long&#45;ID overflow is separately recorded under responsiveness&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspected output/index&#46;html and the output directory, then served output alone&#46; Expected: All local dependencies resolve without a build or repository&#45;root dependency&#46; Observed: The only referenced local dependency is &#46;/three&#46;min&#46;js, present under output; inline CSS/scripts need no further local resources&#46; HTTP resource timing recorded three&#46;min&#46;js with 670184 transferred bytes and the 3D scene rendered&#46; File&#45;protocol execution is excluded&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Exported empty and populated lists, round&#45;tripped through the labeled textarea and explicit import button, and exported while paused and playing&#46; Expected: Exact versioned contract, stable values/order/IDs and no list/live&#45;state mutation; no file picker&#46; Observed: Empty export was &#123;version:1,snapshots:&#91;&#93;&#125;&#46; Populated output contained only numeric version and the five required snapshot fields, preserving types, order and IDs including spaces and case&#45;distinct A/a&#46; No angles, camera offsets or playback fields appeared&#46; Visible mapping named mercury/venus/earth&#46; Paused export left state unchanged; during combined import/export at 2x, time naturally advanced 17&#46;4 to 25&#46;0 over 3&#46;791 seconds with selection, overview and manual framing retained&#46;

#### D5

Status: pass. Rating: 5. Method: http.

Action: Against a nonempty paused list submitted 35 separate invalid payload classes through the textarea, then valid empty/two/three&#45;entry lists, endpoints, fractional time/speed, selected/unselected modes, extra keys, supplied angles, case&#45;distinct and space&#45;only IDs&#46; Repeated valid import while playing with manual overview framing&#46; Expected: Entire&#45;payload validation and atomic list&#45;only replacement, preserving live state/camera and deriving angles&#46; Observed: Malformed JSON; null/array roots and entries; missing version/snapshots/each entry field; wrong/string version; non&#45;array snapshots; empty/non&#45;string IDs; time/speed strings, booleans and range errors; overflow 1e400; invalid/numeric/string&#45;null/empty planet references; invalid mode; null tracking; duplicate IDs; four entries; and valid&#45;prefix/invalid&#45;later payload all showed errors with exact unchanged list and live DOM state&#46; Fixed&#45;view images retained all&#45;body geometry/camera; 34 of the initial 35 scene captures were byte&#45;identical, and the malformed case was additionally inspected with full before/after images after excluding document&#45;scroll differences&#46; Valid inputs retained exact IDs including A, a, surrounding spaces and a single&#45;space ID; ignored extra keys and supplied angles; endpoint 0/0&#46;25 and 120/4 and fractional 17&#46;25/1&#46;5 restored correctly&#46; Paused valid import changed only the list, retaining all three positions and manually orbited/zoomed framing&#46; During playing import/export all three bodies continued moving on the unchanged manually oriented orbit planes, with 17&#46;4 to 25&#46;0 over 3&#46;791 s at 2x&#46; Empty import cleared only snapshots&#46;

#### D6

Status: pass. Rating: 5. Method: http.

Action: Reloaded separately after valid import, deletion and playing capture on the same HTTP origin/context; tested malformed and mixed&#45;validity schema&#45;invalid persisted payloads; filled only a unique quota key within an 8 MiB bound and attempted a larger valid import&#46; Expected: Snapshot identity/order/values/derived angles persist independently of default live startup; corrupt data recovers visibly; actual write failure is reported while memory controls work&#46; Observed: Import IDs/fields and fractional angles survived reload&#46; Deletion persisted exactly; captured obs&#45;mvdhrsvi&#45;1&#45;i1anze retained its 5&#46;1&#45;second displayed record and all angles/ID after reload&#46; Live scenes restarted 1x/overview/no selection/running rather than restoring snapshots&#46; Malformed and invalid&#45;later persisted data yielded empty lists and explicit recovery notices; pause and planet selection worked&#46; A 5242599&#45;character filler reached QuotaExceededError&#46; The actual larger UI import retained its new in&#45;memory snapshot but left the previous persisted payload unchanged and displayed Saving unavailable; snapshots remain in memory only&#46; Playback, selection and Restore remained usable&#46; Removed only the filler and restored the exact saved app&#45;key bytes, verified true&#46; This exercised write failure, not storage read denial&#46;

## Verification and limits

Browser method: http.
Viewports: 1440x900, 390x844.
Coverage: 23 passed, 3 failed, 1 unverified.

### Notes

- User&#45;requested scoring&#45;only migration to solar&#45;system&#45;4&#46;0 at 2026&#45;10&#45;18T08:19:55&#46;883Z&#46; All check observations, ratings, statuses, methods, reviewer identity, and observation timestamp are unchanged&#46; Numeric scores are guaranteed lower endpoints with explicit upper endpoints, not newly observed passes or failures&#46; This recalculation is not a fresh browser review&#46;
- Fresh review of only the assigned saved solar&#45;system&#45;observatory&#45;2 run against solar&#45;system&#45;3&#46;0&#46; Existing reviews were not used as evidence&#46; prompt&#46;md, run&#46;json and every output file were preserved, confirmed by before/after working&#45;memory hashes; no hashes or generation measurements are included here&#46;
- Reviewer model openai&#45;codex/gpt&#45;6&#45;astra comes from current host metadata&#46; Effort high comes from applicable &#46;omp/agents/benchmark&#45;astra&#45;review&#46;md configuration, corroborated by this reviewer session metadata, not inferred from the generator or prior review&#46; Timestamp comes from the current host clock&#46;
- Applied benchmark&#45;review, mandatory aside&#45;browser and explicitly selected poteto&#45;mode under its OMP adapter&#46; Prove It Works changed the review to actual rendered/control evidence instead of source&#45;only inference&#46; Explain the Number changed speed observations to uninterrupted &#62;=3&#45;second measured intervals with visibility and CSS dimensions&#46; No other model, descendant, generation task, build, test, lint or formatter was launched&#46;
- Preregistered tolerances before observations: speed interval &#62;=3 s, error max&#40;0&#46;15 s,5% expected delta&#41;&#43;0&#46;1 s display increment; paused 0&#46;1 s/1 degree; angle 2 degrees&#43;0&#46;1 degree rounding; repeated framing 2 CSS pixels; tracking center within 10% smaller scene dimension&#46; Exact identity/state preservation except natural timed advance&#46;
- Used only connected Aside direct REPL with unchanged output served at 127&#46;0&#46;0&#46;1:9215 and reviewer&#45;only secure viewport wrappers at 127&#46;0&#46;0&#46;1:9225&#46; Children were exactly 1440x900 and 390x844 with isSecureContext=true&#46; This establishes real embedded CSS viewport behavior, not top&#45;level resizing or physical touch&#45;device coverage&#46; No file:// execution, security&#45;setting change or app runtime mutation was used&#46;
- Saved eight&#45;step scenario observed initial all&#45;body motion, scene Earth selection and tracking, pause/seek 30, 2x while paused, 30&#46;5 to 36&#46;6 over 3&#46;103 real seconds on resume, Venus list switch while tracking, overview, and repeat seek 30&#46; An initial pause was inserted to locate the sphere precisely; playing tracking and scene clicks were subsequently observed explicitly&#46; All six saved additional sequences were exercised: immutable A/restore; capacity/delete/export; separate invalid imports; valid null/tracking and playing/empty imports; persistence/corruption/write failure; keyboard operation at both sizes&#46;
- Speed observations were each contained in one REPL call, used the explicit owned tab/frame, and reported document&#46;visibilityState=visible and 1440x900&#46; Samples at 0&#46;25/2/4x met the originally declared tolerance&#46; A parent&#45;reserved hold was respected while another lane timed its scene; no cross&#45;call elapsed&#45;time inference was used&#46; Later measurements used only one owned live artifact tab&#46;
- Diagnostic limitation is the sole unverified check, D3&#46; The fresh shared scratch positive control tested both evaluation&#45;world and actual main&#45;world script console&#46;error/throw after page&#46;on subscriptions&#46; The executed/reached script markers were present but both arrays stayed empty&#46; Therefore empty diagnostic arrays cannot certify clean startup/runtime console or page errors&#46; Required&#45;resource/rendering success was observed independently&#46; No diagnostic pass or numeric total is invented&#46;
- Observed transport limitations were handled as observation issues, not artifact failures&#46; Reload retained stale child frame handles, so fresh snapshots and the newest actual frame were used&#46; Locator screenshots did not reliably crop the requested canvas&#46; Full scene screenshots were used instead&#46; Whole&#45;image comparisons affected by document scrolling/macOS overscroll were excluded; rendered body positions were compared relative to stable scene geometry&#46; Exact same&#45;view repeat restore/seek buffers were used only when the viewport and scroll framing matched&#46;
- D6 boundary fixtures touched only solar&#45;system&#45;observatory&#45;2&#45;snapshots and unique rxm&#45;review&#45;quota&#45;20261018 on this isolated origin&#46; The original app&#45;key value at fixture setup was held in memory&#46; The quota filler was bounded by 8 MiB; actual UI write failure was confirmed by unchanged stored bytes and saving&#45;unavailable feedback&#46; Filler removal and exact saved&#45;key restoration were verified, including final cleanup&#46; Read denial was not forced and is not claimed&#46;
- Concrete mandatory failures are full&#45;system mobile overview clipping &#40;T4&#41; and legal long&#45;ID Restore status causing horizontal overflow and hiding JSON controls at both required sizes &#40;L1,L2&#41;&#46; The long&#45;ID failure was first exposed during quota testing and then repeated after quota removal; valid IDs have no maximum&#45;length restriction&#46; Other quality ratings were not reduced to duplicate these functional penalties&#46;
- Cleanup completed&#46; Closed only the two review&#45;created artifact/viewport tabs used sequentially, stopped both owned servers, removed all 53 distinct transient screenshots identified in this reviewer transcript, and retained no evidence directory or log&#46; Only reviews/review&#46;json and reviews/review&#46;md are published&#46; The working draft is removed after publication&#46;
