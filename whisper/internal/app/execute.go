package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

const usage = `사용법:
  whisper-local [옵션] <음성 또는 영상 파일>

옵션:
  -language auto|ko|ja|zh|en  음성 언어 (기본값: auto)
  -format txt|srt|vtt         출력 형식 (기본값: txt)
  -model <경로>               Whisper 모델 경로
  -output <디렉터리>          출력 디렉터리 (기본값: 입력 파일 위치)
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

	outputPath, err := Transcribe(ctx, options)
	if err != nil {
		fmt.Fprintf(stderr, "전사 실패: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "완료: %s\n", outputPath)
	return 0
}
