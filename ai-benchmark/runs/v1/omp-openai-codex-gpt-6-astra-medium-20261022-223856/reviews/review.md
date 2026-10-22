# 벤치마크 리뷰

## 결론

상태: 완료.
필수 요구사항 판정: 통과.
품질 총점: 94.30 / 100. 확정 점수. 범위: 94.30.
품질 점수에는 상한 제한을 적용하지 않습니다. 필수 요구사항 판정은 품질 점수와 독립적입니다.
필수 항목 실패 (0): 없음.
필수 항목 미검증 (0): 없음.
미검증 항목 (0): 없음.
누락된 리뷰어 메타데이터: 없음.
형식: 5&#46;0. 평가 기준: solar&#45;system&#45;5&#46;0.
리뷰 시각: 2026&#45;10&#45;22T15:09:02&#46;542Z.
리뷰어 모델: openai&#45;codex/gpt&#45;6&#45;astra.
리뷰어 추론 수준: high.

## 점수

| 평가 영역 | 점수 / 5 | 가중치 | 근거 |
| --- | --- | --- | --- |
| 시뮬레이션 (simulation) | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 선택 및 추적 (selection) | 5.00 | 20 | T1, T2, T3, T4, T5; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 사용성 (usability) | 4.60 | 15 | U1, U2, U3, U4, U5; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 시각적 완성도 (visual) | 3.50 | 15 | V1, V2; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 반응형 (responsive) | 5.00 | 15 | L1, L2; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 실행 및 제공 (delivery) | 5.00 | 10 | D2, D3, D4, D5, D6; 확정 점수; 전체 영역 항목을 분모로 사용 |

## 발견 사항

### 확인된 실패

없음.

### 검증하지 못한 항목

없음.

### 검증된 통과 항목

#### S1

상태: 통과. 평점: 5. 방법: HTTP.

Action: HTTP overview에서 startup motion 및 0/12/24/48초 seek 이미지를 관찰하고 17/30/73/120초의 위치·정보를 비교했다&#46; Expected: 태양 하나와 크기·색이 다른 행성 세 개가 고정 phase의 12/24/48초 원 궤도를 따라야 한다&#46; Observed: 노란 태양, 작은 주황 Ember, 중간 파랑 Azure, 큰 보라 Iris가 각각 세 궤도에 보였다&#46; Ember는 12초, Azure는 24초에 기준 위치로 돌아왔고 48초 장면 전체는 0초와 픽셀 단위로 같았다&#46; 표시 phase는 0/80/200°이고 중간 각도와 투영 위치도 수식에 맞았다&#46; 수동 orbit에서 원 궤도의 원근 투영이 변했다&#46;

#### S2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 보완용 기존 부모 페이지에서 새 HTTP load와 reload를 수행하고 startup 및 3&#46;3초 뒤 이미지를 읽었다&#46; Expected: 시간 0/1×/자동 재생/null/overview로 시작한 뒤 세 행성이 움직여야 한다&#46; Observed: 초기 관측 0&#46;15–0&#46;31초/1×/Pause/null/OVERVIEW, tracking disabled 및 명시적 no&#45;selection이었고 가시적 highlight가 없었다&#46; 0&#46;36→3&#46;65초 이미지에서 세 행성이 모두 각 궤도 위의 다른 위치로 이동했다&#46; 관측까지의 정상 로드 경과는 reset 실패로 보지 않았다&#46;

#### S3

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 Pause/seek 30 후 3&#46;3초 기다리고, 정지 중 선택·orbit/zoom 후 Resume했다&#46; Expected: 정지 중 시간과 모든 위치가 고정되고 재개가 현재 위치에서 연속되어야 한다&#46; Observed: 두 정지 캡처의 1019×480 장면 픽셀 차이는 0이었다&#46; 정지 상태의 목록/scene 선택과 카메라 입력도 작동했다&#46; Resume 직후 30&#46;05초/1×와 Iris 65&#46;37°였으며 기존 30초 위치 근처에서 자연스럽게 전진하고 이후 측정 속도로 이어졌다&#46;

#### S4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 정지 30초의 speed 1&#46;5×를 각각 0&#46;25×/2×/4×로 실제 변경하고 재개 및 재생 중 각 속도 변경을 측정했다&#46; Expected: 시간·위치가 도약하지 않고 상대 궤도 속도를 유지하며 지정 배속으로 진행해야 한다&#46; Observed: 각 정지 변경 전후 전체 장면 픽셀 차이 0과 시간 30을 확인했다&#46; 보완 독점 측정에서 3&#46;3024초/0&#46;83초&#40;0&#46;25× resume&#41;, 3&#46;3049초/6&#46;61초&#40;2× resume&#41;, 3&#46;3023초/13&#46;18초&#40;4× resume&#41;, 재생 중 변경은 3&#46;3025초/6&#46;61초&#40;2×&#41;, 3&#46;3025초/13&#46;23초&#40;4×&#41;, 3&#46;3021초/0&#46;83초&#40;0&#46;25×&#41;로 모두 사전 허용오차 내였다&#46; 세 행성의 실제 움직임과 고정 주기 관계도 유지됐다&#46; 초기 약 1초 frame cadence의 짧은 불일치 표본은 notes에 보존했다&#46;

#### S5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 slider로 0/30/120/17/73초를 seek하고 속도·재생·선택 변경 후 및 reload 후 17/30/73을 반복했다&#46; Expected: 즉시 paused time/전체 위치/정보/tracking이 갱신되고 동일 시간이 동일 위치를 내야 한다&#46; Observed: 모든 seek는 Resume 상태와 입력 시간을 보였으며 각도는 0→0/80/200, 30→180/170/65, 120→0/80/20, 17→150/335/327&#46;5, 73→30/95/27&#46;5°였다&#46; 17/30/73의 반복 및 reload 후 장면 차이는 각각 0픽셀이었다&#46; tracking seek 30/73도 즉시 선택이 화면 중심에 있었고 fractional 17&#46;25 restore는 157&#46;5/338&#46;75/329&#46;38°였다&#46;

#### S6

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 time 입력으로 119초 seek, 4× 설정 후 Resume하고 1&#46;8초 기다린 뒤 다시 1&#46;5초 기다렸다&#46; 이후 17초로 seek하고 Resume했다&#46; Expected: 120에서 정지하고 wrap/초과 없이 유지되며 이전 시간부터 다시 재생할 수 있어야 한다&#46; Observed: 두 표본 모두 120&#46;00 s, Resume, Iris 20&#46;00°였고 다시 시작한 표본은 20&#46;97 s, Pause였다&#46;

#### S7

상태: 통과. 평점: 5. 방법: HTTP.

Action: 로드된 local Three&#46;js r160와 index&#46;html:14–20의 WebGLRenderer/PerspectiveCamera/SphereGeometry를 대응시키고 overview/tracking에서 수동 orbit·zoom 이미지를 읽었다&#46; Expected: 실제 Three&#46;js 3D geometry의 원근·깊이·가림이 카메라 변화에 반응해야 한다&#46; Observed: 구체의 입체 음영, 전후 행성 크기 및 궤도 타원의 원근 변화, tracking에서 확대된 행성·태양의 가림과 시점 변화가 보였다&#46; 장면은 실제 WebGL canvas와 로드된 Three&#46;js로 그려지며 단순 import 선언만으로 판정하지 않았다&#46;

#### S8

상태: 통과. 평점: 5. 방법: HTTP.

Action: 정지 및 재생 중 capture 후 모든 live field를 바꾸고, captured&#45;ID를 export→유효 import→reload로 순환시켰으며 fractional time/extra angles import를 관찰했다&#46; Expected: 순간 값과 세 derived angle 및 고유 ID가 불변이고 capture가 live state를 바꾸지 않아야 한다&#46; Observed: A obs&#45;mvjnvodi&#45;zizl1qzl의 30초/2×/iris/tracking 및 180/170/65°가 live 변경 후에도 유지됐다&#46; 삭제하지 않은 captured IDs와 순서는 export/import/reload 후 정확히 동일했다&#46; 재생 capture obs&#45;mvjo6sym&#45;z0h58iq8는 30&#46;025175000001123초/0&#46;25×/iris/tracking 및 180&#46;76/170&#46;38/65&#46;19°를 기록하고 playback이 계속됐다&#46; 정지 capture 및 playing capture 전후 같은 시간으로 비교한 비표준 카메라 장면은 0픽셀 차이였다&#46; imported 17&#46;25초의 157&#46;5/338&#46;75/329&#46;38°는 제공된 angles 999를 무시했다&#46;

#### T1

상태: 통과. 평점: 5. 방법: HTTP.

Action: 세 행성을 정지 및 재생 중 실제 canvas 클릭과 목록 버튼/키보드로 각각 선택했다&#46; Expected: 두 경로가 동일 선택·가시적 highlight·정보를 공유해야 한다&#46; Observed: 보완 canvas 클릭은 Ember/Azure/Iris 모두 두 재생 상태에서 해당 planet ID와 정보를 갱신했다&#46; 실제 이미지에서 선택된 구체의 밝기 강조와 해당 목록 버튼의 청록 강조가 함께 바뀌었고 다른 구체의 강조는 사라졌다&#46; 초기 촬영 불가 상태의 click miss를 결함으로 전용하지 않았다&#46;

#### T2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 각 행성 선택의 playback/seek 정보와 실제 위치를 비교하고 초기 및 null&#45;overview restore를 관찰했다&#46; Expected: 이름·주기·각도가 위치와 일치하고 무선택 상태가 명확해야 한다&#46; Observed: Ember/Azure/Iris는 12/24/48초 및 계산과 일치하는 각도·렌더 위치를 보였다&#46; 예를 들어 30초에서 180/170/65°, 73초에서 30/95/27&#46;5°였다&#46; no&#45;selection restore 뒤 목록·구체 highlight가 모두 없고 &#34;No planet selected&#46; Choose a planet to inspect its orbit&#46;&#34;로 이전 정보가 지워졌다&#46;

#### T3

상태: 통과. 평점: 5. 방법: HTTP.

Action: Iris를 3&#46;3초 이상 tracking하고 pause/seek 30/73, Azure/Ember로 전환했으며 no&#45;selection에서 tracking 상태를 확인했다&#46; Expected: 선택이 실제 장면 중심에 유지되고 target/info/highlight가 함께 바뀌되 time/speed/playback은 보존되어야 한다&#46; Observed: desktop 장면 중심 &#40;532&#46;5,352&#41;에 대해 Iris 렌더 중심 &#40;532,351&#46;5&#41;, Azure &#40;531&#46;5,351&#46;5&#41;였고 mobile seek 후 &#40;187,250&#46;5&#41;, 장면 중심 &#40;187&#46;5,251&#41;로 모두 10% 한도 이내였다&#46; 시간 진행과 seek에서도 중심이 유지되고 대상/정보/강조가 동기화됐다&#46; no&#45;selection에서는 Track selected가 비활성이라 상태를 바꾸지 않았다&#46;

#### T4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 재생 및 정지 tracking에서 Full overview를 눌러 두 viewport의 결과와 전후 live state를 비교했다&#46; Expected: 선택·시간·속도·재생을 보존하며 전체 시스템을 보여야 한다&#46; Observed: Azure 재생 2×에서 overview로 돌아와 자동 시간이 이어졌고 Azure 선택이 유지됐다&#46; 정지 Iris 73초/1&#46;5×도 그대로였다&#46; 두 viewport 이미지에서 태양과 세 행성 및 외곽 궤도 모두 장면 안에 보였고 선택 강조도 유지됐다&#46;

#### T5

상태: 통과. 평점: 5. 방법: HTTP.

Action: unrelated live edits 및 manual orbit/zoom 후 selected tracking A와 null&#45;selection overview를 각각 반복 restore했다&#46; Expected: 저장값·paused·위치·정보·target 및 deterministic standard framing이 함께 복원되어야 한다&#46; Observed: A는 30초/2×/Iris/tracking, null snapshot은 30초/1&#46;5×/null/overview로 복원됐다&#46; null은 이전 정보와 강조를 지웠고 tracking 선택은 중앙이었다&#46; manual view가 실제 바뀐 뒤 반복 restore된 1019×480 장면은 각 첫 표준 구도와 0픽셀 차이로 2px 허용오차 내였다&#46;

#### U1

상태: 통과. 평점: 5. 방법: HTTP.

Action: overview/tracking 각각 playing/paused에서 실제 drag와 wheel을 실행하고 전후 렌더를 읽었다&#46; Expected: 네 조합 모두 현재 target 주위 orbit/zoom이 동작하며 simulation state를 임의로 바꾸지 않아야 한다&#46; Observed: 궤도 투영과 구체 배치/크기가 실제 변했고 tracking target은 중앙에 남았다&#46; 정지에서는 시간·각도가 고정되고 재생 중에는 지정 속도의 자연 진행이 이어졌다&#46;

#### U2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 30초 overview의 Azure 중심 &#40;417,320&#41;에서 빠른 100×40px drag 및 25단계 누적 40×10px slow drag를 실시하고 모든 DOM control 종류를 조작했다&#46; Expected: drag가 클릭 선택이 되지 않고 DOM 입력이 scene/camera에 누출되지 않아야 한다&#46; Observed: Azure에서 시작한 두 drag 모두 이전 Ember 선택을 유지하면서 카메라만 변했다&#46; list/slider/button/snapshot/JSON 조작은 의도한 상태만 바꿨고, 비표준 카메라에서 capture/capacity/export/invalid import/valid import/delete의 장면은 모두 baseline과 0픽셀 차이였다&#46; keyboard에서도 원치 않는 선택/카메라 이동은 없었다&#46;

#### U3

상태: 통과. 평점: 3. 방법: HTTP.

Action: 양쪽 viewport에서 재생·선택·tracking·snapshot·JSON 입력과 용량/오류/복구/저장실패 상태의 실제 DOM 및 접근성 snapshot을 읽고 조작했다&#46; Expected: 목적과 현재 상태, 작업 결과를 이해할 수 있어야 한다&#46; Observed: &#34;Pause&#34;/&#34;Resume&#34;, 시간·속도, &#34;OVERVIEW&#34;/&#34;TRACKING&#34;, 명명된 행성/ID·주기·각도, 전체 snapshot 값, 명시적 Import error/Capacity reached/Recovery/Saving unavailable 문구가 일관되게 구분됐다&#46; 의미 전달은 명료한 보통 수준의 anchor 3으로 평가하며, 보완 실제 이미지에서도 feedback은 읽기 쉬웠다&#46; 다만 capture의 긴 부동소수 표기와 소박한 상태 표현을 고려해 보통 수준 3을 유지했다&#46;

#### U4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 30초 Ember tracking, 17초 Azure overview, 73초 Iris overview를 차례로 capture한 뒤 네 번째 capture, 가운데 Delete 2, 새 capture와 export를 실행했다&#46; Expected: 세 개 한도에서 기존 값과 ID가 유지되고 삭제 후 남은 순서 및 값이 유지되며 새 ID가 충돌 없이 마지막에 추가되어야 한다&#46; Observed: &#34;Capacity reached: delete an observation before capturing another&#46;&#34;가 표시되고 3개 payload와 73초/2×/Iris/overview/paused가 그대로였다&#46; Delete 2 후 첫째와 셋째 ID·값이 순서대로 남았고 새 obs&#45;mvjmp0q1&#45;psrvods0가 마지막에 추가되었다&#46; 추가 키보드 삭제와 quota 중 삭제에서도 시간·속도·선택·모드·재생 상태가 그대로였다&#46; 보완에서 비표준 카메라의 capture·네 번째 capacity rejection·가운데 delete 전후 렌더도 각각 0픽셀 차이로 보존됨을 확인했다&#46;

#### U5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 1440x900 및 390x844에서 Tab/Shift&#43;Tab, Enter, range Home/ArrowRight, 숫자·textarea 키 입력으로 모든 primary DOM control을 작동시켰다&#46; Expected: 의미 있는 accessible name, 실제 보이는 focus와 정상 동작, focus 탈출을 제공해야 한다&#46; Observed: pause/resume, 세 행성, time/speed, tracking/overview, capture/restore/delete, JSON edit/import/export가 모두 작동했다&#46; label/버튼 이름이 명확했고 각 focus의 3px outline 계산값과 :focus&#45;visible, 양쪽 실제 이미지의 금색 button/textarea focus가 확인됐다&#46; Shift&#43;Tab으로 textarea에 돌아왔고 trap이나 가림이 없었다&#46;

#### V1

상태: 통과. 평점: 3. 방법: HTTP.

Action: desktop/mobile의 overview/tracking, 선택 변경, seek 및 camera orbit/zoom 실제 이미지를 비교했다&#46; Expected: scene와 selection이 읽기 쉽고 효과가 관찰을 방해하지 않아야 한다&#46; Observed: 어두운 배경에 세 색과 궤도, 구체 음영, 선택 밝기 차이가 구분되고 tracking 확대가 관찰을 돕는다&#46; 다만 overview의 Ember는 작고 강조는 밝기 변화 정도이며 tracking에서 다른 천체가 상단 caption 뒤에 걸리는 경우가 있다&#46; 선택 자체와 controls는 명료하여 보통의 usable execution인 anchor 3이다&#46;

#### V2

상태: 통과. 평점: 4. 방법: HTTP.

Action: 두 viewport에서 상단 장면, time controls, selection/camera 및 아래 snapshot/JSON panels를 실제 이미지로 읽었다&#46; Expected: typography/spacing/colors/panel hierarchy가 일관되고 읽기 쉬워야 한다&#46; Observed: navy panel, 청록 상태색, 금색 focus/피드백, 고정폭 수치와 구분된 section 계층이 일관된다&#46; 모바일도 같은 계층으로 자연스럽게 한 열로 정리된다&#46; 보조 설명·snapshot 값 글씨가 작고 길게 표시되는 부동소수 capture 시간이 다소 거칠지만 전반적으로 정돈된 anchor 4이다&#46;

#### L1

상태: 통과. 평점: 5. 방법: HTTP.

Action: 매 desktop 호출에 1440x900 CSS metrics를 재설정·확인하고 장면 및 모든 primary controls를 마우스/키보드로 조작하며 상하 panel 이미지를 읽었다&#46; Expected: 수평 스크롤이나 가림 없이 scene와 controls가 usable이어야 한다&#46; Observed: innerWidth/innerHeight=1440/900, scrollWidth=1425였다&#46; scene 1019×480, 오른쪽 selection/camera/capture, 아래 snapshot/JSON panels가 겹치지 않았다&#46; 수직 scroll만으로 restore/delete/textarea/export/import까지 모두 작동했다&#46;

#### L2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 390x844 CSS metrics에서 선택·tracking·seek·speed·pause/resume·capture/restore/delete·JSON을 실제 작동시키고 overview/tracking 및 아래 panels를 읽었다&#46; Expected: 모든 controls와 장면이 수평 숨김/가림 없이 usable이어야 한다&#46; Observed: innerWidth/innerHeight=390/844, scrollWidth=375, scene=349×330이었다&#46; 한 열 배치가 화면 폭 안에 있고 수직 scroll로 모든 control에 도달했다&#46; overview는 전체 궤도를 표시하고 tracking 선택은 중앙에 보였다&#46; focus·snapshot값·textarea·import/export가 서로 가리지 않았다&#46;

#### D2

상태: 통과. 평점: 5. 방법: 정적 검사.

Action: output/index&#46;html의 모든 의존 경로와 output/ 파일 목록을 조사했다&#46; Expected: output/를 HTTP 루트로 할 때 로컬 리소스가 모두 존재하고 빌드나 저장소 루트 의존성이 없어야 한다&#46; Observed: 유일한 외부 스크립트 경로 three&#46;min&#46;js가 output/에 있으며 실제 Three&#46;js 배포본이다&#46; 나머지는 HTML 내부 CSS/JS 및 절차적 geometry로 구성되어 추가 로컬 리소스·모듈·fetch 경로가 없다&#46; file:// 호환성은 평가 범위에서 제외했다&#46;

#### D3

상태: 통과. 평점: 5. 방법: HTTP.

Action: scratch에서 Runtime&#46;consoleAPICalled와 Runtime&#46;exceptionThrown 양성 대조를 확인한 뒤 Runtime/Network를 활성화하여 HTTP 탐색 및 UI 검사에서 이벤트를 수집했다&#46; Expected: 산출물 오류나 필수 리소스 실패가 도구 동작을 막지 않아야 한다&#46; Observed: 필수 three&#46;min&#46;js가 HTTP에서 로드됐고 초기 이벤트는 구형 classic Three&#46;js 배포 경고뿐이었다&#46; quota·반복 UI 검사에서 산출물 console error/exception/필수 로드 실패는 관찰되지 않았다&#46; Page&#46;captureScreenshot timeout은 브라우저 전송 기능의 실패로 분리했다&#46; 호출 사이의 무관측 구간까지 무오류를 보증하지 않는다&#46;

#### D4

상태: 통과. 평점: 5. 방법: HTTP.

Action: empty/populated export, captured&#45;ID 및 valid boundary payload의 UI import/export roundtrip을 실행하고 정지·재생 중 list/live state와 카메라를 비교했다&#46; Expected: labeled textarea와 명시 controls만으로 정확한 versioned fields/type/order/ID를 전송하고 상태를 보존해야 한다&#46; Observed: root는 version:1/snapshots, entry는 id/time/speed/selectedPlanet/viewMode만 있었고 angles/playback/camera는 없었다&#46; 행성 이름&#45;ID 매핑이 JSON 옆에 표시됐다&#46; Case/case/공백 ID와 captured IDs 및 순서가 유지됐고 empty export도 정확했다&#46; 정지 export/import와 playing transfer 후 같은 시간의 비표준 camera 장면은 0픽셀 차이였으며 자연 시간 진행 외 playback/selection/speed/mode 변화가 없었다&#46;

#### D5

상태: 통과. 평점: 5. 방법: HTTP.

Action: nonempty baseline에서 38종 invalid payload를 각각 actual textarea/Import로 제출하고, valid empty/3&#45;entry/endpoint/fractional/case&#45;space&#45;ID/extra&#45;key payload와 playing import를 검사했다&#46; Expected: 전체 validation/atomicity, 정확한 ID·순서·타입 유지 및 목록만 교체가 성립해야 한다&#46; Observed: malformed, root/entry null·array, 모든 required field 누락, 잘못된 version/type/범위/overflow/planet/mode/ID/개수, valid&#45;prefix&#45;invalid&#45;suffix 모두 오류를 표시하고 기존 목록·live DOM을 정확히 유지했다&#46; 38종 각각의 비표준 camera 장면도 baseline과 0픽셀 차이였다&#46; 0/120초, 0&#46;25/4×, 17&#46;25초/1&#46;5×, Case/case/공백 ID, null/selected overview/tracking, extra keys/angles 무시, playing 및 empty list replacement가 요구대로 동작했다&#46;

#### D6

상태: 통과. 평점: 5. 방법: HTTP.

Action: 동일 origin/key에서 import/delete/capture 후 각각 reload, 38개 invalid persisted payload, 8MiB 상한 quota fixture 후 실제 capture와 large valid import를 실행했다&#46; Expected: versioned 목록·ID·순서·값·derived angles가 지속되고 live defaults로 시작하며, invalid saved data는 빈 목록과 recovery notice, 실제 write 실패는 saving&#45;unavailable notice와 계속 사용 가능한 메모리 동작을 제공해야 한다&#46; Observed: import/delete/capture 각각 reload 후 값이 유지되고 0&#46;05–0&#46;07초/1×/null/overview/playing으로 시작했다&#46; 38종 invalid data 모두 empty&#43;Recovery였고 이후 controls가 동작했다&#46; filler 5,242,523문자에서 6번 QuotaExceededError를 관찰했으며 실제 capture 뒤 저장 raw는 그대로인데 메모리만 3개로 늘어나 write failure가 입증됐다&#46; Saving unavailable 문구가 유지된 채 selection/playback/restore/delete 및 큰 valid import가 동작했다&#46; fixture raw value와 filler를 복원했고 마지막 app/filler는 모두 null이다&#46; storage read denial 자체는 시험하지 않았으며 이 증거는 실제 failed write 분기 검증이다&#46;

## 검증 및 한계

브라우저 검증 방식: HTTP.
뷰포트: 1440x900, 390x844.
검증 범위: 통과 27개, 실패 0개, 미검증 0개.

### 참고 사항

- 리뷰어 모델은 현재 주입된 workstation 모델 메타데이터 openai&#45;codex/gpt&#45;6&#45;astra로 확인했다&#46; effort high는 부모가 현재 리뷰 delegate에 적용했다고 확인한 &#46;omp/agents/benchmark&#45;astra&#45;review&#46;md의 thinking&#45;level: high 설정을 직접 읽어 확인했다&#46; 별도 runtime effort scalar는 노출되지 않았고 일반 기본값이 아니다&#46; 생성 run&#46;json이나 폴더 이름에서 추론하지 않았다&#46; 시각은 현재 호스트 Date 시계의 UTC ISO 8601 값이다&#46;
- 사용자가 선택한 benchmark&#45;review와 필수 ego&#45;browser를 적용했다&#46; using&#45;harness&#45;flow를 읽었으나 subagent 면제에 따라 추가 워크플로를 적용하지 않았다&#46; 부모의 필수 harness&#45;flow/caveman 적용은 orchestration 예외로 전달받았으며 본 리뷰에 caveman을 별도로 적용하지 않았다&#46;
- 관찰 전 허용오차: 3초 이상의 실제 시간 측정에서 속도 오차는 max&#40;0&#46;15초, 예상 변화량의 5%&#41;&#43;표시 0&#46;01초; 정지 시간 0&#46;01초와 각도 1°; 예상 각도 2°&#43;표시 반올림 0&#46;01°; 동일 뷰포트·표준 카메라의 반복 seek/restore 중심 2 CSS px; tracking은 3초 이상 표본 및 seek 직후 장면 짧은 변의 10% 이내; JSON 값·순서·타입과 비시간 상태는 정확히 일치해야 한다&#46;
- 번들 probe&#45;terra&#45;xhigh&#46;json과 ID/저장키/조작 구조가 일치하지 않으므로 적용할 수 없다&#46; 실제 DOM &#35;time, &#35;speed, &#35;play, &#35;planet&#45;ember/&#35;planet&#45;azure/&#35;planet&#45;iris, &#35;track, &#35;overview, &#35;capture, &#35;snapshots, &#35;json, &#35;export, &#35;import를 매핑하여 동일 요구 시나리오를 검사한다&#46;
- 검증은 부모가 직렬 배정한 독점 슬롯에서 변경하지 않은 output/ 전용 http://127&#46;0&#46;0&#46;1:19321/index&#46;html로 수행했다&#46; 첫 단계는 자체 TaskSpace 3/p1, 보완 단계는 부모가 명시적으로 배정한 기존 TaskSpace 1/p2였다&#46; 종료된 공간을 재개하거나 새 공간을 만들어 장애를 우회하지 않았다&#46; 각 페이지에서 별도 19322 scratch의 console&#46;error/throw 양성 대조를 모두 포착했다&#46; file://, 다른 브라우저, 모델 실행, 소스/런타임 상태 patch는 사용하지 않았다&#46;
- 초기 TaskSpace 3에서 desktop 기본 및 clip/raw, mobile screenshot이 CdpRequestTimeoutError: CDP request timed out: Page&#46;captureScreenshot으로 실패했다&#46; DOM/UI 독립 검사를 보존한 partial을 먼저 게시했으며 이를 완료로 보지 않았다&#46; 부모가 지정한 기존 공유 p2의 보완 검증에서는 동일 Ego documented screenshot이 정상 동작하여 실제 두 viewport 렌더와 누락 조항을 새로 관찰했다&#46; 이전 장애 원인은 미확정이고 보완 성공은 실제 새 증거이며, 다른 산출물의 이미지를 사용하지 않았다&#46;
- 저장 acceptance 1–8을 초기에는 부분 실행하고, 보완에서는 fresh load→canvas Iris 선택→tracking→pause/30 seek→paused 2×→resume 속도 측정→목록 Azure 전환→overview→pause/30 재방문을 순서대로 다시 실행했다&#46; 실제 target/info/highlight, 전체 시스템, 보존 상태가 일치했고 30초 동일 tracking 구도의 장면 픽셀 차이는 0이었다&#46;
- 추가 시나리오 1–6은 초기 같은 시도에서 순서대로 실행했다&#46; A 불변·복원, 세 개 한도/삭제/재캡처, 38개 JSON 오류, 유효 fractional/null/tracking·playing·empty import, import/delete/capture reload 및 38개 invalid saved payload, 양쪽 전체 keyboard 동작을 관찰했다&#46; 보완에서 실제 시각 restore/camera/selection/focus를 추가하고 38개 JSON 오류 각각의 camera 불변, captured&#45;ID roundtrip, paused/playing capture 및 JSON 전송 불변까지 확인했다&#46;
- 속도 측정은 target scene 하나와 document&#46;visibilityState visible에서 한 호출 안에 before/wait/after로 수행했다&#46; 초기 1×: wall 3&#46;2036/3&#46;2044초, delta 각각 3&#46;00초; 초기 2×: wall 3&#46;2017초, delta 6&#46;00초였다&#46; 후속 6&#46;2초 측정은 기존 허용오차를 유지했다&#46; 별도 read&#45;only requestAnimationFrame 표본 287309&#46;1, 288309&#46;2, 289322&#46;7, 290322&#46;8, 291322&#46;9ms로 약 1초 프레임 cadence를 관찰했다&#46; 이 제한의 원인을 산출물 또는 호스트로 임의 단정하지 않으며 짧은 표본을 삭제하지 않는다&#46; 보완 p2에서는 같은 사전 허용오차를 유지한 3&#46;3초 측정 모두 지정 배속 범위에 들었고 실제 렌더 motion도 확인됐다&#40;S4 수치 참조&#41;&#46; 이는 초기 표본을 숨기거나 허용오차를 늘려 통과시킨 것이 아니라, 정상이 관찰되는 문서에서 필요한 시나리오를 새로 실행한 결과다&#46;
- D6 fixture는 app key orbit&#45;medium&#45;observatory&#45;v1와 고유 filler review&#45;medium&#45;quota&#45;20261022만 변경했고 unrelated storage를 읽거나 지우지 않았다&#46; 각 fixture가 저장 전 raw payload를 정확히 복원했으며 마지막에는 리뷰 시작 때의 app&#45;key absence&#40;null&#41;로 되돌렸다&#46; 마지막 app=null/filler=null을 확인했다&#46; Storage prototype monkeypatch, permission 변경, quota 증가를 하지 않았다&#46;
- 정리: 초기 자체 TaskSpace 3는 override 해제 후 finish&#40;&#123;keep:&#91;&#93;&#125;&#41;를 정확히 한 번 성공했고 소유 서버를 중지했다&#46; 보완 후 app/filler key가 원래 absence&#40;null&#41;임을 재확인하고 override를 해제한 뒤 부모 소유 TaskSpace 1/p2를 about:blank로 반환했다&#46; 부모 요청에 따라 공유 page/space를 finish/close하지 않았으며 전체 작업 후 부모가 관리한다&#46; 보완 19321/19322 서버도 중지했다&#46; 리뷰 소유 임시 scripts와 모든 PNG를 삭제했으며 초안도 최종 게시 후 제거한다&#46; 생성 기록의 검증 메모는 새 행동 증거로 사용하지 않았다&#46;
- prompt&#46;md, run&#46;json, output/index&#46;html, output/three&#46;min&#46;js를 이 시도 시작 시 읽은 byte digest와 비교해 모두 동일함을 작업 메모리에서 확인했다&#46; digest 값은 리뷰에 저장하지 않는다&#46;
- 부모가 screenshot 장애 조사를 위해 상위 필수 systematic&#45;debugging을 추가 적용했다고 전달했다&#46; 본 delegate는 benchmark&#45;review/ego&#45;browser를 적용했고 부모의 추가 적용은 orchestration 예외로 기록한다&#46; 부모의 다른 산출물 캡처 성공은 이 산출물 행동 증거가 아니며, 본 리뷰의 보완은 동일 Medium URL에서 직접 새로 수행했다&#46;
- 초기 all&#45;unverified checkpoint와 도구 제한이 남은 partial을 안전하게 게시한 뒤, 같은 시도의 보완 관찰을 통합하여 최종 27개 check를 판정했다&#46; 필수 항목의 미확인은 남지 않았지만 file:// compatibility, 물리 touch&#45;device 동작, 별도 storage read&#45;denial은 범위 밖 또는 실제 failed&#45;write 분기와 구별되는 미실행 조건으로 명시한다&#46; 저장소 작성 실패·복구는 실제 quota fixture로 검증했다&#46;
- D6는 unavailable storage 또는 failed write를 요구한다&#46; 실제 quota write 실패·표시·메모리 사용·persistence·invalid recovery가 모두 관찰되었으므로 별도 read&#45;denial 부재를 추가 필수 조건으로 삼지 않도록 중간 판정을 정정했다&#46; 이는 새 관찰이 아니라 같은 시도 증거에 대한 요구조건 해석 정정이다&#46;
- 픽셀 비교는 실제 Ego screenshot의 PNG를 외부 작업 메모리에서 decode하여 실제 canvas rectangle &#40;23,112&#41;–&#40;1042,592&#41;의 RGB를 비교했다&#46; 앱의 숨겨진 상태나 Three&#46;js objects를 노출하지 않았다&#46; repeat seek/restore, pause, speed change, capture/delete, 정지·재생 JSON 전송 및 38개 invalid import 모두 해당 비교에서 변화 0픽셀을 관찰했다&#46; tracking 중심은 이미지의 최대 연결된 행성색 영역으로 측정했으며 desktop 오프셋 약 0&#46;7–1&#46;1px, mobile 약 0&#46;7px였다&#46;
- 미설치 Pillow import가 ModuleNotFoundError로 실패하여 의존성을 설치하지 않았다&#46; Node 내장 zlib로 실제 PNG를 읽어 비교하는 임시 검토 코드만 사용했고 파일을 남기지 않았다&#46; build/lint/test/formatter나 산출물 수리는 실행하지 않았다&#46;
