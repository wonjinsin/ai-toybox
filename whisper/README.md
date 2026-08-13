# 로컬 Whisper 명령어

`whisper.cpp`의 다국어 `large-v3` 모델을 사용해 음성·영상 파일을 로컬에서 전사합니다. OpenAI API 키와 API 비용이 필요하지 않으며, 파일을 외부 서버로 전송하지 않습니다.

## 빠른 사용법

```bash
# 언어 자동 감지 후 텍스트 생성
whisper-local "녹음.m4a"

# 일본어 음성을 텍스트로 변환
whisper-local -language ja "인터뷰.mp3"

# 중국어 영상에서 SRT 자막 생성
whisper-local -language zh -format srt "영상.mp4"

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

## 폴더 일괄 처리

입력 경로로 폴더를 주면 폴더 바로 아래의 지원 미디어 파일을 이름순으로 모두 처리합니다. 하위 폴더는 재귀적으로 탐색하지 않습니다.

`-parallel`은 동시에 실행할 최대 전사 수입니다. 한 파일이 끝나면 다음 파일을 바로 시작하며, 폴더의 모든 파일이 성공하거나 실패할 때까지 계속합니다. 파일 하나가 실패해도 나머지는 처리하고 마지막에 성공·실패 개수를 출력합니다. 실패가 하나라도 있으면 종료 코드는 `1`입니다.

기본 병렬도는 `1`입니다. 각 `whisper-cli` 프로세스가 모델을 별도로 로드하므로 메모리와 GPU 사용량이 늘어납니다. 이 장비에서는 `-parallel 2`부터 시작해 처리 속도와 메모리 사용량을 확인하는 것을 권장합니다.

같은 폴더에 `meeting.mp3`와 `meeting.wav`처럼 결과 이름이 같은 파일이 있으면 둘 다 `meeting.txt`를 만들게 됩니다. 이런 충돌은 작업 시작 전에 오류로 차단합니다. 대소문자만 다른 결과 이름도 macOS 파일시스템에서 안전하도록 충돌로 처리합니다.

## 현재 설치 구성

- `whisper.cpp` 1.9.1: Homebrew로 설치
- 다국어 `large-v3` 모델: `models/ggml-large-v3.bin`
- 모델 SHA-256: `64d182b440b98d5203c4f9bd541544d84c605196c4f7b845dfa11fb23594d1e2`
- silero VAD 모델: `models/ggml-silero-v5.1.2.bin` (음성 구간만 전사해 무음·음악 구간의 환각 반복 방지)
- VAD 모델 SHA-256: `29940d98d42b91fbd05ce489f3ecf7c72f0a42f027e4875919a28fb4c04ea2cf`
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
  -o models/ggml-silero-v5.1.2.bin \
  https://huggingface.co/ggml-org/whisper-vad/resolve/main/ggml-silero-v5.1.2.bin

printf '%s  %s\n' \
  '29940d98d42b91fbd05ce489f3ecf7c72f0a42f027e4875919a28fb4c04ea2cf' \
  'models/ggml-silero-v5.1.2.bin' | shasum -a 256 -c -

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

VAD 모델(`ggml-silero-v5.1.2.bin`)은 확정된 Whisper 모델과 같은 디렉터리에서 찾으며, 없으면 VAD 없이 전사합니다.

## 개발 검증

```bash
go test -race -cover ./internal/app
go test ./...
go vet ./...
```

테스트는 실제 셸 명령 문자열을 조합하지 않고 `ffmpeg`와 `whisper-cli`에 인자를 분리해 전달하는지 검증합니다. 공백과 한글이 포함된 파일 경로도 지원합니다.
