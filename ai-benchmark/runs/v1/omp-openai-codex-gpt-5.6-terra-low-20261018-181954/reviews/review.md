# Benchmark review

## Conclusion

Status: partial.
Mandatory verdict: fail.
Published total: 50.93 / 100. Conservative lower bound. Range: 50.93 to 59.18.
Raw total: 50.93 to 59.18.
Confirmed cap: 69.
The confirmed 69 cap is existing review policy for a mandatory failure, not a calibrated quality measurement.
Missing checks (3): S2, S7, D3.
Missing reviewer metadata: reviewer.effort.
Format: 4&#46;0. Rubric: solar&#45;system&#45;4&#46;0.
Reviewed at: 2026&#45;10&#45;18T09:21:45&#46;040625&#43;00:00.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: Unavailable.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 3.13 to 4.38 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |
| selection | 2.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 1.60 | 15 | U1, U2, U3, U4, U5; determined score; full category denominator |
| visual | 3.00 | 15 | V1, V2; determined score; full category denominator |
| responsive | 2.50 | 15 | L1, L2; determined score; full category denominator |
| delivery | 3.00 to 4.00 | 10 | D2, D3, D4, D5, D6; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |

## Findings

### Confirmed failures

#### S8

Status: fail. Rating: 0. Method: http.

Action: At paused time 30, speed 2x, Jade tracking, capture obs&#45;1; change all live fields; import/restore fractional snapshots and reload&#46; Expected: immutable captured fields with angles derived from the same fixed phases/periods as the scene&#46; Observed: the saved 30&#45;second angles were Ember 70&#46;4, Azure 295&#46;1, Jade 341&#46;0 degrees, versus rendered&#45;orbit predictions 180&#46;0, 158&#46;7549, 2&#46;5099 degrees&#46; Saved values stayed fixed and IDs survived export/reload, but angles are wrong&#46; index&#46;html line 8 destructures period/phase one slot early, reading orbital radius/period instead of period/phase&#46;

#### T2

Status: fail. Rating: 0. Method: http.

Action: Select Ember and seek 0, 12, 24, 48; inspect information and rendered returns; inspect all three saved angles at 30&#46; Expected: angle agrees with fixed&#45;phase rendered orbital position&#46; Observed: Ember occupies the same initial position at these whole periods but readouts are 327&#46;5, 224&#46;7, 121&#46;8, and 276&#46;1 degrees rather than the same 0&#45;degree phase&#46; At 30 it reports 70&#46;4 instead of 180&#46; Null&#45;selection restore correctly clears information/highlight, but the angle clause fails; source cause is the line&#45;8 angles destructuring&#46;

#### T4

Status: fail. Rating: 0. Method: http.

Action: Return to Full overview at mobile 390x844, time 0&#46;01, with Jade selected; scroll to the entire scene&#46; Expected: the full solar system is visible with other state preserved&#46; Observed: time/speed/selection were preserved, but Jade and substantial outer&#45;orbit portions were outside the horizontally clipped 357x428 canvas&#46; The full&#45;system overview framing does not adapt to the narrow aspect ratio&#46; Desktop overview was also exercised&#46;

#### T5

Status: fail. Rating: 0. Method: http.

Action: Capture tracking obs&#45;1 at time 30/2x/Jade, change live fields, restore, zoom out with wheel, and restore the same snapshot again&#46; Expected: paused captured state and deterministic standard framing independent of manual zoom&#46; Observed: captured time/speed/selection/mode returned and playback paused, but the second restore retained the zoomed&#45;out scene instead of the original standard framing &#40;sun diameter approximately 91 versus 114 screenshot pixels in the same viewport&#41;&#46; restore&#40;&#41; in line 8 does not reset theta/phi/dist&#46; Null&#45;selection and fractional tracking restores were also exercised; their fields restored but angles failed as noted in S8&#46;

#### U1

Status: fail. Rating: 0. Method: http.

Action: Drag the canvas in paused overview and tracking using real pointer down/moves/up; zoom by wheel&#46; Expected: camera orbit and zoom respond without changing simulation&#46; Observed: wheel changed scale, but a 50&#45;pixel/25&#45;move tracking drag left the viewing orientation unchanged&#46; Independent scratch&#45;page input confirmed pointerdown, pointermoves with buttons=1, and pointerup delivery&#46; Source line 10 accesses renderer&#46;dataset instead of renderer&#46;domElement&#46;dataset, preventing the drag path&#46; All playing/mode combinations were not exhaustively re&#45;run after this confirmed required paused&#45;orbit failure&#46;

#### U2

Status: fail. Rating: 0. Method: http.

Action: At mobile paused time 30 tracking Azure, drag from its center &#40;186,630&#41; to &#40;261,630&#41;, using 25 moves of 3 CSS pixels before releasing over Ember&#46; Expected: cumulative drag must not be interpreted as selection&#46; Observed: selection and tracking changed from Azure to Ember despite no click intention; time stayed 30&#46;00&#46; This is a 75&#45;pixel cumulative drag failure&#46; DOM controls did not visibly select the scene, but the required drag&#45;separation clause fails&#46;

#### U5

Status: fail. Rating: 0. Method: http.

Action: At both 1440x900 and 390x844 use Enter on selection/playback/view/capture/restore/delete/import/export, arrows/Home on ranges, keyboard text insertion, and Tab from the textarea onto a snapshot Restore button&#46; Expected: accessible names, visible persistent focus, operable Tab navigation without lost focus&#46; Observed: primary keyboard actions worked, but focused Restore had focus&#45;visible=true immediately and lost focus to BODY within 120 ms at both sizes, while paused&#46; renderUI&#40;&#41; clears and rebuilds all snapshot buttons every animation frame&#46; Direct locator Enter can activate before replacement, but normal keyboard focus is unstable&#46; A pointer Restore attempt also produced Element is detached from the DOM&#46;

#### L2

Status: fail. Rating: 0. Method: http.

Action: Measure 390x844, exercise keyboard selection/speed/seek/playback/view/snapshot/JSON controls, and scroll to the full overview scene at time 0&#46;01&#46; Expected: scene and primary controls remain usable without horizontal clipping&#46; Observed: actual 390x844 with 373/373 client/scroll width; controls were reachable and no horizontal document overflow occurred, but the fixed camera framing clipped Jade and large outer&#45;orbit portions outside the 357x428 scene&#46; Vertical scrolling could not recover off&#45;canvas content&#46;

#### D6

Status: fail. Rating: 0. Method: http.

Action: Reload after import, deletion and capture; corrupt only the app key; then fill a unique reviewer quota key to 5242628 characters and capture with capacity available&#46; Expected: persistence/recovery plus visible saving&#45;unavailable notice on a real failed write&#46; Observed: import/delete/capture persisted and corrupt data recovered with notice/usable controls&#46; Under quota pressure, capture added obs&#45;2 in memory while persisted solar&#45;observations remained the prior two&#45;entry JSON, yet the visible status said Snapshot captured rather than saving unavailable&#46; save&#40;&#41; sets the error then capture overwrites it with success in line 9&#46; Original storage was restored and filler removed&#46; Unavailable&#45;storage read denial was not tested; the confirmed write&#45;failure notice violation already fails this grouped check&#46;

### Missing proof

#### S2

Status: unverified. Rating: Unverified. Method: http.

Fresh and reloaded pages showed automatic motion, 1&#46;00x, Overview, disabled tracking, and explicit no&#45;selection information, with the first observed times around 5 seconds after navigation&#46; The exact time&#45;zero initialization frame was not observed because openTab returned after that frame&#46; Source defaults alone do not prove the complete startup condition; ordinary load elapsed time is not classified as a failure&#46;

#### S7

Status: unverified. Rating: Unverified. Method: http.

Observed a loaded three&#46;min&#46;js HTTP resource, canvas data&#45;engine three&#46;js r160, actual rendered moving spheres/rings and zoom changes; source creates SphereGeometry, PerspectiveCamera and WebGLRenderer&#46; However camera dragging did not rotate the scene &#40;confirmed under U1&#41;, so the required rendered orbit/depth/occlusion perspective inspection could not be completed&#46; The dependency and source alone are not a full pass&#46;

#### D3

Status: unverified. Rating: Unverified. Method: http.

The page rendered and the required Three&#46;js script was present in resource timing&#46; However console and pageerror subscriptions returned empty arrays for deliberately generated errors on a reviewer&#45;owned scratch page, including a later readback&#46; Subsequent noninterfering app error listeners also yielded no captured errors and could not cover startup&#46; Therefore clean console/resource&#45;error coverage is not established; source&#45;suspected exceptions are not reported as captured diagnostics&#46;

### Verified passes

#### S1

Status: pass. Rating: 5. Method: http.

Action: Observe the sun/three colored planets during motion and seek 0, 12, 24, and 48 at the same desktop camera&#46; Expected: distinct planets on circular 12/24/48&#45;second orbits with fixed phases&#46; Observed: one orange sun, larger red Ember, medium blue Azure, and smaller green Jade; Ember returned at 12, Azure at 24, and all three at 48, with Jade opposite its initial position at 24&#46; Rendered paths followed the projected orbit rings&#46; Source phases are 0, 1&#46;2, and 2&#46;4 radians; displayed&#45;angle defects are separately recorded under T2/S8&#46;

#### S3

Status: pass. Rating: 5. Method: http.

Action: Pause/seek 17, compare rendered output and state over 3&#46;1 seconds, select planets while paused, then resume from sought times&#46; Expected: frozen time/positions while paused and continuous resumption&#46; Observed: time remained exactly 17&#46;00 s, selected information stayed fixed, and before/after screenshots were byte&#45;identical&#46; Both selection routes remained usable&#46; Resumption advanced from the sought time rather than resetting; the 30&#45;second/2x scenario advanced 6&#46;40 seconds over 3&#46;2011 wall seconds&#46; Camera usability is independently failed in U1&#46;

#### S4

Status: pass. Rating: 5. Method: http.

Action: Set 0&#46;25x, 2x, and 4x at paused time 17, resume and sample each for at least 3&#46;1 real seconds, and change speed during playback before pausing again&#46; Expected: current time/positions preserved by the change and proportional simulation rates&#46; Observed: each paused setting retained 17&#46;00 and the same selected readout; measured deltas were 0&#46;78/6&#46;21/12&#46;44 simulation seconds over 3&#46;1014/3&#46;1027/3&#46;1058 wall seconds&#46; No reset occurred, and rendered orbital progression retained the 12/24/48 period relationship&#46;

#### S5

Status: pass. Rating: 5. Method: http.

Action: Seek 0/30/120 and noncanonical 17/73, alter speed/playback/selection, revisit 17/30/73, then reload and repeat using the same camera/selection/speed&#46; Expected: immediate pause/time/position updates and deterministic phases&#46; Observed: seeks paused with exact target time; tracking updated on seek&#46; Repeated full screenshots at 17, 30, and 73 were byte&#45;identical both after intervening edits and after reload&#46; Imported 17&#46;25 restored exactly; incorrect angle readouts are separately failed in T2/S8&#46;

#### S6

Status: pass. Rating: 5. Method: http.

Action: Seek 119, set 4x, resume and wait 1&#46;3 seconds; then seek 73 and resume&#46; Expected: stop at 120 without wrap/overshoot and resume from an earlier time&#46; Observed: 120&#46;00 s with Resume button after the wait, followed by 73&#46;86 s and Pause after seeking/resuming at 73&#46;

#### T1

Status: pass. Rating: 5. Method: http.

Action: For each Ember/Azure/Jade, select from its labeled list, select another planet, then click the target rendered planet with the real pointer; repeat paused and playing&#46; Expected: both routes select/highlight the same planet&#46; Observed: all six canvas cases changed information and list selection to the intended name, agreeing with the corresponding list route; visible emissive highlight moved with selection&#46; Playing readouts advanced naturally without resetting time&#46;

#### T3

Status: pass. Rating: 5. Method: http.

Action: Track Jade during playback, pause/seek 30, switch to Azure, and restore a null&#45;selection overview snapshot&#46; Expected: selected target stays centered, seeks/switches synchronize target and name/highlight while preserving other state, and tracking requires selection&#46; Observed: across two waits totaling 3&#46;304 seconds the mobile Jade center stayed at approximately &#40;168,567&#41; in a scene bounded by x=7&#46;&#46;329,y=374&#46;&#46;760 screenshot pixels, within tolerance of its center&#46; Seeking and switching to Azure placed its highlighted center there at 30&#46;00; desktop switching retained playback/speed&#46; Null&#45;selection restore disabled tracking and cleared selected information&#46; Numeric angle accuracy is failed separately under T2&#46;

#### U3

Status: pass. Rating: 3. Method: http.

Action: Use playback, selection, tracking, snapshot capacity/restore, and JSON controls at both sizes&#46; Expected: understandable purpose and feedback&#46; Observed: plainly labeled controls, explicit paused/running button text, time/speed outputs, stable ID mapping, tracking labels, capacity and JSON errors make normal use understandable&#46; The compact utilitarian interface merits the ordinary clear/usable anchor 3, not exceptional polish; misleading saving success is independently failed under D6&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Capture obs&#45;1/obs&#45;2/obs&#45;3, attempt a fourth, delete middle obs&#45;2 using Enter, capture again, and export&#46; Expected: max three, visible rejection, retained values/order, append&#45;only new capture and no live&#45;state changes&#46; Observed: fourth capture showed the capacity message with three unchanged entries&#46; Deletion retained obs&#45;1/obs&#45;3 and time 73/2x/Jade/overview; recapture appended a noncolliding obs&#45;2 after them&#46; Export order was obs&#45;1, obs&#45;3, obs&#45;2&#46; Reuse of a deleted ID did not collide with any retained snapshot&#46;

#### V1

Status: pass. Rating: 3. Method: http.

Action: Inspect rendered overview/tracking scenes and selected planets at both viewports&#46; Expected: readable scene and selection&#46; Observed: distinct red/blue/green bodies, unobtrusive orbit rings, and brighter selected planets provide ordinary clear visual separation; the sun is immediately recognizable&#46; Shapes/shading are plain and not exceptionally polished&#46; Mobile clipping is assessed in L2/T4 rather than used as an additional visual penalty&#46;

#### V2

Status: pass. Rating: 3. Method: http.

Action: Inspect controls, information, snapshots and JSON panels at desktop/mobile&#46; Expected: coherent legible typography, spacing and hierarchy&#46; Observed: consistent dark&#45;blue panels, restrained borders, readable light text, visible headings and grouped controls; dense small snapshot details remain legible&#46; This is coherent ordinary execution &#40;3&#41;, without exceptional spacing or typographic refinement&#46;

#### L1

Status: pass. Rating: 5. Method: http.

Action: Measure the real desktop iframe and operate every primary control including snapshot/JSON workflows&#46; Expected: usable scene/controls without horizontal scrolling or panel occlusion at 1440x900&#46; Observed: innerWidth/innerHeight were 1440/900 and clientWidth/scrollWidth 1440/1440; the scene and right control panel were separate and all controls reachable&#46; Snapshot activation/focus and camera defects are functional findings under T5/U1/U5, not desktop panel overlap&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspect output/index&#46;html and all output&#45;directory resources against the served root&#46; Expected: referenced local assets present without repository/build dependencies&#46; Observed: the only output file is index&#46;html, containing inline CSS/application JS and one permitted HTTPS CDN reference to Three&#46;js r160; there are no missing local paths or build/root dependencies&#46; File&#45;protocol execution is outside this review&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Export empty/captured/imported lists through the Snapshot JSON textarea; round&#45;trip IDs including spaces and Case/case; export while paused and playing&#46; Expected: only required version/snapshot fields, exact types/order/IDs, exposed planet mapping and unchanged state&#46; Observed: empty export was &#123;&#34;version&#34;:1,&#34;snapshots&#34;:&#91;&#93;&#125;; populated exports contained exactly id/time/speed/selectedPlanet/viewMode, no angles/camera/playback/extra keys, preserved order and case/spaces, and used ember/azure/jade displayed beside names&#46; Export did not change list or displayed live fields; scene&#45;region screenshot was identical before/after paused export&#46;

#### D5

Status: pass. Rating: 5. Method: http.

Action: Run the 34 invalid classes listed in notes against a nonempty paused list; import valid two/three/empty lists, fractional and endpoint values, null/selected overview and tracking, case&#45;distinct/space IDs and extra keys; repeat valid import while playing&#46; Expected: atomic rejection and list&#45;only valid replacement without coercion or scene changes&#46; Observed: all 34 errors were visible and all list/live readouts unchanged, including valid&#45;first/invalid&#45;later&#46; Every invalid case and valid list shape had a byte&#45;identical scene&#45;region screenshot&#46; Valid exports retained exact IDs/order/types, ignored supplied angles/extra keys, empty input cleared only snapshots, and playing time continued without reset&#46; Derived angles themselves are wrong under S8 but were not overridden by imported values&#46;

## Verification and limits

Browser method: http.
Viewports: 1440x900, 390x844.
Coverage: 15 passed, 9 failed, 3 unverified.

### Notes

- User selected benchmark&#45;review for the result of this Terra low invocation&#46; Reviewer model is concrete current&#45;host metadata: openai&#45;codex/gpt&#45;6&#45;astra&#46; Reviewer effort is null: the host exposes generic/default role configuration but no concrete current&#45;execution effort; generic defaults do not establish the review selection&#46;
- Host&#45;required aside&#45;browser was additionally loaded; aside guide and aside guide repl were read&#46; Only connected Aside direct REPL will be used; no separate review model, file navigation, or artifact repair&#46;
- Preregistered tolerances: rate samples at least 3 wall seconds, error &#60;= max&#40;0&#46;15 simulation seconds, 5% expected delta&#41; plus one displayed increment; paused values one displayed increment and 1 degree; predicted orbital angles 2 degrees plus display rounding; repeated standard framing 2 CSS pixels; tracking center within 10% of scene smaller dimension&#46; JSON identity/order/type and paused state preservation are exact&#46;
- Task routing: saved prompt exactly matches the 15&#45;requirement solar&#45;system&#45;observatory&#45;2 contract&#46; Review used format 4&#46;0 / solar&#45;system&#45;4&#46;0, independently of generation status&#46; No previous review findings were reused&#46;
- Viewport checks used unchanged output in the documented secure loopback iframe fixture, not top&#45;level viewport emulation or a touch device&#46; Actual inner sizes were 1440x900 and 390x844; desktop document client/scroll widths 1440/1440 and mobile 373/373 &#40;vertical scrollbar&#41;&#46; Both child contexts reported isSecureContext=true&#46;
- Original eight&#45;step sequence: observed automatic movement; the first moving&#45;scene coordinate attempt did not select, so list selection was used to proceed with tracking&#46; Paused/seeks at 30, paused speed 2x, measured resume, switching Jade to Azure during tracking, overview, and repeat seek were observed&#46; Later all three canvas/list selection routes were independently exercised both paused and playing&#46;
- Additional sequences 1&#45;6 were exercised in their saved order: capture/edit/restore with zoom; three&#45;entry capacity/delete/recapture/export; invalid imports; valid fractional and null/tracking round&#45;trips; import/delete/capture reloads plus corruption/quota fixtures; keyboard controls at both sizes&#46; Confirmed failures and any incomplete clauses are assigned to individual checks below&#46;
- Rate samples were serialized with one review scene open and visible&#46; Observed &#40;speed, wall seconds, simulation seconds&#41;: &#40;0&#46;25,3&#46;1014,0&#46;78&#41;, &#40;2,3&#46;1027,6&#46;21&#41;, &#40;4,3&#46;1058,12&#46;44&#41;; the original 2x scenario measured 3&#46;2011 wall seconds and 6&#46;40 simulation seconds&#46; All fit preregistered tolerances&#46;
- Observation/driver limits: Aside frame evaluation runs in an isolated world &#40;THREE global was inaccessible&#41;, and frames&#40;&#41; returned detached handles after reload; fresh current handles were selected&#46; These were not attributed to the app&#46; Browser console/pageerror subscriptions failed a deliberate scratch&#45;page positive control, so empty diagnostic arrays do not establish a clean console&#46; Noninterfering error listeners also yielded no qualifying errors; no startup diagnostic coverage is claimed&#46;
- Pointer transport was separately exercised on a reviewer&#45;owned scratch page: real pointerdown, 20 intermediate pointermove events with buttons=1, and pointerup were observed&#46; App drags still failed to orbit and a 75&#45;CSS&#45;pixel/25&#45;move drag changed selection&#46; Static source corroborates misuse of renderer&#46;dataset rather than renderer&#46;domElement&#46;dataset, but no captured console exception is claimed&#46;
- JSON rejection matrix covered 34 cases against a nonempty paused list: malformed JSON; null/array roots; missing root fields; string/unsupported version; nonarray snapshots; each missing entry field; null/array entries; empty/numeric IDs; string/boolean time and speed; both time and speed range directions; invalid/numeric/string&#45;null planet IDs; invalid mode; null tracking; duplicates; four entries; valid&#45;first/invalid&#45;later; and 1e400&#46; Every case showed an error and preserved list and displayed live fields&#46; A separate scene&#45;region screenshot comparison was byte&#45;identical for all 34 cases, valid three/empty/two lists, and export&#46;
- Valid JSON probes covered time 0/120/17&#46;25/30, speed 0&#46;25/4/1&#46;5, no&#45;selection overview, selected overview, selected tracking, exact IDs containing spaces and case&#45;distinct Case/case, extra root/entry keys, ignored imported angles, empty clearing, and playing imports&#46; Export contained only the required fields&#46; Wrong derived angle values remain an S8/T2 failure, not evidence that extra supplied angles were accepted&#46;
- Storage fixtures used only solar&#45;observations and the verified&#45;absent reviewer key terra&#45;review&#45;quota&#45;20261018 on the isolated 127&#46;0&#46;0&#46;1:8865 origin&#46; Malformed and schema&#45;invalid saved payloads were exercised&#46; A bounded &#60;=8 MiB search filled 5242628 characters, then a real capture left persisted JSON unchanged while adding an in&#45;memory snapshot&#46; The exact pre&#45;fixture app value was restored and the filler key removal verified&#46; No unrelated storage was read or cleared; unavailable&#45;storage read denial was not exercised&#46;
- Cleanup: all review&#45;created app and scratch tabs were closed, both tracked output/viewport servers were stopped, and the noninterfering error listener was removed&#46; Screenshots were inspected transiently only, never retained in the run&#46; Source prompt, run record, and HTML were not modified&#46; Browser verification used the host&#45;required additional aside&#45;browser skill; isolation from that extra skill is not claimed&#46;
