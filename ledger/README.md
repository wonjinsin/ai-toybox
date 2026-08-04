# ledger

AI 기반 개인 가계부 CLI. 은행/카드 거래내역 CSV를 형식에 상관없이 임포트하고, 자연어로 지출을 분석한다.

AI는 API가 아니라 로컬 CLI(`claude -p` 또는 `codex exec`) 서브프로세스로 호출하므로 추가 과금이 없다. 데이터는 SQLite 파일 하나에 저장된다.

## 요구사항

- Go 1.25+
- `claude` CLI (기본) 또는 `codex` CLI — PATH에 있어야 함

## 설치

```bash
go build -o ledger ./cmd/ledger
```

## 사용법

### 1. 소스(계좌/카드) 등록

```bash
./ledger sources add 신한체크 --kind card    # kind: bank | card
./ledger sources list
```

### 2. CSV 임포트

```bash
./ledger import 내역.csv --source 신한체크
```

- CSV 형식이 처음이면 AI가 컬럼 매핑을 파악한다 (같은 형식은 캐시돼 다음부터 AI 호출 없음).
- EUC-KR 인코딩(은행 CSV에 흔함) 자동 감지·변환.
- 승인취소, 내부이체 추정, 카드대금 출금(카드 내역과 이중 계산 방지), 0원 거래 등 애매한 행은 유형별로 묶어 물어본다:
  - `[i]` 포함 / `[s]` 스킵 / `[I]` 항상 포함 / `[S]` 항상 스킵
  - "항상"을 고르면 규칙으로 저장돼 다음 임포트부터 자동 적용된다.
- `--yes`(`-y`): 질문 전부 생략하고 모두 포함 (스크립트용).
- 같은 파일을 다시 임포트해도 중복 저장되지 않는다.
- 임포트 후 새 가맹점은 AI가 카테고리(식비, 카페/간식, 교통 등)를 자동 분류한다. 분류 결과는 캐시되어 재사용된다.

### 3. 자연어 질의

```bash
./ledger ask "지난달 카테고리별 지출 합계 알려줘"
./ledger ask "이번달 제일 많이 쓴 가맹점 톱5는?"
./ledger ask "커피에 쓴 돈 월별 추세 보여줘"
```

AI가 질문을 SQL로 변환해 실행하고(읽기 전용 — 데이터 변경 불가), 결과를 한국어로 분석해 답한다. 질문 1회당 AI 호출 2회로 수 초~십수 초 걸린다.

### 4. 임포트 규칙 관리

```bash
./ledger rules list          # "항상" 결정으로 생성된 규칙 목록
./ledger rules delete <id>
```

### 5. 거래 조회/삭제

이미 저장된 거래를 확인하고 지울 수 있다. 예: 통장 내역에 포함된 카드대금 출금이 카드 내역과 이중 계산될 때.

```bash
./ledger tx list --match 카드대금   # 가맹점/메모에 해당 문자열이 포함된 거래 검색
./ledger tx delete <id> [<id>...]
```

같은 CSV를 다시 임포트하면 삭제한 행이 되살아나므로, 반복 임포트한다면 임포트 시 `[S] 항상 스킵`으로 규칙을 만드는 편이 안전하다.

## 공통 옵션

| 옵션 | 설명 | 기본값 |
|------|------|--------|
| `--db <path>` | SQLite 파일 경로 | `~/.ledger/ledger.db` |
| `--ai claude\|codex` | AI 백엔드 선택 | `claude` |

## 구조

go-boileplate-web 스타일의 헥사고날 아키텍처:

- `internal/core/` — 도메인, 포트, 서비스 (임포트/질의 파이프라인)
- `internal/adapter/in/cli/` — cobra 커맨드, 터미널 프롬프트
- `internal/adapter/out/persistence/` — ent + SQLite (modernc, CGO 불필요)
- `internal/adapter/out/ai/` — claude/codex 서브프로세스 러너
