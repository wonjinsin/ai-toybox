package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
)

type progressState struct {
	overallStartedAt time.Time
	stageStartedAt   time.Time
	message          string
	heartbeat        bool
	nextHeartbeatAt  time.Time
}

type Progress struct {
	writer   io.Writer
	now      func() time.Time
	mutex    sync.Mutex
	states   map[string]progressState
	stop     chan struct{}
	done     chan struct{}
	once     sync.Once
	closed   bool
	interval time.Duration
	wake     chan struct{}
}

func NewProgressObserver(writer io.Writer) *Progress {
	return newProgressObserver(writer, 10*time.Second, time.Now)
}

func newProgressObserver(writer io.Writer, interval time.Duration, now func() time.Time) *Progress {
	progress := &Progress{writer: writer, now: now, states: make(map[string]progressState),
		stop: make(chan struct{}), done: make(chan struct{}), interval: interval, wake: make(chan struct{}, 1)}
	if interval > 0 {
		go progress.heartbeat()
	} else {
		close(progress.done)
	}
	return progress
}

func (progress *Progress) Observe(event portin.ProgressEvent) {
	progress.mutex.Lock()
	defer progress.mutex.Unlock()
	if progress.closed {
		return
	}
	defer func() {
		select {
		case progress.wake <- struct{}{}:
		default:
		}
	}()
	now := progress.now()
	state, exists := progress.states[event.InputPath]
	if !exists {
		state.overallStartedAt = now
	}
	switch event.Kind {
	case portin.EventStarted:
		state.overallStartedAt = event.StartedAt
		progress.states[event.InputPath] = state
	case portin.EventStage:
		message := stageMessage(event)
		if message == "" {
			return
		}
		state.stageStartedAt = now
		state.message = message
		state.heartbeat = event.Stage != portin.StepClean && event.Stage != portin.StepPublish
		state.nextHeartbeatAt = time.Now().Add(progress.interval)
		progress.states[event.InputPath] = state
		progress.write(event.InputPath, message)
	case portin.EventUpdate:
		if exists {
			state.message = stageMessage(event)
			progress.states[event.InputPath] = state
		}
	case portin.EventIdle:
		if exists {
			state.heartbeat = false
			progress.states[event.InputPath] = state
		}
	case portin.EventResumed:
		state.heartbeat = false
		progress.states[event.InputPath] = state
		progress.write(event.InputPath, fmt.Sprintf("체크포인트 재개: %d/%d개", event.Completed, event.Total))
	case portin.EventSaved:
		if event.Stage == portin.StepRetry {
			progress.write(event.InputPath, fmt.Sprintf("저신뢰 재시도 진행: %d/%d cue", event.Completed, event.Total))
		} else {
			progress.write(event.InputPath, fmt.Sprintf("전사 진행: %d/%d개 (partial SRT 저장됨)", event.Completed, event.Total))
		}
	case portin.EventComplete:
		delete(progress.states, event.InputPath)
		progress.write(event.InputPath, "처리 완료")
	case portin.EventFailed:
		delete(progress.states, event.InputPath)
		progress.write(event.InputPath, fmt.Sprintf("처리 실패: %v", event.Err))
	case portin.EventWarning:
		if event.Stage == portin.StepPublish {
			progress.write(event.InputPath, fmt.Sprintf("경고: 증분 파일 정리 실패: %v", event.Err))
		} else {
			progress.write(event.InputPath, fmt.Sprintf("경고: %v", event.Err))
		}
	}
}

func stageMessage(event portin.ProgressEvent) string {
	switch event.Stage {
	case portin.StepNormalize:
		return "오디오 추출 중..."
	case portin.StepDetect:
		return "음성 구간 탐지 중..."
	case portin.StepExtract:
		return fmt.Sprintf("음성 조각 생성: %d개", event.Count)
	case portin.StepRecognize:
		if event.Format == "srt" {
			return fmt.Sprintf("전사 중: %d개 (완료 %d/%d개)", event.Count, event.Completed, event.Total)
		}
		return fmt.Sprintf("전사 중: %d개", event.Count)
	case portin.StepRetry:
		return fmt.Sprintf("저신뢰 구간 재시도: %d개", event.Count)
	case portin.StepClean:
		return "자막 정리 및 검증 중..."
	case portin.StepPublish:
		return "결과 저장 중..."
	default:
		return ""
	}
}

func (progress *Progress) heartbeat() {
	defer close(progress.done)
	timer := time.NewTimer(progress.interval)
	defer timer.Stop()
	for {
		progress.mutex.Lock()
		if progress.closed {
			progress.mutex.Unlock()
			return
		}
		scheduledAt := time.Now()
		var nextHeartbeatAt time.Time
		for inputPath, state := range progress.states {
			if !state.heartbeat {
				continue
			}
			if !state.nextHeartbeatAt.After(scheduledAt) {
				now := progress.now()
				progress.write(inputPath, fmt.Sprintf("%s (이 단계 경과 %s / 전체 %s)", state.message,
					now.Sub(state.stageStartedAt).Round(time.Second), now.Sub(state.overallStartedAt).Round(time.Second)))
				state.nextHeartbeatAt = scheduledAt.Add(progress.interval)
				progress.states[inputPath] = state
			}
			if nextHeartbeatAt.IsZero() || state.nextHeartbeatAt.Before(nextHeartbeatAt) {
				nextHeartbeatAt = state.nextHeartbeatAt
			}
		}
		progress.mutex.Unlock()

		timer.Stop()
		var heartbeat <-chan time.Time
		if !nextHeartbeatAt.IsZero() {
			timer.Reset(time.Until(nextHeartbeatAt))
			heartbeat = timer.C
		}
		select {
		case <-progress.stop:
			return
		case <-progress.wake:
		case <-heartbeat:
		}
	}
}

func (progress *Progress) write(inputPath, message string) {
	fmt.Fprintf(progress.writer, "[%s] %s\n", filepath.Base(inputPath), message)
}

func (progress *Progress) Close() {
	progress.once.Do(func() {
		progress.mutex.Lock()
		progress.closed = true
		close(progress.stop)
		progress.mutex.Unlock()
	})
	<-progress.done
}
