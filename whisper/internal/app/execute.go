package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `사용법:
  whisper-local [옵션] <음성·영상 파일 또는 폴더>

옵션:
  -language auto|ko|ja|zh|en  음성 언어 (기본값: auto)
  -format txt|srt|vtt         출력 형식 (기본값: txt)
  -model <경로>               Whisper 모델 경로
  -vad-model <경로>           Silero VAD 모델 경로
  -corrections <경로>         문구 교정 JSON 경로
  -output <디렉터리>          출력 디렉터리 (기본값: 입력 파일 위치)
  -parallel <개수>             최대 동시 전사 수 (기본값: 1)
  -force                      기존 결과 파일 덮어쓰기
`

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	options, err := ParseOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return 0
		}
		fmt.Fprintf(stderr, "오류: %v\n\n%s", err, usage)
		return 2
	}

	inputInfo, statErr := os.Stat(options.InputPath)
	inputIsDirectory := statErr == nil && inputInfo.IsDir()

	results, err := transcribeAll(ctx, options, Transcribe)
	if err != nil {
		fmt.Fprintf(stderr, "전사 실패: %v\n", err)
		return 1
	}

	failures := 0
	for _, result := range results {
		if result.Err != nil {
			failures++
			fmt.Fprintf(stderr, "실패: %s: %v\n", result.InputPath, result.Err)
			continue
		}
		fmt.Fprintf(stdout, "완료: %s\n", result.OutputPath)
	}
	if inputIsDirectory {
		fmt.Fprintf(stdout, "요약: 성공 %d개, 실패 %d개\n", len(results)-failures, failures)
	}
	if failures > 0 {
		return 1
	}
	return 0
}
