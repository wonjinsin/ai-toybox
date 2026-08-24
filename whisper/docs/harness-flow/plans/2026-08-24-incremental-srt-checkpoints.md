# 점진적 SRT와 전사 재개 계획

Spec: docs/harness-flow/specs/2026-08-24-incremental-srt-checkpoints.md
Goal: 긴 SRT 전사 중 확인 가능한 partial 자막을 만들고, 중단 후 마지막 완료 배치부터 안전하게 이어간다.
Constraints: 점진적 저장은 SRT에만 적용한다. 배치 크기는 32개로 고정한다. 기존 TXT·VTT, 출력 덮어쓰기 보호, 저신뢰 선택 정확도를 유지한다. 모든 영속 파일은 같은 출력 디렉터리에서 원자적으로 교체한다. 테스트를 먼저 작성하고 `internal/app` 커버리지를 80% 이상 유지한다.

### Task 1: 첫 전사 배치부터 복구 가능한 partial SRT 제공

Delivers: SRT 전사가 두 번째 이후 배치에서 실패해도 첫 32개 조각의 partial SRT와 유효한 체크포인트가 남는다.
Touches: `internal/app/checkpoint.go`, `internal/app/checkpoint_test.go`, `internal/app/pipeline.go`, `internal/app/transcribe.go`, `internal/app/transcribe_test.go`
Blocked by: none
- [x] 33개 이상의 조각에서 배치 호출이 32개로 제한되고 첫 배치 실패 이후가 아니라 첫 배치 성공 직후 partial SRT가 저장됨을 실패하는 테스트로 고정한다.
- [x] 체크포인트 스키마가 원시 cue의 토큰·확률·원본 위치와 음성 조각 경계를 손실 없이 왕복함을 검증한다.
- [x] 배치 슬라이스와 재개 이후에도 각 cue가 전체 조각 목록의 전역 origin 인덱스를 유지한다.
- [x] partial SRT와 체크포인트를 같은 디렉터리의 임시 파일에서 원자적으로 교체하고 불완전한 쓰기 파일을 남기지 않는다.
- [x] TXT와 VTT가 기존 단일 호출 경로를 유지한다.

### Task 2: 유효한 체크포인트에서 완료 배치 재개

Delivers: 같은 SRT 명령을 다시 실행하면 VAD와 완료된 Whisper 배치를 건너뛰고 남은 조각만 처리해 최종 SRT를 만든다.
Touches: `internal/app/checkpoint.go`, `internal/app/checkpoint_test.go`, `internal/app/transcribe.go`, `internal/app/transcribe_test.go`, `internal/app/progress.go`
Blocked by: Task 1
- [x] 첫 실행을 다음 배치에서 중단한 뒤 두 번째 실행이 저장된 음성 조각 경계와 cue를 복원하고 남은 배치만 호출함을 실패하는 통합 테스트로 고정한다.
- [x] 입력·모델·도구·언어·파이프라인 식별 정보가 다르면 새로 시작하고, 손상된 체크포인트는 보존한 채 경로가 포함된 오류로 중단한다.
- [x] 재개 시 완료 개수와 전체 개수를 로그로 알리고 현재 교정 사전으로 partial SRT를 다시 렌더링한다.

### Task 3: 저신뢰 재시도 재개와 최종 정리

Delivers: 저신뢰 재시도 중 중단되어도 완료된 선택을 보존하며, 최종 SRT 성공 후 partial SRT와 체크포인트가 사라진다.
Touches: `internal/app/checkpoint.go`, `internal/app/retry.go`, `internal/app/retry_test.go`, `internal/app/transcribe.go`, `internal/app/transcribe_test.go`, `README.md`
Blocked by: Task 2
- [x] 저신뢰 cue 처리 후 커서와 선택 결과를 저장하고 재실행이 완료된 cue를 재시도하지 않음을 실패하는 테스트로 고정한다.
- [x] 최종 출력 저장 성공 뒤 증분 파일을 정리하고, 정리 오류는 최종 성공을 뒤집지 않고 경고로 표시한다.
- [x] Ctrl-C, 배치 오류, 저장 오류에서는 마지막 유효한 partial SRT와 체크포인트를 보존한다.
- [x] README에 partial 파일명, 자동 재개 조건, 임시 결과라는 점을 한국어로 설명한다.
- [x] `gofmt`, `go vet ./...`, `go test ./...`, `go test -race -cover ./internal/app`, `make build`가 통과한다.
