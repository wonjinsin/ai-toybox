package app

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunWithProgressIntervalReportsStageAndOverallElapsedWhileOperationRuns(t *testing.T) {
	t.Parallel()

	messages := make(chan string, 16)
	releaseOperation := make(chan struct{})
	overallStartedAt := time.Now().Add(-3 * time.Minute)
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseOperation) }) }
	t.Cleanup(release)
	operationDone := make(chan error, 1)
	reporter := func(_ string, message string) {
		messages <- message
	}

	go func() {
		operationDone <- runWithProgressInterval(
			reporter,
			"meeting.mp4",
			"오디오 추출 중...",
			overallStartedAt,
			5*time.Millisecond,
			func() error {
				<-releaseOperation
				return nil
			},
		)
	}()

	if message := <-messages; message != "오디오 추출 중..." {
		t.Fatalf("initial progress = %q, want stage message", message)
	}
	select {
	case message := <-messages:
		if !strings.Contains(message, "오디오 추출 중... (이 단계 경과 0s / 전체 3m0s)") {
			t.Errorf("heartbeat = %q, want stage and overall elapsed time", message)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("progress heartbeat was not reported")
	}

	release()
	if err := <-operationDone; err != nil {
		t.Fatalf("runWithProgressInterval() error = %v", err)
	}
}

func TestRunWithProgressUpdatesIntervalReportsLatestMessage(t *testing.T) {
	t.Parallel()

	messages := make(chan string, 16)
	releaseOperation := make(chan struct{})
	statusUpdated := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseOperation) }) }
	t.Cleanup(release)
	operationDone := make(chan error, 1)
	reporter := func(_ string, message string) {
		messages <- message
	}

	go func() {
		operationDone <- runWithProgressUpdatesInterval(
			reporter,
			"meeting.mp4",
			"전사 중: 84개 (완료 0/84개)",
			time.Now(),
			5*time.Millisecond,
			func(updateMessage func(string)) error {
				updateMessage("전사 중: 84개 (완료 32/84개)")
				close(statusUpdated)
				<-releaseOperation
				return nil
			},
		)
	}()

	if message := <-messages; message != "전사 중: 84개 (완료 0/84개)" {
		t.Fatalf("initial progress = %q, want initial transcription count", message)
	}
	<-statusUpdated
	select {
	case message := <-messages:
		if !strings.Contains(message, "전사 중: 84개 (완료 32/84개)") {
			t.Errorf("heartbeat = %q, want latest completed transcription count", message)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("progress heartbeat was not reported")
	}

	release()
	if err := <-operationDone; err != nil {
		t.Fatalf("runWithProgressUpdatesInterval() error = %v", err)
	}
}
