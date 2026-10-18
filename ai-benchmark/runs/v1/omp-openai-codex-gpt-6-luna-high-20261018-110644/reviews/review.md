# Benchmark review

## Conclusion

Status: complete.
Mandatory verdict: fail.
Published total: 69.00 / 100. Determined score. Range: 69.00.
Raw total: 80.70.
Confirmed cap: 69.
The confirmed 69 cap is existing review policy for a mandatory failure, not a calibrated quality measurement.
Missing checks (0): None.
Missing reviewer metadata: None.
Format: 4&#46;0. Rubric: solar&#45;system&#45;4&#46;0.
Reviewed at: 2026&#45;10&#45;18T07:30:03&#46;157Z.
Reviewer model: openai&#45;codex/gpt&#45;6&#45;astra.
Reviewer effort: high.

## Scores

| Category | Score / 5 | Weight | Basis |
| --- | --- | --- | --- |
| simulation | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; determined score; full category denominator |
| selection | 4.00 | 20 | T1, T2, T3, T4, T5; determined score; full category denominator |
| usability | 4.40 | 15 | U1, U2, U3, U4, U5; determined score; full category denominator |
| visual | 2.50 | 15 | V1, V2; determined score; full category denominator |
| responsive | 5.00 | 15 | L1, L2; determined score; full category denominator |
| delivery | 2.00 | 10 | D2, D3, D4, D5, D6; determined score; full category denominator |

## Findings

### Confirmed failures

#### T4

Status: fail. Rating: 0. Method: http.

Action: Return from tracking to overview at0&#46;1 and30 seconds in1440x900 with default standard framing&#46; Expected: full system visible while preserving selected planet,time,speed and playback&#46; Observed: overview preserved state, but only two planets were in the rendered canvas; the outer Mars body lay outside the frame, and the outer orbital path was clipped on both sides&#46; At30 seconds the sun,Mercury andVenus were visible while Mars was not&#46; Source uses fixed distance11 and field&#45;of&#45;view48 without fitting full orbit to aspect&#46;

#### U3

Status: fail. Rating: 2. Method: http.

Action: Inspect state feedback while restoring30/2x after changing speed to0&#46;25, then restore17&#46;25/1&#46;5 and resume; inspect capture/import/error feedback&#46; Expected: understandable and consistent live controls and status&#46; Observed: labels, names, IDs and JSON messages are discoverable, but restored live speed read2&#46;00x then1&#46;50x while the range still reported0&#46;25; Snapshot restored; playback paused remained visible after PLAYING resumed&#46; Repeated contradictory feedback fits qualitative anchor2, usable with recurring inconsistency&#46; Mobile layout is not penalized here&#46;

#### V1

Status: fail. Rating: 2. Method: http.

Action: Inspect overview/tracking at both viewports and multiple times&#46; Expected: readable bodies and selection with supportive effects&#46; Observed: spheres have distinct cyan,gold,red colors and sizes, restrained stars and clear selected&#45;list state&#46; However standard overview clips outer orbits and can hide Mars completely; mobile30 also clips Venus at the left edge&#46; Zooming out recovers the scene, but recurring clipping compromises observation&#46; Qualitative anchor2&#46;

#### D3

Status: fail. Rating: 0. Method: http.

Action: Attach noninterfering post&#45;load window error/unhandledrejection observers; import valid&#45;shaped snapshot with unknown selectedPlanet constructor and tracking, then click Restore&#46; Expected: invalid reference rejected and no artifact&#45;caused operational exception&#46; Observed: import incorrectly succeeded, then two real error events reported Uncaught TypeError: Cannot read properties of undefined &#40;reading x&#41;&#46; Render/UI remained stale after Restore&#46; Source validList uses inherited byId lookups and updateScene tries to copy an absent mesh position&#46; This observed exception establishes failure without trusting broken console subscriptions&#46; Startup/ordinary console&#46;error coverage remains unavailable&#46;

#### D5

Status: fail. Rating: 0. Method: http.

Action: Against nonempty list run malformed/type/range/missing&#45;field/duplicate/mixed&#45;prefix rejection matrix and valid empty,three&#45;entry,endpoint,fractional/case/space&#45;ID/extra&#45;key imports; also submit selectedPlanet constructor&#46; Expected: reject every unknown planet ID atomically without changes&#46; Observed:31 routine invalid cases showed rejection and exact unchanged visible list/live fields, and valid cases preserved IDs/order and derived angles&#46; But constructor, not mercury/venus/mars, was accepted and replaced the list with unknown&#45;reference; Restore then threw&#46; byId inherits Object&#46;prototype, so validation erroneously accepts inherited names&#46;

#### D6

Status: fail. Rating: 0. Method: http.

Action: Reload after import/delete/capture; inject malformed and version&#45;invalid saved payload; run bounded quota fixture with isolated rh&#45;review&#45;quota&#45;20261018 key, then capture through UI&#46; Expected: persistence with R2 defaults,recovery notices and visible saving&#45;unavailable on failed write while memory remains usable&#46; Observed: ordinary persistence and invalid&#45;data notices worked&#46; Quota bound reached5242804 filler characters with QuotaExceededError; capture appended an in&#45;memory entry but persisted payload stayed the empty list&#46; UI showed Observation captured&#46;, not saving&#45;unavailable, because capture overwrites persist failure notice&#46; Playback/selection/restore remained usable&#46; Removed filler and restored exact prefixture app value, confirmed filler null and original empty versioned object&#46; No read&#45;denial claim&#46;

### Missing proof

None.

### Verified passes

#### S1

Status: pass. Rating: 5. Method: http.

Action: Observe initial and moving scene, zoom out to see entire system, seek0/12/24/48 and compare rendered planet patches; compare all three angle readouts at0/17/30/73/120&#46; Expected: one sun,exactly three different&#45;sized/color planets on circular paths with periods12/24/48 and fixed phases&#46; Observed: sun plus cyan small Mercury,gold larger Venus,red intermediate Mars; circular rings project to ellipses and spheres move along them&#46; Mercury patch at12,Venus at24,Mars at48 were pixel&#45;identical to their0&#45;second patches with fixed camera/selection&#46; Intermediate positions differed&#46; Phases were0,63&#46;025357,126&#46;050715 degrees, matching readouts and the prescribed periods at every sought time&#46;

#### S2

Status: pass. Rating: 5. Method: http.

Action: Open fresh scene and reload after snapshot import, deletion and capture&#46; Expected: time zero, 1x, playing, no selection and overview, followed by motion&#46; Observed: reload reads were 0&#46;0 or 0&#46;1 seconds, 1&#46;00x, PLAYING, None, no angle/period and Full&#45;system overview&#46; Later 5&#46;2 and15&#46;0 second reads and changed rendered positions showed automatic motion&#46; Track was disabled without selection&#46;

#### S3

Status: pass. Rating: 5. Method: http.

Action: Pause during motion, seek30, wait3&#46;1 seconds, orbit/select while paused, and resume&#46; Expected: frozen time and all positions, usable camera/selection and continuous resume&#46; Observed: full screenshots were byte&#45;identical across3&#46;1 seconds at30&#46;0/0&#46;25x/Mars351&#46;0507149287812 degrees&#46; Paused orbit and selection worked, then resume read30&#46;2 and352&#46;48602742882565 degrees with natural continuation rather than reset&#46;

#### S4

Status: pass. Rating: 5. Method: http.

Action: Exercise0&#46;25x,2x,4x paused and playing; after all other review scenes and own viewport duplicates closed, measure each rate in one uninterrupted foreground call using paired DOM/performance&#46;now samples&#46; Expected: unchanged paused state and correct playback rate within max&#40;0&#46;15s,5% expected&#41;&#43;0&#46;1s&#46; Observed: paused17&#46;0 and Mercury150 degrees stayed fixed at all speeds&#46; Final visible1600x1000 single&#45;artifact samples were0&#46;25x:0&#46;9/3&#46;5068 wall seconds &#40;expected0&#46;8767,error0&#46;0233,bound0&#46;25&#41;;2x:7&#46;0/3&#46;5095&#40;expected7&#46;019,error0&#46;019,bound0&#46;45095&#41;;4x:14&#46;1/3&#46;5089&#40;expected14&#46;0356,error0&#46;0644,bound0&#46;80178&#41;&#46; Playing switches2/4/0&#46;25 preserved ongoing time,selection and playback; no reset/jump&#46; Earlier misses under multiple rendered scenes remain disclosed in notes, not attributed to the app&#46; Relative orbital periods were separately verified underS1&#46;

#### S5

Status: pass. Rating: 5. Method: http.

Action: Seek0,30,120,17,73 through the slider; revisit30/17/73 after speed/playback/selection changes at1440x900 with matching overview/Mars selection and framing; reload and repeat17; restore17&#46;25 via import&#46; Expected: paused time and deterministic phase&#45;derived positions&#46; Observed: Mercury/Venus/Mars angles were0/63&#46;025357/126&#46;050715 at0,180/153&#46;025357/351&#46;050715 at30,0/63&#46;025357/306&#46;050715 at120,150/318&#46;025357/253&#46;550715 at17,30/78&#46;025357/313&#46;550715 at73&#46; Repeated30/17/73 screenshots were byte&#45;identical&#46; Reload17 reproduced all three angles;17&#46;25 restore gave157&#46;5/321&#46;8/255&#46;4 rounded degrees&#46;

#### S6

Status: pass. Rating: 5. Method: http.

Action: Seek119 at4x, resume for3&#46;2 seconds, then seek73 and resume&#46; Expected: stop exactly120 without wrapping and resume earlier&#46; Observed:120&#46;0 seconds PAUSED, then77&#46;8 seconds PLAYING after earlier seek, with no overshoot or wrap&#46;

#### S7

Status: pass. Rating: 5. Method: http.

Action: Inspect local Three&#46;js path and renderer/geometry source, load over HTTP, orbit and zoom scene while paused/playing&#46; Expected: genuine three&#45;dimensional rendering, not import&#45;only or flat imitation&#46; Observed: local r160 loaded with670184 transfer bytes; source constructs WebGLRenderer,SphereGeometry,PerspectiveCamera and3D meshes&#46; Actual drags tilted projected orbital circles from shallow ellipses to steep nearly circular views, changing sphere shading, depth order and occlusion; wheel changed apparent size and perspective&#46;

#### S8

Status: pass. Rating: 5. Method: http.

Action: Capture paused30/2x/Mercury/tracking, change all live fields and play; restore/export/import/reload; capture again while playing&#46; Expected: immutable fields, unique stable IDs and time&#45;derived displayed angles, no live&#45;state changes on capture&#46; Observed: obs&#45;mvdhfa7m&#45;9cuii stayed30&#46;0s/2x/mercury/tracking with180&#46;0/153&#46;0/351&#46;1 degrees while live state became17&#46;2s/0&#46;25x/Mars/overview/playing&#46; Later captures had distinct IDs&#46; Fractional imported17&#46;25/1&#46;5/venus/tracking displayed157&#46;5/321&#46;8/255&#46;4 degrees and retained these through reload&#46; Capturing while playing preserved playing/null/overview and derived angles from its captured0&#46;1403 second time&#46;

#### T1

Status: pass. Rating: 5. Method: http.

Action: Click Mercury,Venus,Mars on canvas paused and playing at0&#46;25x after seeking0; choose each from labeled list in paused/playing sequences&#46; Expected: shared selection and visible highlight&#46; Observed: all six canvas results reported the intended name/ID, matching the active list button and enlarged/emissive sphere&#46; List route selected the same IDs and synchronized info even while paused&#46;

#### T2

Status: pass. Rating: 5. Method: http.

Action: Inspect initial no&#45;selection, each selected planet while playing and across seeks, and null&#45;selection restore&#46; Expected: name,period and angle match rendered selection; null clears prior data/highlight&#46; Observed: Mercury12s,Venus24s,Mars48s matched orbit/readouts; at30 their angles were180,153&#46;025357 and351&#46;050715 degrees, and during playback Mars angle advanced255&#46;0072 to261&#46;5063 as the rendered scene moved&#46; Null restore showed None and dashes, disabled tracking and removed highlight&#46;

#### T3

Status: pass. Rating: 5. Method: http.

Action: Track selected planets playing/paused, switch list selection during playback, sample moving Mars three times across3&#46;441 seconds, seek73 and attempt tracking without selection&#46; Expected: centered target and synchronized info/highlight, preserved clock/speed/playback, null selection rejection&#46; Observed: scene bounds were x17&#46;10,y79&#46;36,1195&#46;14x889&#46;77; Mars stayed visually centered at about614&#46;4,524 CSS pixels in all three samples and after73 seek, within2px of scene center614&#46;67,524&#46;24 and well inside88&#46;98px tolerance&#46; Selection switched Mercury toVenus at2x while PLAYING with natural time advance; tracking with null selection was disabled&#46;

#### T5

Status: pass. Rating: 5. Method: http.

Action: Restore selected tracking17&#46;25/1&#46;5/Venus and null overview30/1&#46;5 after unrelated live changes and manual orbit/zoom, then repeat each at1440x900&#46; Expected: exact snapshot state, paused playback, synchronized angles/info/target and deterministic standard framing&#46; Observed: tracking restored321&#46;7753574643905&#45;degree Venus and list angles157&#46;5/321&#46;8/255&#46;4; null restore cleared name/period/angle and disabled tracking&#46; Scene pixels in both repeated&#45;restore pairs were exactly identical after crop excluded sidebar focus, establishing zero pixel displacement within2px tolerance&#46; The restored live speed readout was correct although range thumb retained its former value; that feedback inconsistency is documented under U3&#46;

#### U1

Status: pass. Rating: 5. Method: http.

Action: Orbit by80x40px/25 moves and wheel&#45;zoom in overview/tracking, each paused and playing&#46; Expected: camera responds without altering simulation state beyond natural playback&#46; Observed: all four combinations changed rendered perspective and apparent sphere sizes&#46; Paused17&#46;0/0&#46;25x/Mars253&#46;550715 degrees stayed fixed; playing instances continued17&#46;2 to17&#46;3 with same speed/selection/mode&#46;

#### U2

Status: pass. Rating: 5. Method: http.

Action: Start drag on Mercury while Mars selected, use25 small moves totaling40 horizontal and10 vertical CSS pixels, then a fast115x80 drag; exercise list,ranges,buttons,snapshots and textarea&#46; Expected: drags not interpreted as selection and DOM operations not leak to camera&#46; Observed: slow and fast drags preserved Mars and paused0&#46;0/126&#46;050715 degrees while camera rotated&#46; Across DOM matrix only explicit selection/seek/track/restore effects occurred; text entry and export/import did not shift camera or select canvas&#46;

#### U4

Status: pass. Rating: 5. Method: http.

Action: Capture three, attempt fourth, delete middle, capture replacement, export&#46; Expected: capacity rejection with unchanged entries, ordered deletion and append without collision or live changes&#46; Observed: fourth left full list text identical and showed Capacity reached: delete a snapshot before capturing another&#46; Middle deletion retained first obs&#45;mvdhfa7m&#45;9cuii and third obs&#45;mvdhfh7m&#45;3eiwi in order; new obs&#45;mvdhfj8v&#45;lm5ml appended&#46; Live17&#46;0/2x/Mercury/overview/paused remained unchanged&#46;

#### U5

Status: pass. Rating: 5. Method: http.

Action: At both required sizes Tab into controls, use Enter on selection, pause/resume, tracking/overview, capture/restore/delete and export/import, Home/End/arrows on ranges, type JSON by keyboard, then Tab and Shift&#43;Tab out of import&#46; Expected: meaningful names, visible focus and all actions without trap&#46; Observed: native controls worked, range time0&#46;1 and speed4x changed, selected planets and snapshot operations updated, keyboard JSON empty/imported list accepted, focus left and returned&#46; Focus outline was gold solid2&#46;22222 CSS pixels&#46; Controls were accessible by names including Seek simulation time and Snapshot JSON import and export&#46;

#### V2

Status: pass. Rating: 3. Method: http.

Action: Inspect desktop/mobile scene and control panels including snapshots and JSON&#46; Expected: coherent legible typography,spacing,colors and hierarchy&#46; Observed: dark navy panels,aqua headings,gold numeric values and separated sections form a consistent hierarchy&#46; Mobile cards and textarea remain legible with vertical scroll&#46; Very long unrounded angle strings and dense snapshot text limit polish&#46; Qualitative anchor3, clear ordinary execution&#46;

#### L1

Status: pass. Rating: 5. Method: http.

Action: At actual1440x900 secure child document operate every primary control including keyboard snapshots and JSON&#46; Expected: usable controls without horizontal scrolling or blocking panels&#46; Observed: client/scroll widths1423/1423, all controls worked via sidebar vertical scroll, no panel blocked another control&#46; Canvas responded; full&#45;system default framing defect is owned byT4 rather than horizontal layout&#46;

#### L2

Status: pass. Rating: 5. Method: http.

Action: Repeat all primary control operations at actual390x844 secure child document, inspect scene and scroll between stacked panels&#46; Expected: same controls usable without horizontal scrolling or occlusion&#46; Observed: client/scroll widths373/373; seek/speed/pause/selection/tracking/overview/capture/restore/delete and typed JSON import/export all worked by keyboard&#46; Vertical scrolling reached every panel&#46; No horizontal hidden control or overlay blockage&#46; Default orbital clipping is recorded underT4/V1&#46;

#### D2

Status: pass. Rating: 5. Method: static.

Action: Inspect output/index&#46;html and output/three&#46;min&#46;js and local resource references&#46; Expected: every local dependency under output root with no build or repository&#45;root dependency&#46; Observed: the sole external script reference is three&#46;min&#46;js, present as bundled Three&#46;js r160; all styles and app logic are inline&#46; HTTP resource entry loaded the local script&#46; Direct&#45;file compatibility is excluded&#46;

#### D4

Status: pass. Rating: 5. Method: http.

Action: Export empty and populated lists, roundtrip through actual textarea/import, repeat while playing&#46; Expected: only required fields/types,stable IDs/order,explicit ID mapping and no live/list mutation&#46; Observed: empty export was &#123;version:1,snapshots:&#91;&#93;&#125;; populated root keys were exactly version/snapshots and each entry exactly id/time/speed/selectedPlanet/viewMode&#46; IDs A,a and a single space retained exact values/order; selectedPlanet used mercury/mars or null&#46; Extra imported keys and supplied angles were omitted&#46; Playing export preserved0&#46;25x/Mars/overview/PLAYING with natural17&#46;4 to17&#46;6 advance and unchanged list&#46; Name/ID mapping was visible&#46;

## Verification and limits

Browser method: http.
Viewports: 1600x1000, 1440x900, 390x844.
Coverage: 21 passed, 6 failed, 0 unverified.

### Notes

- User&#45;requested scoring&#45;only migration to solar&#45;system&#45;4&#46;0 at 2026&#45;10&#45;18T08:19:55&#46;883Z&#46; All check observations, ratings, statuses, methods, reviewer identity, and observation timestamp are unchanged&#46; Numeric scores are guaranteed lower endpoints with explicit upper endpoints, not newly observed passes or failures&#46; This recalculation is not a fresh browser review&#46;
- Fresh review of exactly the assigned saved observatory&#45;2 contract, not an earlier review&#46; The saved acceptance scenario and six additional sequences were exercised in order; detailed outcomes appear per check&#46;
- Reviewer model openai&#45;codex/gpt&#45;6&#45;astra comes from current host metadata&#46; Effort high comes from applicable &#46;omp/agents/benchmark&#45;astra&#45;review&#46;md, not independent runtime telemetry&#46; Timestamp is the host JavaScript clock&#46;
- Applied benchmark&#45;review and mandatory aside&#45;browser&#46; Selected poteto&#45;mode evidence principles Prove It Works and Explain the Number led to actual UI observations, preserved discrepant samples, and no score based on broken error subscriptions&#46; No other model or descendants were launched&#46;
- Preregistered tolerances: rate interval &#62;=3 seconds, max&#40;0&#46;15 simulation seconds,5% expected&#41;&#43;0&#46;1 displayed second; paused 0&#46;1 second and 1 degree; predicted angles 2 degrees plus rounding; repeat image center 2 CSS pixels; tracking 10% smaller scene dimension&#46; Exact identity/order/type/state comparisons&#46; No tolerance was widened&#46;
- HTTP only&#46; Output&#45;only server 127&#46;0&#46;0&#46;1:9213 and reviewer wrapper 127&#46;0&#46;0&#46;1:9223&#46; Viewport checks used real secure child documents at 1440x900 and 390x844, not top&#45;level viewport or touch&#45;device proof&#46; Their document client/scroll widths were 1423/1423 and 373/373&#46;
- Shared diagnostic scratch positive controls by Main and RerunLow returned empty console and pageerror arrays even for actual main&#45;world console&#46;error and thrown exceptions&#46; Empty subscriptions cannot prove D3&#46; Post&#45;load noninterfering error listeners, where used, miss startup and console&#46;error&#46;
- Initial visible concurrent rate observations: 0&#46;25x 0&#46;8 simulation seconds /3&#46;224 wall seconds; 2x 5&#46;8/3&#46;212; 4x 12&#46;0/3&#46;209&#46; Earlier acceptance 2x was 6&#46;5/3&#46;210&#46; Two later samples missed fixed tolerances&#46; A reserved exclusive window was requested to distinguish shared browser scheduling from application behavior&#46;
- An early top&#45;level repeated&#45;restore image comparison was invalid because scene width changed from 1212 to1195 with scrollbar/layout changes&#46; No numeric conclusion is based on that comparison&#46;
- Exclusive window coordinated by Main with all other browser lanes held: foreground visible1600x1000 samples were0&#46;25x:0&#46;8 simulation seconds/3&#46;510 wall seconds &#40;expected0&#46;8775,bound0&#46;25&#41;;2x:6&#46;5/3&#46;535&#40;expected7&#46;070,bound0&#46;4535&#41;;4x:11&#46;8/3&#46;527&#40;expected14&#46;108,bound0&#46;8054&#41;&#46; Both higher rates missed unchanged tolerances even with no competing browser tools&#46; Source line13 caps frame dt at0&#46;1 seconds, but no precise cause attribution is claimed without scheduling telemetry&#46;
- A transport defect in child locator&#46;screenshot captured parent top&#45;left bounds including sidebar&#46; Reported through xd://report&#95;issue&#46; Decode of its transient PNGs with standard&#45;library zlib and exact rendered&#45;scene crop x15&#46;&#46;946/y71&#46;&#46;781 showed zero differing rows for repeated tracking and null&#45;overview restores&#46; Sidebar focus, not scene, caused whole&#45;buffer inequality&#46; No tolerance was relaxed&#46;
- Later coordination established that holding browser tool calls does not stop continuously rendered duplicate scenes&#46; Thus the first exclusive interval still contained rendering contention and cannot by itself attribute the rate miss to this artifact&#46; Own desktop/mobile fixtures were then closed and only one assigned artifact page retained for final sampling&#46;
- Tracking evidence was quantitatively checked from transient rendered PNGs&#46; The three moving Mars bounding&#45;box centers were614&#46;444,523&#46;333 CSS pixels, offsets&#40;&#45;0&#46;226,&#45;0&#46;911&#41; from scene center; after73 seek center613&#46;333,523&#46;333, offsets&#40;&#45;1&#46;337,&#45;0&#46;911&#41;&#46; These are well within88&#46;977px tracking bound&#46; Screenshots are scaled0&#46;9 from1600x1000 and were converted back to CSS coordinates&#46;
- S1 period&#45;return patches were exact: Mercury image region640&#46;&#46;684/480&#46;&#46;529 at0 and12,Venus490&#46;&#46;549/545&#46;&#46;604 at0 and24,Mars250&#46;&#46;304/490&#46;&#46;539 at0 and48 had zero differing rows&#46; This measures rendered positions rather than source constants alone&#46;
- D3 observed real later app exceptions&#46; The observational listener was added after load, did not prevent events or modify application behavior, and disappeared on reload&#46; It does not cover startup errors or ordinary console&#46;error&#46; Failure is established by captured UI&#45;triggered exceptions, not by missing diagnostics&#46;
- Storage fixtures touched only observatory&#45;snapshots&#45;v1 and isolated filler rh&#45;review&#45;quota&#45;20261018 on this reviewer&#45;owned origin&#46; Malformed/schema&#45;invalid fixture restored exact saved original value; quota filler removed and exact prequota app payload restored&#46; Ordinary subsequent UI imports intentionally exercised the tool, not unrelated storage&#46; No storage&#45;read denial was induced&#46;
- Final clean timing window: Main confirmed every other review lane had closed its tabs/servers, and own tab listing confirmed only one assigned artifact tab&#46; No competing browser operations or CPU tests occurred&#46; With foreground document&#46;visibilityState visible and1600x1000 unchanged, browser performance&#46;now paired observations were0&#46;25x 0&#46;9/3&#46;5068,2x7&#46;0/3&#46;5095,4x14&#46;1/3&#46;5089&#46; All pass original tolerances&#46; Earlier multiple&#45;scene misses are therefore environment&#45;contaminated measurements, not established artifact speed failures&#46; No claim is made about the precise resource bottleneck&#46;
- Cleanup complete: all three review&#45;created tabs closed; exact own&#45;origin tab listing returned none&#46; Both tracked servers rh&#45;output and rh&#45;viewport stopped&#46; All24 reviewer&#45;generated transient screenshots were deleted after inspection&#46; Observation listeners were removed or destroyed by reload/tab close&#46; No evidence directory, screenshots or logs were retained in the run&#46; Working draft will be removed after final publication&#46; Source preservation was checked by before/after SHA&#45;256 in memory for prompt&#46;md,run&#46;json,index&#46;html and bundled three&#46;min&#46;js; no hashes are recorded here&#46;
