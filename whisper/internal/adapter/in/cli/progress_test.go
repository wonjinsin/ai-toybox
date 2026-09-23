package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
)

type messageWriter chan string

func (writer messageWriter) Write(content []byte) (int, error) {
	writer <- string(content)
	return len(content), nil
}

func readProgressMessage(t *testing.T, messages <-chan string) string {
	t.Helper()
	select {
	case message := <-messages:
		return message
	case <-time.After(time.Second):
		t.Fatal("progress message was not reported")
		return ""
	}
}

func waitForProgressMessage(t *testing.T, messages <-chan string, expected string) {
	t.Helper()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case message := <-messages:
			if strings.Contains(message, expected) {
				return
			}
		case <-timer.C:
			t.Fatalf("progress did not report %q", expected)
		}
	}
}

func TestProgressReportsStageAndOverallElapsed(t *testing.T) {
	t.Parallel()

	messages := make(messageWriter, 64)
	var elapsed atomic.Int64
	progress := newProgressObserver(messages, 5*time.Millisecond, func() time.Time {
		return time.Unix(0, elapsed.Load())
	})
	t.Cleanup(progress.Close)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepNormalize})
	if message := readProgressMessage(t, messages); message != "[meeting.mp4] 오디오 추출 중...\n" {
		t.Fatalf("initial progress = %q, want audio stage message", message)
	}
	elapsed.Store(int64(3 * time.Minute))
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepDetect})
	waitForProgressMessage(t, messages, "[meeting.mp4] 음성 구간 탐지 중...\n")
	elapsed.Store(int64(3*time.Minute + 2*time.Second))
	waitForProgressMessage(t, messages, "음성 구간 탐지 중... (이 단계 경과 2s / 전체 3m2s)")
}

func TestProgressHeartbeatReportsLatestRecognitionCount(t *testing.T) {
	t.Parallel()

	messages := make(messageWriter, 64)
	progress := newProgressObserver(messages, 5*time.Millisecond, time.Now)
	t.Cleanup(progress.Close)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Format: "srt", Kind: portin.EventStage,
		Stage: portin.StepRecognize, Count: 84, Total: 84})
	if message := readProgressMessage(t, messages); message != "[meeting.mp4] 전사 중: 84개 (완료 0/84개)\n" {
		t.Fatalf("initial progress = %q, want initial transcription count", message)
	}
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Format: "srt", Kind: portin.EventUpdate,
		Stage: portin.StepRecognize, Count: 84, Completed: 32, Total: 84})
	waitForProgressMessage(t, messages, "전사 중: 84개 (완료 32/84개)")
}

func TestProgressPreservesLegacyMessages(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		event portin.ProgressEvent
		want  string
	}{
		{event: portin.ProgressEvent{Kind: portin.EventStage, Stage: portin.StepExtract, Count: 5}, want: "음성 조각 생성: 5개"},
		{event: portin.ProgressEvent{Kind: portin.EventStage, Stage: portin.StepRecognize, Format: "txt", Count: 5}, want: "전사 중: 5개"},
		{event: portin.ProgressEvent{Kind: portin.EventStage, Stage: portin.StepRecognize, Format: "vtt", Count: 5}, want: "전사 중: 5개"},
		{event: portin.ProgressEvent{Kind: portin.EventStage, Stage: portin.StepRetry, Count: 2}, want: "저신뢰 구간 재시도: 2개"},
		{event: portin.ProgressEvent{Kind: portin.EventStage, Stage: portin.StepClean}, want: "자막 정리 및 검증 중..."},
		{event: portin.ProgressEvent{Kind: portin.EventStage, Stage: portin.StepPublish}, want: "결과 저장 중..."},
		{event: portin.ProgressEvent{Kind: portin.EventResumed, Completed: 4, Total: 5}, want: "체크포인트 재개: 4/5개"},
		{event: portin.ProgressEvent{Kind: portin.EventSaved, Stage: portin.StepRecognize, Completed: 4, Total: 5}, want: "전사 진행: 4/5개 (partial SRT 저장됨)"},
		{event: portin.ProgressEvent{Kind: portin.EventSaved, Stage: portin.StepRetry, Completed: 4, Total: 5}, want: "저신뢰 재시도 진행: 4/5 cue"},
		{event: portin.ProgressEvent{Kind: portin.EventComplete}, want: "처리 완료"},
		{event: portin.ProgressEvent{Kind: portin.EventFailed, Err: errors.New("bad media")}, want: "처리 실패: bad media"},
		{event: portin.ProgressEvent{Kind: portin.EventWarning, Stage: portin.StepPublish, Err: errors.New("permission denied")}, want: "경고: 증분 파일 정리 실패: permission denied"},
		{event: portin.ProgressEvent{Kind: portin.EventWarning, Err: errors.New("cleanup failed")}, want: "경고: cleanup failed"},
	} {
		t.Run(test.want, func(t *testing.T) {
			var output bytes.Buffer
			progress := newProgressObserver(&output, 0, time.Now)
			defer progress.Close()
			event := test.event
			event.InputPath = "/recordings/meeting.mp4"
			progress.Observe(event)
			if got, want := output.String(), "[meeting.mp4] "+test.want+"\n"; got != want {
				t.Errorf("progress = %q, want %q", got, want)
			}
		})
	}
}

func TestProgressIncludesPreparationInOverallElapsed(t *testing.T) {
	t.Parallel()

	messages := make(messageWriter, 64)
	progress := newProgressObserver(messages, 5*time.Millisecond, func() time.Time { return time.Unix(182, 0) })
	t.Cleanup(progress.Close)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStarted, StartedAt: time.Unix(0, 0)})
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepNormalize})
	readProgressMessage(t, messages)
	if message := readProgressMessage(t, messages); !strings.Contains(message, "이 단계 경과 0s / 전체 3m2s") {
		t.Errorf("heartbeat = %q, want overall elapsed including preparation", message)
	}
}

func TestProgressStopsHeartbeatForFinishedFiles(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{portin.EventComplete, portin.EventFailed} {
		t.Run(kind, func(t *testing.T) {
			messages := make(messageWriter, 64)
			progress := newProgressObserver(messages, 5*time.Millisecond, time.Now)
			t.Cleanup(progress.Close)
			progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepNormalize})
			readProgressMessage(t, messages)
			progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: kind, Err: errors.New("bad media")})
			waitForProgressMessage(t, messages, "처리 ")
			select {
			case message := <-messages:
				t.Errorf("unexpected progress after terminal event: %q", message)
			case <-time.After(20 * time.Millisecond):
			}
		})
	}
}

func TestProgressCloseStopsWritesAndIsIdempotent(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	progress := newProgressObserver(&output, 5*time.Millisecond, time.Now)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepNormalize})
	progress.Close()
	before := output.String()
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventComplete})
	progress.Close()
	<-time.After(20 * time.Millisecond)
	if got := output.String(); got != before {
		t.Errorf("progress after Close() = %q, want unchanged %q", got, before)
	}
}

func TestProgressSerializesConcurrentFileMessages(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	progress := newProgressObserver(&output, 0, time.Now)
	defer progress.Close()
	var workers sync.WaitGroup
	for index := range 50 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			inputPath := fmt.Sprintf("meeting-%d.mp4", index)
			progress.Observe(portin.ProgressEvent{InputPath: inputPath, Kind: portin.EventStage, Stage: portin.StepNormalize})
			progress.Observe(portin.ProgressEvent{InputPath: inputPath, Kind: portin.EventComplete})
		}()
	}
	workers.Wait()
	if lines := strings.Count(output.String(), "\n"); lines != 100 {
		t.Fatalf("progress has %d lines, want 100 complete messages", lines)
	}
	for index := range 50 {
		for _, message := range []string{"오디오 추출 중...", "처리 완료"} {
			want := fmt.Sprintf("[meeting-%d.mp4] %s\n", index, message)
			if !strings.Contains(output.String(), want) {
				t.Errorf("missing complete concurrent progress line %q", want)
			}
		}
	}
}

func TestProgressStartsEachStageHeartbeatOnItsOwnInterval(t *testing.T) {
	t.Parallel()

	messages := make(messageWriter, 64)
	progress := newProgressObserver(messages, 200*time.Millisecond, time.Now)
	t.Cleanup(progress.Close)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepNormalize})
	waitForProgressMessage(t, messages, "오디오 추출 중... (이 단계 경과")
	<-time.After(150 * time.Millisecond)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepDetect})
	waitForProgressMessage(t, messages, "음성 구간 탐지 중...\n")
	select {
	case message := <-messages:
		t.Fatalf("new stage heartbeat arrived before its interval: %q", message)
	case <-time.After(100 * time.Millisecond):
	}
	waitForProgressMessage(t, messages, "음성 구간 탐지 중... (이 단계 경과")
}

func TestProgressIdlePausesHeartbeatUntilNextStage(t *testing.T) {
	t.Parallel()

	messages := make(messageWriter, 64)
	progress := newProgressObserver(messages, 5*time.Millisecond, time.Now)
	t.Cleanup(progress.Close)
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepNormalize})
	waitForProgressMessage(t, messages, "오디오 추출 중... (이 단계 경과")
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventIdle})
	for draining := true; draining; {
		select {
		case message := <-messages:
			if !strings.Contains(message, "오디오 추출 중... (이 단계 경과") {
				t.Errorf("idle event emitted output: %q", message)
			}
		default:
			draining = false
		}
	}
	select {
	case message := <-messages:
		t.Fatalf("heartbeat continued while idle: %q", message)
	case <-time.After(20 * time.Millisecond):
	}
	progress.Observe(portin.ProgressEvent{InputPath: "meeting.mp4", Kind: portin.EventStage, Stage: portin.StepDetect})
	waitForProgressMessage(t, messages, "음성 구간 탐지 중...\n")
	waitForProgressMessage(t, messages, "음성 구간 탐지 중... (이 단계 경과")
}
