package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wonjinsin/ai-toybox/whisper/internal/port"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
)

type batchRunnerFunc func(context.Context, portin.Request) (portin.BatchResult, error)

func (runner batchRunnerFunc) Run(ctx context.Context, request portin.Request) (portin.BatchResult, error) {
	return runner(ctx, request)
}

func TestExecuteReportsDirectoryResults(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	factory := func(observer portin.ProgressObserver) portin.BatchRunner {
		return batchRunnerFunc(func(_ context.Context, request portin.Request) (portin.BatchResult, error) {
			if request.InputPath != "recordings" || request.Parallel != 2 || request.Format != "srt" {
				t.Fatalf("Run() request = %#v, want parsed CLI options", request)
			}
			observer(portin.ProgressEvent{InputPath: "recordings/a.mp3", Kind: portin.EventComplete})
			return portin.BatchResult{Directory: true, Results: []portin.Result{
				{InputPath: "recordings/a.mp3", OutputPath: "recordings/a.srt"},
				{InputPath: "recordings/b.mp3", Err: port.NewOutputExistsError("recordings/b.srt")},
				{InputPath: "recordings/c.mp3", Err: errors.New("bad media")},
			}}, nil
		})
	}
	exitCode := Execute(context.Background(), []string{"-format", "srt", "-parallel", "2", "recordings"}, &stdout, &stderr, factory)
	if exitCode != 1 {
		t.Fatalf("Execute() exit code = %d, want partial-failure exit 1", exitCode)
	}
	want := "완료: recordings/a.srt\n건너뜀: recordings/b.mp3\n요약: 성공 1개, 건너뜀 1개, 실패 1개\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
	if got, want := stderr.String(), "[a.mp3] 처리 완료\n실패: recordings/c.mp3: bad media\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestExecuteShowsUsageAndHelpWithoutStartingService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		args []string
		code int
	}{
		{name: "missing input", code: 2},
		{name: "help", args: []string{"-help"}, code: 0},
		{name: "unknown flag", args: []string{"-unknown"}, code: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			factory := func(portin.ProgressObserver) portin.BatchRunner {
				t.Fatal("service was constructed while showing usage")
				return nil
			}
			if got := Execute(context.Background(), test.args, &stdout, &stderr, factory); got != test.code {
				t.Fatalf("Execute() exit code = %d, want %d", got, test.code)
			}
			output := stderr.String()
			if test.code == 0 {
				output = stdout.String()
				if stderr.Len() != 0 {
					t.Errorf("stderr = %q, want empty help errors", stderr.String())
				}
			}
			if !strings.Contains(output, "사용법:") || !strings.Contains(output, "-vad-model") || !strings.Contains(output, "-corrections") {
				t.Errorf("output = %q, want complete usage", output)
			}
		})
	}
}

func TestExecuteReportsDiscoveryFailure(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	factory := func(portin.ProgressObserver) portin.BatchRunner {
		return batchRunnerFunc(func(context.Context, portin.Request) (portin.BatchResult, error) {
			return portin.BatchResult{}, errors.New("input not found")
		})
	}
	if got := Execute(context.Background(), []string{"missing.mp4"}, &stdout, &stderr, factory); got != 1 {
		t.Errorf("Execute() exit code = %d, want 1", got)
	}
	if got, want := stderr.String(), "전사 실패: input not found\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want no completion", stdout.String())
	}
}

func TestExecuteOmitsDirectorySummaryForSingleFile(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	factory := func(portin.ProgressObserver) portin.BatchRunner {
		return batchRunnerFunc(func(context.Context, portin.Request) (portin.BatchResult, error) {
			return portin.BatchResult{Results: []portin.Result{{InputPath: "meeting.mp3", OutputPath: "meeting.txt"}}}, nil
		})
	}
	if got := Execute(context.Background(), []string{"meeting.mp3"}, &stdout, &stderr, factory); got != 0 {
		t.Errorf("Execute() exit code = %d, want 0", got)
	}
	if got, want := stdout.String(), "완료: meeting.txt\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}
