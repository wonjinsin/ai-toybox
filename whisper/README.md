# 로컬 Whisper 명령어

`whisper.cpp`의 다국어 `small` 모델을 사용해 음성·영상 파일을 로컬에서 전사합니다. OpenAI API 키와 API 비용이 필요하지 않으며, 파일을 외부 서버로 전송하지 않습니다.

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
- 입력: FFmpeg가 읽을 수 있는 MP3, M4A, WAV, MP4, MOV, MKV 등
- 영상: 영상 속 음성 트랙을 전사하며 화면 내용은 분석하지 않음

## 현재 설치 구성

- `whisper.cpp` 1.9.1: Homebrew로 설치
- 다국어 `small` 모델: `models/ggml-small.bin`
- 모델 SHA-256: `1be3a9b2063867b937e64e2ec7483364a79917e157fa98c5d94b5c1fffea987b`
- Go 실행 파일: `bin/whisper-local`
- 전역 명령 링크: `/opt/homebrew/bin/whisper-local`

모델은 용량이 약 488MB이므로 Git에 포함되지 않습니다.

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
  -o models/ggml-small.bin \
  https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin

printf '%s  %s\n' \
  '1be3a9b2063867b937e64e2ec7483364a79917e157fa98c5d94b5c1fffea987b' \
  'models/ggml-small.bin' | shasum -a 256 -c -

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
3. 현재 디렉터리의 `models/ggml-small.bin`
4. 실행 파일 기준 `../models/ggml-small.bin`
5. 사용자 캐시 디렉터리의 `whisper-local/models/ggml-small.bin`

## 개발 검증

```bash
go test -race -cover ./internal/app
go test ./...
go vet ./...
```

테스트는 실제 셸 명령 문자열을 조합하지 않고 `ffmpeg`와 `whisper-cli`에 인자를 분리해 전달하는지 검증합니다. 공백과 한글이 포함된 파일 경로도 지원합니다.
