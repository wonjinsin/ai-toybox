# 벤치마크 리뷰

## 결론

상태: 완료.
필수 요구사항 판정: 실패.
품질 총점: 61.55 / 100. 확정 점수. 범위: 61.55.
품질 점수에는 상한 제한을 적용하지 않습니다. 필수 요구사항 판정은 품질 점수와 독립적입니다.
필수 항목 실패 (9): S7, S8, T2, T4, T5, U1, U2, U5, D6.
필수 항목 미검증 (0): 없음.
미검증 항목 (0): 없음.
누락된 리뷰어 메타데이터: 없음.
형식: 5&#46;0. 평가 기준: solar&#45;system&#45;5&#46;0.
리뷰 시각: 2026&#45;10&#45;22T16:00:24&#46;806Z.
리뷰어 모델: openai&#45;codex/gpt&#45;6&#45;astra.
리뷰어 추론 수준: high.

## 점수

| 평가 영역 | 점수 / 5 | 가중치 | 근거 |
| --- | --- | --- | --- |
| 시뮬레이션 (simulation) | 3.75 | 25 | S1, S2, S3, S4, S5, S6, S7, S8; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 선택 및 추적 (selection) | 2.00 | 20 | T1, T2, T3, T4, T5; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 사용성 (usability) | 1.60 | 15 | U1, U2, U3, U4, U5; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 시각적 완성도 (visual) | 3.00 | 15 | V1, V2; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 반응형 (responsive) | 5.00 | 15 | L1, L2; 확정 점수; 전체 영역 항목을 분모로 사용 |
| 실행 및 제공 (delivery) | 3.00 | 10 | D2, D3, D4, D5, D6; 확정 점수; 전체 영역 항목을 분모로 사용 |

## 발견 사항

### 확인된 실패

#### S7

상태: 실패. 평점: 0. 방법: HTTP.

Action: 로드된 unpkg Three&#46;js r160 HTTP 200, index&#46;html의 WebGLRenderer/SphereGeometry/TorusGeometry/PerspectiveCamera 및 실제 렌더 장면을 대조했다&#46; 정지 overview에서 &#40;500,450&#41;부터 &#40;620,530&#41;까지 30회 이동 드래그하고 tracking/재생에서도 반복했으며 휠 zoom도 수행했다&#46; Expected: 실제 3D 기하를 렌더하고 카메라 orbit에 따라 깊이/가림/원근 시점이 변해야 한다&#46; Observed: Three&#46;js 리소스와 실제 행성/궤도 렌더 및 휠에 따른 원근 크기 변화는 확인되었다&#46; 그러나 정지 overview 드래그 전후 장면 RGB 차이는 0이고 orbit 시점이 전혀 바뀌지 않았다&#46; pointerdown은 TypeError: Cannot set properties of undefined &#40;setting &#39;x&#39;&#41;, pointermove는 Cannot read properties of undefined &#40;reading &#39;x&#39;&#41;를 발생시켰다&#40;index&#46;html:10:531,10:724&#41;&#46; 소스는 canvas가 아니라 WebGLRenderer의 renderer&#46;dataset을 사용한다&#46; 따라서 R1의 orbit으로 기하 시점이 변해야 한다는 조건이 실제 실패했으며 기존 미검증을 pass로 추정하지 않았다&#46; 동작하는 orbit의 깊이/가림 변화는 이 확정 결함 때문에 관찰할 수 없지만 체크는 미검증이 아닌 fail이다&#46;

#### S8

상태: 실패. 평점: 0. 방법: HTTP.

Action: 정지 30초/2x/Ember/tracking의 obs&#45;1 및 재생 중 snapshot을 Capture하고 모든 live 필드를 바꾼 뒤 목록/JSON/reload를 비교했다&#46; 17&#46;25초 fractional snapshot도 import했다&#46; Expected: capture 상태/ID가 안정적이며 세 각도가 저장 시각의 실제 궤도와 일치한다&#46; Observed: ID와 time/speed/selection/mode는 변경 없이 유지되고 capture는 live 상태를 보존했다&#46; 그러나 30초 목록은 Ember 70&#46;4°, Azure 295&#46;1°, Jade 341&#46;0°를 표시했다&#46; 실제 렌더/고정 위상의 기대값은 180&#46;0°,158&#46;8°,2&#46;5°이다&#46; 17&#46;25초도 134&#46;7°/196&#46;1°/347&#46;9°로 실제 157&#46;5°/327&#46;5°/266&#46;9°와 다르다&#46; index&#46;html:8의 angles 구조 분해가 orbit 반경을 period로, period를 phase로 읽는다&#46; 저장 목록 각도 정확성의 확정 실패이며 import/reload 후에도 같은 오류가 유지된다&#46;

#### T2

상태: 실패. 평점: 0. 방법: HTTP.

Action: 각 행성을 선택하고 0/30초 seek 및 재생 중 정보 패널과 측정된 궤도를 대조했다&#46; null&#45;selection snapshot도 복원했다&#46; Expected: 이름/주기/각도가 실제 행성과 일치하고 선택 없음은 명시적으로 비워진다&#46; Observed: 이름과 12/24/48초 주기 및 No planet selected&#46; 복원은 맞았지만 0초 Ember 각도가 327&#46;5°&#40;실제 0°&#41;, 30초 70&#46;4°&#40;실제 180°&#41;였다&#46; Azure 30초도 295&#46;1°&#40;실제158&#46;8°&#41;였다&#46; 재생 시 값은 바뀌지만 실제 위치와 불일치한다&#46; 원인은 index&#46;html:8 angles의 잘못된 배열 인덱스다&#46;

#### T4

상태: 실패. 평점: 0. 방법: HTTP.

Action: desktop 및 390×844에서 tracking 후 Full overview로 돌아왔다&#46; 특히 mobile에서 0초/1&#46;5x/정지/Jade 선택을 유지한 채 전환했다&#46; Expected: 선택/시간/속도/재생 유지와 전체 태양계 프레이밍&#46; Observed: 상태는 정확히 유지되었으나 mobile canvas &#40;8,408&#46;19&#41;–&#40;367,836&#46;19&#41;의 overview에는 태양/Ember/Azure만 보이고 선택된 Jade는 왼쪽 장면 밖으로 사라졌다&#46; 궤도도 양옆으로 잘렸다&#46; desktop 기본 카메라도 바깥 궤도 하단이 잘린다&#46; overview 핸들러는 mode만 바꾸고 크기에 맞는 전체계 프레이밍을 설정하지 않는다&#46;

#### T5

상태: 실패. 평점: 0. 방법: HTTP.

Action: obs&#45;1&#40;30초,2x,Ember,tracking&#41;을 live 상태 변경 뒤 Restore하고, 실제 드래그/휠 &#43;300을 수행한 뒤 다시 Restore했다&#46; null&#45;selection overview 및 17&#46;25초 tracking import도 복원했다&#46; Expected: 모든 상태 적용/정지와 반복 가능한 표준 카메라 프레이밍&#46; Observed: 시간/속도/선택/모드/정지는 적용되고 null restore는 정보/강조를 지웠다&#46; 하지만 반복 tracking restore에서 zoom이 남아 태양 중심이 약&#40;699,629&#41;에서&#40;671,595&#41;로, 지름이 약181px에서147px로 달라졌다&#46; 동일 크기/시각/선택이며 2px 오차를 크게 넘는다&#46; restore는 theta/phi/dist를 초기화하지 않는다&#46; overview null restore 역시 기존 zoom을 유지했다&#46; 각도 정보 불일치는 T2/S8과 동일하게 남는다&#46;

#### U1

상태: 실패. 평점: 0. 방법: HTTP.

Action: overview/tracking 각각 재생/정지에서 120×80px 드래그와 wheel zoom을 실행했다&#46; Expected: 두 카메라 조작이 네 상태 모두 작동하고 시뮬레이션 상태를 바꾸지 않는다&#46; Observed: wheel zoom은 크기를 바꿨지만 모든 orbit 드래그가 renderer&#46;dataset TypeError를 내며 시점을 바꾸지 않았다&#46; 정지 overview의 드래그 전후 장면 픽셀은 동일했다&#46; 시간/선택/모드 자체는 유지되었다&#46;

#### U2

상태: 실패. 평점: 0. 방법: HTTP.

Action: 정지 0초에 Azure를 목록 선택한 뒤 Ember 중심에서 30px를 30개의 1px 이동으로 느리게 드래그했다&#46; 빠른 20px 드래그와 각 DOM 컨트롤도 사용했다&#46; Expected: 드래그는 행성 선택이 아니며 DOM 입력은 canvas로 새지 않는다&#46; Observed: 느린 드래그와 빠른 드래그 모두 끝나자 선택이 Azure에서 Ember로 바뀌었다&#46; pointerdown/move의 renderer&#46;dataset 예외 때문에 drag 플래그가 설정되지 않고 pointerup이 raycast 선택을 실행했다&#46; 시간은 0&#46;00초로 유지되었다&#46; 목록/slider/button/snapshot/JSON 조작 자체에서 추가 canvas 선택은 관찰하지 않았으며 JSON 전후 장면 RGB 차이도0이었다&#46;

#### U5

상태: 실패. 평점: 0. 방법: HTTP.

Action: 두 실제 뷰포트에서 Tab/Shift&#43;Tab, 버튼 Enter, range Home/End/화살표, textarea 키보드 입력으로 모든 주 컨트롤을 시도했다&#46; Restore/Delete를 focus한 뒤 100ms 후 Enter도 사용했다&#46; Expected: 이름/가시적 focus/정상 키 조작과 trap 없음&#46; Observed: 일반 컨트롤은 Tab 이동과 노란3px focus가 보이고 selection/pause/speed/seek/tracking/overview/capture/export/import가 작동했다&#46; 그러나 양쪽 크기 모두 snapshot Restore/Delete focus가 100ms 안에 BODY로 사라져 Enter가 아무 동작도 하지 않았다&#46; renderUI가 매 프레임 snapshot DOM을 재생성한다&#46; 포인터 좌표 클릭으로는 두 동작이 성공했으므로 키보드 접근성의 확정 실패다&#46; 일반 컨트롤에서 나가는 Tab/Shift&#43;Tab은 가능했다&#46;

#### D3

상태: 실패. 평점: 0. 방법: HTTP.

Action: 별도 reviewer&#45;owned scratch에서 console&#46;error와 비동기 throw 양성 대조를 실행하고 Runtime을 탐색 전에 활성화하여 대상 navigation/resource/실행 오류를 수집했다&#46; Expected: intended operation을 방해하는 산출물 예외/필수 리소스 실패가 없다&#46; Observed: Runtime&#46;consoleAPICalled와 Runtime&#46;exceptionThrown 두 양성 신호를 모두 받았다&#46; 실제 Three&#46;js HTTP200과 문서200을 확인했다&#46; canvas pointerdown/move는 index&#46;html:10:531/724의 undefined dataset TypeError를 반복 발생시켜 orbit을 방해했다&#46; 비필수 favicon404 및 CDN의 deprecation warning은 기능 실패의 근거로 삼지 않았다&#46; 진단 transport 미확인을 빈 로그 통과로 바꾸지 않았다&#46;

#### D6

상태: 실패. 평점: 0. 방법: HTTP.

Action: 같은127&#46;0&#46;0&#46;1:8791 origin에서 import/delete/capture 각각 뒤 reload, malformed/schema&#45;invalid 저장소 복구,8MiB 상한 quota filler와 실제 Capture를 수행했다&#46; Expected: 동일 schema/ID/순서 지속성, 잘못된 저장 데이터의 복구 알림, 실제 쓰기 실패의 가시 saving&#45;unavailable 알림과 in&#45;memory 사용성&#46; Observed: 세 persistence round&#45;trip의 JSON은 정확히 일치하고 reload live는 기본값으로 시작했다&#46; malformed 및 speed 문자열 저장 데이터는 빈 목록과 Saved snapshots were invalid and were cleared&#46;를 표시하고 선택/seek가 가능했다&#46; quota fixture는 최대5242805문자에서 저장 한계에 도달했고 QuotaExceededError를5회 관측했다&#46; 그 상태에서 Capture 후 in&#45;memory obs&#45;1이 생겼지만 저장 값은 &#123;&#34;version&#34;:1,&#34;snapshots&#34;:&#91;&#93;&#125; 그대로여서 실제 app 쓰기 실패가 확인됐다&#46; 화면은 Saving unavailable 대신 Snapshot captured&#46;만 표시했다&#46; save&#40;&#41;의 오류 알림을 capture의 성공 알림이 덮는다&#46; 재생/선택은 계속 작동했다&#46; filler 및 app key는 원래 absence로 복구했다&#46; read&#45;denial은 이 허용된 실제 write&#45;failure 분기를 대체하기 위한 필수 추가 조건이 아니다&#46;

### 검증하지 못한 항목

없음.

### 검증된 통과 항목

#### S1

상태: 통과. 평점: 5. 방법: HTTP.

Action: HTTP 장면에서 태양/세 행성을 관찰하고 실제 time 슬라이더로 0, 3, 6, 12, 24, 48초 및 17, 30, 73초를 선택했다&#46; 카메라는 같은 overview와 실제 휠 &#43;600 위치로 유지하고 스크린샷의 색상별 중심을 측정했다&#46; Expected: 서로 다른 크기/색의 세 행성이 고정 위상과 12/24/48초 원 궤도를 따른다&#46; Observed: 노란 태양 하나와 적색 Ember, 청색 Azure, 녹색 Jade 세 개가 보였다&#46; 0초 중심은 각각 &#40;650&#46;41,571&#46;12&#41;, &#40;373&#46;82,617&#46;67&#41;, &#40;251&#46;77,387&#46;05&#41; CSS px였다&#46; Ember는 12초, Azure는 24초, Jade는 48초에 동일 중심으로 돌아왔고 중간 시각에는 각 궤도를 따라 이동했다&#46; 실제 렌더 중심과 위상 0/1&#46;2/2&#46;4 rad, 반경 7/10/13, 주기 12/24/48의 원 궤도 투영 잔차는 0&#46;68–1&#46;04px였다&#46; 화면상의 기울어진 타원을 원 궤도의 원근 투영으로 비교했으며 화면 각도를 공전 각도로 오인하지 않았다&#46;

#### S2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 저장소가 원래 비어 있는 origin에 새로 탐색하고 여러 reload 뒤 시작 상태와 3&#46;2033초 자동 재생을 읽었다&#46; Expected: 0초/1x/재생/선택 없음/overview로 시작하고 자동 이동한다&#46; Observed: 초기 관측은 로드 경과 0&#46;15초, 1&#46;00x, Pause, Overview, No planet selected&#46;였고 3&#46;2033 실제 초 뒤 3&#46;35초로 증가했다&#46; 이미지에서 행성 이동을 확인했다&#46; import/delete/capture 후 reload도 0&#46;22–0&#46;36초의 정상 로드 경과만 가진 동일 기본 상태로 시작했으며 저장된 snapshot을 자동 적용하지 않았다&#46; 선택 없음에서는 track&#46;disabled=true였다&#46;

#### S3

상태: 통과. 평점: 5. 방법: HTTP.

Action: 움직이는 장면을 Pause로 정지하고 30초로 seek한 뒤 3&#46;2032초 대기, 선택 및 카메라 입력, Resume를 수행했다&#46; Expected: 정지 시간/위치 유지, 재개 시 현재 시각에서 연속 이동&#46; Observed: 30&#46;00초가 그대로였고 동일 뷰포트에서 정지 전후 장면 이미지의 RGB 차이가 0이었다&#46; 정지 중 세 행성 선택은 작동했다&#46; 2x 재개 뒤 17&#46;11→23&#46;51초/3&#46;2031 실제 초 등 연속 증가를 관측했다&#46; 카메라 orbit 자체의 예외는 독립 U1/S7에 기록하며 정지 시뮬레이션은 움직이지 않았다&#46;

#### S4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 0&#46;25x, 2x, 4x를 각각 정지 중 설정 후 재개하고, 재생 중에도 각 속도로 변경하여 3초 이상 측정했다&#46; Expected: 현재 시간/위치가 유지되고 증가율이 설정 속도를 따르며 상대 궤도 주기가 유지된다&#46; Observed: 정지 설정 모두 17&#46;00초로 유지되었다&#46; 재개 측정은 0&#46;25x:0&#46;80/3&#46;2031초, 2x:6&#46;40/3&#46;2031초, 4x:12&#46;80/3&#46;2030초였다&#46; 재생 중 변경 측정은 0&#46;25x:0&#46;78/3&#46;1043초, 2x:6&#46;24/3&#46;1094초, 4x:12&#46;42/3&#46;1049초였다&#46; 모두 사전 허용오차 내이며 document&#46;visibilityState=visible, 유일한 대상 장면이었다&#46; 속도 변경 후 같은 시각의 세 렌더 중심이 재현되었고 재개 시 seek/reset이 발생하지 않았다&#46;

#### S5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 슬라이더로 0,30,120,17,73초를 seek하고 속도/재생/선택 변경 후 재방문했으며 reload 후 같은 카메라 설정으로 재측정했다&#46; 17&#46;25초는 실제 JSON import와 Restore로 설정했다&#46; Expected: seek가 정지하고 즉시 세 위치/시각을 갱신하며 같은 시각의 위상이 재현된다&#46; Observed: 모든 seek의 버튼은 Resume였고 시각이 지정값과 일치했다&#46; 0/17/30/73초의 reload 전후 세 행성 중심 차이는 각각 0&#46;00px였다&#46; 선택 변화로 강조 색이 변한 재방문에서도 중심 차이는 최대 약 0&#46;1px였고 30초 정지 이미지도 동일했다&#46; 17&#46;25초 Restore가 실제 시각을 17&#46;25로 적용했다&#46; 잘못 계산된 정보/저장 각도는 T2/S8의 별도 위반이다&#46;

#### S6

상태: 통과. 평점: 5. 방법: HTTP.

Action: 119초에서 4x 재생 후 0&#46;7초 및 추가 0&#46;5초 대기하고 17초로 seek하여 재개했다&#46; Expected: 정확히 120초에서 멈추고 wrap/초과 없이 이전 시각에서 다시 재생된다&#46; Observed: 두 관측 모두 120&#46;00초와 Resume였으며 이후 17초에서 재개하여 18&#46;39초, Pause로 진행했다&#46;

#### T1

상태: 통과. 평점: 5. 방법: HTTP.

Action: Ember/Azure/Jade 각각을 실제 목록과 canvas 클릭으로 정지/재생 상태에서 선택했다&#46; Expected: 두 경로의 선택과 강조 표시가 일치한다&#46; Observed: 여섯 scene 선택과 여섯 목록 선택 모두 지정 ID/이름으로 전환되었고 aria&#45;pressed도 일치했다&#46; 예를 들어 정지 0초 canvas 좌표 Ember&#40;697&#46;91,627&#46;90&#41;, Azure&#40;289&#46;07,700&#46;36&#41;, Jade&#40;142&#46;45,364&#46;90&#41; 클릭이 각각 해당 행성을 선택했다&#46; 실제 이미지에서 적색/청색/녹색 선택의 밝은 강조와 목록 강조가 확인되었다&#46; pointerdown 예외는 있으나 pointerup 선택은 작동하며 드래그 오선택은 U2에 별도 기록했다&#46;

#### T3

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제 Track 버튼으로 재생/정지 추적, 73초 seek, Ember→Jade 전환 및 선택 없는 시작 상태를 검사했다&#46; Expected: 선택 중심이 장면 중심 10% 안에 있고 seek/선택 전환에 즉시 동기화되며 무선택 추적은 불가능하다&#46; Observed: 실제 장면 약1072×874&#46;36px의 중심 &#40;549,450&#46;18&#41; 대비 3&#46;54초에 걸친 네 Ember 이미지 중심은 &#40;548&#46;44–548&#46;56,449&#46;40–449&#46;48&#41;로 1px 안이었다&#46; 73초 seek 및 Jade 전환도 중심 추적과 이름/강조가 함께 전환되고 73&#46;00초/1x/정지가 유지되었다&#46; 재생 중 Azure 전환은 2x/재생을 유지했다&#46; 시작 track&#46;disabled=true로 무선택 추적이 차단되었다&#46; 각도 수치 정확성은 T2 실패로 분리한다&#46;

#### U3

상태: 통과. 평점: 3. 방법: HTTP.

Action: 재생/정지/속도/선택/추적, snapshot 용량, JSON 성공/거부 상태를 두 크기에서 사용했다&#46; Expected: 상태와 작업 목적 및 피드백이 이해 가능하다&#46; Observed: Pause/Resume, Time, Global speed, Overview/Tracking, 명시적 ID 매핑과 snapshot 필드/오류 문구는 일관되고 일반적인 관측 도구 수준으로 이해하기 쉬웠다&#46; 별도 카메라 조작 설명이 없고 저장 관련 메시지 영역을 공유하여 정보가 쉽게 덮인다&#46; 특별히 뛰어난 안내 수준은 아니므로 정성 앵커3을 부여했다&#46; 잘못된 공전 각도/저장 실패의 의무 위반은 각 소유 체크에 기록했다&#46;

#### U4

상태: 통과. 평점: 5. 방법: HTTP.

Action: obs&#45;1/obs&#45;2/obs&#45;3을 차례로 캡처하고 네 번째 캡처, 가운데 obs&#45;2 삭제, 다시 캡처 및 export를 수행했다&#46; Expected: 3개 제한, 명시적 거절, 기존 값/순서 보존, 삭제 뒤 새 캡처 가능과 충돌 없는 ID&#46; Observed: 네 번째는 Snapshot capacity reached; delete one before capturing&#46;을 표시하고 export된 세 항목은 그대로였다&#46; 삭제 후 obs&#45;1/obs&#45;3 순서를 유지하고 새 obs&#45;2가 맨 뒤에 추가되었다&#46; 현재 목록 내 충돌은 없고 잔존 snapshot ID/값과 live 17&#46;00초/2x/Azure/overview/정지는 변하지 않았다&#46; 삭제된 ID의 재사용은 생존 snapshot의 안정성을 깨지 않는다&#46;

#### V1

상태: 통과. 평점: 3. 방법: HTTP.

Action: 두 크기의 실제 overview/tracking 및 선택 이미지를 확인했다&#46; Expected: 장면/선택이 읽기 쉽고 효과가 관측을 방해하지 않는다&#46; Observed: 검정 배경, 서로 다른 색/크기, 밝은 선택 강조와 얇은 궤도선으로 기본 가독성은 명확했다&#46; 과한 효과는 없지만 구의 조명이 거의 평면색처럼 보여 깊이 단서가 약하고 궤도 대비도 낮다&#46; 전체계 잘림은 T4에 기록했으며 장식이나 소스 분량 대신 관측된 보통 수준의 명료성으로3을 부여했다&#46;

#### V2

상태: 통과. 평점: 3. 방법: HTTP.

Action: desktop/mobile의 controls, snapshot 목록과 JSON 영역을 실제 이미지로 검사했다&#46; Expected: 타이포그래피/여백/색상/패널 구조가 일관되고 읽을 수 있다&#46; Observed: 짙은 파랑 패널과 얇은 테두리, 읽기 쉬운 본문/노란 focus 및 상태 색은 일관됐다&#46; snapshot의 작은 글씨는 정보 밀도가 높고 mobile은 컨트롤과 장면을 보려면 많이 세로 스크롤해야 한다&#46; 기본적으로 정돈된 일반적 수준이므로3이며 mobile의 허용된 세로 스크롤 자체를 실패로 처리하지 않았다&#46;

#### L1

상태: 통과. 평점: 5. 방법: HTTP.

Action: 매 관련 invocation에서 실제1440×900을 검증하고 selection, tracking/overview, seek, speed, pause/resume, capture/restore/delete, textarea import/export를 실행했다&#46; Expected: 가로 스크롤/패널 가림 없이 장면과 주 컨트롤에 접근할 수 있다&#46; Observed: scrollWidth는1440 또는 세로 스크롤바를 제외한1425였고 viewport를 넘지 않았다&#46; controls는 우측 패널 안에서 접근 가능하며 snapshot 좌표 클릭도 성공했다&#46; 패널이 canvas 입력 또는 버튼을 덮지 않았다&#46; snapshot 키보드 소실과 overview 프레이밍은 크기와 무관한 U5/T4 결함으로 따로 판정했다&#46;

#### L2

상태: 통과. 평점: 5. 방법: HTTP.

Action: 실제390×844, scrollWidth375에서 전체 DOM 조작을 키보드 및 실제 포인터로 실행하고 canvas를 세로 스크롤하여 관찰했다&#46; Expected: 수평 스크롤/컨트롤 가림 없이 동일 주요 작업에 접근 가능&#46; Observed: 선택/추적/seek/2x/재생, capture, JSON export/import가 동작했고 mobile snapshot Restore&#40;64&#46;77,618&#46;69&#41;, Delete&#40;128&#46;43,618&#46;69&#41;는 elementFromPoint가 해당 버튼이며 실제 클릭도 성공했다&#46; 복원된17&#46;25초/1&#46;5x/Azure/tracking에서 재생 후17&#46;87초로 증가했다&#46; controls는 폭 안에 있고 패널은 장면 위 별도 행에 배치되어 겹치지 않았다&#46; 세로 스크롤은 필요하지만 허용된다&#46; 전체계가 viewport에 맞게 축소되지 않는 별도 T4 실패와 snapshot 키보드 U5 실패를 이 레이아웃 통과로 면제하지 않는다&#46;

#### D2

상태: 통과. 평점: 5. 방법: 정적 검사.

Action: output 디렉터리 전체와 index&#46;html의 script/style/자산 경로를 정적으로 검사했다&#46; Expected: 로컬 참조가 output HTTP root 안에서 모두 해결되고 build/repository&#45;root 의존성이 없다&#46; Observed: output/index&#46;html 하나뿐이며 스타일/앱 코드는 인라인, 유일한 외부 실행 의존성은 허용된 https://unpkg&#46;com/three@0&#46;160&#46;0/build/three&#46;min&#46;js였다&#46; 추가 로컬 참조/누락 자산/빌드 요구가 없다&#46; file:// 호환성은 요청된 리뷰 범위에서 제외했다&#46;

#### D4

상태: 통과. 평점: 5. 방법: HTTP.

Action: 비어 있는 목록과 세 항목 목록을 실제 Export JSON 버튼으로 출력하고 실제 Import로 round&#45;trip했다&#46; 정지 및 재생 상태를 비교했다&#46; Expected: 명시적 textarea/UI와 정확한 필드/타입/순서/ID, 목록/live 불변, planet&#45;ID 매핑&#46; Observed: 빈 출력은 &#123;&#34;version&#34;:1,&#34;snapshots&#34;:&#91;&#93;&#125;였다&#46; populated JSON은 version/snapshots와 id/time/speed/selectedPlanet/viewMode만 포함하고 angles/playback/camera는 없었다&#46; obs&#45;1,obs&#45;3,obs&#45;2 및 공백/대소문자 ID의 순서와 numeric1&#46;5/null/string ID가 그대로였다&#46; Snapshot JSON accessible name과 IDs: ember, azure, jade가 보이며 OS picker 없이 작동했다&#46; 정지30초/4x/Azure/tracking의 import/export 전후 장면 RGB 차이는0, live 필드도 동일했다&#46; 재생 중 자연 시간 증가 외 상태 변경은 없었다&#46;

#### D5

상태: 통과. 평점: 5. 방법: HTTP.

Action: 비어 있지 않은 목록에서 추가 acceptance의11개 invalid 사례와 보완32개 matrix 사례를 각각 textarea와 Import 버튼으로 제출했다&#46; 이어 endpoint/17&#46;25초/1&#46;5x/공백 및 대소문자 ID/추가키를 포함한 유효 입력과 empty list, 재생 중 import를 수행했다&#46; Expected: 전체 입력 검증/원자성/정확한 타입·범위·ID 보존 및 list&#45;only 교체&#46; Observed: malformed JSON, root null/array, 루트·각 entry 필드 누락, entry null/array, version 문자열/boolean/unsupported, non&#45;array, numeric string/boolean, 음수·상한 초과 time/speed,1e400 두 필드, 빈/숫자 ID, unknown/숫자/빈/null문자 planet, unknown/boolean mode, tracking&#45;null,4개/중복 및 valid&#45;prefix&#43;invalid&#45;later가 각각 가시 오류와 전체 목록/live 불변을 보였다&#46; 유효0/120초,0&#46;25/4x,17&#46;25/1&#46;5, null overview/selected overview/tracking의3개 목록이 정확히 들어갔다&#46; ID &#34; A &#34;,&#34;A&#34;,&#34;a&#34;,&#34; &#34;를 trim/재생성하지 않았고 extra keys/angles999는 export에서 제외됐다&#46; fractional Restore와 재생 중 import 및 empty&#45;list clearing도 live 상태를 보존했다&#46; 파생 각도 산식 자체 오류는 S8이며 외부 angles가 권위를 갖지는 않았다&#46;

## 검증 및 한계

브라우저 검증 방식: HTTP.
뷰포트: 1440x900, 390x844.
검증 범위: 통과 17개, 실패 10개, 미검증 0개.

### 참고 사항

- 리뷰어 모델 openai&#45;codex/gpt&#45;6&#45;astra는 현재 호스트 메타데이터로 확인했다&#46; effort high는 현재 적용 &#46;omp/agents/benchmark&#45;astra&#45;review&#46;md thinking&#45;level의 명시적 설정에서 확인했고 별도 런타임 scalar는 제공되지 않았다&#46; 생성 모델/이전 리뷰 결과는 근거로 사용하지 않았다&#46; 실제 호스트 시계 ISO 시각으로 기록했다&#46;
- 사전 허용오차: 속도는 3초 이상 측정하고 max&#40;0&#46;15초, 기대 증가량 5%&#41;&#43;표시 한 증분; 정지 시간 한 표시 증분 및 각도 1도; 예측 각도 2도&#43;표시 반올림; 동일 뷰포트/카메라 반복 위치 2 CSS px; 3초 이상 추적 중심은 장면 짧은 변 10% 이내; JSON 타입/순서/ID 및 상태 보존은 정확 일치&#46;
- 사용 스킬: benchmark&#45;review, 호스트 필수 ego&#45;browser&#46; using&#45;harness&#45;flow는 읽었으나 서브에이전트 면제 규칙을 적용했다&#46; benchmark&#45;record 및 다른 모델/하위 에이전트는 사용하지 않았다&#46;
- 저장된 solar&#45;system&#45;observatory&#45;2 R1–R15 및 JSON 계약을 확인하여 format5&#46;0/solar&#45;system&#45;5&#46;0로 라우팅했다&#46; 소스 검사 전 metadata/허용오차를 확정하고 전부 미검증 체크포인트를 finalizer로 저장했다&#46; 기존 리뷰는 새로운 증거로 읽거나 재사용하지 않았다&#46;
- 원래 여덟 단계는 자동 재생→scene Ember 선택/추적→정지/30초 seek→정지2x→2x 재개→목록 Azure 전환→overview→정지/30초 재방문의 순서로 실행했다&#46; 시간/선택/추적은 전환됐고 각도 오류 및 overview 잘림이 관측됐다&#46; 추가 여섯 시퀀스는 snapshot A/반복복원, 용량/삭제/export, invalid import, valid/null/fractional/playing import, persistence/invalid storage/quota, 두 크기의 keyboard 순서로 실행한 후 나머지 matrix를 보완했다&#46;
- 번들 probe&#45;terra&#45;xhigh&#46;json은 이 산출물의 &#35;time/&#35;speed/&#35;pause 및 solar&#45;observations와 일치하지 않아 사용하지 않았다&#46; 실제 소스/DOM에 근거해 selector를 매핑하고 실제 Page input만 사용했다&#46; 상태 함수/Three 객체를 노출하거나 패치하지 않았다&#46; 위치 측정은 실제 일시 스크린샷의 색상 픽셀 중심과 고정 카메라 투영을 이용했다&#46;
- 오류 후 continuation: reviewer helper의 인자 누락은 이미 수행된 capture를 유지하고 다음 단계부터 계속했다&#46; range에 &#46;25 fill은 malformed value를 반환하면서 입력값만 변경하여 상태를 다시 확인한 뒤 정상 ArrowRight/Home 입력으로 완료했다&#46; 동적 snapshot DOM의 page&#46;click은 html intercept timeout을 내어 실제 사각형과 hit target을 확인한 뒤 정상 mouse&#46;click으로 성공했다&#46; 이 오류들을 앱의 import/복원 실패로 혼동하지 않았다&#46; 별도 라이브러리 설치 없이 Node 내장 기능으로 일시 PNG를 분석했다&#46;
- 27개 체크 모두 이 시도의 실제 근거로 판정했다&#46; S7은 screenshot transport 미확인이 아니라 실제 orbit 결함으로 fail이다&#46; 성공하는 orbit의 깊이/가림 시점 변화와 snapshot 버튼의 정상 keyboard restore/delete는 확정 기능 결함 때문에 관측 불가능했으며 관련 체크를 pass 처리하지 않았다&#46; 다른 실패가 독립 시나리오를 중단시키지 않았다&#46;
- viewport는 각 관련 ego&#45;browser nodejs invocation에서 Emulation metrics를 적용하고 실제 innerWidth/innerHeight를 읽었다&#46; desktop1440×900/mobile390×844를 확인했고 측정 중 대상 scene은 parent&#45;owned TaskSpace5/p1 하나뿐이었다&#46; 이 결과는 CSS viewport 검증이며 실제 터치 장치나 file:// 실행을 인증하지 않는다&#46;
- app&#45;owned solar&#45;observations의 처음 값은 null이었다&#46; malformed/schema&#45;invalid와 quota fixture는 이 key와 사전 부재 확인된 terra&#45;low&#45;quota&#45;1792683672115만 건드렸다&#46; unrelated storage/cookies/cache는 읽거나 정리하지 않았다&#46; 최종 두 key가 null임을 확인하고 emulation 해제, p1 about:blank 이동 및8791/8792 서버 종료를 완료했다&#46; parent&#45;owned page/space를 닫거나 finish하지 않고 부모에게 슬롯을 반환했다&#46;
- prompt&#46;md,run&#46;json,output/index&#46;html의 사전/사후 working&#45;memory SHA&#45;256이 모두 동일했다&#46; output 전체 파일은 index&#46;html 하나였고 원본은 수정하지 않았다&#46; 일시 browser script/PNG 67개를 삭제했다&#46; 작업 draft도 finalizer 출판 후 제거한다&#46; 저장 결과는 reviews/review&#46;json 및 review&#46;md 두 파일뿐이며 원본 생성 기록/다른 run은 변경하지 않는다&#46;
- 상위 조정 작업에는 호스트가 필수 주입한 using&#45;harness&#45;flow 및 caveman을 적용했습니다&#46; 브라우저 검증에는 호스트 필수 ego&#45;browser를 사용했습니다&#46; 이전 요청의 캡처 장애에 사용한 systematic&#45;debugging은 이번 Terra 재리뷰에 자동 적용하지 않았습니다&#46; 두 리뷰 후 상위 작업이 공유 TaskSpace5에 finish&#40;&#123;keep:&#91;&#93;&#125;&#41;를 실행하고 finished:true를 직접 확인했습니다&#46;
