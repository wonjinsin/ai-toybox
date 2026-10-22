# 벤치마크 리뷰

## 결론

상태: 완료.
필수 요구사항 판정: 통과.
품질 총점: 95.80 / 100. 확정 점수. 범위: 95.80.
품질 점수에는 상한 제한을 적용하지 않습니다. 필수 요구사항 판정은 품질 점수와 독립적입니다.
필수 항목 실패 (0): 없음.
필수 항목 미검증 (0): 없음.
미검증 항목 (0): 없음.
누락된 리뷰어 메타데이터: 없음.
형식: 5&#46;0. 평가 기준: solar&#45;system&#45;5&#46;0.
리뷰 시각: 2026&#45;10&#45;22T14:56:32&#46;143878&#43;00:00.
리뷰어 모델: openai&#45;codex/gpt&#45;6&#45;astra.
리뷰어 추론 수준: high.

## 점수

| 평가 영역 | 점수 / 5 | 가중치 | 근거 |
| --- | --- | --- | --- |
| 시뮬레이션 (simulation) | 5.00 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 선택 및 추적 (selection) | 5.00 | 20 | T1, T2, T3, T4, T5; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 사용성 (usability) | 4.60 | 15 | U1, U2, U3, U4, U5; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 시각적 완성도 (visual) | 4.00 | 15 | V1, V2; 확정 점수; 전체 영역 항목을 분모로 사용 |
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

Action: 초기 자동 재생과0/3/6/12/24/48/30/17/73/120초 실제 seek 화면 및 세 정보 각도를 관찰했다&#46; Expected: 태양1개·서로 다른 크기/색의 행성3개, 원궤도와 고정 위상 및12/24/48초 주기가 보여야 한다&#46; Observed: 주황/청록/보라 구체와 태양을 확인했다&#46; Ember 중심은0/12초 모두&#40;615&#46;5,393&#46;5&#41;, Azure는0/24초 모두&#40;445,466&#41;, Iris는0/48초 모두&#40;464&#46;5,289&#46;5&#41;였다&#46; 표준 카메라 투영을 역산한 실제 이미지 중심 각도는 위상0/100/225도 및 주기12/24/48초 식과 최대0&#46;74도 이내로 일치했다&#46; 중간 시각의 궤도상 위치와 자동 운동도 이미지로 관찰했다&#46;

#### S2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 새 HTTP 탐색과 reload 직후 DOM을 읽고 초기 및3초 이후 장면 이미지를 비교했다&#46; Expected: 0초·1×·재생·무선택·overview로 시작하고 자동으로 움직여야 한다&#46; Observed: 최초 준비 직후0&#46;17초/1×/Pause/무선택/overview, 초기 이미지0&#46;25초와 이후3&#46;6초대 이미지에서 세 행성이 실제 이동했다&#46; 초기 highlight가 없고 명시적인 무선택 정보와 disabled Track가 보였다&#46; reload도0&#46;14&#126;0&#46;27초의 로드 경과만 있는 같은 기본값이었다&#46;

#### S3

상태: 통과. 평점: 5. 방법: HTTP.

Action: 30초에서 정지한 장면을3&#46;3초 이상 떨어져 촬영하고 짧게 resume한 뒤 다시 정지했다&#46; 정지 중 선택·orbit·zoom도 조작했다&#46; Expected: 시간과 모든 궤도 위치가 동결되고 resume은 위치 점프 없이 이어져야 한다&#46; Observed: 30초 정지 전후 scene 픽셀이 완전히 같았다&#46; 재생 후30&#46;10초 장면의 세 위치는 그 시간의 예측 각도와 최대0&#46;37도 차이였고 불연속 점프가 없었다&#46; 정지 상태의 선택 및 두 모드 카메라 조작도 실제 highlight·원근/확대 변화를 보였다&#46;

#### S4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 정지17초에서0&#46;25×/2×/4×를 설정하고 각각3&#46;3초 이상 재생했다&#46; 재생 중에도2×/4×/0&#46;25×로 변경했다&#46; Expected: 현재 시간·위치를 보존하며 모든 궤도의 상대 속도와 설정 속도가 유지되어야 한다&#46; Observed: 정지 세 속도의 scene 픽셀이 baseline과 완전히 같았다&#46; 실제/시뮬레이션 증분은0&#46;25×:3&#46;3039/0&#46;83초,2×:3&#46;3088/6&#46;62초,4×:3&#46;3217/13&#46;28초로 사전 오차 안이었다&#46; 재생 중 변경에서 시간이 reset되지 않았고 old/new 속도 사이의 자연 진행만 있었다&#46; 각 rate 종료와 변경 뒤 세 렌더 위치를 읽은 시간과 비교한 각도 오차는 최대0&#46;66도였다&#46;

#### S5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 time range의 클릭·Home/End·방향키로0/30/120/17/73초를 seek하고 속도/선택/재생 변경과 reload 후 반복했다&#46; Expected: seek가 정지 및 즉시 위치/정보 갱신을 수행하고 같은 시각은 같은 위치를 재현해야 한다&#46; Observed: 모든 seek에서 Resume 상태였다&#46;30초 각도180/190/90도,17초150/355/352&#46;5도,73초30/115/52&#46;5도가 실제 위치와 맞았다&#46;30/17/73초의 변경 전후 및0/30/17/73초 reload 전후 scene 픽셀이 완전히 같았다&#46;17&#46;25초 imported restore도157&#46;50/358&#46;75/354&#46;38도와 해당 렌더 위치를 재현했다&#46;

#### S6

상태: 통과. 평점: 5. 방법: HTTP.

Action: 119초·4×에서 재생하고3&#46;3초 뒤 및 추가3&#46;1초 뒤 관찰한 다음17초로 seek하여 재생했다&#46; Expected: 120초에서 wrap/초과 없이 정지하고 이전 시각에서 재생 가능해야 한다&#46; Observed: 두 endpoint 관찰 모두 정확히120초/Resume이었고 이미지도120초 위치였다&#46; 이전 seek 후 재생하여18&#46;79초 및 새로운 궤도 위치로 진행했다&#46;

#### S7

상태: 통과. 평점: 5. 방법: HTTP.

Action: 로컬 Three&#46;js 자원과 WebGLRenderer/SphereGeometry/PerspectiveCamera 구현을 검사하고 실제 HTTP scene의 orbit·zoom을 네 playback/mode 조합에서 촬영했다&#46; Expected: 진짜3D 기하의 원근·가림·깊이 변화가 보여야 한다&#46; Observed: three&#46;min&#46;js가669884 decoded bytes로 로드됐다&#46; 실제 drag에서 궤도 타원의 기울기와 구체 간 상대 배치·가림이 바뀌었고 wheel로 투영 크기가 증가했다&#46; tracking에서는 선택 구체 뒤의 궤도와 다른 구체가 상대적으로 이동했다&#46; flat label/import만을 근거로 하지 않았다&#46;

#### S8

상태: 통과. 평점: 5. 방법: HTTP.

Action: 정지·재생 중 capture하고 모든 live 필드를 변경한 뒤 저장값을 비교했다&#46; 캡처 ID를 export→UI import→reload→restore하고 fractional imported 항목도 관찰했다&#46; Expected: 다섯 필드와 세 derived angles를 표시하고 고유 ID·값이 유지되며 capture가 live state를 바꾸지 않아야 한다&#46; Observed: 캡처 obs&#45;mvjncsji&#45;p638sx의30초/2×/ember/tracking 및180/190/90도가 live 편집과 전체 round&#45;trip 후 같았다&#46; 재생 캡처31&#46;707149999997824초/1&#46;5×/azure/overview 및231&#46;21/215&#46;61/102&#46;80도도 이후 편집에 불변이었다&#46; 안정화한 capture 전후 scene 픽셀이 같았다&#46;17&#46;25초 fractional 각도는157&#46;50/358&#46;75/354&#46;38도였으며 reload 후도 같은 이미지와 값이었다&#46;

#### T1

상태: 통과. 평점: 5. 방법: HTTP.

Action: 세 행성을 각각 canvas와 labeled select로 선택하고 정지 및0&#46;25×재생에서 두 경로를 모두 촬영했다&#46; Expected: 두 선택 경로가 같은 선택·정보·가시 highlight를 갱신해야 한다&#46; Observed: Ember/Azure/Iris 모두 실제 canvas 클릭으로 선택됐고 select값·이름/주기/각도 및 해당 구체의 민트 halo가 일치했다&#46; 별도 list 변경도 정지/재생 각각 세 행성으로 halo를 옮겼다&#46;

#### T2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 세 행성의 정보와 렌더 위치를 재생·seek에서 대조하고 null overview snapshot을 restore했다&#46; Expected: 이름·주기·각도가 렌더와 맞으며 무선택은 정보와 highlight가 지워져야 한다&#46; Observed: Ember12초/Azure24초/Iris48초 정보와 실제 위치 각도 오차가 최대0&#46;74도였다&#46; 재생에서 각도가 계속 갱신됐다&#46; null restore 후 No planet selected&#46; Choose a planet to inspect its orbit&#46;와 disabled Track가 나타났고 이미지에서 halo가 없어졌다&#46;

#### T3

상태: 통과. 평점: 5. 방법: HTTP.

Action: desktop/mobile에서 tracking 재생을3초 이상3회 촬영하고 seek 및 다른 행성 선택을 수행했다&#46; Expected: 선택 구체가 scene 중심 근처에 유지되고 seek/선택 변경이 target·정보·highlight를 함께 갱신해야 한다&#46; Observed: desktop Iris는3&#46;94&#126;7&#46;63초 사이 중심&#40;534,376&#41;에 유지됐다&#46; mobile17&#46;04&#126;20&#46;63초는&#40;188,275&#46;5&#126;276&#41;였고73초 seek 직후도 같았다&#46; 각각 작은 scene변520/350px의10% 기준보다 훨씬 작았다&#46; Azure로 변경하면 halo·정보·중심 구체가 함께 바뀌면서 time/speed/playback을 보존했다&#46; 무선택에서는 Track disabled가 유지됐다&#46;

#### T4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 재생/정지 tracking 및 수동 카메라 변경에서 Overview로 돌아오고 양 viewport의 여러 시각을 촬영했다&#46; Expected: full&#45;system framing을 복원하면서 선택·시간·속도·재생을 보존해야 한다&#46; Observed: desktop에서 구체3개와 전체 궤도가 다시 보였고 정지17초/선택Iris/0&#46;25×가 유지됐다&#46; 재생은 자연 시간만 진행했다&#46; mobile5&#46;2/17/73/120초에서도 전체 시스템이 scene 안에 보였으며 selected와 playback이 의도치 않게 바뀌지 않았다&#46;

#### T5

상태: 통과. 평점: 5. 방법: HTTP.

Action: tracking과null overview snapshot을 live 편집 및 orbit/zoom 뒤 restore하고 같은 viewport에서 반복했다&#46; Expected: 저장 필드를 함께 적용하고 정지하며 정보/위치/target과 결정적 표준 framing을 복원해야 한다&#46; Observed: 30초/2×/ember tracking 및30초/1&#46;5×/null overview,17&#46;25초 tracking을 복원했다&#46; null은 정보를 지웠고 tracking은 선택 구체를 중심에 두었다&#46; orbit/zoom 뒤 restore의 중심 차이는 사전2px 이내였으며 안정화 후 repeated tracking restore와 fractional reload restore의 scene 픽셀은 완전히 같았다&#46;

#### U1

상태: 통과. 평점: 5. 방법: HTTP.

Action: overview/tracking×playing/paused 네 조합에서 실제 pointer drag와 wheel 전후 이미지를 비교했다&#46; Expected: 각 조합에서 카메라 orbit·zoom이 작동하고 simulation state를 바꾸지 않아야 한다&#46; Observed: 모든 조합에서 궤도 투영 방향·구체의 상대 배치가 변했고 wheel로 구체가 확대됐다&#46; 정지17초는 그대로였고 재생은 자연 시간만 진행했다&#46; 선택과mode/speed도 유지됐다&#46;

#### U2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 행성 위치에서 빠른80×40px drag와1&#46;6px씩20회=32px 느린 drag를 수행하고 list/range/button/snapshot/JSON 실제 입력의 scene을 비교했다&#46; Expected: drag가 click selection이 되지 않고 DOM 입력이 scene 선택/카메라로 누출되지 않아야 한다&#46; Observed: 두 drag는 카메라 방향만 바꾸고 무선택을 유지했다&#46; list는 의도한 halo만 변경했다&#46; 속도 변경, capture/delete, export/import 및 반복 seek에서 보존 대상 scene 프레이밍이 같았다&#46;35개 invalid import 전후 scene 픽셀도 모두 같았고 입력 누출은 없었다&#46;

#### U3

상태: 통과. 평점: 3. 방법: HTTP.

Action: 양 viewport에서 playback·selection·tracking·archive·JSON·용량/저장 오류 피드백을 관찰했다&#46; Expected: 주요 목적과 상태·동작 결과가 이해 가능해야 한다&#46; Observed: 명확한 제목,Pause/Resume,속도/시간,planet&#45;ID mapping,restore·capacity·Import error·Saving unavailable 피드백이 읽혔다&#46; 다만 정지 중 LIVE SIMULATION 헤더와 매우 긴 캡처 소수 표현은 덜 정제됐다&#46; 명료하고 사용 가능한 보통 수준의3점을 부여한다&#46;

#### U4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 세 캡처 뒤 네 번째를 시도하고 가운데 항목을 삭제한 다음 다시 캡처했다&#46; Expected: 최대3개·가시 용량 거부·append 순서·나머지 ID/값·live state 보존 및 ID 비충돌을 만족해야 한다&#46; Observed: 네 번째는 Archive full: delete an observation before capturing another&#46;를 표시하고 세 export 항목을 보존했다&#46; middle 삭제 뒤 나머지 순서/값/ID를 유지하고 새 obs&#45;mvjncvzz&#45;l4hhw0를 끝에 추가했다&#46; time/speed/selection/mode/playback은 그대로였고 안정화한 capture/delete 전후 scene 픽셀도 같았다&#46;

#### U5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 1440×900과390×844에서 Tab/Shift&#43;Tab,Enter,range 방향키,select typeahead로 모든 primary DOM control을 조작하고 각 동작 직후 결과와 focus 이미지를 관찰했다&#46; Expected: 의미 있는 accessible name,가시 focus,정상 keyboard 결과와 trap 없는 이동을 제공해야 한다&#46; Observed: play/resume/seek/speed/selection/track/overview/capture/restore/delete/textarea/import/export를 모두 수행했다&#46; capture 뒤 시간·선택·mode를 바꾸고 keyboard restore가 원래4필드와정지를 복원했다&#46;각 input 연결 label과 snapshot별 aria&#45;label이 있었고 실제 이미지에서 금색 outline이 보였다&#46; Shift&#43;Tab으로 textarea 밖을 이동해 trap이 없었다&#46;

#### V1

상태: 통과. 평점: 4. 방법: HTTP.

Action: 양 viewport의 overview/tracking,선택·seek·camera 상태와 캡처 이미지를 직접 확인했다&#46; Expected: 관측 대상과 선택이 읽기 쉬워야 한다&#46; Observed: 서로 구별되는 주황·청록·보라 구체와 밝은 태양,민트 halo가 선명했다&#46; tracking에서는 선택을 크게 보고 overview는 전체를 볼 수 있었다&#46; 얇은 궤도선의 대비가 다소 약한 점을 제외하면 상태가 읽기 쉬워4점을 부여한다&#46;

#### V2

상태: 통과. 평점: 4. 방법: HTTP.

Action: 양 viewport의 full&#45;page panel 배치와 keyboard focus·archive·JSON 영역을 직접 확인했다&#46; Expected: 타이포·간격·색·계층이 일관되고 읽기 쉬워야 한다&#46; Observed: 어두운 카드,민트 transport 강조,명확한 섹션 제목과 일관된 간격이 desktop2열/mobile1열에 유지됐다&#46; snapshot 값과 JSON은 조밀하고 긴 소수는 다소 산만하지만 전반적으로 정돈된 계층이어서4점을 부여한다&#46;

#### L1

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제1440×900 document에서 모든 primary control과 snapshot/JSON을 pointer·keyboard로 조작하고 일반/full&#45;page 이미지를 확인했다&#46; Expected: 가로 스크롤이나 panel 가림 없이 scene과control을 사용 가능해야 한다&#46; Observed: innerWidth1440/innerHeight900/scrollWidth1425였다&#46; scene,transport,우측 selection/JSON은 겹치지 않았고 아래 archive의restore/delete는 수직 스크롤로 접근했다&#46; 모든 기본 동작이 성공했고 가로로 숨은 control은 없었다&#46;

#### L2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제390×844에서 selection/tracking/overview/seek/speed/play/capture/restore/delete/JSON을 조작하고 focus 및 전체 panel 이미지를 확인했다&#46; Expected: mobile에서도 가로 스크롤·차단 panel 없이 같은 기능을 제공해야 한다&#46; Observed: innerWidth390/innerHeight844/scrollWidth375였다&#46; scene은350×360 정도의 영역에서 overview 전체와 tracking 선택을 읽을 수 있었고 panel들은 수직으로 쌓였다&#46; 3개 archive와 JSON이 긴 페이지를 만들지만 수직 이동으로 모든 control이 정상 사용됐고 가로 overflow나 겹침은 없었다&#46;

#### D2

상태: 통과. 평점: 5. 방법: 정적 검사.

Action: output/index&#46;html의 모든 자원 참조와 output 디렉터리를 검사했다&#46; Expected: 모든 로컬 자산이 output HTTP root 아래에 있고 build/root 의존성이 없어야 한다&#46; Observed: 유일한 로컬 script 참조 three&#46;min&#46;js가 output/three&#46;min&#46;js로 존재했다&#46; CSS와 앱 script는 인라인이며 추가 font/texture/build/root 의존성은 없었다&#46;

#### D3

상태: 통과. 평점: 5. 방법: HTTP.

Action: Ego의 Runtime/Network 진단을 탐색 전에 켜고 scratch 및보완 p2 about:blank에서 console&#46;error/비동기 exception 양성 대조를 수행한 뒤각 시퀀스 events를 수집했다&#46; Expected: 의도한 작업을 막는 artifact 원인의 page/console/resource 오류가 없어야 한다&#46; Observed: Runtime&#46;consoleAPICalled&#40;error&#41;와Runtime&#46;exceptionThrown을 모두 수신해 진단 전송을 확인했다&#46; artifact에서는 오류/Network&#46;loadingFailed가 수집되지 않았고 로컬 Three&#46;js가 로드됐다&#46; classic&#45;build deprecation warning만 있었다&#46; 앞선 screenshot transport 장애는 artifact 오류로 세지 않았으며 보완에서는 캡처와 기능 관찰이 정상 완료됐다&#46;

#### D4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 명시 Export/Import와 labeled textarea로 empty/populated 목록을 교환하고 캡처 payload의 export→import→reload를 수행했다&#46; Expected: 정확한 version1 및5필드만 내보내고 ID/타입/순서·목록·live state를 보존해야 한다&#46; Observed: 빈 export는&#123;version:1,snapshots:&#91;&#93;&#125;였고 populated export에는id/time/speed/selectedPlanet/viewMode만 있었다&#46;각도/playback/camera offset은 없었다&#46; 실제 행성 ID mapping과stable IDs·순서가 유지됐고 캡처 ID도 전체 round&#45;trip에서 같았다&#46; export/import 전후 정지 scene 픽셀과 live UI가 같았고 재생 교환은 자연 진행만 유지했다&#46;

#### D5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 textarea/import로35개 invalid 사례 및 empty/3&#45;entry·경계값·fraction·extra&#45;key 유효 payload를 각각 적용하고 목록/live UI/scene 이미지를 비교했다&#46; Expected: 전체 atomic validation과list&#45;only replacement,정확한 ID·순서·derived angles 및camera 보존을 만족해야 한다&#46; Observed: malformed,root/entry null·array,누락필드,version문자열/2,숫자문자열/boolean,범위초과,1e400,non&#45;array list,4개/중복/빈·숫자ID,잘못된planet/mode/null tracking,mixed&#45;validity를 모두 visible error로 거부했다&#46; 기존 비어있지 않은 payload/live state 및35개 scene 이미지가 모두 같았다&#46;공백/A/a ID,0·120초/&#46;25·4×/17&#46;25초·1&#46;5×를 보존하고extra keys/999도 입력은 무시했다&#46; 정지 valid/empty import 전후 scene도 같았고 재생 import는0&#46;2044초에0&#46;30초&#40;1&#46;5×&#41;만 진행하며camera framing을 유지했다&#46;

#### D6

상태: 통과. 평점: 5. 방법: HTTP.

Action: 같은8891 origin/context에서 import/delete/capture 각각 reload하고 app 전용 key의invalid payload와bounded quota fixture를 시험했다&#46; Expected: 목록 persistence·R2 live 기본값·invalid 복구·저장실패 안내와 memory 사용성을 제공해야 한다&#46; Observed: 각 mutation의 ID/순서/값/각도가reload에 유지됐고 live는0&#46;14&#126;0&#46;33초/1×/재생/무선택/overview였다&#46; malformed/version2/문자열time 저장값은 빈목록과가시 recovery notice를 보였다&#46; quota filler5242803문자에서 capture는 memory에 추가됐지만 저장 raw는빈목록으로남고 Saving unavailable&#46; Observations remain in memory only&#46;가 보였다&#46;재생/선택/복원/삭제는계속가능했다&#46; fixture 원값 복원과filler 제거,마지막 최초 key absence 복원을확인했다&#46;별도read&#45;denial은시험하지않았고허용된실제write&#45;failure경로를입증했다&#46;

## 검증 및 한계

브라우저 검증 방식: HTTP.
뷰포트: 1440x900, 390x844.
검증 범위: 통과 27개, 실패 0개, 미검증 0개.

### 참고 사항

- 범위는 지정된 solar&#45;system&#45;observatory&#45;2 Low 실행 하나이다&#46; 처음 all&#45;unverified 체크포인트와 partial 검토 뒤 사용자 요구에 따라 같은 작업의 누락 검증을 실질 보완했다&#46; 이번 최종 판단은 새 UI·이미지 관찰을 근거로 하며 이전 리뷰 JSON을 증거로 복사하지 않았다&#46; 생성 run&#46;json의 failed 측정 상태나 generation notes는 품질 증거/감점 근거로 사용하지 않았다&#46;
- 리뷰어 모델 openai&#45;codex/gpt&#45;6&#45;astra는 현재 호스트 주입 메타데이터,effort high는 이 리뷰 대리 작업에 명시 적용된 구성에 근거한다&#46; 독립 runtime effort scalar는 노출되지 않았으며 generic default나 생성 폴더에서 추정하지 않았다&#46; 최종시각은 현재 호스트 Python UTC 실제 시계이다&#46;
- 명시된 benchmark&#45;review를 적용했다&#46; using&#45;harness&#45;flow는 대리 작업 예외를 읽었고,호스트 필수 ego&#45;browser를 적용하여 설치된 전체 가이드를 읽었다&#46; screenshot 장애 조사 때문에 부모와 이 대리 작업은 higher&#45;priority mandatory systematic&#45;debugging을 추가 적용했다&#46; 모델 변경·새 모델·descendant·Aside·다른 브라우저·dependency 설치는 없었다&#46;
- 사전 허용오차를 유지했다: 속도는3초 이상, max&#40;0&#46;15초,기대delta의5%&#41;&#43;표시0&#46;01초;정지0&#46;01초/각도1도;예측각도2도&#43;표시0&#46;01도;동일상태 중심2 CSS px;tracking은scene 작은변의10%;ID/순서/타입/보존필드는정확비교&#46; 실패 뒤 허용오차를 넓히지 않았다&#46;
- 원래8단계를 보완에서 canvas Iris선택부터 올바른순서로 재실행했다: 자동운동→scene선택/추적→정지30초→정지2×→2×재생3초이상→list Azure전환→overview→정지30초반복&#46; 이어 저장된추가1&#126;6&#40;snapshot restore,capacity,invalid JSON,valid round&#45;trip,persistence/recovery,양viewport keyboard&#41;를수행하고남은전체matrix를관찰했다&#46;
- 원래p1의 첫1440×900 screenshot만성공하고후속 Page&#46;captureScreenshot timeout 및 fromSurface:false의 Unable to capture screenshot이있었다&#46; 근본원인은미확정이다&#46; 부모는동일TaskSpace1의기존빈공간에서p2를생성하여blank/desktop/mobile 캡처정상화를확인했고,승인받은보완은그p2만사용했다&#46; 다른공간/런타임으로우회하거나원인을확정했다고주장하지않는다&#46; Browser&#46;getWindowForTarget 실패는반복하지않았다&#46;
- 보완은Ego nodejs의문서화Page/keyboard/mouse/evaluate/cdp만사용하고127&#46;0&#46;0&#46;1:8891이변경없는output만제공했다&#46; 매호출viewport를다시설정하고실제1440×900/390×844를확인했다&#46;스크린샷을직접읽었고임시contact sheet는실제scene/focus pixels를crop/배치한것이다&#46; 설치된ffmpeg와표준Python으로색구체중심을측정하고검사된표준카메라투영을역산했다&#46;앱내부camera/mesh state를노출하거나DOM/Three&#46;js를mock/수정하지않았다&#46;
- 보완의정지repeat&#45;seek/reload/속도변경,안정화capture/delete/repeated restore,export/import 및35개invalid scene 비교는픽셀동일이었다&#46; 일부초기자동스크롤직후이미지는페이지전체가1&#126;2px이동했으므로동일픽셀이라고보고하지않았고,표준framing 반복중심은사전2px이내였다&#46; 별도안정화된관찰에서는snapshot/camera보존의정확한scene동일성도확인했다&#46;
- 초기시도의비동기1×표본3&#46;2042초→3&#46;01초는사전오차밖이었고보존한다&#46; 당시느린갱신과캡처장애원인을앱결함으로확정하지않았다&#46; 보완의단독scene에서는&#46;25/2/4×에3&#46;3039/3&#46;3088/3&#46;3217초동안&#46;83/6&#46;62/13&#46;28초진행했고visibilityState는visible이었다&#46; time range의기존thumb안클릭으로값이안바뀌어리뷰helper가중단된두경우는현재DOM을확인한뒤실제Home/End와range입력으로해결했다&#46;앱오류나숨은state수정이아니다&#46;
- 저장소는app key orbit&#45;observatory&#45;2&#45;snapshots와독립review filler만접근했다&#46;8MiB이하bounded search로실제QuotaExceededError를만들었으며prototype/security설정은변경하지않았다&#46;각fixture원raw값을복원했고최종app key 최초부재및이전/새filler부재를확인했다&#46; 직접file://실행,터치실기기,별도storage read&#45;denial은범위의인증에포함하지않는다&#46;
- 정리: 부모소유p2는최신지시에따라닫거나finish하지않고about:blank로돌려scene렌더를종료했으며emulation을해제하고부모에게반환했다&#46;부모소유proc://low&#45;capture&#45;diagnosis 서버종료를요청했다&#46;임시스크립트/이미지/metadata/draft는최종게시후삭제한다&#46;초기와보완후prompt&#46;md/run&#46;json/output두파일의SHA&#45;256일치를작업메모리에서확인했으며해시자체는리뷰파일에보존하지않는다&#46;
