// Package out defines capabilities required by transcription use cases.
package out

import "context"

type SessionRequest struct {
	InputPath       string
	Language        string
	Format          string
	ModelPath       string
	VADModelPath    string
	CorrectionsPath string
	OutputDir       string
	Force           bool
}

type Job struct {
	InputPath   string
	OutputPath  string
	Language    string
	Format      string
	Corrections map[string]string
}

type InputSet struct {
	Paths     []string
	Directory bool
}

type InputCatalog interface {
	Discover(string) (InputSet, error)
}

// SessionFactory binds per-job resources while retaining shared engine limits.
type SessionFactory interface {
	Open(context.Context, SessionRequest) (Session, error)
}

type Session struct {
	Job         Job
	Audio       AudioProcessor
	Detector    SpeechDetector
	Recognizer  Recognizer
	Checkpoints CheckpointStore
	Transcripts TranscriptStore
	Close       func() error
}
