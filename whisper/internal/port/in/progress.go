package in

import "time"

// ProgressEvent is a progress response for an active transcription request.
type ProgressEvent struct {
	Format    string
	StartedAt time.Time
	InputPath string
	Kind      string
	Stage     string
	Count     int
	Completed int
	Total     int
	Err       error
}

type ProgressObserver func(ProgressEvent)

const (
	EventStarted  = "started"
	EventIdle     = "idle"
	EventStage    = "stage"
	EventUpdate   = "update"
	EventResumed  = "resumed"
	EventSaved    = "saved"
	EventComplete = "complete"
	EventFailed   = "failed"
	EventWarning  = "warning"
	StepNormalize = "normalize"
	StepDetect    = "detect"
	StepExtract   = "extract"
	StepRecognize = "recognize"
	StepRetry     = "retry"
	StepClean     = "clean"
	StepPublish   = "publish"
)
