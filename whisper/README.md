# 로컬 Whisper 명령어

`whisper.cpp`의 다국어 `large-v3`와 Silero VAD를 사용해 음성·영상 파일을 로컬에서 전사합니다. OpenAI API 키와 API 비용이 필요하지 않으며, 파일을 외부 서버로 전송하지 않습니다. 성능 모드는 나누지 않고 항상 최고 품질 파이프라인을 사용합니다.

## 빠른 사용법

```bash
# 언어 자동 감지 후 텍스트 생성
whisper-local "녹음.m4a"

# 일본어 음성을 텍스트로 변환
whisper-local -language ja "인터뷰.mp3"

# 중국어 영상에서 SRT 자막 생성
whisper-local -language zh -format srt "영상.mp4"

# 일본어 고유명사 교정 사전을 적용해 SRT 생성
whisper-local -language ja -format srt -corrections ./corrections-ja.json "영상.mp4"

# 별도 디렉터리에 VTT 자막 생성
whisper-local -format vtt -output ./자막 "회의.mov"

# 폴더 안의 모든 미디어 파일을 최대 2개씩 병렬 처리
whisper-local -parallel 2 ./recordings
```

결과 파일은 기본적으로 입력 파일 옆에 생성됩니다. 기존 결과 파일은 보호되며, 의도적으로 덮어쓰려면 `-force`를 사용합니다.

```bash
whisper-local -force -language ko "회의.mp4"
```

전체 옵션은 다음 명령으로 확인합니다.

```bash
whisper-local -help
```

## 지원 범위

- 언어: 자동 감지, 한국어(`ko`), 일본어(`ja`), 중국어(`zh`), 영어(`en`)
- 출력: 일반 텍스트(`txt`), SRT 자막(`srt`), WebVTT 자막(`vtt`)
- 단일 파일 입력: FFmpeg가 읽을 수 있는 음성·영상 파일
- 폴더 입력 확장자: `3g2`, `3gp`, `aac`, `ac3`, `aif`, `aiff`, `alac`, `amr`, `ape`, `au`, `avi`, `caf`, `dts`, `flac`, `m2ts`, `m4a`, `m4v`, `mka`, `mkv`, `mov`, `mp3`, `mp4`, `mpeg`, `mpg`, `mts`, `mxf`, `oga`, `ogg`, `ogv`, `opus`, `ts`, `vob`, `wav`, `webm`, `wma`, `wmv`
- 영상: 영상 속 음성 트랙을 전사하며 화면 내용은 분석하지 않음
- 음성 전처리: 16kHz 모노 PCM 변환과 동적 음량 정규화
- 발화 탐지: Silero VAD 필수. VAD 모델이 없으면 명확한 오류로 중단
- 자막 시간: 원본 타임라인 기준 최대 20초 조각으로 전사해 긴 무음 후에도 시간 밀림 방지
- 자막 정리: 토큰 신뢰도, 근접 반복, 일본어 비언어 발성 필터링
- 가독성: 자막당 최대 5.5초, 비겹침, 언어별 최대 줄 길이 적용

## 폴더 일괄 처리

입력 경로로 폴더를 주면 폴더 바로 아래의 지원 미디어 파일을 이름순으로 모두 처리합니다. 하위 폴더는 재귀적으로 탐색하지 않습니다.

`-parallel`은 동시에 실행할 최대 전사 수입니다. 한 파일이 끝나면 다음 파일을 바로 시작하며, 폴더의 모든 파일이 성공하거나 실패할 때까지 계속합니다. 파일 하나가 실패해도 나머지는 처리하고 마지막에 성공·실패 개수를 출력합니다. 실패가 하나라도 있으면 종료 코드는 `1`입니다.

기본 병렬도는 `1`입니다. 각 `whisper-cli` 프로세스가 모델을 별도로 로드하므로 메모리와 GPU 사용량이 늘어납니다. 이 장비에서는 `-parallel 2`부터 시작해 처리 속도와 메모리 사용량을 확인하는 것을 권장합니다.

같은 폴더에 `meeting.mp3`와 `meeting.wav`처럼 결과 이름이 같은 파일이 있으면 둘 다 `meeting.txt`를 만들게 됩니다. 이런 충돌은 작업 시작 전에 오류로 차단합니다. 대소문자만 다른 결과 이름도 macOS 파일시스템에서 안전하도록 충돌로 처리합니다.

## 고품질 자막 파이프라인

파일당 다음 순서를 사용합니다.

1. FFmpeg로 전체 오디오를 정규화한 16kHz 모노 WAV로 변환합니다.
2. Silero VAD로 원본 타임라인의 발화 구간을 찾습니다.
3. 경계 여유 150ms를 포함한 최대 20초 WAV 조각을 만듭니다.
4. `large-v3`를 한 번 로드한 `whisper-cli` 다중 파일 실행으로 전체 JSON을 생성합니다.
5. 각 JSON 시간에 원본 조각 시작 시간을 더해 절대 시간을 복원합니다.
6. 신뢰도·반복·비언어 필터와 가독성 규칙을 적용한 뒤 TXT, SRT, VTT를 렌더링합니다.
7. 모든 자막이 입력 미디어 길이 안에 있고 서로 겹치지 않는지 검증합니다.

빠른 모드, 균형 모드, 양자화 모델 자동 선택은 제공하지 않습니다.

## 교정 사전

`-corrections`는 인식 후 문구를 치환할 UTF-8 JSON 객체를 받습니다. 긴 원문을 먼저 적용합니다.

```json
{
  "レッサン": "レッスン",
  "FC2 PPV": "FC2PPV"
}
```

교정 사전은 선택 사항입니다. 잘못된 JSON이나 빈 원문 키가 있으면 실행을 중단합니다.

## 현재 설치 구성

- `whisper.cpp` 1.9.1: Homebrew로 설치
- 다국어 `large-v3` 모델: `models/ggml-large-v3.bin`
- 모델 SHA-256: `64d182b440b98d5203c4f9bd541544d84c605196c4f7b845dfa11fb23594d1e2`
- silero VAD 모델: `models/ggml-silero-v6.2.0.bin` (작은 발화 탐지를 개선하고 무음·음악 구간의 환각 반복 방지)
- VAD 모델 SHA-256: `2aa269b785eeb53a82983a20501ddf7c1d9c48e33ab63a41391ac6c9f7fb6987`
- Go 실행 파일: `bin/whisper-local`
- 전역 명령 링크: `/opt/homebrew/bin/whisper-local`

모델은 용량이 약 3.1GB이므로 Git에 포함되지 않습니다.

## 설치 요구 사항

- Apple Silicon 기반 macOS
- Homebrew
- FFmpeg와 `whisper.cpp`
- 소스에서 Go 명령을 다시 빌드할 경우 Go 1.25 이상

## 다시 설치하기

```bash
brew install ffmpeg whisper-cpp

mkdir -p models bin
curl -L \
  -o models/ggml-large-v3.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin

printf '%s  %s\n' \
  '64d182b440b98d5203c4f9bd541544d84c605196c4f7b845dfa11fb23594d1e2' \
  'models/ggml-large-v3.bin' | shasum -a 256 -c -

curl -L \
  -o models/ggml-silero-v6.2.0.bin \
  https://huggingface.co/ggml-org/whisper-vad/resolve/main/ggml-silero-v6.2.0.bin

printf '%s  %s\n' \
  '2aa269b785eeb53a82983a20501ddf7c1d9c48e33ab63a41391ac6c9f7fb6987' \
  'models/ggml-silero-v6.2.0.bin' | shasum -a 256 -c -

go build -o bin/whisper-local ./cmd/whisper-local
ln -s "$(pwd)/bin/whisper-local" /opt/homebrew/bin/whisper-local
```

이미 전역 링크가 있으면 기존 링크가 올바른 프로젝트 실행 파일을 가리키는지 먼저 확인합니다.

```bash
ls -l /opt/homebrew/bin/whisper-local
```

## 모델 검색 순서

명령은 다음 순서로 모델을 찾습니다.

1. `-model` 옵션 경로
2. `WHISPER_MODEL` 환경 변수
3. 현재 디렉터리의 `models/ggml-large-v3.bin`
4. 실행 파일 기준 `../models/ggml-large-v3.bin`
5. 사용자 캐시 디렉터리의 `whisper-local/models/ggml-large-v3.bin`

VAD 모델은 다음 순서로 찾습니다.

1. `-vad-model` 옵션 경로
2. `WHISPER_VAD_MODEL` 환경 변수
3. 확정된 Whisper 모델과 같은 디렉터리의 `ggml-silero-v6.2.0.bin`

VAD 모델은 필수입니다. 찾지 못하면 시간 복원 품질을 낮춘 대체 경로로 진행하지 않고 실행을 중단합니다. 구형 v5.1.2 모델은 사용하지 않습니다.

## 개발 검증

```bash
go test -race -cover ./internal/app
go test ./...
go vet ./...
```

테스트는 실제 셸 명령 문자열을 조합하지 않고 `ffmpeg`와 `whisper-cli`에 인자를 분리해 전달하는지 검증합니다. 공백과 한글이 포함된 파일 경로도 지원합니다.
