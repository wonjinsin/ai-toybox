package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	"github.com/wonjinsin/ai-toybox/whisper/internal/port"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

type Service struct {
	catalog  portout.InputCatalog
	factory  portout.SessionFactory
	progress portin.ProgressObserver
}

func NewService(catalog portout.InputCatalog, factory portout.SessionFactory, progress portin.ProgressObserver) *Service {
	return &Service{catalog: catalog, factory: factory, progress: progress}
}

func (service *Service) report(event portin.ProgressEvent) {
	if service.progress != nil {
		service.progress(event)
	}
}

func (service *Service) Transcribe(ctx context.Context, request portin.Request) (outputPath string, returnErr error) {
	startedAt := time.Now()
	progressInput := request.InputPath
	defer func() {
		if returnErr != nil && !port.IsOutputExistsError(returnErr) {
			service.report(portin.ProgressEvent{InputPath: progressInput, Kind: portin.EventFailed, Err: returnErr})
		}
	}()
	if err := portin.ValidateRequest(request); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	session, err := service.factory.Open(ctx, portout.SessionRequest{
		InputPath: request.InputPath, Language: request.Language, Format: request.Format,
		ModelPath: request.ModelPath, VADModelPath: request.VADModelPath,
		CorrectionsPath: request.CorrectionsPath, OutputDir: request.OutputDir, Force: request.Force,
	})
	if err != nil {
		return "", err
	}
	defer func() {
		if session.Close != nil {
			if err := session.Close(); err != nil {
				service.report(portin.ProgressEvent{InputPath: session.Job.InputPath, Kind: portin.EventWarning, Err: err})
			}
		}
	}()
	job := session.Job
	progressInput = job.InputPath
	service.report(portin.ProgressEvent{InputPath: job.InputPath, Kind: portin.EventStarted, StartedAt: startedAt, Format: job.Format})
	stage := func(step string, count, completed, total int) {
		service.report(portin.ProgressEvent{InputPath: job.InputPath, Kind: portin.EventStage, Stage: step, Format: job.Format, Count: count, Completed: completed, Total: total})
	}
	idle := func() { service.report(portin.ProgressEvent{InputPath: job.InputPath, Kind: portin.EventIdle}) }
	state, restored, err := service.loadState(session)
	if err != nil {
		return "", err
	}
	stage(portin.StepNormalize, 0, 0, 0)
	if err := session.Audio.Normalize(ctx); err != nil {
		return "", err
	}
	idle()
	if restored {
		service.report(portin.ProgressEvent{InputPath: job.InputPath, Kind: portin.EventResumed, Completed: state.CompletedChunks, Total: len(state.Chunks)})
		if state.CompletedChunks > 0 {
			if err := persistProgress(session, state); err != nil {
				return "", err
			}
		}
	} else {
		duration, err := session.Audio.Duration(ctx)
		if err != nil {
			return "", err
		}
		stage(portin.StepDetect, 0, 0, 0)
		segments, err := session.Detector.Detect(ctx)
		if err != nil {
			return "", err
		}
		idle()
		state = portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: duration, Chunks: domain.BuildSpeechChunks(segments, duration)}
		if job.Format == "srt" {
			if err := session.Checkpoints.Save(state); err != nil {
				return "", err
			}
		}
	}
	remaining, firstIndex := state.Chunks, 0
	if job.Format == "srt" {
		if state.Stage == portout.StageTranscribing {
			remaining, firstIndex = state.Chunks[state.CompletedChunks:], state.CompletedChunks
		} else {
			remaining = nil
		}
	}
	stage(portin.StepExtract, len(remaining), 0, 0)
	batch, err := session.Audio.Extract(ctx, remaining, firstIndex)
	if err != nil {
		return "", err
	}
	idle()
	if job.Format != "srt" || state.Stage == portout.StageTranscribing {
		stage(portin.StepRecognize, len(batch.Chunks), state.CompletedChunks, len(state.Chunks))
		state, err = service.recognize(ctx, session, state, batch)
		if err != nil {
			return "", err
		}
		idle()
	} else {
		service.releaseAudio(job.InputPath, batch)
	}
	if job.Format != "srt" || state.Stage == portout.StageTranscribing {
		state = portout.ResumeState{Stage: portout.StageRetrying, MediaDuration: state.MediaDuration,
			Chunks: state.Chunks, CompletedChunks: len(state.Chunks), Cues: domain.ReconcileChunkBoundaries(state.Cues)}
		if job.Format == "srt" {
			if err := persistProgress(session, state); err != nil {
				return "", err
			}
		}
	}
	if state.Stage == portout.StageRetrying {
		stage(portin.StepRetry, domain.CountLowConfidenceCues(state.Cues[state.RetryCursor:]), 0, 0)
		state, err = service.retry(ctx, session, state)
		if err != nil {
			return "", err
		}
		idle()
		state.Stage = portout.StageRetryComplete
		state.RetryCursor = len(state.Cues)
		if job.Format == "srt" {
			if err := persistProgress(session, state); err != nil {
				return "", err
			}
		}
	}
	stage(portin.StepClean, 0, 0, 0)
	cues := domain.CleanSubtitleCues(state.Cues, job.Language, job.Corrections, state.MediaDuration)
	if err := domain.ValidateSubtitleCues(cues, state.MediaDuration); err != nil {
		return "", err
	}
	stage(portin.StepPublish, 0, 0, 0)
	if err := session.Transcripts.Publish(cues); err != nil {
		return "", err
	}
	if job.Format == "srt" {
		if err := session.Transcripts.RemoveProgress(); err != nil {
			service.report(portin.ProgressEvent{InputPath: job.InputPath, Kind: portin.EventWarning, Stage: portin.StepPublish, Err: err})
		}
	}
	service.report(portin.ProgressEvent{InputPath: job.InputPath, Kind: portin.EventComplete})
	return job.OutputPath, nil
}

func (service *Service) loadState(session portout.Session) (portout.ResumeState, bool, error) {
	if session.Job.Format != "srt" {
		return portout.ResumeState{}, false, nil
	}
	state, restored, err := session.Checkpoints.Load()
	if err != nil {
		return portout.ResumeState{}, false, err
	}
	if restored {
		if err := state.Validate(); err != nil {
			return portout.ResumeState{}, false, fmt.Errorf("validate transcription checkpoint: %w", err)
		}
	}
	return state, restored, nil
}

func (service *Service) recognize(ctx context.Context, session portout.Session, state portout.ResumeState, batch portout.AudioBatch) (portout.ResumeState, error) {
	defer service.releaseAudio(session.Job.InputPath, batch)
	chunks := batch.Chunks
	batchSize := len(chunks)
	if session.Job.Format == "srt" {
		batchSize = portout.TranscriptionBatchSize
	}
	for start := 0; start < len(chunks); start += batchSize {
		if err := ctx.Err(); err != nil {
			return portout.ResumeState{}, err
		}
		end := min(start+batchSize, len(chunks))
		completed := state.CompletedChunks
		var progress func(int)
		if session.Job.Format == "srt" {
			progress = func(count int) {
				service.report(portin.ProgressEvent{InputPath: session.Job.InputPath, Kind: portin.EventUpdate, Stage: portin.StepRecognize, Format: session.Job.Format,
					Count: len(chunks), Completed: completed + count, Total: len(state.Chunks)})
			}
		}
		results, err := session.Recognizer.Recognize(ctx, chunks[start:end], portout.RecognitionStandard, progress)
		if err != nil {
			return portout.ResumeState{}, err
		}
		if len(results) != end-start {
			return portout.ResumeState{}, fmt.Errorf("recognition returned %d chunk results, want %d", len(results), end-start)
		}
		cues := append([]domain.Cue(nil), state.Cues...)
		for _, chunkCues := range results {
			cues = append(cues, chunkCues...)
		}
		state = portout.ResumeState{Stage: state.Stage, MediaDuration: state.MediaDuration, Chunks: state.Chunks,
			CompletedChunks: completed + end - start, Cues: cues}
		if session.Job.Format == "srt" {
			if err := persistProgress(session, state); err != nil {
				return portout.ResumeState{}, err
			}
			service.report(portin.ProgressEvent{InputPath: session.Job.InputPath, Kind: portin.EventUpdate, Stage: portin.StepRecognize, Format: session.Job.Format,
				Count: len(chunks), Completed: state.CompletedChunks, Total: len(state.Chunks)})
			service.report(portin.ProgressEvent{InputPath: session.Job.InputPath, Kind: portin.EventSaved, Stage: portin.StepRecognize,
				Completed: state.CompletedChunks, Total: len(state.Chunks)})
		}
	}
	return state, nil
}

func (service *Service) releaseAudio(inputPath string, batch portout.AudioBatch) {
	if batch.Release != nil {
		if err := batch.Release(); err != nil {
			service.report(portin.ProgressEvent{InputPath: inputPath, Kind: portin.EventWarning, Err: err})
		}
	}
}

func persistProgress(session portout.Session, state portout.ResumeState) error {
	cues := domain.CleanSubtitleCues(state.Cues, session.Job.Language, session.Job.Corrections, state.MediaDuration)
	if err := domain.ValidateSubtitleCues(cues, state.MediaDuration); err != nil {
		return fmt.Errorf("validate partial transcript: %w", err)
	}
	if err := session.Transcripts.WritePartial(cues); err != nil {
		return err
	}
	return session.Checkpoints.Save(state)
}

var _ portin.Transcriber = (*Service)(nil)
var _ portin.BatchRunner = (*Service)(nil)
