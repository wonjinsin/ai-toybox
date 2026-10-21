# Benchmark review

## Conclusion

Status: partial.
Mandatory verdict: fail.
Quality total: 92.40 / 100. Conservative lower bound. Range: 92.40 to 94.40.
Quality is uncapped. The mandatory verdict is independent of the quality score.
Mandatory failures (1): D6.
Mandatory unverified (0): None.
Missing checks (1): D3.
Missing reviewer metadata: None.
Format: 5&#46;0. Rubric: solar&#45;system&#45;5&#46;0.
Reviewed at: 2026&#45;10&#45;20T16:01:19&#46;472296&#43;00:00.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: high.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; determined score; full category denominator |
| selection | 5.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 4.80 | 15 | U1, U2, U3, U4, U5; determined score; full category denominator |
| visual | 4.00 | 15 | V1, V2; determined score; full category denominator |
| responsive | 5.00 | 15 | L1, L2; determined score; full category denominator |
| delivery | 3.00 to 4.00 | 10 | D2, D3, D4, D5, D6; conservative lower bound; unknown ratings range from 0 to 5 with the full category denominator |

## Findings

### Confirmed failures

#### D6

Status: fail. Rating: 0. Method: http.

Action: Reload after capture, deletion and valid import; replace only orbit&#45;observatory&#45;v1 with each invalid&#45;payload class; exercise bounded real quota exhaustion and restore fixtures&#46; Expected: stable persisted IDs/order/values/angles with fresh live defaults; every invalid saved payload produces an empty log plus a recovery notice; failed writes visibly report saving unavailable while memory controls work&#46; Observed: normal capture/delete/import persistence and defaults passed, and all 35 invalid saved payloads produced empty lists&#46; However saved empty&#45;string IDs, numeric IDs and duplicate IDs displayed Saving unavailable; observations are in memory only&#46; They did not display an invalid&#45;data recovery notice, although storage was available&#46; Reproduce with version 1 and a snapshot whose id is empty &#40;other fields valid&#41;, or duplicate two valid IDs, then reload&#46; index&#46;html:24,34 misclassifies the ID&#45;validation error; other tested invalid classes showed Recovered invalid saved observations; the log is empty&#46; Quota fixture reached 5,242,405 filler characters under the 8&#45;MiB bound; a real capture failed to persist, kept the prior raw payload, added an in&#45;memory card and visibly said Saving unavailable; changes remain in memory only&#46; Playback, selection and deletion remained usable&#46; Exact original storage absence was restored and filler removed&#46; Read&#45;denial itself was not induced; the permitted real write&#45;failure fixture covered saving unavailability&#46;

### Missing proof

#### D3

Status: unverified. Rating: Unverified. Method: http.

Console/page&#45;error coverage is unverified&#46; A reviewer&#45;owned scratch page registered console and pageerror subscriptions, then deliberately called console&#46;error and threw an asynchronous exception; the diagnostic array remained empty&#46; Therefore empty subscriptions cannot prove a clean artifact, and startup/ordinary&#45;console coverage was unavailable&#46; A brief post&#45;load listener experiment cannot supply the missing startup coverage and was discarded on reload&#46; Required local resources were observed rendering and returned HTTP 200 with matching final source bytes; no qualifying artifact error was established&#46; Review&#45;triggered stale&#45;frame, navigation&#45;readiness and isolated&#45;world ReferenceError messages were tool/evaluation limitations, not artifact errors&#46;

### Verified passes

#### S1

Status: pass. Rating: 5. Method: http.

Action: Inspect the final bundled&#45;renderer overview, seek 0/17/30/73/120, and sample each selected orbital angle during playback&#46; Expected: one sun, three differently sized/colored planets, circular fixed&#45;phase 12/24/48&#45;second orbits&#46; Observed: gold sun and small mint Aster, medium blue Azure, larger orange Ember on three circular paths; perspective projection and scene clicks agreed with their positions&#46; At time 0 the angles were 0/80/190 degrees; at 30 they were 180/170/55; at 17 they were 150/335/317&#46;50&#46; Playback samples: Aster 6&#46;38 to 99&#46;57 degrees over 0&#46;21 to 3&#46;32 simulation s &#40;3&#46;102 wall s&#41;; Azure 83&#46;20 to 129&#46;81 over 0&#46;21 to 3&#46;32 &#40;3&#46;103 wall s&#41;; Ember 191&#46;71 to 215&#46;01 over 0&#46;23 to 3&#46;34 &#40;3&#46;103 wall s&#41;, matching 30/15/7&#46;5 degrees per simulation second&#46; Final shaded rendering and changing orbital positions were directly inspected, not inferred solely from constants&#46;

#### S2

Status: pass. Rating: 5. Method: http.

Action: Reload the same&#45;origin artifact and observe immediately and after 3&#46;106 wall seconds&#46; Expected: time starts at 0, 1x, running, no selection, overview, followed by automatic motion&#46; Observed: earliest completed&#45;load readouts were 0&#46;02–0&#46;04 s, speed 1, Pause button, empty selected list, explicit No planet selected, and FULL SYSTEM overview&#46; Final replay advanced from 0&#46;04 to 3&#46;15 s in 3&#46;106 wall s with rendered motion&#46; The small initial elapsed interval is normal load time; inspected initialization starts at zero&#46;

#### S3

Status: pass. Rating: 5. Method: http.

Action: Pause during playback, wait 3&#46;1 seconds, select/orbit while paused, and resume&#46; Expected: frozen time and all positions, usable selection/camera, continuous resumption&#46; Observed: paused 73&#46;00&#45;s scene&#45;region image was byte&#45;identical after 3&#46;1 seconds and selected information unchanged; paused 17&#46;00&#45;s rate setups also stayed exactly 17&#46;00&#46; Canvas/list selection and orbit/zoom remained usable&#46; Resuming the sought 30&#45;s state advanced smoothly through 30&#46;40 to 36&#46;61 s at 2x without a reset or visible position jump&#46;

#### S4

Status: pass. Rating: 5. Method: http.

Action: Set 0&#46;25x, 2x and 4x paused and playing, then resume each paused setting and measure at least 3 seconds&#46; Expected: rate changes only, preserved positions/time on speed changes, unchanged relative speeds&#46; Observed: at paused 73&#46;00 s every setting preserved identical rendered scene bytes&#46; Paused&#45;resume samples over 3&#46;104 wall s were 17&#46;06→17&#46;83 at 0&#46;25x, 17&#46;43→23&#46;64 at 2x and 17&#46;85→30&#46;28 at 4x&#46; After refreshing the exact bundled dependency, playing&#45;change samples were 17&#46;22→17&#46;99 over 3&#46;105 s, 17&#46;99→24&#46;21 over 3&#46;106 s and 24&#46;26→36&#46;69 over 3&#46;107 s respectively&#46; Transition reads differed only by natural frame advance; orbital information remained consistent with fixed phases and periods&#46;

#### S5

Status: pass. Rating: 5. Method: http.

Action: Use the actual time slider for 0,30,120,17,73; revisit 30/17/73 after speed, playback and selection changes, then reload and revisit; restore imported 17&#46;25&#46; Expected: seeking pauses and updates time, positions, selected angle and active tracking immediately; fixed phases survive reload&#46; Observed: all seeks paused at the requested value&#46; All&#45;three angle vectors were 0:&#91;0,80,190&#93;, 30:&#91;180,170,55&#93;, 120:&#91;0,80,10&#93;, 17:&#91;150,335,317&#46;50&#93;, 73:&#91;30,95,17&#46;50&#93;&#46; Final exact&#45;bundle scene images at 30/17/73 were identical both after intervening edits and after reload at matched overview/selection/viewport&#46; Tracking sought 73 immediately centered Azure with angle 95&#46;00&#46; Imported 17&#46;25 displayed derived &#91;157&#46;50,338&#46;75,319&#46;38&#93;&#46;

#### S6

Status: pass. Rating: 5. Method: http.

Action: Seek 119, set 4x, resume, wait 3&#46;1 seconds; seek 73 and resume&#46; Expected: stop exactly at 120 without wrapping and allow earlier playback&#46; Observed: 120&#46;00 s, Resume button and Ember 10&#46;00 degrees remained at the endpoint; earlier playback resumed to 73&#46;85 s at 4x, then paused at 74&#46;12&#46; No overshoot or wrap was observed&#46;

#### S7

Status: pass. Rating: 5. Method: http.

Action: Inspect the loaded local Three&#46;js dependency and index renderer/geometry/camera code, verify exact HTTP bytes, and orbit/zoom the final rendered scene&#46; Expected: genuine Three&#46;js 3D geometry, perspective and occlusion&#46; Observed: index&#46;html:35–54 constructs WebGLRenderer, PerspectiveCamera, spheres and torus highlights&#46; Exact served index and bundled library matched the saved files; final loaded resource size matched the bundle after cache refresh&#46; Orbiting changed orbital ellipse perspective, shaded sphere lighting and torus front/back occlusion; zoom changed apparent sphere size while the selected target remained centered&#46; This was live rendering rather than a flat image or import&#45;only claim&#46;

#### S8

Status: pass. Rating: 5. Method: http.

Action: Capture paused snapshot A at 30/2x/Ember/tracking, change every live field and play, restore it; capture while playing; export, re&#45;import and reload captured IDs; restore fractional imported time&#46; Expected: immutable complete captured values, unique stable IDs, derived displayed angles, no live mutation&#46; Observed: A ID obs&#45;mvguskr6&#45;1 retained 30&#46;00 s, 2x, ember, tracking and &#91;180,170,55&#93; degrees after unrelated edits&#46; Captured values and IDs survived actual export/import/reload equality checks and A restored the same fields&#46; A playing capture at approximately 0&#46;10 s retained its displayed angles across reload&#46; Later paused capture at 17&#46;29 s preserved exact live fields and identical camera/scene bytes; snapshot list showed every field and all three angles&#46; Fractional 17&#46;25 imports ignored supplied angles and displayed &#91;157&#46;50,338&#46;75,319&#46;38&#93;&#46;

#### T1

Status: pass. Rating: 5. Method: http.

Action: Select Aster, Azure and Ember through both labeled list and canvas, paused and playing&#46; Expected: one shared selection with visible highlight&#46; Observed: final exact&#45;bundle canvas clicks at predicted and visually inspected centers changed from a different prior selection to select&#45;aster, select&#45;azure and select&#45;ember in both playback states&#46; List buttons selected the same IDs&#46; Information and mint highlight rings agreed; each planet’s tracking highlight was visually inspected&#46;

#### T2

Status: pass. Rating: 5. Method: http.

Action: Inspect each selected information panel during playback and seeks, then restore a null&#45;selection overview snapshot&#46; Expected: correct name, period and rendered angle; explicit cleared no&#45;selection state&#46; Observed: Aster/Azure/Ember showed 12/24/48 simulation s, their live angles advanced at the measured rates, and at 30 s showed 180/170/55 degrees matching rendered locations&#46; Startup and null restore showed No planet selected, no pressed planet button, no highlight and disabled tracking; prior Azure information was cleared&#46;

#### T3

Status: pass. Rating: 5. Method: http.

Action: Track Azure over more than 3 seconds at 2x, seek 73 while tracking, switch planets, and inspect no&#45;selection tracking control&#46; Expected: actual selected center stays within 10% of scene’s smaller dimension; target/info/highlight switch without live resets; no&#45;selection tracking unavailable&#46; Observed: in the 347&#46;12x340 CSS&#45;pixel scene, Azure’s visible center stayed at approximately the scene center &#40;visually within about 3 CSS pixels, versus 34 allowed&#41; at displayed times 30&#46;53,33&#46;89,37&#46;25 and after seeking 73&#46;00&#46; Aster, Ember and Azure each centered with matching information/highlight&#46; Playing switch to Ember retained 2x and running time 37&#46;20; paused switches retained 73&#46;00/2x&#46; No&#45;selection tracking was disabled&#46;

#### T4

Status: pass. Rating: 5. Method: http.

Action: Return from tracking each planet to overview, including while playing&#46; Expected: full system visible and selection/time/speed/playback preserved&#46; Observed: all three orbits, the sun and planets fit standard overview at both requested CSS sizes&#46; Each paused return retained 73&#46;00 s/2x/Resume and the chosen planet&#46; In the ordered playing scenario, Ember remained selected, speed 2 and playback running while time naturally advanced 37&#46;20→37&#46;76&#46; Overview reset framing without clearing selection&#46;

#### T5

Status: pass. Rating: 5. Method: http.

Action: Restore selected tracking and unselected overview snapshots after changing live fields and manually orbiting/zooming; repeat at the same viewport&#46; Expected: atomic captured&#45;state restore, paused playback, correct positions/info/target and deterministic standard framing&#46; Observed: overview restored 30&#46;00/1&#46;5x/null with explicit no&#45;selection information; tracking restored 17&#46;25/0&#46;25x/Azure and angle 338&#46;75&#46; In final exact&#45;bundle checks, both repeated restored scene&#45;region images were byte&#45;identical after camera/time/selection changes&#46; Original captured A also restored 30/2x/Ember/tracking after export/import/reload&#46;

#### U1

Status: pass. Rating: 5. Method: http.

Action: Orbit by real pointer drags and zoom by wheel in overview and tracking, both playing and paused&#46; Expected: camera responds in all four combinations without unintended simulation mutation&#46; Observed: final exact&#45;bundle images changed in every combination; orbit changed path perspective and sphere/ring occlusion, and zoom visibly enlarged/reduced planets&#46; Paused cases stayed at 30&#46;00 s; playing cases advanced naturally to 31&#46;47 and 31&#46;44 at 2x while selection and view mode stayed correct&#46;

#### U2

Status: pass. Rating: 5. Method: http.

Action: Start a fast drag on unselected Aster and a 36&#45;CSS&#45;pixel cumulative drag on unselected Ember using 24 small moves with 15&#45;ms intervals, with Azure selected; exercise list/range/button/snapshot/JSON controls&#46; Expected: drags do not select planets; DOM actions do not leak into scene input&#46; Observed: both final&#45;bundle drags left Azure selected while changing camera framing&#46; Paused capture/delete/export and all 35 invalid imports preserved exact live fields and scene&#45;region image bytes; valid import also preserved them&#46; List and seeking produced only their intended selection/position changes&#46; Tracking/overview changed camera only as requested&#46;

#### U3

Status: pass. Rating: 4. Method: http.

Action: Inspect playback, speed, selection, mode, snapshot and JSON feedback throughout both viewport workflows&#46; Expected: clear purpose and understandable current state and action results&#46; Observed: numbered observation panels, Pause/Resume, two&#45;decimal clock, numeric speed, explicit selected/no&#45;selection information, named tracking banner, captured values/angles, capacity message and detailed import rejection messages form a polished and coherent tool&#46; Rating 4: feedback is strong, but storage&#45;ID recovery is misleading as detailed under D6, and selection/information require considerable vertical travel on the narrow layout&#46; No separate mobile obstruction penalty is applied here&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Capture three distinct observations, attempt a fourth, delete the middle entry, recapture and export&#46; Expected: append&#45;only capacity three, visible rejection without replacement, stable remaining order/values, unique replacement ID and no live changes&#46; Observed: fourth capture displayed Capacity reached: delete an observation before capturing a fourth; all three cards were unchanged&#46; Deleting obs&#45;mvgut4ye&#45;2 preserved A then obs&#45;mvgut5ed&#45;3; recapture appended obs&#45;mvgut62q&#45;4 without collision&#46; Live paused fields remained stable, and independent final exact&#45;bundle capture/delete checks also preserved the rendered scene and live state&#46;

#### U5

Status: pass. Rating: 5. Method: http.

Action: At both actual viewports, navigate by Tab and Shift&#43;Tab and activate all primary controls with Enter, ranges with Home/End/ArrowRight, and textarea with keyboard select&#45;all/type&#46; Expected: meaningful names, visible focus, working actions and no trap&#46; Observed: selection, both pause/resume transitions, seek, speed, tracking/overview, capture/restore/delete and JSON export/edit/import all worked&#46; Seek reached 0 then 0&#46;01, speed reached 4, restore returned 30/1&#46;5x/null, and JSON import restored two entries&#46; Focus&#45;visible was true with a solid visible outline for each control; Shift&#43;Tab exited import to Export&#46; Accessible snapshots supplied meaningful names for the label&#45;associated ranges, textarea and numbered snapshot buttons&#46;

#### V1

Status: pass. Rating: 4. Method: http.

Action: Inspect exact bundled&#45;renderer overview/tracking, selected/unselected states, orbit/zoom and both CSS sizes&#46; Expected: readable scene and selection, supportive effects&#46; Observed: distinct shaded mint/blue/orange spheres, subtle orbit paths and stars, warm sun and high&#45;contrast mint selection rings stayed readable&#46; Tracking makes the selected sphere prominent and demonstrates depth&#46; Rating 4: polished and unobscured, but the smallest overview planet is quite small and some dark unlit surfaces have limited contrast&#46;

#### V2

Status: pass. Rating: 4. Method: http.

Action: Inspect desktop and narrow layouts, controls, information, three&#45;card snapshots, textarea and visible keyboard focus&#46; Expected: coherent legible typography, spacing, colors and panel hierarchy&#46; Observed: consistent dark panels, mint numeric accents, muted uppercase section labels, rounded controls and clear observation&#45;card hierarchy&#46; Mobile cards stack and labels wrap legibly; desktop cards share a row&#46; Rating 4: cohesive and polished with minor density/small&#45;text limitations in detailed snapshot and JSON content rather than exceptionally refined typography&#46;

#### L1

Status: pass. Rating: 5. Method: http.

Action: Verify child viewport and exercise every primary control, including keyboard snapshot/JSON operations, at desktop size; inspect scene and scrolled panels&#46; Expected: usable 1440x900 CSS&#45;pixel layout without horizontal scrolling or blocked controls&#46; Observed: inner size 1440x900, clientWidth=scrollWidth=1423 &#40;vertical scrollbar&#41;, secure context, scene approximately 1007x510&#46; Selection, tracking, pause/resume, seek, speed, capture/restore/delete, export and textarea import all operated&#46; Vertical scrolling exposed the log and JSON; no control was covered or horizontally hidden&#46;

#### L2

Status: pass. Rating: 5. Method: http.

Action: Load a fresh fixed&#45;size mobile iframe and operate every primary control, including the full keyboard sequence and JSON round&#45;trip&#46; Expected: usable 390x844 CSS&#45;pixel layout without horizontal scrolling or blocked controls&#46; Observed: inner size 390x844, clientWidth=scrollWidth=373, secure context and scene approximately 347x340&#46; Controls wrapped into a single column and snapshots stacked&#46; Selection, tracking, seek, speed, pause/resume, capture/restore/delete and JSON edit/export/import worked via real input; vertical scrolling was necessary but no panel blocked the controls&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspect output/index&#46;html and all local dependency paths beneath output/&#46; Expected: complete HTTP&#45;root artifact without build or repository&#45;root dependencies&#46; Observed: the only referenced local asset is three&#46;min&#46;js, present alongside index&#46;html&#46; Styles and application code are inline; no other local assets or build outputs are required&#46; Final HTTP reads of both files matched saved bytes&#46; Direct&#45;file execution is deliberately outside review scope&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Export empty/populated lists through the labeled textarea, inspect fields/types/order/IDs, re&#45;import and export while paused and playing&#46; Expected: only version and five required snapshot fields, stable actual IDs, no file picker or live/list mutation&#46; Observed: empty export was version 1 with an empty snapshots array&#46; Populated captured IDs and order round&#45;tripped exactly through import/export/reload&#46; Exports omitted derived angles, playback and camera fields; Aster=aster/Azure=azure/Ember=ember mapping was visible&#46; Paused export preserved exact scene bytes; playing export/import continued naturally, and an exact&#45;time comparison at 17&#46;25 confirmed camera framing preservation&#46;

#### D5

Status: pass. Rating: 5. Method: http.

Action: Submit 35 separate invalid payload classes against a nonempty list; valid empty/two/three&#45;entry payloads paused and playing; restore imported entries and export&#46; Expected: atomic complete validation, visible errors without list/live/camera changes, exact ID/order preservation, endpoints/fractions and extra keys accepted, angles derived&#46; Observed: malformed JSON; null/array roots and entries; every missing required field; string/unsupported version; non&#45;array snapshots; empty/numeric IDs; numeric&#45;string/boolean time or speed; both time/speed range violations; numeric overflow; unknown/numeric/empty/string&#45;null planet references; bad mode; null tracking; four entries; duplicate IDs; and valid&#45;prefix/invalid&#45;suffix all rejected visibly with unchanged cards/live state&#46; Final exact&#45;bundle repeats also preserved complete scene&#45;region bytes for every rejection&#46; Valid IDs Case, case and a single space survived exactly in order, with time 0/120/17&#46;25 and speeds 0&#46;25/4/1&#46;5; null/selected overview and tracking restored correctly&#46; Extra keys and supplied angles were ignored&#46; Empty import cleared only the list; playing import retained playback and custom camera &#40;confirmed by returning to the same exact time&#41;&#46;

## Verification and limits

Browser method: http.
Viewports: 1440x900, 390x844.
Coverage: 25 passed, 1 failed, 1 unverified.

### Notes

- Reviewer identity: current host metadata names openai&#45;codex/gpt&#45;6&#45;astra&#46; Effort high is the explicit applicable worker configuration supplied for this review session; no independent runtime override telemetry was available&#46; Host clock supplied the timezone&#45;qualified review timestamp&#46; Generator identity and recording measurements were not used as review evidence&#46;
- Scope: saved prompt matches solar&#45;system&#45;observatory&#45;2, all R1–R15, the eight&#45;step scenario and six additional sequences&#46; Rubric solar&#45;system&#45;5&#46;0 / format 5&#46;0 applied&#46; An all&#45;unverified 0&#46;&#46;100 checkpoint was finalized before artifact source or browser inspection&#46; No previous review was used&#46;
- Applied skills: benchmark&#45;review plus host&#45;required aside&#45;browser; read the actual Aside and direct&#45;REPL guides&#46; Connected Aside MCP direct repl only, no other model or descendants, no exec/CLI browser session, no security changes, no file:// test, and no repository build/lint/tests/formatters&#46; The Terra&#45;focused control profile does not match this artifact; observed control selectors were used instead&#46;
- Preregistered before browser observations: rates measured for &#62;=3 wall seconds away from endpoint, tolerance max&#40;0&#46;15 simulation seconds, 5% expected delta&#41;&#43;0&#46;01 displayed seconds; pause tolerance 0&#46;01 s and 1 degree with one&#45;frame transition allowance; angle predictions phases Aster 0, Azure 80, Ember 190 degrees and periods 12/24/48, tolerance 2 degrees&#43;0&#46;01 rounding; repeat seek/restore centers 2 CSS px at identical camera/viewport; tracking within 10% of smaller scene dimension at several samples over &#62;=3 s and immediately after seek; JSON/state identity exact except natural measured playback advance&#46; Probe profile selectors and storage key do not match this artifact; use observed native control mapping, not the Terra profile&#46;
- Browser setup: exclusive lease ReviewSolMedium held for all browser work, one target scene at a time&#46; Output&#45;only HTTP origin 127&#46;0&#46;0&#46;1:8922 and fixture&#45;only server 127&#46;0&#46;0&#46;1:8932&#46; Fresh iframe CSS viewports were measured at 1440x900 and 390x844 with isSecureContext=true&#46; Desktop client/scroll widths were 1423/1423 and mobile 373/373; vertical scrollbars explain the difference from innerWidth&#46; This is real embedded&#45;document CSS layout verification, not top&#45;level viewport or touch&#45;device proof&#46;
- Ordered acceptance replay: automatic 1x motion; canvas Azure selection and tracking; pause/seek 30; set 2x paused; resume and measure 30&#46;40→36&#46;61 over 3&#46;103 wall seconds; switch to Ember while tracking without reset; return to full overview preserving Ember/2x/running; pause and repeat 30&#46; Subsequent matched&#45;camera comparisons established identical 30&#45;second positions&#46; The full sequence was replayed after the cache identity issue described below was resolved&#46;
- Additional sequences ran in saved order: &#40;1&#41; immutable A capture, unrelated live edits and repeated restore; &#40;2&#41; three snapshots, fourth&#45;capacity rejection, middle deletion, unique recapture and export; &#40;3&#41; invalid import rejection/atomicity matrix; &#40;4&#41; fractional/null/tracking valid import, restore, round&#45;trip, playing import and empty&#45;list clearing; &#40;5&#41; capture/delete/import reloads, malformed/schema&#45;invalid storage and real bounded quota failure; &#40;6&#41; full keyboard controls at both CSS sizes&#46; The remaining fixed matrix then covered endpoint rates, noncanonical seeks, phase reload stability, all canvas/list selections, slow drag separation, camera mode/playback combinations, extended validation and diagnostic capability&#46; Confirmed D6 failure did not stop independent checks&#46;
- Fresh&#45;cache limitation and correction: the first owned navigation showed a stale cached document inconsistent with the served index; it was excluded&#46; A cache&#45;busted index navigation loaded the assigned artifact&#46; Later a default library fetch exposed a stale cached dependency; no&#45;store byte checks identified the exact saved files, then a cache&#45;reload GET refreshed that owned&#45;origin library and a document reload reported the correct loaded resource size&#46; No source or browser security setting was changed&#46; The final genuine shaded renderer, ordered acceptance, phase/repeat&#45;seek/restore images, tracking, camera modes, drag separation, selection, speed rates, JSON/capture camera isolation, both layouts and keyboard workflows were observed again with the exact bundled dependency&#46; Earlier target&#45;HTML DOM/storage results are retained where independent of the library; initial stale&#45;document findings are not used&#46;
- Transport limits: initial navigation readiness timed out even though a later snapshot showed the correct loaded page; frames&#40;&#41; retained a stale frame handle, resolved to the live child before observations&#46; Isolated&#45;world evaluation could not access application lexical globals, so no internal state was exposed or mutated&#46; An early range fill of &#46;25 was rejected by the driver; valid 0&#46;25 input was used&#46; These are reviewer transport limitations, not artifact failures&#46;
- Image evidence: nonzero screenshot clip offsets and locator screenshots unexpectedly included out&#45;of&#45;element header content&#46; Those crops were not treated as claimed element coverage&#46; Final visual comparisons used full viewport images or explicitly inspected top&#45;left regions containing the complete scene, excluding the changing scrollbar&#46; Whole&#45;page byte differences during error messages were not camera failures because scrollbar geometry changed&#46; A playing comparison based on a rounded 17&#46;29 readout was not exact&#45;time proof; the independent exact 17&#46;25 replay produced identical framing&#46; No tolerances were widened&#46; Transient images were inspected, not retained as run evidence&#46;
- D3 remains unverified: deliberately generated console and exception signals were absent from subscriptions on a reviewer&#45;owned scratch page&#46; Empty arrays cannot establish diagnostic cleanliness; a post&#45;load listener cannot establish startup or ordinary console coverage&#46; Functional/resource observations remain usable, and no artifact&#45;caused error was invented from this instrumentation gap&#46;
- Storage fixtures: only app key orbit&#45;observatory&#45;v1 and isolated reviewer key reviewSolMedium&#45;quota&#45;20261020 were touched&#46; Original app key was absent and saved in working memory&#46; All 35 rejection classes were also exercised as persisted payloads; ID errors exposed the confirmed recovery&#45;message defect&#46; Quota search never exceeded 8 MiB of string content and observed expected QuotaExceededError exceptions&#46; Actual capture failed to save while memory stayed usable&#46; Quota filler was removed, interim raw payload restored after each fixture, and final original key absence restored without a later reload altering it&#46; No unrelated storage was inspected or cleared; read&#45;denial recovery was not separately claimed&#46;
- Preservation and cleanup: working&#45;memory before/after comparisons confirm prompt&#46;md, run&#46;json, output/index&#46;html and output/three&#46;min&#46;js were unchanged&#46; Final HTTP source bytes matched the saved output files&#46; All owned scene/fixture/scratch tabs were closed; review&#45;sol&#45;medium&#45;http and review&#45;sol&#45;medium&#45;viewport both reported exited, and the exclusive lease was released&#46; No denied operation occurred&#46; Only reviews/review&#46;json and reviews/review&#46;md are retained review outputs; working draft is removed after final publication&#46;
