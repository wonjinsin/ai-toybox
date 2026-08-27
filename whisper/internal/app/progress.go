package app

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"
)

const progressHeartbeatInterval = 10 * time.Second

type progressReporter func(inputPath, message string)

func newProgressReporter(writer io.Writer) progressReporter {
	var writeMutex sync.Mutex
	return func(inputPath, message string) {
		writeMutex.Lock()
		defer writeMutex.Unlock()
		fmt.Fprintf(writer, "[%s] %s\n", filepath.Base(inputPath), message)
	}
}

func reportProgress(reporter progressReporter, inputPath, message string) {
	if reporter != nil {
		reporter(inputPath, message)
	}
}

func runWithProgress(reporter progressReporter, inputPath, message string, overallStartedAt time.Time, operation func() error) error {
	return runWithProgressInterval(reporter, inputPath, message, overallStartedAt, progressHeartbeatInterval, operation)
}

func runWithProgressInterval(reporter progressReporter, inputPath, message string, overallStartedAt time.Time, interval time.Duration, operation func() error) error {
	return runWithProgressMessageInterval(reporter, inputPath, func() string { return message }, overallStartedAt, interval, operation)
}

func runWithProgressUpdates(reporter progressReporter, inputPath, initialMessage string, overallStartedAt time.Time, operation func(updateMessage func(string)) error) error {
	return runWithProgressUpdatesInterval(reporter, inputPath, initialMessage, overallStartedAt, progressHeartbeatInterval, operation)
}

func runWithProgressUpdatesInterval(reporter progressReporter, inputPath, initialMessage string, overallStartedAt time.Time, interval time.Duration, operation func(updateMessage func(string)) error) error {
	var messageMutex sync.RWMutex
	currentMessage := initialMessage
	loadMessage := func() string {
		messageMutex.RLock()
		defer messageMutex.RUnlock()
		return currentMessage
	}
	updateMessage := func(message string) {
		messageMutex.Lock()
		defer messageMutex.Unlock()
		currentMessage = message
	}
	return runWithProgressMessageInterval(reporter, inputPath, loadMessage, overallStartedAt, interval, func() error {
		return operation(updateMessage)
	})
}

func runWithProgressMessageInterval(reporter progressReporter, inputPath string, message func() string, overallStartedAt time.Time, interval time.Duration, operation func() error) error {
	stageStartedAt := time.Now()
	reportProgress(reporter, inputPath, message())
	if reporter == nil || interval <= 0 {
		return operation()
	}

	stopHeartbeat := make(chan struct{})
	heartbeatStopped := make(chan struct{})
	go func() {
		defer close(heartbeatStopped)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopHeartbeat:
				return
			case <-ticker.C:
				select {
				case <-stopHeartbeat:
					return
				default:
				}
				reportedAt := time.Now()
				stageElapsed := reportedAt.Sub(stageStartedAt).Round(time.Second)
				overallElapsed := reportedAt.Sub(overallStartedAt).Round(time.Second)
				reportProgress(reporter, inputPath, fmt.Sprintf("%s (이 단계 경과 %s / 전체 %s)", message(), stageElapsed, overallElapsed))
			}
		}
	}()

	err := operation()
	close(stopHeartbeat)
	<-heartbeatStopped
	return err
}
